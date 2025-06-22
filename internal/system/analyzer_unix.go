// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

//go:build !windows
// +build !windows
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
	shell := os.Getenv("SHELL")
	if shell == "" {
		info.Shell = "bash" // default
		return nil
	}

	info.Shell = filepath.Base(shell)
	return nil
}

func (a *Analyzer) detectPackageManager(info *Info) error {
	managers := []string{"apt", "yum", "dnf", "pacman", "brew", "zypper"}

	for _, manager := range managers {
		if _, err := exec.LookPath(manager); err == nil {
			info.PackageManager = manager
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

	info.PathDirectories = strings.Split(pathEnv, ":")
	return nil
}

func (a *Analyzer) getAvailableCommands(info *Info) error {
	commandSet := make(map[string]bool)

	// Get commands from PATH directories
	for _, dir := range info.PathDirectories {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				commandSet[entry.Name()] = true
			}
		}
	}

	// Convert map to slice
	for cmd := range commandSet {
		info.AvailableCommands = append(info.AvailableCommands, cmd)
	}

	return nil
}

func (a *Analyzer) parseAliases(info *Info) error {
	rcFiles := []string{
		filepath.Join(info.HomeDir, ".bashrc"),
		filepath.Join(info.HomeDir, ".zshrc"),
		filepath.Join(info.HomeDir, ".bash_aliases"),
	}

	for _, rcFile := range rcFiles {
		if err := a.parseAliasFile(rcFile, info.Aliases); err != nil {
			// Continue with other files if one fails
			continue
		}
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
