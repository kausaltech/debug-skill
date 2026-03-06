package main

import (
	"fmt"
	"os"

	dap "github.com/AlmogBaku/debug-skill"
)

var version = "dev"

func main() {
	if err := dap.NewRootCmd(version).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
