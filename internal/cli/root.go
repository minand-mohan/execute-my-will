// File: internal/cli/root.go
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/minand-mohan/execute-my-will/internal/ai"
	"github.com/minand-mohan/execute-my-will/internal/config"
	"github.com/minand-mohan/execute-my-will/internal/system"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "execute-my-will [intent]",
	Short: "Your faithful digital knight, ready to execute your commands",
	Long:  "A CLI application that interprets your natural language intent and executes the appropriate system commands with your permission, my lord.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  executeWill,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.execute-my-will.yaml)")
	rootCmd.PersistentFlags().String("provider", "gemini", "AI provider (gemini, openai, anthropic)")
	rootCmd.PersistentFlags().String("api-key", "", "API key for the AI provider")
	rootCmd.PersistentFlags().String("model", "", "Model to use (default varies by provider)")
	rootCmd.PersistentFlags().Int("max-tokens", 1000, "Maximum tokens for AI response")
	rootCmd.PersistentFlags().Float32("temperature", 0.1, "Temperature for AI response")

	// Bind flags to viper
	viper.BindPFlag("ai.provider", rootCmd.PersistentFlags().Lookup("provider"))
	viper.BindPFlag("ai.api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("ai.model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("ai.max_tokens", rootCmd.PersistentFlags().Lookup("max-tokens"))
	viper.BindPFlag("ai.temperature", rootCmd.PersistentFlags().Lookup("temperature"))
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".execute-my-will" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".execute-my-will")
	}

	// Environment variables
	viper.SetEnvPrefix("EXECUTE_MY_WILL")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("ai.provider", "gemini")
	viper.SetDefault("ai.max_tokens", 1000)
	viper.SetDefault("ai.temperature", 0.1)
	viper.SetDefault("ai.model", getDefaultModel(viper.GetString("ai.provider")))

	// Read config file
	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("🔧 Using config file: %s\n", viper.ConfigFileUsed())
	}
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

func executeWill(cmd *cobra.Command, args []string) error {
	// Load configuration from viper
	cfg := config.FromViper()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error, sire: %w", err)
	}

	// Join all arguments as the user's intent
	intent := strings.Join(args, " ")

	fmt.Printf("🛡️  Your faithful knight has received your command: \"%s\"\n", intent)
	fmt.Println("🔍 Analyzing your noble request...")

	// Initialize system analyzer
	analyzer := system.NewAnalyzer()

	// Perform system analysis
	sysInfo, err := analyzer.AnalyzeSystem()
	if err != nil {
		return fmt.Errorf("failed to analyze the realm's systems, my lord: %w", err)
	}

	// Validate the intent
	validator := system.NewValidator(sysInfo)
	if err := validator.ValidateIntent(intent); err != nil {
		fmt.Printf("⚠️  Forgive me sire, but your request needs clarification: %s\n", err.Error())
		return nil
	}

	// Initialize AI client
	aiClient, err := ai.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to summon the oracle, my lord: %w", err)
	}

	// Generate command
	fmt.Println("🧙 Consulting with the ancient oracles...")
	command, err := aiClient.GenerateCommand(intent, sysInfo)
	if err != nil {
		return fmt.Errorf("the oracles have failed us, sire: %w", err)
	}

	// Display the command for confirmation
	fmt.Printf("\n⚔️  I propose to execute this command on your behalf:\n")
	fmt.Printf("   %s\n\n", command)

	// Ask for confirmation
	fmt.Print("🤴 Do you wish me to proceed with this quest? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read your royal decree: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("🙏 I misunderstood your will, sire. Please try again with clearer instructions.")
		return nil
	}

	// Execute the command
	fmt.Println("⚡ Executing your command with honor...")
	executor := system.NewExecutor()
	if err := executor.Execute(command); err != nil {
		return fmt.Errorf("the quest has encountered difficulties, my lord: %w", err)
	}

	fmt.Println("✅ Your command has been executed successfully, sire!")
	return nil
}
