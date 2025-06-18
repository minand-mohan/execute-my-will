// Copyright (c) 2025 Minand Nellipunath Manomohanan
// 
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: internal/cli/configure.go
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/minand-mohan/execute-my-will/internal/config"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure your digital knight's settings",
	Long:  "Set up or modify the configuration for your faithful digital assistant.",
	RunE:  runConfigure,
}

func init() {
	// Add flags for non-interactive configuration
	configureCmd.Flags().String("provider", "", "AI provider (gemini, openai, anthropic)")
	configureCmd.Flags().String("api-key", "", "API key for the AI provider")
	configureCmd.Flags().String("model", "", "Model to use (uses provider defaults if not specified)")
	configureCmd.Flags().Int("max-tokens", 0, "Maximum tokens for AI response")
	configureCmd.Flags().Float32("temperature", -1, "Temperature for AI response (0.0-1.0)")
	configureCmd.Flags().String("mode", "", "Execution mode: monarch or royal-heir")
}

func runConfigure(cmd *cobra.Command, args []string) error {
	fmt.Println("🔧 Configuring your digital knight...")
	fmt.Println()

	// Check if any flags were provided for non-interactive mode
	hasFlags := cmd.Flags().Changed("provider") ||
		cmd.Flags().Changed("api-key") ||
		cmd.Flags().Changed("model") ||
		cmd.Flags().Changed("max-tokens") ||
		cmd.Flags().Changed("temperature") ||
		cmd.Flags().Changed("mode")

	// Load existing config or create new one
	cfg, err := config.Load()
	if err != nil && !config.IsConfigNotFound(err) {
		return fmt.Errorf("failed to load existing configuration: %w", err)
	}
	if cfg == nil {
		cfg = config.New()
	}

	if hasFlags {
		// Non-interactive mode: update specific values from flags
		if cmd.Flags().Changed("provider") {
			provider, _ := cmd.Flags().GetString("provider")
			cfg.AIProvider = provider
		}

		if cmd.Flags().Changed("api-key") {
			apiKey, _ := cmd.Flags().GetString("api-key")
			cfg.APIKey = apiKey
		}

		if cmd.Flags().Changed("model") {
			model, _ := cmd.Flags().GetString("model")
			cfg.Model = model
		}

		if cmd.Flags().Changed("max-tokens") {
			maxTokens, _ := cmd.Flags().GetInt("max-tokens")
			cfg.MaxTokens = maxTokens
		}

		if cmd.Flags().Changed("temperature") {
			temperature, _ := cmd.Flags().GetFloat32("temperature")
			cfg.Temperature = temperature
		}

		if cmd.Flags().Changed("mode") {
			mode, _ := cmd.Flags().GetString("mode")
			cfg.Mode = mode
		}

		fmt.Println("📝 Updating configuration with provided values...")
	} else {
		// Interactive mode
		fmt.Println("⚙️  Interactive configuration mode")
		fmt.Println("📋 Press Enter to use default values shown in [brackets]")
		fmt.Println()

		if err := runInteractiveConfiguration(cfg); err != nil {
			return fmt.Errorf("interactive configuration failed: %w", err)
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Display final configuration
	fmt.Println()
	fmt.Println("✅ Configuration saved successfully!")
	fmt.Println()
	displayConfiguration(cfg)

	return nil
}

func runInteractiveConfiguration(cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	// Configure AI Provider
	fmt.Printf("🤖 AI Provider [%s]: ", cfg.AIProvider)
	if input := readInput(reader); input != "" {
		cfg.AIProvider = input
	}

	// Update model default based on provider
	if cfg.Model == "" || !isValidModelForProvider(cfg.Model, cfg.AIProvider) {
		cfg.Model = config.GetDefaultModel(cfg.AIProvider)
	}

	// Configure API Key (mandatory)
	for {
		fmt.Printf("🔑 API Key [%s]: ", maskAPIKey(cfg.APIKey))
		if input := readInput(reader); input != "" {
			cfg.APIKey = input
			break
		} else if cfg.APIKey != "" {
			// Keep existing API key
			break
		}
		fmt.Println("❌ API Key is required. Please provide a valid API key.")
	}

	// Configure Model
	fmt.Printf("🧠 Model [%s]: ", cfg.Model)
	if input := readInput(reader); input != "" {
		cfg.Model = input
	}

	// Configure Max Tokens
	fmt.Printf("📊 Max Tokens [%d]: ", cfg.MaxTokens)
	if input := readInput(reader); input != "" {
		if tokens, err := parseIntInput(input); err == nil {
			cfg.MaxTokens = tokens
		} else {
			fmt.Printf("⚠️  Invalid number format, using default: %d\n", cfg.MaxTokens)
		}
	}

	// Configure Temperature
	fmt.Printf("🌡️  Temperature [%.1f]: ", cfg.Temperature)
	if input := readInput(reader); input != "" {
		if temp, err := parseFloatInput(input); err == nil && temp >= 0.0 && temp <= 1.0 {
			cfg.Temperature = temp
		} else {
			fmt.Printf("⚠️  Invalid temperature (must be 0.0-1.0), using default: %.1f\n", cfg.Temperature)
		}
	}

	// Configure Mode (new)
	fmt.Println()
	fmt.Println("👑 Execution Mode Configuration:")
	fmt.Println("   🤴 monarch     - For experienced rulers who know their domain well")
	fmt.Println("                   Commands are shown without detailed explanations")
	fmt.Println("   👑 royal-heir  - For heirs still learning the ways of the realm")
	fmt.Println("                   Commands are shown with detailed explanations of each part")
	fmt.Println()

	for {
		currentMode := cfg.Mode
		if currentMode == "" {
			currentMode = "not set"
		}
		fmt.Printf("🎯 Choose your mode [%s]: ", currentMode)
		if input := readInput(reader); input != "" {
			input = strings.ToLower(strings.TrimSpace(input))
			if input == "monarch" || input == "royal-heir" {
				cfg.Mode = input
				break
			} else {
				fmt.Println("❌ Invalid mode. Please choose either 'monarch' or 'royal-heir'")
			}
		} else if cfg.Mode != "" {
			// Keep existing mode
			break
		} else {
			fmt.Println("❌ Mode is required. Please choose either 'monarch' or 'royal-heir'")
		}
	}

	return nil
}

func readInput(reader *bufio.Reader) string {
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(input)
}

func parseIntInput(input string) (int, error) {
	var result int
	_, err := fmt.Sscanf(input, "%d", &result)
	return result, err
}

func parseFloatInput(input string) (float32, error) {
	var result float32
	_, err := fmt.Sscanf(input, "%f", &result)
	return result, err
}

func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "not set"
	}
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}

