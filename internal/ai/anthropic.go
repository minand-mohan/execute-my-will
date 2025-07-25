// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: internal/ai/anthropic.go
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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

type AnthropicModelsResponse struct {
	Data []struct {
		ID string `json:"id"` // The model ID, e.g., "claude-3-opus-20240229"
	} `json:"data"`
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

// List Models
func (a *AnthropicProvider) ListModels() ([]string, error) {
	fmt.Println("Fetching Claude models...")
	const maxRetries = 5
	initialDelay := 100 * time.Millisecond

	var body []byte
	var err error

	for i := 0; i < maxRetries; i++ {
		client := &http.Client{}
		req, httpErr := http.NewRequest("GET", "https://api.anthropic.com/v1/models", nil) // Note: This endpoint might not exist for listing all models
		if httpErr != nil {
			err = fmt.Errorf("failed to create Claude request: %w", httpErr)
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", i+1, err, initialDelay)
			time.Sleep(initialDelay)
			initialDelay *= 2 // Exponential backoff
			continue
		}
		req.Header.Add("x-api-key", a.apiKey)             // IMPORTANT: Use the provider's API key
		req.Header.Add("anthropic-version", "2023-06-01") // Specify the API version

		resp, httpErr := client.Do(req)
		if httpErr != nil {
			err = fmt.Errorf("failed to make HTTP request to Claude: %w", httpErr)
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", i+1, err, initialDelay)
			time.Sleep(initialDelay)
			initialDelay *= 2 // Exponential backoff
			continue
		}
		defer resp.Body.Close() // Ensure body is closed on each iteration

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			err = fmt.Errorf("Claude API returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", i+1, err, initialDelay)
			time.Sleep(initialDelay)
			initialDelay *= 2 // Exponential backoff
			continue
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("failed to read Claude response body: %w", err)
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", i+1, err, initialDelay)
			time.Sleep(initialDelay)
			initialDelay *= 2 // Exponential backoff
			continue
		}
		break // Success, exit retry loop
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch Claude models after %d retries: %w", maxRetries, err)
	}

	var claudeResp AnthropicModelsResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse Claude models response: %w", err)
	}

	var models []string
	for _, model := range claudeResp.Data {
		models = append(models, model.ID)
	}

	fmt.Println("Claude models fetched and parsed successfully.")
	return models, nil
}
