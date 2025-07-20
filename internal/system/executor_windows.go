// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

//go:build windows
// +build windows

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

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(command string, shell string) error {
	ui.PrintExecutionHeader(fmt.Sprintf("Executing thy will, my lord: %s", command))

	cmd := exec.Command(shell, "/C", command)

	// Create pipes to capture output for highlighting
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	cmd.Stdin = os.Stdin

	// Ensure it runs in the same console
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    false,
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Create output highlighter
	highlighter := ui.NewOutputHighlighter(false, 1)

	// Stream outputs concurrently
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

	return err
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

	// Generate script filename with timestamp and appropriate extension
	timestamp := time.Now().Format("20060102_150405")
	var scriptPath string
	var scriptWithExecutor string

	if shell == "powershell" || shell == "pwsh" {
		scriptPath = filepath.Join(tmpDir, fmt.Sprintf("script_%s.ps1", timestamp))
		scriptWithExecutor = e.createPowerShellScript(scriptContent, showComments)
	} else {
		// Default to cmd
		scriptPath = filepath.Join(tmpDir, fmt.Sprintf("script_%s.bat", timestamp))
		scriptWithExecutor = e.createCmdScript(scriptContent, showComments)
	}

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

	// Execute the script
	var cmd *exec.Cmd
	if shell == "powershell" || shell == "pwsh" {
		cmd = exec.Command(shell, "-File", scriptPath)
	} else {
		cmd = exec.Command("cmd", "/C", scriptPath)
	}

	// Create pipes for enhanced output capture
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
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    false,
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

// createPowerShellScript creates a PowerShell script with error handling and comment display
func (e *Executor) createPowerShellScript(scriptContent string, showComments bool) string {
	lines := strings.Split(scriptContent, "\n")
	var result strings.Builder

	// PowerShell script header with error handling
	result.WriteString("$ErrorActionPreference = 'Stop'\n")
	result.WriteString("$LineNumber = 0\n\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		result.WriteString("$LineNumber++\n")

		if strings.HasPrefix(line, "#") && showComments {
			// Display comment
			comment := strings.TrimPrefix(line, "#")
			result.WriteString(fmt.Sprintf("Write-Host '%s' -ForegroundColor Yellow\n", strings.TrimSpace(comment)))
		} else if !strings.HasPrefix(line, "#") {
			// Execute command with error handling
			result.WriteString("try {\n")
			result.WriteString(fmt.Sprintf("    %s\n", line))
			result.WriteString("} catch {\n")
			result.WriteString(fmt.Sprintf("    Write-Host \"Line $LineNumber failed: %s - $($_.Exception.Message)\" -ForegroundColor Red\n", line))
			result.WriteString("    exit 1\n")
			result.WriteString("}\n")
		}
	}

	return result.String()
}

// createCmdScript creates a CMD batch script with error handling and comment display
func (e *Executor) createCmdScript(scriptContent string, showComments bool) string {
	lines := strings.Split(scriptContent, "\n")
	var result strings.Builder

	// CMD script header with error handling
	result.WriteString("@echo off\n")
	result.WriteString("setlocal enabledelayedexpansion\n")
	result.WriteString("set LINE=0\n\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		result.WriteString("set /a LINE+=1\n")

		if strings.HasPrefix(line, "REM") && showComments {
			// Display comment
			comment := strings.TrimPrefix(line, "REM")
			result.WriteString(fmt.Sprintf("echo %s\n", strings.TrimSpace(comment)))
		} else if !strings.HasPrefix(line, "REM") {
			// Execute command with error handling
			result.WriteString(fmt.Sprintf("%s\n", line))
			result.WriteString("if !errorlevel! neq 0 (\n")
			result.WriteString(fmt.Sprintf("    echo Line !LINE! failed: %s - Error code !errorlevel!\n", line))
			result.WriteString("    exit /b !errorlevel!\n")
			result.WriteString(")\n")
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
