package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/google/uuid"
)

var sessions sync.Map

type Session struct {
	id          string
	targetUrl   *url.URL
	writer      http.ResponseWriter
	flusher     http.Flusher
	initialized bool
}

func (s *Session) Initialize() {
	s.initialized = true
}

const (
	eventEndpoint = "endpoint"
	eventMessage  = "message"
)

func newSession(w http.ResponseWriter, targetUrl *url.URL) *Session {
	return &Session{
		id:          uuid.NewString(),
		targetUrl:   targetUrl,
		writer:      w,
		flusher:     w.(http.Flusher),
		initialized: false,
	}
}

func (s *Session) WriteEvent(event string, data any) error {
	if !s.initialized {
		return fmt.Errorf("session not initialized")
	}
	return s.forceWriteEvent(event, data)
}

func (s *Session) forceWriteEvent(event string, data any) error {
	var b []byte
	if s, ok := data.(string); ok {
		b = []byte(s)
	} else {
		var err error
		b, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %v", err)
		}
	}
	_, err := s.writer.Write([]byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event, b)))
	if err != nil {
		return fmt.Errorf("failed to write event: %v", err)
	}
	s.flusher.Flush()
	return nil
}

// {"jsonrpc":"2.0","id":2,"method":"tools/call","params": {}}
type RpcRequest struct {
	Jsonrpc string         `json:"jsonrpc"`
	Id      any            `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
}

// {"result":{"content":[{"type":"text","text":"Echo: hello"}]},"jsonrpc":"2.0","id":2}
type RpcResponse struct {
	Jsonrpc string         `json:"jsonrpc"`
	Id      any            `json:"id"`
	Result  map[string]any `json:"result,omitempty"`
	Error   *RpcError      `json:"error,omitempty"`
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func startServe(localUrl, targetUrl string) error {
	parsedLocalUrl, err := url.Parse(localUrl)
	if err != nil {
		return fmt.Errorf("invalid local_url: %v", err)
	}

	parsedTargetUrl, err := url.Parse(targetUrl)
	if err != nil {
		return fmt.Errorf("invalid target_url: %v", err)
	}

	listener, err := net.Listen("tcp4", parsedLocalUrl.Host)
	if err != nil {
		return fmt.Errorf("listen local_url failed: %v", err)
	}

	return http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path != "/message" {
			handleConnect(w, r, parsedTargetUrl)
		} else if r.Method == "POST" && r.URL.Path == "/message" {
			handleMessage(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
}

func handleConnect(w http.ResponseWriter, r *http.Request, parsedTargetUrl *url.URL) {
	session := newSession(w, parsedTargetUrl)
	sessions.Store(session.id, session)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err := session.forceWriteEvent(eventEndpoint, fmt.Sprintf("/message?sessionId=%s", session.id))
	if err != nil {
		log.Printf("Failed to write event: sessionId=%s, %v", session.id, err)
		return
	}

	log.Printf("Connected: sessionId=%s", session.id)

	// Create a channel to wait for connection close
	done := make(chan bool)

	// Notify channel when client closes connection
	go func() {
		<-r.Context().Done()
		sessions.Delete(session.id)
		log.Printf("Disconnected: sessionId=%s", session.id)
		done <- true
	}()

	// Wait for connection close
	<-done
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	sessionId := r.URL.Query().Get("sessionId")
	session, ok := sessions.Load(sessionId)
	if !ok {
		http.NotFound(w, r)
		return
	}

	req := &RpcRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case "notifications/initialized":
		log.Printf("Initialize: sessionId=%s", sessionId)
		session.(*Session).Initialize()
	default:
		log.Printf("Accepted: sessionId=%s, method=%s", sessionId, req.Method)
		go convertMessageToRequest(session.(*Session), req)
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Accepted"))
	}
}

// convertMessageToRequest converts a message to a request
// and sends it to the target URL.
func convertMessageToRequest(session *Session, req *RpcRequest) {
	// target request:
	// - url: http://target-host/path/$req.method
	// - body: $req.params
	targetUrl := *session.targetUrl
	targetUrl.Path = path.Join(targetUrl.Path, req.Method)
	// tools/call/$name
	if req.Method == "tools/call" {
		name, ok := req.Params["name"]
		if ok {
			targetUrl.Path = path.Join(targetUrl.Path, name.(string))
		}
	}
	query := targetUrl.Query()
	query.Set("sessionId", session.id)
	targetUrl.RawQuery = query.Encode()
	body, err := json.Marshal(req.Params)
	if err != nil {
		log.Printf("Failed to marshal request params: %v", err)
		// FIXME: send error to the client
		return
	}
	httpReq, err := http.NewRequest(http.MethodPost, targetUrl.String(), bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		// FIXME: send error to the client
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Method = http.MethodPost

	log.Printf("Sending request: sessionId=%s, id=%v, %s", session.id, req.Id, httpReq.URL)
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		// FIXME: send error to the client
		return
	}
	defer resp.Body.Close()
	log.Printf("Received response: sessionId=%s, id=%v, %s", session.id, req.Id, resp.Status)

	// target response:
	// - success:
	//   - status: 200
	//   - body: map[string]any
	// - error:
	//   - status: 500
	//   - body: {code, message, data}

	rpcResp := RpcResponse{
		Jsonrpc: req.Jsonrpc,
		Id:      req.Id,
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Failed to decode response: %v", err)
			// FIXME: send error to the client
			return
		}
		rpcResp.Result = result
	case http.StatusInternalServerError:
		var rpcErr RpcError
		if err := json.NewDecoder(resp.Body).Decode(&rpcErr); err != nil {
			log.Printf("Failed to decode response: %v", err)
			// FIXME: send error to the client
			return
		}
		rpcResp.Error = &rpcErr
	default:
		log.Printf("Unexpected response: sessionId=%s, id=%v, %s", session.id, req.Id, resp.Status)
		// FIXME: send error to the client
		return
	}
	if req.Method == "initialize" {
		err = session.forceWriteEvent(eventMessage, rpcResp)
	} else {
		err = session.WriteEvent(eventMessage, rpcResp)
	}
	if err != nil {
		log.Printf("Failed to write response: %v", err)
		return
	}
}
