package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
)

const (
	yellow     = "\033[33m"
	green      = "\033[32m"
	resetColor = "\033[0m"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <command> [options]\n", os.Args[0])
		fmt.Printf("Commands:\n")
		fmt.Printf("    inspect <local_url> <target_url>")
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "inspect":
		if len(os.Args) < 3 {
			fmt.Println("missing <local_url>")
			os.Exit(1)
		}
		if len(os.Args) < 4 {
			fmt.Println("missing <target_url>")
			os.Exit(1)
		}
		localUrl := os.Args[2]
		targetUrl := os.Args[3]
		if err := startInspect(localUrl, targetUrl); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func dumpHeaders(headers http.Header) string {
	var b strings.Builder
	for key, values := range headers {
		for _, value := range values {
			b.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}
	return b.String()
}

func readBody(body io.ReadCloser) (string, io.ReadCloser, error) {
	if body == nil {
		return "", nil, nil
	}

	var buf bytes.Buffer
	bodyData, err := io.ReadAll(body)
	if err != nil {
		return "", nil, err
	}
	buf.Write(bodyData)

	// Create new reader for the body
	return buf.String(), io.NopCloser(bytes.NewReader(bodyData)), nil
}

func colorPrintf(color string, format string, a ...interface{}) {
	fmt.Printf(color+format+resetColor, a...)
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

		// Set response status code
		w.WriteHeader(resp.StatusCode)

		// Create a pipe for streaming response body
		pr, pw := io.Pipe()

		// Create a multi-writer to write to both the client and our pipe
		mw := io.MultiWriter(w, pw)

		// Start a goroutine to read and print chunks from the pipe
		go func() {
			defer pr.Close()
			reader := bufio.NewReader(pr)
			buffer := make([]byte, 4096)
			chunkNum := 0

			for {
				n, err := reader.Read(buffer)
				if n > 0 {
					chunkNum++
					chunk := buffer[:n]
					colorPrintf(green, "=== Response %d Body Chunk %d ===\n%s\n", reqNum, chunkNum, string(chunk))
				}
				if err == io.EOF {
					colorPrintf(green, "=== Response %d Body Complete ===\n\n", reqNum)
					break
				}
				if err != nil {
					colorPrintf(green, "Error reading response chunk: %v\n", err)
					break
				}
			}
		}()

		// Copy response body to multi-writer
		if _, err := io.Copy(mw, resp.Body); err != nil {
			colorPrintf(green, "Failed to copy response body: %v\n", err)
		}
		pw.Close()
	}))
	if err != nil {
		return fmt.Errorf("listen local_url failed: %v", err)
	}
	return nil
}
