// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: internal/cli/configure.go
package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/minand-mohan/execute-my-will/internal/config"
	"github.com/minand-mohan/execute-my-will/internal/system"
)

type Client interface {
	GenerateCommand(intent string, sysInfo *system.Info) (string, error)
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

func (c *clientImpl) GenerateCommand(intent string, sysInfo *system.Info) (string, error) {
	prompt := buildCommandPrompt(intent, sysInfo)
	return exponentialRetryForAiRespone(c.provider.GenerateResponse, prompt, 5, 1*time.Second)
}

func (c *clientImpl) ExplainCommand(command string, sysInfo *system.Info) (string, error) {
	prompt := buildExplanationPrompt(command, sysInfo)
	return exponentialRetryForAiRespone(c.provider.GenerateResponse, prompt, 3, 1*time.Second)
}

func (c *clientImpl) ListModels() ([]string, error) {
	return c.provider.ListModels()
}

func buildCommandPrompt(intent string, sysInfo *system.Info) string {
	primaryPackageManager := "the detected package manager"
	if len(sysInfo.PackageManagers) > 0 {
		primaryPackageManager = sysInfo.PackageManagers[0]
	}

	prompt := fmt.Sprintf(`You are a command line expert for %s systems. Generate a single, safe command based on the user's intent.

SYSTEM INFORMATION:
- OS: %s
- Shell: %s
- Available Package Managers: %s
- Home Directory: %s
- Current Directory: %s
- Installed Packages: %s
- Available Commands: %s

USER INTENT: %s

REQUIREMENTS:
1. Output must be a SINGLE shell command with NO formatting or enclosure — no backticks, no quotes, no markdown.
2. The command must be ONE LINE ONLY and ready to paste directly into a terminal.
3. First, check the "Installed Packages" and "Available Commands" lists to see if the required application is already available.
4. If a required application is NOT available, prepend the proper installation command. Use the primary package manager '%s' for the installation (e.g., 'brew install htop && htop').
5. If the application IS already installed, generate the command directly without an installer.
6. If the task is too complex, respond only with: FAILURE: Intent too complex for a single shell command.
7. If any directory reference is vague (e.g., “some folder”), respond only with: FAILURE: Directory reference too vague.
8. Use safe and non-destructive flags where possible.
9. Return only the command — no comments, no explanations, no headers.

COMMAND:`,
		sysInfo.OS,
		sysInfo.OS,
		sysInfo.Shell,
		joinSlice(sysInfo.PackageManagers),
		sysInfo.HomeDir,
		sysInfo.CurrentDir,
		joinSlice(sysInfo.InstalledPackages),
		joinSlice(sysInfo.AvailableCommands),
		intent,
		primaryPackageManager,
	)

	return prompt
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

func exponentialRetryForAiRespone(fn func(string) (string, error), prompt string, maxRetries int, delay time.Duration) (string, error) {
	var resp string
	var err error

	for range maxRetries {
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
