// File: test/command_executor_test.go
package test

import (
	"strings"
	"testing"

	"github.com/minand-mohan/execute-my-will/internal/system"
)

func TestExecutor_Interface(t *testing.T) {
	// Test that NewExecutor returns the CommandExecutor interface
	var executor system.CommandExecutor = system.NewExecutor()

	// Test interface methods exist (we can't actually execute in tests)
	if executor == nil {
		t.Error("NewExecutor should return a non-nil CommandExecutor")
	}
}

func TestMockExecutor_Execute(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}

	err := mockExecutor.Execute("ls -la", "bash")
	if err != nil {
		t.Errorf("Mock executor should not error by default: %v", err)
	}

	// Check that command was recorded
	if len(mockExecutor.ExecutedCommands) != 1 {
		t.Errorf("Expected 1 executed command, got %d", len(mockExecutor.ExecutedCommands))
	}

	if mockExecutor.ExecutedCommands[0] != "ls -la" {
		t.Errorf("Expected command 'ls -la', got '%s'", mockExecutor.ExecutedCommands[0])
	}

	if mockExecutor.LastShell != "bash" {
		t.Errorf("Expected shell 'bash', got '%s'", mockExecutor.LastShell)
	}
}

func TestMockExecutor_ExecuteScript(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}

	scriptContent := "#!/bin/bash\necho 'hello'\nls -la"
	err := mockExecutor.ExecuteScript(scriptContent, "bash", true)
	if err != nil {
		t.Errorf("Mock executor should not error by default: %v", err)
	}

	// Check that script was recorded
	if len(mockExecutor.ExecutedScripts) != 1 {
		t.Errorf("Expected 1 executed script, got %d", len(mockExecutor.ExecutedScripts))
	}

	if mockExecutor.ExecutedScripts[0] != scriptContent {
		t.Errorf("Expected script content '%s', got '%s'", scriptContent, mockExecutor.ExecutedScripts[0])
	}

	if mockExecutor.LastShell != "bash" {
		t.Errorf("Expected shell 'bash', got '%s'", mockExecutor.LastShell)
	}

	if !mockExecutor.LastShowComments {
		t.Error("Expected LastShowComments to be true")
	}
}

func TestMockExecutor_MultipleCommands(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}

	commands := []string{"ls -la", "pwd", "whoami"}
	for _, cmd := range commands {
		err := mockExecutor.Execute(cmd, "bash")
		if err != nil {
			t.Errorf("Unexpected error for command '%s': %v", cmd, err)
		}
	}

	// Check all commands were recorded
	if len(mockExecutor.ExecutedCommands) != len(commands) {
		t.Errorf("Expected %d executed commands, got %d", len(commands), len(mockExecutor.ExecutedCommands))
	}

	for i, expectedCmd := range commands {
		if mockExecutor.ExecutedCommands[i] != expectedCmd {
			t.Errorf("Expected command '%s' at index %d, got '%s'",
				expectedCmd, i, mockExecutor.ExecutedCommands[i])
		}
	}
}

func TestMockExecutor_ErrorHandling(t *testing.T) {
	mockExecutor := &MockCommandExecutor{
		ShouldError: true,
	}

	// Test Execute error
	err := mockExecutor.Execute("test command", "bash")
	if err == nil {
		t.Error("Expected error when ShouldError is true")
	}

	if !strings.Contains(err.Error(), "mock execution error") {
		t.Errorf("Expected 'mock execution error', got '%s'", err.Error())
	}

	// Test ExecuteScript error
	err = mockExecutor.ExecuteScript("test script", "bash", false)
	if err == nil {
		t.Error("Expected error when ShouldError is true")
	}

	if !strings.Contains(err.Error(), "mock script execution error") {
		t.Errorf("Expected 'mock script execution error', got '%s'", err.Error())
	}

	// Commands should still be recorded even when errors occur
	if len(mockExecutor.ExecutedCommands) != 1 {
		t.Errorf("Expected 1 executed command even with error, got %d", len(mockExecutor.ExecutedCommands))
	}

	if len(mockExecutor.ExecutedScripts) != 1 {
		t.Errorf("Expected 1 executed script even with error, got %d", len(mockExecutor.ExecutedScripts))
	}
}

func TestMockExecutor_ShellTracking(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}

	shells := []string{"bash", "zsh", "fish", "sh"}

	for i, shell := range shells {
		if i%2 == 0 {
			err := mockExecutor.Execute("test command", shell)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		} else {
			err := mockExecutor.ExecuteScript("test script", shell, false)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		}

		// Check that LastShell is updated correctly
		if mockExecutor.LastShell != shell {
			t.Errorf("Expected LastShell '%s', got '%s'", shell, mockExecutor.LastShell)
		}
	}
}

func TestMockExecutor_ShowCommentsTracking(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}

	testCases := []struct {
		name         string
		showComments bool
	}{
		{"with comments", true},
		{"without comments", false},
		{"with comments again", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := mockExecutor.ExecuteScript("test script", "bash", tc.showComments)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if mockExecutor.LastShowComments != tc.showComments {
				t.Errorf("Expected LastShowComments %v, got %v",
					tc.showComments, mockExecutor.LastShowComments)
			}
		})
	}
}

func TestMockExecutor_StatefulBehavior(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}

	// Execute multiple operations and verify state is maintained
	mockExecutor.Execute("first command", "bash")
	mockExecutor.ExecuteScript("first script", "zsh", true)
	mockExecutor.Execute("second command", "fish")

	// Check final state
	expectedCommands := []string{"first command", "second command"}
	expectedScripts := []string{"first script"}

	if len(mockExecutor.ExecutedCommands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(mockExecutor.ExecutedCommands))
	}

	if len(mockExecutor.ExecutedScripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(mockExecutor.ExecutedScripts))
	}

	for i, expectedCmd := range expectedCommands {
		if mockExecutor.ExecutedCommands[i] != expectedCmd {
			t.Errorf("Expected command '%s' at index %d, got '%s'",
				expectedCmd, i, mockExecutor.ExecutedCommands[i])
		}
	}

	if mockExecutor.ExecutedScripts[0] != expectedScripts[0] {
		t.Errorf("Expected script '%s', got '%s'",
			expectedScripts[0], mockExecutor.ExecutedScripts[0])
	}

	// Final shell should be from last operation
	if mockExecutor.LastShell != "fish" {
		t.Errorf("Expected final shell 'fish', got '%s'", mockExecutor.LastShell)
	}

	// ShowComments should be from last script execution
	if !mockExecutor.LastShowComments {
		t.Error("Expected LastShowComments to be true from script execution")
	}
}
