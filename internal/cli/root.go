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
		fmt.Println("🤔 Please provide an intent, my lord!")
		fmt.Println("Example:")
		fmt.Println("  execute-my-will 'create a new file named 'my-file.txt' in the current directory'")
		return nil
	}

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
			fmt.Println("  execute-my-will configure --api-key your-key --provider gemini --mode monarch")
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

	// Generate response (command or script)
	fmt.Println("🧙 Consulting with the ancient oracles...")
	response, err := aiClient.GenerateResponse(intent, sysInfo)
	if err != nil {
		return fmt.Errorf("the oracles have failed us, sire: %w", err)
	}

	var taskContent string
	var isScript bool

	// Handle different response types
	switch response.Type {
	case ai.ResponseTypeFailure:
		fmt.Printf("\n❌ Alas, I cannot fulfill this quest: %s\n", response.Error)
		return nil
		
	case ai.ResponseTypeCommand:
		// Display the command for confirmation
		fmt.Printf("\n⚔️  I propose to execute this command on your behalf:\n")
		fmt.Printf("   %s\n", response.Content)
		taskContent = response.Content
		isScript = false

		// If in royal-heir mode, provide detailed explanation for commands only
		if cfg.Mode == "royal-heir" {
			fmt.Println("================================================")
			fmt.Println("")
			fmt.Println("📚 As you are still learning the ways of the realm, allow me to explain:")
			explanation, err := aiClient.ExplainCommand(response.Content, sysInfo)
			if err != nil {
				fmt.Printf("⚠️  I encountered difficulty explaining the command, but it should still work, my lord: %v\n\n", err)
			} else {
				fmt.Printf("%s\n", explanation)
			}
			fmt.Println("================================================")
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
		fmt.Printf("\n📜 I propose to execute this script on your behalf:\n")
		fmt.Println("================================================")
		
		// Display script with or without comments based on mode
		showComments := cfg.Mode == "royal-heir"
		displayScript(response.Content, showComments)
		
		fmt.Println("================================================")
		taskContent = response.Content
		isScript = true
		
		if cfg.Mode == "royal-heir" {
			fmt.Println("📚 This script will execute each command in sequence, maintaining context between steps.")
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
		fmt.Println("🙏 I misunderstood your will, sire. Please try again with clearer instructions.")
		return nil
	}

	// Execute the task with enhanced interactive support
	fmt.Println("⚡ Executing your quest with honor...")
	fmt.Println("") // Add some space before execution

	executor := system.NewExecutor()
	var execErr error
	
	if isScript {
		showComments := cfg.Mode == "royal-heir"
		execErr = executor.ExecuteScript(taskContent, sysInfo.Shell, showComments)
	} else {
		execErr = executor.Execute(taskContent, sysInfo.Shell)
	}
	
	if execErr != nil {
		fmt.Printf("\n⚔️  Alas! The quest has encountered difficulties, my lord: %v\n", execErr)

		// Check if it's a common issue and provide helpful suggestions
		if strings.Contains(execErr.Error(), "permission denied") {
			fmt.Println("💡 This might require elevated privileges. Consider adding 'sudo' to your request if appropriate.")
		} else if strings.Contains(execErr.Error(), "command not found") {
			fmt.Println("💡 The command appears to be missing. The system may need to install required packages first.")
		} else if strings.Contains(execErr.Error(), "no such file or directory") {
			fmt.Println("💡 Please ensure all file paths in your request are correct and accessible.")
		}

		return nil // Don't return the error to avoid double error messages
	}

	if isScript {
		fmt.Printf("\n🏆 Your script has been executed successfully, sire!\n")
	} else {
		fmt.Printf("\n🏆 Your command has been executed successfully, sire!\n")
	}
	return nil
}

// displayScript shows the script content with or without comments
func displayScript(scriptContent string, showComments bool) {
	lines := strings.Split(scriptContent, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Check if line is a comment
		isComment := strings.HasPrefix(line, "#") || strings.HasPrefix(line, "REM")
		
		if isComment && showComments {
			// Display comment in a different style
			fmt.Printf("   💬 %s\n", strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "#"), "REM")))
		} else if !isComment {
			// Display command
			fmt.Printf("   %s\n", line)
		}
	}
}
