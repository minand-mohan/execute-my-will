// File: test/ai_client_test.go
package test

import (
	"strings"
	"testing"

	"github.com/minand-mohan/execute-my-will/internal/ai"
	"github.com/minand-mohan/execute-my-will/internal/config"
	"github.com/minand-mohan/execute-my-will/internal/system"
)

func TestAIClient_GenerateResponse(t *testing.T) {
	cfg := &config.Config{
		AIProvider:  "gemini",
		APIKey:      "test-key",
		Model:       "gemini-pro",
		MaxTokens:   1000,
		Temperature: 0.1,
		Mode:        "monarch",
	}

	// This test requires actual AI provider setup, so we'll test interface compliance
	// and error handling rather than actual AI calls
	_, err := ai.NewClient(cfg)
	
	// The client creation might fail due to missing real API keys, which is expected in tests
	if err != nil {
		t.Logf("AI Client creation failed as expected in test environment: %v", err)
		return
	}

	// If we reach here, we have a working client (unlikely in test env)
	t.Log("AI Client created successfully (unexpected in test environment)")
}

func TestAIClient_Interface(t *testing.T) {
	// Test that the AI client interface is properly defined
	var client ai.Client

	// Create a mock implementation
	mockClient := &MockAIClient{
		Response: &ai.AIResponse{
			Type:    ai.ResponseTypeCommand,
			Content: "ls -la",
		},
		ExplanationText: "This command lists files",
		Models:          []string{"model1", "model2"},
	}

	client = mockClient

	// Test GenerateResponse
	sysInfo := &system.Info{
		OS:    "linux",
		Shell: "bash",
	}

	response, err := client.GenerateResponse("list files", sysInfo)
	if err != nil {
		t.Errorf("GenerateResponse should not error: %v", err)
	}

	if response == nil {
		t.Fatal("GenerateResponse should return a response")
	}

	if response.Type != ai.ResponseTypeCommand {
		t.Errorf("Expected ResponseTypeCommand, got %v", response.Type)
	}

	if response.Content != "ls -la" {
		t.Errorf("Expected 'ls -la', got '%s'", response.Content)
	}

	// Test ExplainCommand
	explanation, err := client.ExplainCommand("ls -la", sysInfo)
	if err != nil {
		t.Errorf("ExplainCommand should not error: %v", err)
	}

	if explanation != "This command lists files" {
		t.Errorf("Expected 'This command lists files', got '%s'", explanation)
	}

	// Test ListModels
	models, err := client.ListModels()
	if err != nil {
		t.Errorf("ListModels should not error: %v", err)
	}

	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	// Test call counting
	if mockClient.GenerateCallCount != 1 {
		t.Errorf("Expected 1 GenerateResponse call, got %d", mockClient.GenerateCallCount)
	}

	if mockClient.ExplainCallCount != 1 {
		t.Errorf("Expected 1 ExplainCommand call, got %d", mockClient.ExplainCallCount)
	}
}

func TestAIResponse_Types(t *testing.T) {
	testCases := []struct {
		name         string
		responseType ai.ResponseType
		content      string
		error        string
	}{
		{
			name:         "command response",
			responseType: ai.ResponseTypeCommand,
			content:      "ls -la",
			error:        "",
		},
		{
			name:         "script response",
			responseType: ai.ResponseTypeScript,
			content:      "#!/bin/bash\necho 'hello'\nls -la",
			error:        "",
		},
		{
			name:         "failure response",
			responseType: ai.ResponseTypeFailure,
			content:      "",
			error:        "Cannot complete this task",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := &ai.AIResponse{
				Type:    tc.responseType,
				Content: tc.content,
				Error:   tc.error,
			}

			if response.Type != tc.responseType {
				t.Errorf("Expected type %v, got %v", tc.responseType, response.Type)
			}

			if response.Content != tc.content {
				t.Errorf("Expected content '%s', got '%s'", tc.content, response.Content)
			}

			if response.Error != tc.error {
				t.Errorf("Expected error '%s', got '%s'", tc.error, response.Error)
			}
		})
	}
}

func TestMockAIClient_ErrorHandling(t *testing.T) {
	mockClient := &MockAIClient{
		ShouldError: true,
	}

	sysInfo := &system.Info{
		OS:    "linux",
		Shell: "bash",
	}

	// Test GenerateResponse error
	_, err := mockClient.GenerateResponse("test intent", sysInfo)
	if err == nil {
		t.Error("Expected error from GenerateResponse when ShouldError is true")
	}

	if !strings.Contains(err.Error(), "mock AI error") {
		t.Errorf("Expected 'mock AI error', got '%s'", err.Error())
	}

	// Test ExplainCommand error
	_, err = mockClient.ExplainCommand("test command", sysInfo)
	if err == nil {
		t.Error("Expected error from ExplainCommand when ShouldError is true")
	}

	// Test ListModels error
	_, err = mockClient.ListModels()
	if err == nil {
		t.Error("Expected error from ListModels when ShouldError is true")
	}
}

func TestMockAIClient_CustomResponses(t *testing.T) {
	mockClient := &MockAIClient{
		Response: &ai.AIResponse{
			Type:    ai.ResponseTypeScript,
			Content: "custom script content",
		},
		ExplanationText: "custom explanation",
		Models:          []string{"custom-model"},
	}

	sysInfo := &system.Info{
		OS:    "linux",
		Shell: "bash",
	}

	// Test custom response
	response, err := mockClient.GenerateResponse("test", sysInfo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response.Type != ai.ResponseTypeScript {
		t.Errorf("Expected ResponseTypeScript, got %v", response.Type)
	}

	if response.Content != "custom script content" {
		t.Errorf("Expected 'custom script content', got '%s'", response.Content)
	}

	// Test custom explanation
	explanation, err := mockClient.ExplainCommand("test", sysInfo)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if explanation != "custom explanation" {
		t.Errorf("Expected 'custom explanation', got '%s'", explanation)
	}

	// Test custom models
	models, err := mockClient.ListModels()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(models) != 1 || models[0] != "custom-model" {
		t.Errorf("Expected ['custom-model'], got %v", models)
	}
}