func isValidModelForProvider(model, provider string) bool {
	// Simple validation - can be expanded
	switch provider {
	case "gemini":
		return strings.HasPrefix(model, "gemini")
	case "openai":
		return strings.HasPrefix(model, "gpt") || strings.HasPrefix(model, "text-")
	case "anthropic":
		return strings.HasPrefix(model, "claude")
	default:
		return true
	}
}

func displayConfiguration(cfg *config.Config) {
	fmt.Println("📋 Current Configuration:")
	fmt.Println("┌─────────────────────────────────────────┐")
	fmt.Printf("│ Provider:     %-25s │\n", cfg.AIProvider)
	fmt.Printf("│ API Key:      %-25s │\n", maskAPIKey(cfg.APIKey))
	fmt.Printf("│ Model:        %-25s │\n", cfg.Model)
	fmt.Printf("│ Max Tokens:   %-25d │\n", cfg.MaxTokens)
	fmt.Printf("│ Temperature:  %-25.1f │\n", cfg.Temperature)
	fmt.Printf("│ Mode:         %-25s │\n", cfg.Mode)
	fmt.Println("└─────────────────────────────────────────┘")
	fmt.Println()

	if cfg.Mode == "monarch" {
		fmt.Println("🤴 You have chosen the path of the experienced monarch!")
		fmt.Println("   Commands will be shown without detailed explanations.")
	} else {
		fmt.Println("👑 You have chosen the path of the learning heir!")
		fmt.Println("   Commands will be shown with detailed explanations to aid your learning.")
	}

	fmt.Println()
	fmt.Println("🎯 Your knight is now ready to serve!")
	fmt.Println("💡 Try: execute-my-will \"list my files\"")
}
