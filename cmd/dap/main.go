package main

import (
	"fmt"
	"os"

	dap "github.com/AlmogBaku/debug-skill"
)

func main() {
	if err := dap.NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
