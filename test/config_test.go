// File: test/config_test.go
package test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/minand-mohan/execute-my-will/internal/config"
)

func TestConfig_New(t *testing.T) {
	cfg := config.New()

	if cfg == nil {
		t.Fatal("New() should return a config")
	}

	// Test default values
	if cfg.AIProvider != "gemini" {
		t.Errorf("Expected default AIProvider 'gemini', got '%s'", cfg.AIProvider)
	}

	if cfg.Model != "gemini-2.5-pro" {
		t.Errorf("Expected default Model 'gemini-2.5-pro', got '%s'", cfg.Model)
	}

	if cfg.MaxTokens != 1000 {
		t.Errorf("Expected default MaxTokens 1000, got %d", cfg.MaxTokens)
	}

	if cfg.Temperature != 0.1 {
		t.Errorf("Expected default Temperature 0.1, got %f", cfg.Temperature)
	}

	if cfg.APIKey != "" {
		t.Errorf("Expected empty default APIKey, got '%s'", cfg.APIKey)
	}

	if cfg.Mode != "" {
		t.Errorf("Expected empty default Mode, got '%s'", cfg.Mode)
	}
}

func TestConfig_Validate(t *testing.T) {
	testCases := []struct {
		name           string
		config         *config.Config
		shouldError    bool
		errorSubstring string
	}{
		{
			name: "valid config",
			config: &config.Config{
				AIProvider:  "gemini",
				APIKey:      "test-key",
				Model:       "gemini-pro",
				MaxTokens:   1000,
				Temperature: 0.1,
				Mode:        "monarch",
			},
			shouldError: false,
		},
		{
			name: "missing API key",
			config: &config.Config{
				AIProvider:  "gemini",
				APIKey:      "",
				Model:       "gemini-pro",
				MaxTokens:   1000,
				Temperature: 0.1,
				Mode:        "monarch",
			},
			shouldError:    true,
			errorSubstring: "API key is required",
		},
		{
			name: "missing mode",
			config: &config.Config{
				AIProvider:  "gemini",
				APIKey:      "test-key",
				Model:       "gemini-pro",
				MaxTokens:   1000,
				Temperature: 0.1,
				Mode:        "",
			},
			shouldError:    true,
			errorSubstring: "mode is required",
		},
		{
			name: "invalid mode",
			config: &config.Config{
				AIProvider:  "gemini",
				APIKey:      "test-key",
				Model:       "gemini-pro",
				MaxTokens:   1000,
				Temperature: 0.1,
				Mode:        "invalid-mode",
			},
			shouldError:    true,
			errorSubstring: "invalid mode",
		},
		{
			name: "royal-heir mode valid",
			config: &config.Config{
				AIProvider:  "openai",
				APIKey:      "test-key",
				Model:       "gpt-3.5-turbo",
				MaxTokens:   1000,
				Temperature: 0.1,
				Mode:        "royal-heir",
			},
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for config, but got none")
					return
				}
				if tc.errorSubstring != "" && !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for valid config, but got: %v", err)
				}
			}
		})
	}
}

func TestConfig_ValidateDefaults(t *testing.T) {
	// Test that Validate sets default values for missing fields
	cfg := &config.Config{
		APIKey:      "test-key",
		Mode:        "monarch",
		MaxTokens:   0,     // This should trigger default
		Temperature: -0.5,  // This should trigger default (invalid range)
		// Missing other fields
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate should not error when setting defaults: %v", err)
	}

	// Check that defaults were set
	if cfg.AIProvider != "gemini" {
		t.Errorf("Expected AIProvider to be set to 'gemini', got '%s'", cfg.AIProvider)
	}

	if cfg.MaxTokens != 1000 {
		t.Errorf("Expected MaxTokens to be set to 1000, got %d", cfg.MaxTokens)
	}

	if cfg.Temperature != 0.1 {
		t.Errorf("Expected Temperature to be set to 0.1, got %f", cfg.Temperature)
	}

	if cfg.Model == "" {
		t.Error("Expected Model to be set to default")
	}
}

