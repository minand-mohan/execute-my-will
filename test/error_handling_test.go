// File: test/error_handling_test.go
package test

import (
	"strings"
	"testing"

	"github.com/minand-mohan/execute-my-will/internal/ai"
	"github.com/minand-mohan/execute-my-will/internal/system"
)

// Test comprehensive error handling scenarios
func TestErrorHandlingScenarios(t *testing.T) {
	testCases := []struct {
		name                string
		systemAnalyzerError bool
		validatorError      bool
		aiClientError       bool
		envValidatorError   bool
		executorError       bool
		expectedPhase       string // Which phase should fail
	}{
		{
			name:                "system analysis failure",
			systemAnalyzerError: true,
			expectedPhase:       "system_analysis",
		},
		{
			name:           "intent validation failure",
			validatorError: true,
			expectedPhase:  "intent_validation",
		},
		{
			name:          "ai client failure",
			aiClientError: true,
			expectedPhase: "ai_generation",
		},
		{
			name:          "all systems working",
			expectedPhase: "success",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock components with configured errors
			mockAnalyzer := &MockSystemAnalyzer{
				ShouldError: tc.systemAnalyzerError,
			}

			mockValidator := &MockIntentValidator{
				ShouldError: tc.validatorError,
			}

			mockAIClient := &MockAIClient{
				ShouldError: tc.aiClientError,
				Response: &ai.AIResponse{
					Type:    ai.ResponseTypeCommand,
					Content: "ls -la",
				},
			}

			mockEnvValidator := &MockEnvironmentValidator{
				ShouldError: tc.envValidatorError,
				InvalidCommands: map[string]string{
					"ls -la": "test_error", // Make the default command fail when ShouldError is true
				},
			}

			mockExecutor := &MockCommandExecutor{
				ShouldError: tc.executorError,
			}

			// Simulate the execution flow
			phase, err := simulateExecutionFlow(
				mockAnalyzer,
				mockValidator,
				mockAIClient,
				mockEnvValidator,
				mockExecutor,
				"test intent",
			)

			if tc.expectedPhase == "success" {
				if err != nil {
					t.Errorf("Expected success but got error in phase %s: %v", phase, err)
				}
				if phase != "success" {
					t.Errorf("Expected success phase, got %s", phase)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error in phase %s, but execution succeeded", tc.expectedPhase)
				}
				if phase != tc.expectedPhase {
					t.Errorf("Expected error in phase %s, but got error in phase %s", tc.expectedPhase, phase)
				}
			}
		})
	}
}

// Test complex AI response handling
func TestComplexAIResponseHandling(t *testing.T) {
	testCases := []struct {
		name            string
		response        *ai.AIResponse
		expectExecution bool
		expectError     bool
	}{
		{
			name: "valid command response",
			response: &ai.AIResponse{
				Type:    ai.ResponseTypeCommand,
				Content: "ls -la",
			},
			expectExecution: true,
			expectError:     false,
		},
		{
			name: "valid script response",
			response: &ai.AIResponse{
				Type:    ai.ResponseTypeScript,
				Content: "#!/bin/bash\necho 'hello'\nls -la",
			},
			expectExecution: true,
			expectError:     false,
		},
		{
			name: "failure response",
			response: &ai.AIResponse{
				Type:  ai.ResponseTypeFailure,
				Error: "Task too dangerous",
			},
			expectExecution: false,
			expectError:     false, // Failure is handled gracefully
		},
		{
			name: "empty command response",
			response: &ai.AIResponse{
				Type:    ai.ResponseTypeCommand,
				Content: "",
			},
			expectExecution: true,
			expectError:     false, // Empty command should be handled
		},
		{
			name: "command blocked by environment validator",
			response: &ai.AIResponse{
				Type:    ai.ResponseTypeCommand,
				Content: "export PATH=$PATH:/usr/local/bin",
			},
			expectExecution: false,
			expectError:     false, // Blocked commands are handled gracefully
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAIClient := &MockAIClient{
				Response: tc.response,
			}

			mockEnvValidator := &MockEnvironmentValidator{
				InvalidCommands: map[string]string{
					"export PATH=$PATH:/usr/local/bin": "export",
				},
			}

			mockExecutor := &MockCommandExecutor{}

			executed, err := simulateResponseHandling(mockAIClient, mockEnvValidator, mockExecutor, "test intent")

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tc.expectExecution && !executed {
				t.Error("Expected execution but nothing was executed")
			}

			if !tc.expectExecution && executed {
				t.Error("Expected no execution but something was executed")
			}
		})
	}
}

