// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: internal/system/analyzer_windows.go
//go:build windows
// +build windows
package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

type Info struct {
	OS                string
	Shell             string
	PackageManager    string
	CurrentDir        string
	HomeDir           string
	PathDirectories   []string
	AvailableCommands []string
	Aliases           map[string]string
}

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) AnalyzeSystem() (*Info, error) {
	info := &Info{
		Aliases: make(map[string]string),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Detect OS
	info.OS = runtime.GOOS

	// Get current and home directories
	currentDir, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	info.CurrentDir = currentDir
	info.HomeDir = homeDir

	// Concurrent system analysis
	tasks := []func() error{
		func() error { return a.detectShell(info) },
		func() error { return a.detectPackageManager(info) },
		func() error { return a.getPathDirectories(info) },
		func() error { return a.getAvailableCommands(info) },
		func() error { return a.parseAliases(info) },
	}

	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(t func() error) {
			defer wg.Done()
			if err := t(); err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}(task)
	}

	wg.Wait()

	if len(errors) > 0 {
		return info, fmt.Errorf("system analysis completed with warnings: %v", errors[0])
	}

	return info, nil
}

func (a *Analyzer) detectShell(info *Info) error {
	// Check if running in PowerShell
	if os.Getenv("PSModulePath") != "" {
		// Further check for PowerShell Core vs Windows PowerShell
		if psVersion := os.Getenv("PSVersionTable"); psVersion != "" {
			info.Shell = "pwsh" // PowerShell Core
		} else {
			info.Shell = "powershell" // Windows PowerShell
		}
		return nil
	}

	// Check for Windows Terminal or other shells
	comspec := os.Getenv("COMSPEC")
	if comspec != "" {
		shellName := filepath.Base(comspec)
		info.Shell = strings.ToLower(strings.TrimSuffix(shellName, ".exe"))
	} else {
		info.Shell = "cmd" // default
	}

	// Check for Git Bash or other Unix-like shells on Windows
	if _, err := exec.LookPath("bash.exe"); err == nil {
		if os.Getenv("MSYSTEM") != "" { // MSYS2/Git Bash environment
			info.Shell = "bash"
		}
	}

	return nil
}

func (a *Analyzer) detectPackageManager(info *Info) error {
	// Windows package managers in order of preference/commonality
	managers := []struct {
		name string
		cmd  string
	}{
		{"chocolatey", "choco"},
		{"winget", "winget"},
		{"scoop", "scoop"},
		{"nuget", "nuget"},
		{"vcpkg", "vcpkg"},
		{"pip", "pip"}, // Python package manager
		{"npm", "npm"}, // Node.js package manager
	}

	for _, manager := range managers {
		if _, err := exec.LookPath(manager.cmd + ".exe"); err == nil {
			info.PackageManager = manager.name
			return nil
		}
		// Also check without .exe extension
		if _, err := exec.LookPath(manager.cmd); err == nil {
			info.PackageManager = manager.name
			return nil
		}
	}

	info.PackageManager = "unknown"
	return nil
}

func (a *Analyzer) getPathDirectories(info *Info) error {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil
	}

	// Windows uses semicolon as PATH separator
	info.PathDirectories = strings.Split(pathEnv, ";")

	// Remove empty entries
	var cleanPaths []string
	for _, path := range info.PathDirectories {
		if strings.TrimSpace(path) != "" {
			cleanPaths = append(cleanPaths, path)
		}
	}
	info.PathDirectories = cleanPaths

	return nil
}

func (a *Analyzer) getAvailableCommands(info *Info) error {
	commandSet := make(map[string]bool)

	// Common Windows executable extensions
	execExtensions := []string{".exe", ".bat", ".cmd", ".com", ".ps1"}

	// Get commands from PATH directories
	for _, dir := range info.PathDirectories {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				// Add both with and without extension
				for _, ext := range execExtensions {
					if strings.HasSuffix(strings.ToLower(name), ext) {
						// Add with extension
						commandSet[name] = true
						// Add without extension
						baseName := strings.TrimSuffix(name, ext)
						commandSet[baseName] = true
						break
					}
				}
			}
		}
	}

	// Add built-in commands for different shells
	builtinCommands := a.getBuiltinCommands(info.Shell)
	for _, cmd := range builtinCommands {
		commandSet[cmd] = true
	}

	// Convert map to slice
	for cmd := range commandSet {
		info.AvailableCommands = append(info.AvailableCommands, cmd)
	}

	return nil
}

