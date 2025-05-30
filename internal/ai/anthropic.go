package ai

import (
	"fmt"

	"github.com/minand-mohan/execute-my-will/internal/config"
)

// Anthropic Provider (placeholder)
type AnthropicProvider struct {
	apiKey string
	model  string
}

func NewAnthropicProvider(cfg *config.Config) (*AnthropicProvider, error) {
	return &AnthropicProvider{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
	}, nil
}

func (a *AnthropicProvider) GenerateResponse(prompt string) (string, error) {
	// Implementation for Anthropic API
	return "", fmt.Errorf("anthropic provider not yet implemented")
}
