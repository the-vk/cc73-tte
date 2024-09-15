package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("failed to make raw:", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	b := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(b)
		if err != nil {
			break
		}
		os.Stdout.Write([]byte(fmt.Sprintf("input (%d): %s\n", n, string(b))))
	}
}
