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

// ExecuteScript runs a script with comments displayed during execution
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

	// Create executable script
	scriptWithExecutor := e.createExecutableScript(scriptContent, showComments)

	if err := ioutil.WriteFile(scriptPath, []byte(scriptWithExecutor), 0755); err != nil {
		return fmt.Errorf("failed to write script file: %v", err)
	}

	// Clean up script file after execution
	defer func() {
		os.Remove(scriptPath)
		// Clean up old script files (older than 1 hour)
		e.cleanupOldScripts(tmpDir)
	}()

	fmt.Printf("🗡️  Executing thy script, my lord\n")
	fmt.Println("═══════════════════════════════════════════════════════")

	// Execute the script with the specified shell
	cmd := exec.Command(shell, scriptPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Foreground: true,
		Pgid:       0,
	}

	err = cmd.Run()

	fmt.Println("═══════════════════════════════════════════════════════")

	return err
}

// createExecutableScript creates a bash script with error handling and comment display
func (e *Executor) createExecutableScript(scriptContent string, showComments bool) string {
	lines := strings.Split(scriptContent, "\n")
	var result strings.Builder

	// Bash script header with error handling
	result.WriteString("#!/bin/bash\n")
	result.WriteString("set -e\n\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") && showComments {
			// Display comment
			comment := strings.TrimPrefix(line, "#")
			result.WriteString(fmt.Sprintf("echo '%s'\n", strings.TrimSpace(comment)))
		} else if !strings.HasPrefix(line, "#") {
			// Execute command
			result.WriteString(fmt.Sprintf("%s\n", line))
		}
	}

	return result.String()
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
