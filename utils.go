package main

import "fmt"

const (
	yellow     = "\033[33m"
	green      = "\033[32m"
	resetColor = "\033[0m"
)

func colorPrintf(color string, format string, a ...interface{}) {
	fmt.Printf(color+format+resetColor, a...)
}
