// File: internal/config/config.go
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	AIProvider  string  `mapstructure:"provider"`
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float32 `mapstructure:"temperature"`
}

// FromViper creates a Config from viper settings
func FromViper() *Config {
	var cfg Config

	// Map the nested "ai" section to our config struct
	if err := viper.UnmarshalKey("ai", &cfg); err != nil {
		// If unmarshal fails, create config manually
		cfg = Config{
			AIProvider:  viper.GetString("ai.provider"),
			APIKey:      viper.GetString("ai.api_key"),
			Model:       viper.GetString("ai.model"),
			MaxTokens:   viper.GetInt("ai.max_tokens"),
			Temperature: float32(viper.GetFloat64("ai.temperature")),
		}
	}

	// Set default model if not provided
	if cfg.Model == "" {
		cfg.Model = getDefaultModel(cfg.AIProvider)
	}

	return &cfg
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.APIKey == "" {
		// Check environment variable as fallback
		if envKey := os.Getenv("EXECUTE_MY_WILL_API_KEY"); envKey != "" {
			c.APIKey = envKey
		} else {
			return fmt.Errorf("API key not configured. Set it via:\n" +
				"  1. Flag: --api-key YOUR_KEY\n" +
				"  2. Config file: ~/.execute-my-will.yaml\n" +
				"  3. Environment: EXECUTE_MY_WILL_API_KEY=YOUR_KEY\n\n" +
				"Example config file:\n" +
				"ai:\n" +
				"  provider: gemini\n" +
				"  api_key: your-api-key-here\n" +
				"  model: gemini-pro")
		}
	}

	if c.AIProvider == "" {
		c.AIProvider = "gemini"
	}

	if c.MaxTokens <= 0 {
		c.MaxTokens = 1000
	}

	return nil
}

func getDefaultModel(provider string) string {
	switch provider {
	case "gemini":
		return "gemini-2.0-flash"
	case "openai":
		return "gpt-3.5-turbo"
	case "anthropic":
		return "claude-3-sonnet-20240229"
	default:
		return "gemini-2.0-flash"
	}
}
