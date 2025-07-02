// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AIProvider  string  `yaml:"provider"`
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float32 `yaml:"temperature"`
	Mode        string  `yaml:"mode"` // field for monarch/royal-heir modes
}

type ConfigFile struct {
	AI Config `yaml:"ai"`
}

// New creates a new config with default values
func New() *Config {
	return &Config{
		AIProvider:  "gemini",
		APIKey:      "",
		Model:       "gemini-pro",
		MaxTokens:   1000,
		Temperature: 0.1,
		Mode:        "", // Empty by default, requires configuration
	}
}

// Load loads configuration from file
func Load() (*Config, error) {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, &ConfigNotFoundError{Path: configPath}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configFile ConfigFile
	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg := configFile.AI

	// Set default model if not provided
	if cfg.Model == "" {
		cfg.Model = GetDefaultModel(cfg.AIProvider)
	}

	return &cfg, nil
}

// Save saves configuration to file
func Save(cfg *Config) error {
	configPath := getConfigPath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := ConfigFile{AI: *cfg}

	data, err := yaml.Marshal(&configFile)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required. Run 'execute-my-will configure' to set it up")
	}

	if c.Mode == "" {
		return fmt.Errorf("mode is required. I must know who I serve. Run 'execute-my-will configure' to set your preferred mode (monarch or royal-heir)")
	}

	if c.Mode != "monarch" && c.Mode != "royal-heir" {
		return fmt.Errorf("invalid mode '%s'. I only serve the 'monarch' or the 'royal-heir'", c.Mode)
	}

	if c.AIProvider == "" {
		c.AIProvider = "gemini"
	}

	if c.MaxTokens <= 0 {
		c.MaxTokens = 1000
	}

	if c.Temperature < 0 || c.Temperature > 1 {
		c.Temperature = 0.1
	}

	// Set default model if not provided
	if c.Model == "" {
		c.Model = GetDefaultModel(c.AIProvider)
	}

	return nil
}

// GetDefaultModel returns the default model for a provider
func GetDefaultModel(provider string) string {
	switch provider {
	case "gemini":
		return "gemini-pro"
	case "openai":
		return "gpt-3.5-turbo"
	case "anthropic":
		return "claude-3-sonnet-20240229"
	default:
		return "gemini-pro"
	}
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return "config.yaml"
	}
	return filepath.Join(home, ".config/execute-my-will/config.yaml")
}

// ConfigNotFoundError represents a missing config file error
type ConfigNotFoundError struct {
	Path string
}

func (e *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("config file not found at %s", e.Path)
}

// IsConfigNotFound checks if the error is a config not found error
func IsConfigNotFound(err error) bool {
	_, ok := err.(*ConfigNotFoundError)
	return ok
}