func (a *Analyzer) getBuiltinCommands(shell string) []string {
	switch shell {
	case "powershell", "pwsh":
		return []string{
			"Get-ChildItem", "Set-Location", "Get-Location", "Copy-Item", "Move-Item",
			"Remove-Item", "New-Item", "Get-Content", "Set-Content", "Select-String",
			"Where-Object", "ForEach-Object", "Sort-Object", "Group-Object",
			"Measure-Object", "Compare-Object", "Get-Process", "Stop-Process",
			"Get-Service", "Start-Service", "Stop-Service", "Get-EventLog",
			"ls", "cd", "pwd", "cp", "mv", "rm", "mkdir", "cat", "grep",
		}
	case "cmd":
		return []string{
			"dir", "cd", "copy", "move", "del", "md", "rd", "type", "find",
			"findstr", "sort", "more", "attrib", "xcopy", "robocopy",
			"tasklist", "taskkill", "net", "sc", "reg", "wmic",
		}
	case "bash":
		return []string{
			"ls", "cd", "pwd", "cp", "mv", "rm", "mkdir", "rmdir", "cat",
			"grep", "find", "sort", "head", "tail", "wc", "chmod", "chown",
			"ps", "kill", "jobs", "bg", "fg", "history", "alias", "export",
		}
	default:
		return []string{}
	}
}

func (a *Analyzer) parseAliases(info *Info) error {
	switch info.Shell {
	case "powershell", "pwsh":
		return a.parsePowerShellAliases(info)
	case "cmd":
		return a.parseCmdAliases(info)
	case "bash":
		return a.parseBashAliases(info)
	default:
		return nil
	}
}

func (a *Analyzer) parsePowerShellAliases(info *Info) error {
	// PowerShell profile locations
	profilePaths := []string{
		filepath.Join(info.HomeDir, "Documents", "PowerShell", "Profile.ps1"),
		filepath.Join(info.HomeDir, "Documents", "WindowsPowerShell", "Profile.ps1"),
		filepath.Join(os.Getenv("PROGRAMFILES"), "PowerShell", "7", "Profile.ps1"),
	}

	for _, profilePath := range profilePaths {
		a.parsePowerShellProfileFile(profilePath, info.Aliases)
	}

	// Get built-in PowerShell aliases via command execution
	a.getPowerShellBuiltinAliases(info.Aliases)

	return nil
}

func (a *Analyzer) parsePowerShellProfileFile(filename string, aliases map[string]string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Parse Set-Alias commands
		if strings.HasPrefix(strings.ToLower(line), "set-alias ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				alias := strings.Trim(parts[1], `'"`)
				command := strings.Trim(parts[2], `'"`)
				aliases[alias] = command
			}
		}

		// Parse function aliases (simplified)
		if strings.HasPrefix(strings.ToLower(line), "function ") && strings.Contains(line, "{") {
			funcDef := strings.TrimPrefix(line, "function ")
			if idx := strings.Index(funcDef, "{"); idx != -1 {
				funcName := strings.TrimSpace(funcDef[:idx])
				// Simple extraction - this could be more sophisticated
				aliases[funcName] = "function"
			}
		}
	}

	return scanner.Err()
}

func (a *Analyzer) getPowerShellBuiltinAliases(aliases map[string]string) {
	// Try to get PowerShell aliases via command execution
	cmd := exec.Command("powershell", "-Command", "Get-Alias | ForEach-Object { $_.Name + '=' + $_.Definition }")
	output, err := cmd.Output()
	if err != nil {
		return // Silently fail if PowerShell is not available
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				aliases[parts[0]] = parts[1]
			}
		}
	}
}

func (a *Analyzer) parseCmdAliases(info *Info) error {
	// CMD doesn't have traditional aliases, but we can check for DOSKEY macros
	cmd := exec.Command("doskey", "/macros")
	output, err := cmd.Output()
	if err != nil {
		return nil // DOSKEY might not be available or no macros defined
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				info.Aliases[parts[0]] = parts[1]
			}
		}
	}

	return nil
}

func (a *Analyzer) parseBashAliases(info *Info) error {
	// For Git Bash or similar Unix-like shells on Windows
	rcFiles := []string{
		filepath.Join(info.HomeDir, ".bashrc"),
		filepath.Join(info.HomeDir, ".bash_profile"),
		filepath.Join(info.HomeDir, ".profile"),
		filepath.Join(info.HomeDir, ".bash_aliases"),
	}

	for _, rcFile := range rcFiles {
		a.parseAliasFile(rcFile, info.Aliases)
	}

	return nil
}

func (a *Analyzer) parseAliasFile(filename string, aliases map[string]string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "alias ") {
			// Parse alias line: alias name='command'
			aliasLine := strings.TrimPrefix(line, "alias ")
			parts := strings.SplitN(aliasLine, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				command := strings.Trim(strings.TrimSpace(parts[1]), `'"`)
				aliases[name] = command
			}
		}
	}

	return scanner.Err()
}

// Windows-specific helper function to get system information via Windows API
func (a *Analyzer) getWindowsVersion() (string, error) {
	var mod = syscall.NewLazyDLL("kernel32.dll")
	var proc = mod.NewProc("GetVersion")

	ret, _, _ := proc.Call()
	major := byte(ret)
	minor := byte(ret >> 8)

	return fmt.Sprintf("%d.%d", major, minor), nil
}

// Helper function to check if running in Windows Subsystem for Linux
func (a *Analyzer) isWSL() bool {
	if _, err := os.Stat("/proc/version"); err != nil {
		return false
	}

	content, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(content)), "microsoft")
}