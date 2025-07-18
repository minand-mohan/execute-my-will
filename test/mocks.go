// File: test/mocks.go
package test

import (
	"errors"
	"fmt"

	"github.com/minand-mohan/execute-my-will/internal/ai"
	"github.com/minand-mohan/execute-my-will/internal/config"
	"github.com/minand-mohan/execute-my-will/internal/system"
)

// Mock implementations for testing

// MockSystemAnalyzer
type MockSystemAnalyzer struct {
	ShouldError bool
	SystemInfo  *system.Info
}

func (m *MockSystemAnalyzer) AnalyzeSystem() (*system.Info, error) {
	if m.ShouldError {
		return nil, errors.New("mock system analysis error")
	}
	if m.SystemInfo != nil {
		return m.SystemInfo, nil
	}
	return &system.Info{
		OS:                "linux",
		Shell:             "bash",
		PackageManagers:   []string{"apt"},
		CurrentDir:        "/home/user",
		HomeDir:           "/home/user",
		PathDirectories:   []string{"/usr/bin", "/bin"},
		InstalledPackages: []string{"vim", "curl"},
		AvailableCommands: []string{"ls", "cat", "grep"},
	}, nil
}

// MockCommandExecutor
type MockCommandExecutor struct {
	ShouldError       bool
	ExecutedCommands  []string
	ExecutedScripts   []string
	LastShell         string
	LastShowComments  bool
}

func (m *MockCommandExecutor) Execute(command string, shell string) error {
	m.ExecutedCommands = append(m.ExecutedCommands, command)
	m.LastShell = shell
	if m.ShouldError {
		return errors.New("mock execution error")
	}
	return nil
}

func (m *MockCommandExecutor) ExecuteScript(scriptContent string, shell string, showComments bool) error {
	m.ExecutedScripts = append(m.ExecutedScripts, scriptContent)
	m.LastShell = shell
	m.LastShowComments = showComments
	if m.ShouldError {
		return errors.New("mock script execution error")
	}
	return nil
}

// MockEnvironmentValidator
type MockEnvironmentValidator struct {
	ShouldError     bool
	InvalidCommands map[string]string // command -> reason
}

func (m *MockEnvironmentValidator) ValidateEnvironmentCommand(command string) error {
	if m.ShouldError {
		return errors.New("mock validation error")
	}
	if m.InvalidCommands != nil {
		if reason, exists := m.InvalidCommands[command]; exists {
			return &system.EnvironmentCommandError{
				Command:     command,
				Reason:      reason,
				Explanation: "mock validation error",
			}
		}
	}
	return nil
}

// MockIntentValidator
type MockIntentValidator struct {
	ShouldError    bool
	InvalidIntents map[string]string // intent -> error message
}

func (m *MockIntentValidator) ValidateIntent(intent string) error {
	if m.ShouldError {
		return errors.New("mock intent validation error")
	}
	if m.InvalidIntents != nil {
		if errMsg, exists := m.InvalidIntents[intent]; exists {
			return errors.New(errMsg)
		}
	}
	return nil
}

// MockAIClient
type MockAIClient struct {
	ShouldError       bool
	Response          *ai.AIResponse
	ExplanationText   string
	Models            []string
	GenerateCallCount int
	ExplainCallCount  int
}

func (m *MockAIClient) GenerateResponse(intent string, sysInfo *system.Info) (*ai.AIResponse, error) {
	m.GenerateCallCount++
	if m.ShouldError {
		return nil, errors.New("mock AI error")
	}
	if m.Response != nil {
		return m.Response, nil
	}
	return &ai.AIResponse{
		Type:    ai.ResponseTypeCommand,
		Content: fmt.Sprintf("mock command for: %s", intent),
	}, nil
}

func (m *MockAIClient) ExplainCommand(command string, sysInfo *system.Info) (string, error) {
	m.ExplainCallCount++
	if m.ShouldError {
		return "", errors.New("mock explanation error")
	}
	if m.ExplanationText != "" {
		return m.ExplanationText, nil
	}
	return fmt.Sprintf("This command does: %s", command), nil
}

func (m *MockAIClient) ListModels() ([]string, error) {
	if m.ShouldError {
		return nil, errors.New("mock list models error")
	}
	if m.Models != nil {
		return m.Models, nil
	}
	return []string{"model1", "model2"}, nil
}

// MockConfig
type MockConfig struct {
	AIProvider  string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float32
	Mode        string
	ShouldError bool
}

func (m *MockConfig) Validate() error {
	if m.ShouldError {
		return errors.New("mock config validation error")
	}
	if m.APIKey == "" {
		return errors.New("API key is required")
	}
	if m.Mode == "" {
		return errors.New("mode is required")
	}
	return nil
}

func (m *MockConfig) ToConfig() *config.Config {
	return &config.Config{
		AIProvider:  m.AIProvider,
		APIKey:      m.APIKey,
		Model:       m.Model,
		MaxTokens:   m.MaxTokens,
		Temperature: m.Temperature,
		Mode:        m.Mode,
	}
}