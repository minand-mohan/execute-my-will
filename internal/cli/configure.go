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

	"github.com/minand-mohan/execute-my-will/internal/ai"
	"github.com/minand-mohan/execute-my-will/internal/config"
	"github.com/minand-mohan/execute-my-will/internal/ui"
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
	ui.PrintKnightMessage("Configuring your digital knight...")
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

		ui.PrintInfoMessage("Updating configuration with provided values...")
	} else {
		// Interactive mode
		ui.PrintInfoMessage("Interactive configuration mode")
		ui.PrintInfoMessage("Press Enter to use default values shown in [brackets]")
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
	ui.PrintSuccessMessage("Configuration saved successfully!")
	fmt.Println()
	displayConfiguration(cfg)

	return nil
}

func runInteractiveConfiguration(cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	providers := map[string]string{
		"1": "gemini",
		"2": "openai",
		"3": "anthropic",
	}

	// List AI  Providers
	ui.PrintInfoMessage("AI Providers:")
	fmt.Println(ui.Cyan.Sprint("1. Gemini"))
	fmt.Println(ui.Cyan.Sprint("2. OpenAI"))
	fmt.Println(ui.Cyan.Sprint("3. Anthropic"))
	fmt.Print(ui.Gold.Sprint("Enter the number of the provider you want to use: "))

	if input := readInput(reader); input != "" {
		cfg.AIProvider = providers[input]
	}

	// Update model default based on provider
	if cfg.Model == "" || !isValidModelForProvider(cfg.Model, cfg.AIProvider) {
		cfg.Model = config.GetDefaultModel(cfg.AIProvider)
	}

	// Configure API Key (mandatory)
	for {
		fmt.Printf("%s API Key [%s]: ", ui.Gold.Sprint("🔑"), ui.Gray.Sprint(maskAPIKey(cfg.APIKey)))
		if input := readInput(reader); input != "" {
			cfg.APIKey = input
			break
		} else if cfg.APIKey != "" {
			// Keep existing API key
			break
		}
		ui.PrintErrorMessage("API Key is required. Please provide a valid API key.")
	}

	// Get Models for provider
	aiClient, err := ai.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client")
	}
	models, err := aiClient.ListModels()
	if err != nil {
		return fmt.Errorf("failed to get models: %w", err)
	}
	ui.PrintInfoMessage("Available Models:")
	for _, model := range models {
		fmt.Printf("  - %s\n", ui.Cyan.Sprint(model))
	}
	// Configure Model
	fmt.Printf("%s Select Model [%s]: ", ui.Gold.Sprint("🧠"), ui.Gray.Sprint(cfg.Model))
	if input := readInput(reader); input != "" {
		cfg.Model = input
	}

	// Configure Max Tokens
	fmt.Printf("%s Max Tokens [%s]: ", ui.Gold.Sprint("📊"), ui.Gray.Sprint(fmt.Sprintf("%d", cfg.MaxTokens)))
	if input := readInput(reader); input != "" {
		if tokens, err := parseIntInput(input); err == nil {
			cfg.MaxTokens = tokens
		} else {
			ui.PrintWarningMessage(fmt.Sprintf("Invalid number format, using default: %d", cfg.MaxTokens))
		}
	}

	// Configure Temperature
	fmt.Printf("%s Temperature [%s]: ", ui.Gold.Sprint("🌡️"), ui.Gray.Sprint(fmt.Sprintf("%.1f", cfg.Temperature)))
	if input := readInput(reader); input != "" {
		if temp, err := parseFloatInput(input); err == nil && temp >= 0.0 && temp <= 1.0 {
			cfg.Temperature = temp
		} else {
			ui.PrintWarningMessage(fmt.Sprintf("Invalid temperature (must be 0.0-1.0), using default: %.1f", cfg.Temperature))
		}
	}

	// Configure Mode
	fmt.Println()
	ui.PrintInfoMessage("Execution Mode Configuration:")
	fmt.Printf(" %s   %s - For experienced rulers who know their domain well\n", ui.Gold.Sprint("1."), ui.Gold.Sprint("🤴 monarch"))
	fmt.Printf("                   %s\n", ui.Gray.Sprint("Commands are shown without detailed explanations"))
	fmt.Printf(" %s   %s - For heirs still learning the ways of the realm\n", ui.Gold.Sprint("2."), ui.Gold.Sprint("👑 royal-heir"))
	fmt.Printf("                   %s\n", ui.Gray.Sprint("Commands are shown with detailed explanations of each part"))
	fmt.Println()

	modeMap := map[string]string{
		"1": "monarch",
		"2": "royal-heir",
	}

	for {
		currentMode := cfg.Mode
		if currentMode == "" {
			currentMode = "not set"
		}
		fmt.Printf("%s Choose the number of the mode you want to use [%s]: ", ui.Gold.Sprint("🎯"), ui.Gray.Sprint(currentMode))
		if input := readInput(reader); input != "" {
			if mode, ok := modeMap[input]; ok {
				cfg.Mode = mode
				break
			} else {
				ui.PrintErrorMessage("Invalid mode. Please enter either '1'(monarch) or '2'(royal-heir)")
			}
		} else if cfg.Mode != "" {
			// Keep existing mode
			break
		} else {
			ui.PrintErrorMessage("Mode is required. Please enter either '1'(monarch) or '2'(royal-heir)")
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
	return apiKey[:4] + strings.Repeat("*", 6)
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
	// Create config map for structured display
	configs := map[string]string{
		"Provider":    ui.Cyan.Sprint(cfg.AIProvider),
		"API Key":     ui.Gray.Sprint(maskAPIKey(cfg.APIKey)),
		"Model":       ui.Cyan.Sprint(cfg.Model),
		"Max Tokens":  ui.Blue.Sprint(fmt.Sprintf("%d", cfg.MaxTokens)),
		"Temperature": ui.Blue.Sprint(fmt.Sprintf("%.1f", cfg.Temperature)),
		"Mode":        ui.Purple.Sprint(cfg.Mode),
	}
	
	ui.PrintConfigBox(configs)

	// Mode-specific message
	var modeMsg string
	if cfg.Mode == "monarch" {
		modeMsg = "You have chosen the path of the experienced monarch!\nCommands will be shown without detailed explanations."
	} else {
		modeMsg = "You have chosen the path of the learning heir!\nCommands will be shown with detailed explanations to aid your learning."
	}
	
	ui.PrintStatusBox("CONFIGURATION COMPLETE", modeMsg, "success")
	
	// Final message
	finalMsg := "Your knight is now ready to serve!\n\n💡 Try: " + ui.CommandText("execute-my-will \"list my files\"")
	ui.PrintStatusBox("READY TO SERVE", finalMsg, "info")
}
