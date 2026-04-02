package main

import (
	"os"

	"github.com/mzaran/w9s/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
