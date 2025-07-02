// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

//go:build windows
// +build windows

package system

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type Info struct {
	OS                string
	Shell             string
	PackageManagers   []string
	CurrentDir        string
	HomeDir           string
	PathDirectories   []string
	InstalledPackages []string
	AvailableCommands []string
}

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) AnalyzeSystem() (*Info, error) {
	info := &Info{
		PackageManagers:   make([]string, 0),
		InstalledPackages: make([]string, 0),
		AvailableCommands: make([]string, 0),
	}

	var wg sync.WaitGroup
	errors := make(chan error, 5)

	info.OS = runtime.GOOS
	currentDir, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	info.CurrentDir = currentDir
	info.HomeDir = homeDir

	initial_tasks := []func(*Info) error{
		func(*Info) error { return a.detectShell(info) },
		func(*Info) error { return a.detectPackageManagers(info) },
		func(*Info) error { return a.getPathDirectories(info) },
	}

	wg.Add(len(initial_tasks))
	for _, task := range initial_tasks {
		go func(t func(*Info) error) {
			defer wg.Done()
			if err := t(info); err != nil {
				errors <- err
			}
		}(task)
	}
	wg.Wait()

	secondary_tasks := []func(*Info) error{
		func(*Info) error { return a.getInstalledPackages(info) },
		func(*Info) error { return a.getAvailableCommands(info) },
	}

	wg.Add(len(secondary_tasks))
	for _, task := range secondary_tasks {
		go func(t func(*Info) error) {
			defer wg.Done()
			if err := t(info); err != nil {
				errors <- err
			}
		}(task)
	}

	wg.Wait()

	close(errors)
	if len(errors) > 0 {
		err := <-errors
		return info, fmt.Errorf("system analysis completed with warnings: %v", err)
	}

	return info, nil
}

func (a *Analyzer) detectShell(info *Info) error {
	if os.Getenv("PSModulePath") != "" {
		info.Shell = "powershell"
		if _, err := exec.LookPath("pwsh.exe"); err == nil {
			info.Shell = "pwsh"
		}
		return nil
	}
	if comspec := os.Getenv("COMSPEC"); comspec != "" {
		shellName := filepath.Base(comspec)
		info.Shell = strings.ToLower(strings.TrimSuffix(shellName, ".exe"))
	} else {
		info.Shell = "cmd"
	}
	return nil
}

func (a *Analyzer) detectPackageManagers(info *Info) error {
	managers := []struct {
		name string
		cmd  string
	}{
		{"winget", "winget.exe"},
		{"chocolatey", "choco.exe"},
		{"scoop", "scoop.cmd"},
	}

	for _, manager := range managers {
		if _, err := exec.LookPath(manager.cmd); err == nil {
			info.PackageManagers = append(info.PackageManagers, manager.name)
		}
	}
	if len(info.PackageManagers) == 0 {
		info.PackageManagers = append(info.PackageManagers, "unknown")
	}
	return nil
}

func (a *Analyzer) getPathDirectories(info *Info) error {
	pathEnv := os.Getenv("PATH")
	if pathEnv != "" {
		info.PathDirectories = strings.Split(pathEnv, ";")
	}
	return nil
}

func (a *Analyzer) getInstalledPackages(info *Info) error {
	var wg sync.WaitGroup
	packageChan := make(chan string, 100)

	for _, manager := range info.PackageManagers {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			var cmd *exec.Cmd
			var parser func(string) []string

			switch m {
			case "winget":
				cmd = exec.Command("winget", "list", "--source", "winget", "--disable-interactivity", "--accept-source-agreements")
				parser = parseWingetOutput
			case "chocolatey":
				cmd = exec.Command("choco", "list", "--local-only", "--limit-output", "--no-progress")
				parser = parseChocoOutput
			case "scoop":
				cmd = exec.Command("scoop", "list")
				parser = parseScoopOutput
			default:
				return
			}

			var out bytes.Buffer
			cmd.Stdout = &out
			if err := cmd.Run(); err == nil {
				for _, p := range parser(out.String()) {
					packageChan <- p
				}
			}
		}(manager)
	}

	go func() {
		wg.Wait()
		close(packageChan)
	}()

	packageSet := make(map[string]bool)
	for p := range packageChan {
		packageSet[p] = true
	}

	for p := range packageSet {
		info.InstalledPackages = append(info.InstalledPackages, p)
	}

	return nil
}

func parseWingetOutput(output string) []string {
	packages := []string{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Name") || strings.HasPrefix(trimmed, "---") {
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}
	return packages
}

func parseChocoOutput(output string) []string {
	packages := []string{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) == 2 {
			packages = append(packages, strings.TrimSpace(parts[0]))
		}
	}
	return packages
}

func parseScoopOutput(output string) []string {
	packages := []string{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Name") || strings.HasPrefix(trimmed, "----") || strings.HasPrefix(trimmed, "Installed") {
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}
	return packages
}

func (a *Analyzer) getAvailableCommands(info *Info) error {
	commandSet := make(map[string]bool)
	execExtensions := []string{".exe", ".bat", ".cmd", ".com", ".ps1"}

	// Get commands from PATH directories
	for _, dir := range info.PathDirectories {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip unreadable directories
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				lowerName := strings.ToLower(name)
				for _, ext := range execExtensions {
					if strings.HasSuffix(lowerName, ext) {
						commandSet[name] = true                               // e.g., git.exe
						commandSet[strings.TrimSuffix(lowerName, ext)] = true // e.g., git
						break
					}
				}
			}
		}
	}

	// Add built-in commands for the detected shell
	for _, cmd := range a.getBuiltinCommands(info.Shell) {
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
