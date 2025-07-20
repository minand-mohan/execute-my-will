// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

//go:build !windows
// +build !windows

package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/minand-mohan/execute-my-will/internal/ui"
)

type Executor struct{}

// NewExecutor creates a new executor instance
func NewExecutor() CommandExecutor {
	return &Executor{}
}

// Execute runs the command with enhanced real-time output display
func (e *Executor) Execute(command string, shell string) error {
	ui.PrintExecutionHeader(fmt.Sprintf("Executing thy will, my lord: %s", command))

	cmd := exec.Command(shell, "-c", command)

	// Create pipes to capture output for highlighting while still showing real-time
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	cmd.Stdin = os.Stdin

	// Ensure the command runs in the foreground
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Foreground: true,
		Pgid:       0,
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Create output highlighter
	highlighter := ui.NewOutputHighlighter(false, 1)

	// Stream stdout and stderr concurrently
	done := make(chan error, 2)

	go func() {
		done <- highlighter.StreamOutput(stdoutPipe, "")
	}()

	go func() {
		done <- highlighter.StreamOutput(stderrPipe, "")
	}()

	// Wait for both streams to complete
	for i := 0; i < 2; i++ {
		if streamErr := <-done; streamErr != nil {
			ui.PrintWarningMessage(fmt.Sprintf("Stream error: %v", streamErr))
		}
	}

	// Wait for command to complete
	err = cmd.Wait()

	ui.PrintSeparator()

	if err != nil {
		return err
	}
	return nil
}

// ExecuteScript runs a script with enhanced real-time output and comment display
func (e *Executor) ExecuteScript(scriptContent string, shell string, showComments bool) error {
	// Create temp directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %v", err)
	}

	tmpDir := filepath.Join(configDir, "execute-my-will", "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create tmp directory: %v", err)
	}

	// Generate script filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	scriptPath := filepath.Join(tmpDir, fmt.Sprintf("script_%s.sh", timestamp))

	// Create executable script with enhanced output
	scriptWithExecutor := e.createExecutableScriptWithOutput(scriptContent, showComments)

	if err := ioutil.WriteFile(scriptPath, []byte(scriptWithExecutor), 0755); err != nil {
		return fmt.Errorf("failed to write script file: %v", err)
	}

	// Clean up script file after execution
	defer func() {
		os.Remove(scriptPath)
		// Clean up old script files (older than 1 hour)
		e.cleanupOldScripts(tmpDir)
	}()

	ui.PrintExecutionHeader("Executing thy script, my lord")

	// Execute the script with enhanced output capture
	cmd := exec.Command(shell, scriptPath)

	// Create pipes for output capture
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	cmd.Stdin = os.Stdin

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Foreground: true,
		Pgid:       0,
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Create output highlighter with timestamps for scripts
	highlighter := ui.NewOutputHighlighter(true, 1)

	// Stream outputs concurrently
	done := make(chan error, 2)

	go func() {
		done <- highlighter.StreamOutput(stdoutPipe, "")
	}()

	go func() {
		done <- highlighter.StreamOutput(stderrPipe, "")
	}()

	// Wait for both streams
	for i := 0; i < 2; i++ {
		if streamErr := <-done; streamErr != nil {
			ui.PrintWarningMessage(fmt.Sprintf("Stream error: %v", streamErr))
		}
	}

	// Wait for command completion
	err = cmd.Wait()

	ui.PrintSeparator()

	return err
}

// createExecutableScriptWithOutput creates a bash script with enhanced output and error handling
func (e *Executor) createExecutableScriptWithOutput(scriptContent string, showComments bool) string {
	lines := strings.Split(scriptContent, "\n")
	var result strings.Builder

	// Bash script header with error handling
	result.WriteString("#!/bin/bash\n")
	result.WriteString("set -e\n")
	result.WriteString("set -o pipefail\n\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") && showComments {
			// Display comment with medieval emoji
			comment := strings.TrimPrefix(line, "#")
			comment = strings.TrimSpace(comment)
			result.WriteString(fmt.Sprintf("echo '💬 %s'\n", comment))
		} else if !strings.HasPrefix(line, "#") {
			// Execute command with step indication
			result.WriteString(fmt.Sprintf("echo '⚔️  Executing: %s'\n", line))
			result.WriteString(fmt.Sprintf("%s\n", line))
			result.WriteString("echo ''\n") // Add spacing between commands
		}
	}

	return result.String()
}

// createExecutableScript creates a bash script with error handling and comment display (legacy method)
func (e *Executor) createExecutableScript(scriptContent string, showComments bool) string {
	return e.createExecutableScriptWithOutput(scriptContent, showComments)
}

// cleanupOldScripts removes script files older than 1 hour
func (e *Executor) cleanupOldScripts(tmpDir string) {
	files, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-1 * time.Hour)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "script_") && file.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(tmpDir, file.Name()))
		}
	}
}
