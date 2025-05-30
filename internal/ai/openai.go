package ai

import (
	"fmt"

	"github.com/minand-mohan/execute-my-will/internal/config"
)

// OpenAI Provider (placeholder)
type OpenAIProvider struct {
	apiKey string
	model  string
}

func NewOpenAIProvider(cfg *config.Config) (*OpenAIProvider, error) {
	return &OpenAIProvider{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
	}, nil
}

func (o *OpenAIProvider) GenerateResponse(prompt string) (string, error) {
	// Implementation for OpenAI API
	return "", fmt.Errorf("OpenAI provider not yet implemented")
}
