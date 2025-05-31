// File: cmd/execute-my-will/main.go
package main

import (
	"log"
	"os"

	"github.com/minand-mohan/execute-my-will/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Printf("Noble quest has failed!")
		os.Exit(1)
	}
}
