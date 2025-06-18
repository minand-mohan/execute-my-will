// File: internal/ai/openai.go
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
