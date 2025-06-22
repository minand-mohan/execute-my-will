// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

//go:build windows
// +build windows

package system

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(command string, shell string) error {

	fmt.Printf("🗡️  Executing thy will, my lord: %s\n", command)
	fmt.Println("═══════════════════════════════════════════════════════")

	cmd := exec.Command(shell, "/C", command)

	// Hook I/O streams
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Ensure it runs in the same console
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    false,
	}

	err := cmd.Run()

	fmt.Println("═══════════════════════════════════════════════════════")
	return err
}
