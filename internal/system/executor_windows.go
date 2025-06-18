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

func (e *Executor) Execute(command string) error {
	// Use the standard Windows shell
	shell := os.Getenv("ComSpec")
	if shell == "" {
		shell = "powershell.exe"
	}

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
