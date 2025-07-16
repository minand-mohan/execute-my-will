// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: internal/cli/configure.go
package ai

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/minand-mohan/execute-my-will/internal/config"
	"github.com/minand-mohan/execute-my-will/internal/system"
)

type Client interface {
	GenerateResponse(intent string, sysInfo *system.Info) (*AIResponse, error)
	ExplainCommand(command string, sysInfo *system.Info) (string, error)
	ListModels() ([]string, error)
}

type clientImpl struct {
	provider AIProvider
}

func NewClient(cfg *config.Config) (Client, error) {
	var provider AIProvider
	var err error

	switch cfg.AIProvider {
	case "gemini":
		provider, err = NewGeminiProvider(cfg)
	case "openai":
		provider, err = NewOpenAIProvider(cfg)
	case "anthropic":
		provider, err = NewAnthropicProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.AIProvider)
	}

	if err != nil {
		return nil, err
	}

	return &clientImpl{provider: provider}, nil
}

func (c *clientImpl) GenerateResponse(intent string, sysInfo *system.Info) (*AIResponse, error) {
	prompt := buildCommandPrompt(intent, sysInfo)
	response, err := exponentialRetryForAiResponse(c.provider.GenerateResponse, prompt, 5, 1*time.Second)
	if err != nil {
		return nil, err
	}
	return parseAIResponse(response), nil
}

func (c *clientImpl) ExplainCommand(command string, sysInfo *system.Info) (string, error) {
	prompt := buildExplanationPrompt(command, sysInfo)
	return exponentialRetryForAiResponse(c.provider.GenerateResponse, prompt, 3, 1*time.Second)
}

func (c *clientImpl) ListModels() ([]string, error) {
	return c.provider.ListModels()
}

func buildCommandPrompt(intent string, sysInfo *system.Info) string {
	primaryPackageManager := "the detected package manager"
	if len(sysInfo.PackageManagers) > 0 {
		primaryPackageManager = sysInfo.PackageManagers[0]
	}

	// Determine script format based on shell
	scriptFormat, commentPrefix := getScriptFormat(sysInfo.Shell)

	prompt := fmt.Sprintf(`You are a command line expert for %s systems. Generate a single, safe command or a safe script based on the user's intent.

SYSTEM INFORMATION:
- OS: %s
- Shell: %s
- Available Package Managers: %s
- Home Directory: %s
- Current Directory: %s
- Installed Packages: %s
- Available Commands: %s

USER INTENT: %s

RESPONSE FORMAT:
You must respond with exactly ONE of these three formats:

1. For simple single commands:
COMMAND: [single shell command with no formatting]

2. For complex multi-step tasks:
SCRIPT:
`+"```"+`%s
%s Brief description of what this command does
command1
%s Brief description of what this command does  
command2
`+"```"+`

3. For impossible/unsafe tasks:
FAILURE: [Brief reason why task cannot be completed]

REQUIREMENTS:
1. All commands and scripts must be SAFE and non-destructive.
2. First, check the "Installed Packages" and "Available Commands" lists to see if required applications are available.
3. If a required application is NOT available, include installation using the primary package manager '%s' (e.g., 'brew install htop', 'apt install htop', 'winget install htop').
4. For SCRIPT responses: Each command must have a brief one-line comment above it explaining what it does.
5. For SCRIPT responses: Use %s syntax for comments and ensure commands work in %s shell.
6. For SCRIPT responses: Use proper %s syntax and ensure commands can run in sequence in the same shell session.
7. Use safe and non-destructive flags where possible (e.g., 'cp -i' for interactive copy, 'rm -i' for interactive removal).
8. If any directory reference is vague (e.g., "some folder"), respond with FAILURE: Directory reference too vague.
9. Choose SCRIPT over COMMAND when the task requires multiple steps, environment setup, or variable usage.

RESPONSE:`,
		sysInfo.OS,                           // systems
		sysInfo.OS,                           // OS
		sysInfo.Shell,                        // Shell
		joinSlice(sysInfo.PackageManagers),   // Available Package Managers
		sysInfo.HomeDir,                      // Home Directory
		sysInfo.CurrentDir,                   // Current Directory
		joinSlice(sysInfo.InstalledPackages), // Installed Packages
		joinSlice(sysInfo.AvailableCommands), // Available Commands
		intent,                               // USER INTENT
		scriptFormat,                         // script format (```bash)
		commentPrefix,                        // comment prefix (first comment)
		commentPrefix,                        // comment prefix (second comment)
		primaryPackageManager,                // primary package manager
		commentPrefix,                        // comment syntax
		sysInfo.Shell,                        // shell name
		scriptFormat,                         // script format (proper bash syntax)
	)

	return prompt
}

