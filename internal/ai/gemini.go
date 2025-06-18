// Copyright (c) 2025 Minand Nellipunath Manomohanan
// 
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: internal/ai/gemini.go
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/minand-mohan/execute-my-will/internal/config"
)

// Gemini Provider
type GeminiProvider struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float32
}

type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens"`
	Temperature     float32 `json:"temperature"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

func NewGeminiProvider(cfg *config.Config) (*GeminiProvider, error) {
	return &GeminiProvider{
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
	}, nil
}

func (g *GeminiProvider) GenerateResponse(prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.model, g.apiKey)

	request := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			MaxOutputTokens: g.maxTokens,
			Temperature:     g.temperature,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response GeminiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	responseText := response.Candidates[0].Content.Parts[0].Text

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

	return response.Candidates[0].Content.Parts[0].Text, nil
}
