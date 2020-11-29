package main

import "fmt"

var (
	// Version can be set with the Go linker.
	Version string = "master"
)

func main() {
	fmt.Printf("app version %s\n", Version)
}
