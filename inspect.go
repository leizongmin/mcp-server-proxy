package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
)

type chunkedResponseWriter struct {
	origin   http.ResponseWriter
	flusher  http.Flusher
	reqNum   uint64
	chunkNum uint64
}

func newChunkedResponseWriter(origin http.ResponseWriter, reqNum uint64) *chunkedResponseWriter {
	return &chunkedResponseWriter{
		origin:  origin,
		flusher: origin.(http.Flusher),
		reqNum:  reqNum,
	}
}

func (w *chunkedResponseWriter) Write(p []byte) (int, error) {
	chunkNum := atomic.AddUint64(&w.chunkNum, 1)
	colorPrintf(green, "=== Response %d Body Chunk %d ===\n%s\n", w.reqNum, chunkNum, string(p))
	n, err := w.origin.Write(p)
	if err != nil {
		return n, err
	}
	if w.flusher != nil {
		w.flusher.Flush()
	}
	return n, nil
}

func startInspect(localUrl, targetUrl string) error {
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

	var num uint64
	err = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqNum := atomic.AddUint64(&num, 1)

		// Print request details
		bodyStr, newBody, err := readBody(r.Body)
		if err != nil {
			colorPrintf(yellow, "Failed to read request body: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Body = newBody

		colorPrintf(yellow, "=== Request %d ===\n", reqNum)
		colorPrintf(yellow, "URL: %s\n", r.URL.String())
		colorPrintf(yellow, "Headers:\n%s", dumpHeaders(r.Header))
		if bodyStr != "" {
			colorPrintf(yellow, "Body:\n%s\n", bodyStr)
		}
		fmt.Println()

		// Build new request URL
		targetPath := strings.TrimPrefix(r.URL.Path, parsedLocalUrl.Path)
		newPath := parsedTargetUrl.Path + targetPath
		if r.URL.RawQuery != "" {
			newPath += "?" + r.URL.RawQuery
		}

		// Create new request
		proxyReq, err := http.NewRequest(r.Method, parsedTargetUrl.Scheme+"://"+parsedTargetUrl.Host+newPath, r.Body)
		if err != nil {
			colorPrintf(yellow, "Failed to create proxy request: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Copy original request headers
		for key, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Send request to target server
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			colorPrintf(green, "Failed to send proxy request: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Print response status and headers
		colorPrintf(green, "=== Response %d ===\n", reqNum)
		colorPrintf(green, "Status: %s\n", resp.Status)
		colorPrintf(green, "Headers:\n%s\n", dumpHeaders(resp.Header))

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Set chunked encoding if the response is chunked
		if len(resp.TransferEncoding) > 0 && resp.TransferEncoding[0] == "chunked" {
			w.Header().Set("Transfer-Encoding", "chunked")
			colorPrintf(green, "=== Response %d Transfer-Encoding: chunked ===\n", reqNum)
		}

		// Set response status code
		w.WriteHeader(resp.StatusCode)

		// Create chunked response writer for logging
		chunkedWriter := newChunkedResponseWriter(w, reqNum)

		// Copy response body while maintaining chunked encoding
		if _, err := io.Copy(chunkedWriter, resp.Body); err != nil {
			colorPrintf(green, "Failed to copy response body: %v\n", err)
			return
		}

		colorPrintf(green, "=== Response %d Body Complete ===\n\n", reqNum)
	}))
	if err != nil {
		return fmt.Errorf("listen local_url failed: %v", err)
	}
	return nil
}
