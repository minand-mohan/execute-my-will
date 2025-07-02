// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: cmd/execute-my-will/main.go
package main

import (
	"log"
	"os"

	"github.com/minand-mohan/execute-my-will/internal/cli"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	/// Set the build information in the cli package
	cli.SetBuildInfo(version, commit, buildTime)

	if err := cli.Execute(); err != nil {
		log.Printf("Noble quest has failed!")
		os.Exit(1)
	}
}
