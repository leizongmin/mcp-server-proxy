package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
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
	case "serve":
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
		if err := startServe(localUrl, targetUrl); err != nil {
			log.Fatal(err)
		}
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func printUsage() {
	exec := path.Base(os.Args[0])
	fmt.Printf("Usage: %s <command> [options]\n", exec)
	fmt.Printf("Commands:\n")
	fmt.Printf("    inspect <local_url> <target_url>\n")
	fmt.Printf("    serve <local_url> <target_url>\n")
	fmt.Printf("    help\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("    %s inspect http://localhost:8080 http://example.com\n", exec)
	fmt.Printf("           => Inspect the request and response of http://example.com\n")
	fmt.Printf("    %s serve http://localhost:8080 http://example.com\n", exec)
	fmt.Printf("           => Start a server that convert MCP protocol's SSE transport layer to a standard HTTP request/response\n")
}
