package main

import (
	"os"

	"github.com/firewood-buck-3000/wiz/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