// Test edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("very long intent", func(t *testing.T) {
		longIntent := strings.Repeat("do something complex ", 1000)

		mockValidator := &MockIntentValidator{}
		mockAIClient := &MockAIClient{
			Response: &ai.AIResponse{
				Type:    ai.ResponseTypeCommand,
				Content: "echo 'processed'",
			},
		}

		// Should handle very long intents gracefully
		err := mockValidator.ValidateIntent(longIntent)
		if err != nil {
			t.Errorf("Should handle long intents: %v", err)
		}

		_, err = mockAIClient.GenerateResponse(longIntent, &system.Info{})
		if err != nil {
			t.Errorf("AI client should handle long intents: %v", err)
		}
	})

	t.Run("special characters in intent", func(t *testing.T) {
		specialIntent := "create file with name 'test@#$%^&*()_+-={}[]|\\:;\"'<>?,./' and content"

		mockValidator := &MockIntentValidator{}
		err := mockValidator.ValidateIntent(specialIntent)
		if err != nil {
			t.Errorf("Should handle special characters: %v", err)
		}
	})

	t.Run("unicode characters in intent", func(t *testing.T) {
		unicodeIntent := "créer un fichier nommé 'тест' avec du contenu 中文"

		mockValidator := &MockIntentValidator{}
		err := mockValidator.ValidateIntent(unicodeIntent)
		if err != nil {
			t.Errorf("Should handle unicode characters: %v", err)
		}
	})

	t.Run("empty system info", func(t *testing.T) {
		emptyInfo := &system.Info{}

		mockAIClient := &MockAIClient{
			Response: &ai.AIResponse{
				Type:    ai.ResponseTypeCommand,
				Content: "ls",
			},
		}

		_, err := mockAIClient.GenerateResponse("test", emptyInfo)
		if err != nil {
			t.Errorf("Should handle empty system info: %v", err)
		}
	})
}

// Test concurrent operations
func TestConcurrentOperations(t *testing.T) {
	mockAnalyzer := &MockSystemAnalyzer{}

	// Test concurrent system analysis calls
	results := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := mockAnalyzer.AnalyzeSystem()
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < 10; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Concurrent system analysis failed: %v", err)
		}
	}
}

// Helper functions for testing

func simulateExecutionFlow(
	analyzer system.SystemAnalyzer,
	validator system.IntentValidator,
	aiClient ai.Client,
	envValidator system.EnvironmentValidatorInterface,
	executor system.CommandExecutor,
	intent string,
) (string, error) {

	// Phase 1: System Analysis
	_, err := analyzer.AnalyzeSystem()
	if err != nil {
		return "system_analysis", err
	}

	// Phase 2: Intent Validation
	err = validator.ValidateIntent(intent)
	if err != nil {
		return "intent_validation", err
	}

	// Phase 3: AI Generation
	sysInfo := &system.Info{OS: "linux", Shell: "bash"}
	response, err := aiClient.GenerateResponse(intent, sysInfo)
	if err != nil {
		return "ai_generation", err
	}

	// Handle failure responses
	if response.Type == ai.ResponseTypeFailure {
		return "success", nil // Failure responses are handled gracefully
	}

	// Phase 4: Environment Validation
	if response.Type == ai.ResponseTypeCommand {
		err = envValidator.ValidateEnvironmentCommand(response.Content)
		if err != nil {
			// Check if it's an EnvironmentCommandError (handled gracefully) or other error
			if _, ok := err.(*system.EnvironmentCommandError); ok {
				return "success", nil // Environment blocks are handled gracefully
			}
			return "env_validation", err // Other validation errors are failures
		}
	}

	// Phase 5: Execution
	if response.Type == ai.ResponseTypeCommand {
		err = executor.Execute(response.Content, "bash")
	} else if response.Type == ai.ResponseTypeScript {
		err = executor.ExecuteScript(response.Content, "bash", false)
	}

	if err != nil {
		return "execution", err
	}

	return "success", nil
}

func simulateResponseHandling(
	aiClient ai.Client,
	envValidator system.EnvironmentValidatorInterface,
	executor system.CommandExecutor,
	intent string,
) (bool, error) {

	sysInfo := &system.Info{OS: "linux", Shell: "bash"}
	response, err := aiClient.GenerateResponse(intent, sysInfo)
	if err != nil {
		return false, err
	}

	// Handle failure responses
	if response.Type == ai.ResponseTypeFailure {
		return false, nil
	}

	// Environment validation
	if response.Type == ai.ResponseTypeCommand {
		err = envValidator.ValidateEnvironmentCommand(response.Content)
		if err != nil {
			return false, nil // Blocked by environment validator
		}
	}

	// Execution
	executed := false
	if response.Type == ai.ResponseTypeCommand {
		err = executor.Execute(response.Content, "bash")
		executed = true
	} else if response.Type == ai.ResponseTypeScript {
		err = executor.ExecuteScript(response.Content, "bash", false)
		executed = true
	}

	return executed, err
}
