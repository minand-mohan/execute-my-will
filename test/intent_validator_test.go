// File: test/intent_validator_test.go
package test

import (
	"strings"
	"testing"

	"github.com/minand-mohan/execute-my-will/internal/system"
)

func TestValidator_ValidateIntent(t *testing.T) {
	sysInfo := &system.Info{
		OS:         "linux",
		Shell:      "bash",
		CurrentDir: "/home/user",
		HomeDir:    "/home/user",
	}

	validator := system.NewValidator(sysInfo)

	testCases := []struct {
		name        string
		intent      string
		shouldError bool
		errorSubstr string
	}{
		{
			name:        "simple command intent",
			intent:      "list files",
			shouldError: false,
		},
		{
			name:        "install package intent",
			intent:      "install docker",
			shouldError: false,
		},
		{
			name:        "current directory reference",
			intent:      "list files in current directory",
			shouldError: false,
		},
		{
			name:        "home directory reference",
			intent:      "navigate to home directory",
			shouldError: false,
		},
		{
			name:        "valid existing path",
			intent:      "list files in /usr/bin",
			shouldError: false,
		},
		{
			name:        "non-existent path",
			intent:      "navigate to /non/existent/path",
			shouldError: true,
			errorSubstr: "does not exist",
		},
		{
			name:        "non-existent directory in copy operation",
			intent:      "copy file to /invalid/directory/path",
			shouldError: true,
			errorSubstr: "does not exist",
		},
		{
			name:        "tilde home reference to non-existent",
			intent:      "move file to ~/Documents",
			shouldError: true,
			errorSubstr: "does not exist",
		},
		{
			name:        "relative path reference to non-existent",
			intent:      "list files in ./src",
			shouldError: true,
			errorSubstr: "does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateIntent(tc.intent)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for intent '%s', but got none", tc.intent)
					return
				}
				if tc.errorSubstr != "" && !strings.Contains(err.Error(), tc.errorSubstr) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorSubstr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for intent '%s', but got: %v", tc.intent, err)
				}
			}
		})
	}
}

func TestValidator_DirectoryOperationDetection(t *testing.T) {
	sysInfo := &system.Info{
		OS:         "linux",
		Shell:      "bash",
		CurrentDir: "/home/user",
		HomeDir:    "/home/user",
	}

	validator := system.NewValidator(sysInfo)

	testCases := []struct {
		name                string
		intent              string
		shouldCheckDir      bool // Whether directory validation should be triggered
		expectValidation    bool // Whether we expect validation to pass
	}{
		{
			name:                "move operation with path",
			intent:              "move file.txt to /tmp/folder",
			shouldCheckDir:      true,
			expectValidation:    false, // "/tmp/folder" likely doesn't exist
		},
		{
			name:                "copy operation with path", 
			intent:              "copy data.txt to ./backup",
			shouldCheckDir:      true,
			expectValidation:    false, // "./backup" doesn't exist
		},
		{
			name:                "list operation",
			intent:              "list contents of directory",
			shouldCheckDir:      true,
			expectValidation:    true, // "directory" is a common word, should be ignored
		},
		{
			name:                "navigate operation with path",
			intent:              "navigate to /some/folder",
			shouldCheckDir:      true,
			expectValidation:    false, // "/some/folder" doesn't exist  
		},
		{
			name:                "cd operation",
			intent:              "cd to home",
			shouldCheckDir:      true,
			expectValidation:    true, // "home" is a known directory
		},
		{
			name:                "non-directory operation",
			intent:              "show system information",
			shouldCheckDir:      false,
			expectValidation:    true, // No directory validation needed
		},
		{
			name:                "install package",
			intent:              "install nginx package",
			shouldCheckDir:      false,
			expectValidation:    true, // No directory validation needed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateIntent(tc.intent)

			if tc.shouldCheckDir && !tc.expectValidation {
				if err == nil {
					t.Errorf("Expected directory validation error for intent '%s', but got none", tc.intent)
				}
			} else if tc.expectValidation {
				if err != nil {
					t.Errorf("Expected no error for intent '%s', but got: %v", tc.intent, err)
				}
			}
		})
	}
}

func TestValidator_KnownDirectories(t *testing.T) {
	sysInfo := &system.Info{
		OS:         "linux",
		Shell:      "bash", 
		CurrentDir: "/home/user",
		HomeDir:    "/home/user",
	}

	validator := system.NewValidator(sysInfo)

	knownDirIntents := []string{
		"move file to home",
		"navigate to current directory",
		"list files in present directory",
		"copy data to here",
		"go to ~",
		"list contents of .",
		"navigate to ..",
		"move file to /",
	}

	for _, intent := range knownDirIntents {
		t.Run(intent, func(t *testing.T) {
			err := validator.ValidateIntent(intent)
			if err != nil {
				t.Errorf("Known directory intent '%s' should not error, got: %v", intent, err)
			}
		})
	}
}

func TestValidator_Interface(t *testing.T) {
	sysInfo := &system.Info{
		OS:         "linux",
		Shell:      "bash",
		CurrentDir: "/home/user",
		HomeDir:    "/home/user",
	}

	// Test that NewValidator returns the IntentValidator interface
	var validator system.IntentValidator = system.NewValidator(sysInfo)

	err := validator.ValidateIntent("test intent")
	if err != nil {
		// Error is ok here, just testing interface works
		t.Logf("Interface method works (error expected for some cases): %v", err)
	}
}

func TestValidator_EdgeCases(t *testing.T) {
	sysInfo := &system.Info{
		OS:         "linux",
		Shell:      "bash",
		CurrentDir: "/home/user",
		HomeDir:    "/home/user",
	}

	validator := system.NewValidator(sysInfo)

	edgeCases := []struct {
		name   string
		intent string
	}{
		{"empty intent", ""},
		{"whitespace only", "   "},
		{"single word", "help"},
		{"very long intent", strings.Repeat("word ", 100)},
		{"special characters", "do something with @#$%^&*()"},
		{"unicode characters", "créer un fichier"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic or crash
			err := validator.ValidateIntent(tc.intent)
			// We don't care about the specific result, just that it doesn't crash
			t.Logf("Intent '%s' validation result: %v", tc.intent, err)
		})
	}
}