func TestConfig_ValidateTemperatureRange(t *testing.T) {
	testCases := []struct {
		name        string
		temperature float32
		expected    float32
	}{
		{"valid temperature 0.0", 0.0, 0.0},
		{"valid temperature 0.5", 0.5, 0.5},
		{"valid temperature 1.0", 1.0, 1.0},
		{"invalid temperature -0.1", -0.1, 0.1},
		{"invalid temperature 1.1", 1.1, 0.1},
		{"invalid temperature 2.0", 2.0, 0.1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				APIKey:      "test-key",
				Mode:        "monarch",
				Temperature: tc.temperature,
			}

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Validate should not error: %v", err)
			}

			if cfg.Temperature != tc.expected {
				t.Errorf("Expected temperature %f, got %f", tc.expected, cfg.Temperature)
			}
		})
	}
}

func TestConfig_GetDefaultModel(t *testing.T) {
	testCases := []struct {
		provider      string
		expectedModel string
	}{
		{"gemini", "gemini-pro"},
		{"openai", "gpt-3.5-turbo"},
		{"anthropic", "claude-3-sonnet-20240229"},
		{"unknown", "gemini-pro"}, // fallback
	}

	for _, tc := range testCases {
		t.Run(tc.provider, func(t *testing.T) {
			model := config.GetDefaultModel(tc.provider)
			if model != tc.expectedModel {
				t.Errorf("Expected model '%s' for provider '%s', got '%s'", 
					tc.expectedModel, tc.provider, model)
			}
		})
	}
}

func TestConfig_GetModels(t *testing.T) {
	testCases := []struct {
		provider       string
		shouldError    bool
		expectedModels []string
	}{
		{
			provider:       "gemini",
			shouldError:    false,
			expectedModels: []string{"gemini-pro", "gemini-2.5-pro"},
		},
		{
			provider:       "openai",
			shouldError:    false,
			expectedModels: []string{"gpt-3.5-turbo", "gpt-4"},
		},
		{
			provider:       "anthropic",
			shouldError:    false,
			expectedModels: []string{"claude-3-sonnet-20240229"},
		},
		{
			provider:    "unsupported",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.provider, func(t *testing.T) {
			models, err := config.GetModels(tc.provider)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for unsupported provider '%s'", tc.provider)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for provider '%s': %v", tc.provider, err)
				return
			}

			if len(models) != len(tc.expectedModels) {
				t.Errorf("Expected %d models, got %d", len(tc.expectedModels), len(models))
				return
			}

			for i, expectedModel := range tc.expectedModels {
				if models[i] != expectedModel {
					t.Errorf("Expected model '%s' at index %d, got '%s'", 
						expectedModel, i, models[i])
				}
			}
		})
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up test config
	originalConfig := &config.Config{
		AIProvider:  "openai",
		APIKey:      "test-api-key",
		Model:       "gpt-4",
		MaxTokens:   2000,
		Temperature: 0.2,
		Mode:        "royal-heir",
	}

	// This test would require mocking the file system operations
	// since the config package uses specific paths
	// For now, we'll test the structure and validation
	t.Logf("Config save/load test would require filesystem mocking")
	t.Logf("Original config: %+v", originalConfig)
}

func TestConfigNotFoundError(t *testing.T) {
	err := &config.ConfigNotFoundError{Path: "/test/path"}

	expectedMsg := "config file not found at /test/path"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test IsConfigNotFound function
	if !config.IsConfigNotFound(err) {
		t.Error("IsConfigNotFound should return true for ConfigNotFoundError")
	}

	// Test with different error type
	otherErr := errors.New("not a config error")
	if config.IsConfigNotFound(otherErr) {
		t.Error("IsConfigNotFound should return false for non-ConfigNotFoundError")
	}
}