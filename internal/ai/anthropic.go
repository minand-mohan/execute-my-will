// File: internal/ai/anthropic.go
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/minand-mohan/execute-my-will/internal/config"
)

// Anthropic Provider
type AnthropicProvider struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float32
}

type AnthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float32            `json:"temperature"`
	Messages    []AnthropicMessage `json:"messages"`
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	Content []AnthropicContent `json:"content"`
	Error   *AnthropicError    `json:"error,omitempty"`
}

type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AnthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewAnthropicProvider(cfg *config.Config) (*AnthropicProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("anthropic API key is required")
	}

	return &AnthropicProvider{
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
	}, nil
}

func (a *AnthropicProvider) GenerateResponse(prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	request := AnthropicRequest{
		Model:       a.model,
		MaxTokens:   a.maxTokens,
		Temperature: a.temperature,
		Messages: []AnthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var response AnthropicResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if response.Error != nil {
		return "", fmt.Errorf("anthropic API error: %s", response.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	responseText := response.Content[0].Text

	// Handle failure cases as defined in the prompt
	if responseText == "FAILURE: Intent too complex for a single shell command." {
		return "", fmt.Errorf("intent too complex for a single shell command, might need merlin")
	}

	if responseText == "FAILURE: Directory reference too vague." {
		return "", fmt.Errorf("directory reference too vague - please specify exact paths. the map instructions are not clear")
	}

	// Check for any other FAILURE responses
	if len(responseText) >= 8 && responseText[:8] == "FAILURE:" {
		return "", fmt.Errorf("command generation failed: %s", responseText[9:])
	}

	return responseText, nil
}
