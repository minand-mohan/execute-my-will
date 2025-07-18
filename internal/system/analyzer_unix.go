// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

//go:build !windows
// +build !windows

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

func NewAnalyzer() SystemAnalyzer {
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
	shell := os.Getenv("SHELL")
	if shell == "" {
		info.Shell = "bash" // default
		return nil
	}

	info.Shell = filepath.Base(shell)
	return nil
}

func (a *Analyzer) detectPackageManagers(info *Info) error {
	managers := []string{"apt", "yum", "dnf", "pacman", "brew", "zypper"}
	for _, manager := range managers {
		if _, err := exec.LookPath(manager); err == nil {
			info.PackageManagers = append(info.PackageManagers, manager)
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
		info.PathDirectories = strings.Split(pathEnv, ":")
	}
	return nil
}

func (a *Analyzer) getInstalledPackages(info *Info) error {
	var wg sync.WaitGroup

	packageChan := make(chan string, 50)

	for _, manager := range info.PackageManagers {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			var cmd *exec.Cmd
			switch m {
			case "apt":
				cmd = exec.Command("sh", "-c", "apt-mark showmanual")
			case "yum", "dnf":
				cmd = exec.Command("sh", "-c", "dnf repoquery --userinstalled --queryformat '%{name}'")
			case "brew":
				cmd = exec.Command("brew", "list", "--formula", "-1")
			case "pacman":
				cmd = exec.Command("pacman", "-Qqe")
			default:
				return
			}

			var out bytes.Buffer
			cmd.Stdout = &out
			if err := cmd.Run(); err == nil {
				packages := strings.Split(out.String(), "\n")
				for _, p := range packages {
					if pkgName := strings.TrimSpace(p); pkgName != "" {
						packageChan <- pkgName
					}
				}
			}
		}(manager)
	}

	go func() {
		wg.Wait()
		close(packageChan)
	}()

	// Use a map to prevent duplicates
	packageSet := make(map[string]bool)
	for p := range packageChan {
		packageSet[p] = true
	}

	for p := range packageSet {
		info.InstalledPackages = append(info.InstalledPackages, p)
	}

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
			// On Unix, any file that is not a directory could be an executable script
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
