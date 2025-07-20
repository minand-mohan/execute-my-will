// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

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
	"github.com/minand-mohan/execute-my-will/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Build info - will be set by SetBuildInfo function
	appVersion   string
	appCommit    string
	appBuildTime string
	versionFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "execute-my-will [intent]",
	Short: "Your faithful digital knight, ready to execute your commands",
	Long:  "A CLI application that interprets your natural language intent and executes the appropriate system commands with your permission, my lord.",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  executeWill,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// SetVersion sets the application version.
func SetBuildInfo(version, commit, buildTime string) {
	appVersion = version
	appCommit = commit
	appBuildTime = buildTime
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display application version")

	// Add configure subcommand
	rootCmd.AddCommand(configureCmd)

	// Add mode flag
	rootCmd.Flags().String("mode", "", "Execution mode: monarch (no explanations) or royal-heir (detailed explanations)")
}

func executeWill(cmd *cobra.Command, args []string) error {
	if versionFlag {
		fmt.Print("execute-my-will\n")
		fmt.Printf("Version: %s\n", appVersion)
		if appCommit != "" && appCommit != "unknown" {
			fmt.Printf("Commit: %s\n", appCommit)
		}
		if appBuildTime != "" && appBuildTime != "unknown" {
			fmt.Printf("Build Time: %s\n", appBuildTime)
		}
		return nil
	}

	// Check if there are any arguments
	if len(args) == 0 {
		ui.PrintStatusBox("QUEST REQUIRED", "Please provide an intent, my lord!\n\nExample:\n  execute-my-will 'create a new file named my-file.txt in the current directory'", "info")
		return nil
	}

	// Check if config file exists, if not prompt user to configure
	cfg, err := config.Load()
	if err != nil {
		if config.IsConfigNotFound(err) {
			ui.PrintStatusBox("🔧 CONFIGURATION REQUIRED", "Configuration file not found, my lord!\n\n📋 Please run 'execute-my-will configure' to set up your configuration first.\n\nExample:\n  execute-my-will configure\n  # or set specific values:\n  execute-my-will configure --api-key your-key --provider gemini --mode monarch", "warning")
			return nil
		}
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override mode from flag if provided
	if cmd.Flags().Changed("mode") {
		mode, _ := cmd.Flags().GetString("mode")
		cfg.Mode = mode
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error, sire: %w", err)
	}

	// Join all arguments as the user's intent
	intent := strings.Join(args, " ")

	ui.PrintKnightMessage(fmt.Sprintf("Your faithful knight has received your command: \"%s\"", intent))
	ui.PrintInfoMessage("Analyzing your noble request...")

	ui.PrintPhaseHeader("🧙", "Consulting with the ancient oracles...")

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
		ui.PrintStatusBox("⚠️  REQUEST CLARIFICATION NEEDED", fmt.Sprintf("Forgive me sire, but your request needs clarification: %s", err.Error()), "warning")
		return nil
	}

	// Initialize AI client
	aiClient, err := ai.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to summon the oracle, my lord: %w", err)
	}

	// Generate response (command or script)
	response, err := aiClient.GenerateResponse(intent, sysInfo)
	if err != nil {
		return fmt.Errorf("the oracles have failed us, sire: %w", err)
	}

	var taskContent string
	var isScript bool

	// Handle different response types
	switch response.Type {
	case ai.ResponseTypeFailure:
		ui.PrintStatusBox("❌ QUEST CANNOT BE COMPLETED", fmt.Sprintf("Alas, I cannot fulfill this quest: %s", response.Error), "error")
		return nil

	case ai.ResponseTypeCommand:
		// Display the command for confirmation
		ui.PrintCommandBox(response.Content)
		taskContent = response.Content
		isScript = false

		// If in royal-heir mode, provide detailed explanation for commands only
		if cfg.Mode == "royal-heir" {
			explanation, err := aiClient.ExplainCommand(response.Content, sysInfo)
			if err != nil {
				ui.PrintStatusBox("⚠️  EXPLANATION DIFFICULTY", fmt.Sprintf("I encountered difficulty explaining the command, but it should still work, my lord: %v", err), "warning")
			} else {
				ui.PrintStatusBox("📚 COMMAND EXPLANATION", fmt.Sprintf("As you are still learning the ways of the realm, allow me to explain:\n\n%s", explanation), "info")
			}
		}

		// Validate if the command affects the environment
		envValidator := system.NewEnvironmentValidator(sysInfo)
		if err := envValidator.ValidateEnvironmentCommand(response.Content); err != nil {
			if envErr, ok := err.(*system.EnvironmentCommandError); ok {
				fmt.Println()
				fmt.Println(envErr.GetKnightlyMessage())
				return nil
			}
			return fmt.Errorf("environment validation failed: %w", err)
		}

	case ai.ResponseTypeScript:
		// Display the script for confirmation  
		showComments := cfg.Mode == "royal-heir"
		scriptLines := strings.Split(response.Content, "\n")
		
		// Filter and format script lines based on mode
		var displayLines []string
		displayLines = append(displayLines, "") // Empty line at start
		
		for _, line := range scriptLines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			
			// Check if line is a comment
			isComment := strings.HasPrefix(line, "#") || strings.HasPrefix(line, "REM")
			
			if isComment && showComments {
				// Display comment with proper formatting
				comment := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "#"), "REM"))
				displayLines = append(displayLines, ui.CommentText("• "+comment))
			} else if !isComment {
				// Display command with arrow prefix
				displayLines = append(displayLines, ui.CommandText("→ "+line))
			}
		}
		displayLines = append(displayLines, "") // Empty line at end
		
		template := ui.DefaultTemplate()
		template.PrintBox("📜 PROPOSED SCRIPT", displayLines)
		taskContent = response.Content
		isScript = true

		if cfg.Mode == "royal-heir" {
			ui.PrintStatusBox("📚 SCRIPT INFORMATION", "This script will execute each command in sequence, maintaining context between steps.", "info")
		}
	}

	// Ask for confirmation
	if cfg.Mode == "monarch" {
		fmt.Print("🤴 Do you wish me to proceed with this quest? (y/N): ")
	} else {
		fmt.Print("👑 Do you wish me to proceed with this quest, young heir? (y/N): ")
	}

	reader := bufio.NewReader(os.Stdin)
	userResponse, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read your royal decree: %w", err)
	}

	userResponse = strings.TrimSpace(strings.ToLower(userResponse))
	if userResponse != "y" && userResponse != "yes" {
		ui.PrintStatusBox("🙏 QUEST DECLINED", "I understand, sire. Please try again when you're ready.", "info")
		return nil
	}

	// Execute the task with enhanced interactive support
	fmt.Println("🛡️  Executing your quest with honor...")
	fmt.Println()

	executor := system.NewExecutor()
	var execErr error

	if isScript {
		showComments := cfg.Mode == "royal-heir"
		execErr = executor.ExecuteScript(taskContent, sysInfo.Shell, showComments)
	} else {
		execErr = executor.Execute(taskContent, sysInfo.Shell)
	}

	if execErr != nil {
		var suggestionMsg string
		
		// Check if it's a common issue and provide helpful suggestions
		if strings.Contains(execErr.Error(), "permission denied") {
			suggestionMsg = "\n\n💡 This might require elevated privileges. Consider adding 'sudo' to your request if appropriate."
		} else if strings.Contains(execErr.Error(), "command not found") {
			suggestionMsg = "\n\n💡 The command appears to be missing. The system may need to install required packages first."
		} else if strings.Contains(execErr.Error(), "no such file or directory") {
			suggestionMsg = "\n\n💡 Please ensure all file paths in your request are correct and accessible."
		}
		
		ui.PrintStatusBox("⚔️  QUEST DIFFICULTIES", fmt.Sprintf("Alas! The quest has encountered difficulties, my lord: %v%s", execErr, suggestionMsg), "error")
		return nil // Don't return the error to avoid double error messages
	}

	if isScript {
		ui.PrintStatusBox("🏆 QUEST COMPLETED", "Your script has been executed successfully, sire!", "success")
	} else {
		ui.PrintStatusBox("🏆 QUEST COMPLETED", "Your command has been executed successfully, sire!", "success")
	}
	return nil
}

