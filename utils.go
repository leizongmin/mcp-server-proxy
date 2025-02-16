package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	yellow     = "\033[33m"
	green      = "\033[32m"
	resetColor = "\033[0m"
)

func colorPrintf(color string, format string, a ...interface{}) {
	fmt.Printf(color+format+resetColor, a...)
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
