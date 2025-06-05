// File: internal/cli/configure.go
package ai

import (
	"fmt"

	"github.com/minand-mohan/execute-my-will/internal/config"
	"github.com/minand-mohan/execute-my-will/internal/system"
)

type Client interface {
	GenerateCommand(intent string, sysInfo *system.Info) (string, error)
	ExplainCommand(command string, sysInfo *system.Info) (string, error) // New method
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
	return c.provider.GenerateResponse(prompt)
}

func (c *clientImpl) ExplainCommand(command string, sysInfo *system.Info) (string, error) {
	prompt := buildExplanationPrompt(command, sysInfo)
	return c.provider.GenerateResponse(prompt)
}

func buildCommandPrompt(intent string, sysInfo *system.Info) string {
	prompt := fmt.Sprintf(`You are a command line expert for %s systems. Generate a single, safe command based on the user's intent.

SYSTEM INFORMATION:
- OS: %s
- Shell: %s
- Package Manager: %s
- Available Commands: %s
- PATH directories: %s
- Current Directory: %s
- Home Directory: %s

USER ALIASES:
%s

USER INTENT: %s

REQUIREMENTS:
1. Output must be a SINGLE shell command with NO formatting or enclosure — no backticks, no quotes, no markdown.
2. The command must be ONE LINE ONLY and ready to paste directly into a terminal.
3. Do NOT return scripts or use line continuations ('\') or multiple commands on separate lines.
4. If the task is too complex to express as one command, respond only with: FAILURE: Intent too complex for a single shell command.
5. If the intent involves directories (e.g., copy, move, list, extract), all directory paths must be absolute and clearly specified.
6. If any directory reference is vague (e.g., “some folder”, “the project directory”), respond only with: FAILURE: Directory reference too vague.
7. Use known aliases if applicable.
8. If the command is not available, prepend it with an install command using the detected package manager.
9. Use safe and non-destructive flags when possible.
10. Use proper syntax for the detected shell.
11. Return only the command — no comments, no explanations, no headers.

COMMAND:`,
		sysInfo.OS,
		sysInfo.OS,
		sysInfo.Shell,
		sysInfo.PackageManager,
		joinSlice(sysInfo.AvailableCommands),
		joinSlice(sysInfo.PathDirectories),
		sysInfo.CurrentDir,
		sysInfo.HomeDir,
		formatAliases(sysInfo.Aliases),
		intent,
	)

	return prompt
}

func buildExplanationPrompt(command string, sysInfo *system.Info) string {
	prompt := fmt.Sprintf(`You are a patient teacher explaining command line operations to a learning student. Provide a clear, educational explanation of the given command.

SYSTEM INFORMATION:
- OS: %s
- Shell: %s
- Current Directory: %s
- Home Directory: %s

COMMAND TO EXPLAIN: %s

REQUIREMENTS:
1. Break down each part of the command and explain what it does
2. Use clear, beginner-friendly language
3. Explain any flags, options, or arguments used
4. Mention any important safety considerations
5. If the command involves multiple parts (pipes, &&, etc.), explain the flow
6. Keep the explanation concise but thorough
7. Use a teaching tone that is encouraging and informative

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
	result := ""
	for i, item := range slice {
		if i > 0 {
			result += ", "
		}
		result += item
		if i >= 20 { // Limit to prevent overly long prompts
			result += "..."
			break
		}
	}
	return result
}

func formatAliases(aliases map[string]string) string {
	if len(aliases) == 0 {
		return "No aliases found"
	}

	result := ""
	count := 0
	for alias, command := range aliases {
		if count > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("%s='%s'", alias, command)
		count++
		if count >= 10 { // Limit aliases to prevent long prompts
			result += "\n... (and more)"
			break
		}
	}
	return result
}
