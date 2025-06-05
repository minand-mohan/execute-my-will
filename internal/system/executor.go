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
func (e *Executor) Execute(command string) error {

	// Determine shell to use
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	fmt.Printf("🗡️  Executing thy will, my lord: %s\n", command)
	fmt.Println("═══════════════════════════════════════════════════════")

	// For commands that need full terminal emulation, we can use a pseudo-terminal
	// This requires the golang.org/x/term package
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
