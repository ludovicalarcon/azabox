package main

import (
	"fmt"
	"os"

	"gitlab.com/ludovic-alarcon/azabox/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}
}
