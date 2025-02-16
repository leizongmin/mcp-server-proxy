package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <command> [options]\n", path.Base(os.Args[0]))
		fmt.Printf("Commands:\n")
		fmt.Printf("    inspect <local_url> <target_url>\n")
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
