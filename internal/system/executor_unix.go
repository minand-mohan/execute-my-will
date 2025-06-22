// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

//go:build !windows
// +build !windows

package system

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Executor struct{}

// NewExecutor creates a new executor instance
func NewExecutor() *Executor {
	return &Executor{}
}

// Execute runs the command with full interactive terminal support
func (e *Executor) Execute(command string, shell string) error {

	fmt.Printf("🗡️  Executing thy will, my lord: %s\n", command)
	fmt.Println("═══════════════════════════════════════════════════════")

	cmd := exec.Command(shell, "-c", command)

	// Direct I/O connection - simplest and most compatible approach
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Ensure the command runs in the foreground
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Foreground: true,
		Pgid:       0,
	}

	err := cmd.Run()

	fmt.Println("═══════════════════════════════════════════════════════")

	if err != nil {
		return err
	}
	return nil
}
