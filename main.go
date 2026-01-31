package main

import (
	"os"

	"github.com/buck3000/wiz/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
