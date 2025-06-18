// Copyright (c) 2025 Minand Nellipunath Manomohanan
// 
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: internal/system/validator.go
package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Validator struct {
	sysInfo *Info
}

func NewValidator(sysInfo *Info) *Validator {
	return &Validator{sysInfo: sysInfo}
}

func (v *Validator) ValidateIntent(intent string) error {
	// Check for directory-related operations
	if v.containsDirectoryOperation(intent) {
		return v.validateDirectoryReferences(intent)
	}

	return nil
}

func (v *Validator) containsDirectoryOperation(intent string) bool {
	keywords := []string{"move", "copy", "list", "cd", "navigate", "directory", "folder", "file"}
	lowerIntent := strings.ToLower(intent)

	for _, keyword := range keywords {
		if strings.Contains(lowerIntent, keyword) {
			return true
		}
	}

	return false
}

func (v *Validator) validateDirectoryReferences(intent string) error {
	// Extract potential directory references
	words := strings.Fields(intent)

	for _, word := range words {
		// Skip common words and known special directories
		if v.isKnownDirectory(word) || v.isCommonWord(word) {
			continue
		}

		// Check if word looks like a path
		if strings.Contains(word, "/") || strings.Contains(word, "\\") {
			// Validate that the directory exists
			if !v.pathExists(word) {
				return fmt.Errorf("the directory '%s' does not exist in your realm. Please specify an existing path or use specific directory names", word)
			}
		}
	}

	return nil
}

func (v *Validator) isKnownDirectory(word string) bool {
	known := []string{"home", "current", "present", "here", "pwd", "~", ".", "..", "/"}
	lowerWord := strings.ToLower(word)

	for _, dir := range known {
		if lowerWord == dir {
			return true
		}
	}

	return false
}

func (v *Validator) isCommonWord(word string) bool {
	common := []string{"the", "to", "from", "in", "at", "of", "for", "with", "by", "a", "an", "and", "or", "but"}
	lowerWord := strings.ToLower(word)

	for _, common := range common {
		if lowerWord == common {
			return true
		}
	}

	return false
}

func (v *Validator) pathExists(path string) bool {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		path = filepath.Join(v.sysInfo.HomeDir, path[1:])
	}

	_, err := os.Stat(path)
	return err == nil
}
