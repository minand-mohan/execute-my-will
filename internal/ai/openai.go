// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: internal/ai/openai.go
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

// OpenAI Provider
type OpenAIProvider struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float32
}

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float32         `json:"temperature"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []OpenAIChoice `json:"choices"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

type OpenAIChoice struct {
	Message OpenAIMessage `json:"message"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type OpenAIModelsResponse struct {
	Data []struct {
		ID string `json:"id"` // The model ID, e.g., "gpt-4o"
	} `json:"data"`
}

func NewOpenAIProvider(cfg *config.Config) (*OpenAIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	return &OpenAIProvider{
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
	}, nil
}

func (o *OpenAIProvider) GenerateResponse(prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	request := OpenAIRequest{
		Model: o.model,
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   o.maxTokens,
		Temperature: o.temperature,
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

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

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if response.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	responseText := response.Choices[0].Message.Content

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

func (o *OpenAIProvider) ListModels() ([]string, error) {
	fmt.Println("Fetching OpenAI models...")
	const maxRetries = 5
	initialDelay := 100 * time.Millisecond

	var body []byte
	var err error

	for i := 0; i < maxRetries; i++ {
		client := &http.Client{}
		req, httpErr := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
		if httpErr != nil {
			err = fmt.Errorf("failed to create OpenAI request: %w", httpErr)
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", i+1, err, initialDelay)
			time.Sleep(initialDelay)
			initialDelay *= 2 // Exponential backoff
			continue
		}
		req.Header.Add("Authorization", "Bearer "+o.apiKey) // IMPORTANT: Use the provider's API key

		resp, httpErr := client.Do(req)
		if httpErr != nil {
			err = fmt.Errorf("failed to make HTTP request to OpenAI: %w", httpErr)
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", i+1, err, initialDelay)
			time.Sleep(initialDelay)
			initialDelay *= 2 // Exponential backoff
			continue
		}
		defer resp.Body.Close() // Ensure body is closed on each iteration

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			err = fmt.Errorf("OpenAI API returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", i+1, err, initialDelay)
			time.Sleep(initialDelay)
			initialDelay *= 2 // Exponential backoff
			continue
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("failed to read OpenAI response body: %w", err)
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", i+1, err, initialDelay)
			time.Sleep(initialDelay)
			initialDelay *= 2 // Exponential backoff
			continue
		}
		break // Success, exit retry loop
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenAI models after %d retries: %w", maxRetries, err)
	}

	var openAIResp OpenAIModelsResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI models response: %w", err)
	}

	var models []string
	for _, model := range openAIResp.Data {
		models = append(models, model.ID)
	}

	fmt.Println("OpenAI models fetched and parsed successfully.")
	return models, nil
}
