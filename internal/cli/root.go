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
)

var rootCmd = &cobra.Command{
	Use:   "execute-my-will [intent]",
	Short: "Your faithful digital knight, ready to execute your commands",
	Long:  "A CLI application that interprets your natural language intent and executes the appropriate system commands with your permission, my lord.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  executeWill,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add configure subcommand
	rootCmd.AddCommand(configureCmd)
}

func executeWill(cmd *cobra.Command, args []string) error {
	// Check if config file exists, if not prompt user to configure
	cfg, err := config.Load()
	if err != nil {
		if config.IsConfigNotFound(err) {
			fmt.Println("🔧 Configuration file not found, my lord!")
			fmt.Println("📋 Please run 'execute-my-will configure' to set up your configuration first.")
			fmt.Println()
			fmt.Println("Example:")
			fmt.Println("  execute-my-will configure")
			fmt.Println("  # or set specific values:")
			fmt.Println("  execute-my-will configure --api-key your-key --provider gemini")
			return nil
		}
		return fmt.Errorf("failed to load configuration: %w", err)
	}

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