func getScriptFormat(shell string) (scriptFormat, commentPrefix string) {
	switch shell {
	case "powershell", "pwsh":
		return "powershell", "#"
	case "cmd":
		return "cmd", "REM"
	case "bash", "zsh", "fish", "sh":
		return "bash", "#"
	default:
		// Default to bash for unknown shells
		return "bash", "#"
	}
}

func buildExplanationPrompt(command string, sysInfo *system.Info) string {
	prompt := fmt.Sprintf(`You are an expert explaining command-line instructions to someone new to the terminal.

SYSTEM INFO:
- OS: %s
- Shell: %s
- Current Dir: %s
- Home Dir: %s

COMMAND: %s

INSTRUCTIONS:
Explain what this command does in one clear, simple paragraph. Break down the parts in plain English, avoiding technical jargon where possible. Focus on what the command does, what each part means, and why someone might use it. Be friendly, helpful, and avoid assuming any prior knowledge of the shell.

EXPLANATION:`,
		sysInfo.OS,
		sysInfo.Shell,
		sysInfo.CurrentDir,
		sysInfo.HomeDir,
		command,
	)

	return prompt
}

func joinSlice(slice []string) string {
	if len(slice) == 0 {
		return "none"
	}
	// Limit to prevent overly long prompts
	const limit = 100
	if len(slice) > limit {
		return strings.Join(slice[:limit], ", ") + "..."
	}
	return strings.Join(slice, ", ")
}

func parseAIResponse(response string) *AIResponse {
	response = strings.TrimSpace(response)

	if strings.HasPrefix(response, "COMMAND:") {
		content := strings.TrimSpace(strings.TrimPrefix(response, "COMMAND:"))
		return &AIResponse{
			Type:    ResponseTypeCommand,
			Content: content,
		}
	}

	if strings.HasPrefix(response, "SCRIPT:") {
		scriptContent := strings.TrimSpace(strings.TrimPrefix(response, "SCRIPT:"))

		// Extract content from markdown code block - support multiple script types
		re := regexp.MustCompile("(?s)```(?:bash|sh|cmd|bat|powershell|ps1)?\n(.*?)```")
		matches := re.FindStringSubmatch(scriptContent)
		if len(matches) > 1 {
			scriptContent = strings.TrimSpace(matches[1])
		}
		return &AIResponse{
			Type:    ResponseTypeScript,
			Content: scriptContent,
		}
	}

	if strings.HasPrefix(response, "FAILURE:") {
		errorMsg := strings.TrimSpace(strings.TrimPrefix(response, "FAILURE:"))
		return &AIResponse{
			Type:  ResponseTypeFailure,
			Error: errorMsg,
		}
	}

	// Default fallback - treat as command for backward compatibility
	return &AIResponse{
		Type:    ResponseTypeCommand,
		Content: response,
	}
}

func exponentialRetryForAiResponse(fn func(string) (string, error), prompt string, maxRetries int, delay time.Duration) (string, error) {
	var resp string
	var err error

	for i := 0; i < maxRetries; i++ {
		resp, err = fn(prompt)
		if err == nil {
			return resp, nil
		}
		fmt.Println("🌀" + " " + "The oracles have rejected us, sire. I will try again...")
		time.Sleep(delay)
		delay *= 2
		if delay > 10*time.Second {
			delay = 10 * time.Second
		}

	}

	return "", fmt.Errorf("failed to get response after %d attempts: %v", maxRetries, err)

}
