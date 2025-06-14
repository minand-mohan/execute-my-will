// File: internal/system/env_validator.go
package system

import (
	"fmt"
	"regexp"
	"strings"
)

type EnvironmentValidator struct {
	sysInfo *Info
}

func NewEnvironmentValidator(sysInfo *Info) *EnvironmentValidator {
	return &EnvironmentValidator{sysInfo: sysInfo}
}

// ValidateEnvironmentCommand checks if a command would affect the parent shell environment
// and returns an error with the command if it would be ineffective in a subshell
func (ev *EnvironmentValidator) ValidateEnvironmentCommand(command string) error {
	// Clean the command for analysis
	cleanCmd := strings.TrimSpace(command)
	if cleanCmd == "" {
		return nil
	}

	// Check for environment-affecting patterns
	if envCmd := ev.detectEnvironmentCommand(cleanCmd); envCmd != "" {
		return &EnvironmentCommandError{
			Command:     command,
			Reason:      envCmd,
			Explanation: "this application cannot modify your terminal session",
		}
	}

	return nil
}

// detectEnvironmentCommand analyzes the command and returns the type of environment command detected
func (ev *EnvironmentValidator) detectEnvironmentCommand(command string) string {
	// Convert to lowercase for case-insensitive matching
	lowerCmd := strings.ToLower(command)

	// Remove leading sudo, && chains, and pipes for core command analysis
	coreCmd := ev.extractCoreCommand(lowerCmd)

	// Check for different types of environment-affecting commands
	checks := []struct {
		name     string
		detector func(string, string) bool
	}{
		{"path_modification", ev.detectPathModification}, // Check path_modification before export
		{"conda_env", ev.detectCondaEnvironment},         // Check conda before virtual_env
		{"source", ev.detectSourceCommand},
		{"export", ev.detectExportCommand},
		{"alias", ev.detectAliasCommand},
		{"cd", ev.detectCdCommand},
		{"virtual_env", ev.detectVirtualEnvCommand},
		{"shell_function", ev.detectShellFunctionCommand},
		{"environment_module", ev.detectEnvironmentModuleCommand},
		{"shell_options", ev.detectShellOptions},
		{"docker_env", ev.detectDockerEnvironment},
		{"rbenv_pyenv", ev.detectVersionManagers},
	}

	for _, check := range checks {
		if check.detector(coreCmd, command) {
			return check.name
		}
	}

	return ""
}

// extractCoreCommand removes common prefixes and extracts the main command
func (ev *EnvironmentValidator) extractCoreCommand(command string) string {
	// Remove sudo and common prefixes
	prefixes := []string{"sudo ", "sudo -E ", "sudo -i ", "nohup "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(command, prefix) {
			command = strings.TrimPrefix(command, prefix)
			break
		}
	}

	// Handle command chaining - analyze commands to find environment-affecting ones
	// Split by && and || and ; to get individual commands
	separators := []string{" && ", " || ", " ; ", "; "}
	for _, sep := range separators {
		if strings.Contains(command, sep) {
			parts := strings.Split(command, sep)
			// Look for environment-affecting commands, prioritizing them over install commands
			var firstNonInstall string

			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed == "" {
					continue
				}

				// Check if this command looks like it affects environment
				if ev.looksLikeEnvironmentCommand(trimmed) {
					return trimmed
				}

				// Keep track of first non-install command as fallback
				if firstNonInstall == "" && !ev.isInstallCommand(trimmed) {
					firstNonInstall = trimmed
				}
			}

			// If we found a non-install command, use it
			if firstNonInstall != "" {
				command = firstNonInstall
			}
			break
		}
	}

	return strings.TrimSpace(command)
}

// looksLikeEnvironmentCommand does a quick check to see if a command might affect environment
func (ev *EnvironmentValidator) looksLikeEnvironmentCommand(command string) bool {
	envKeywords := []string{
		"export", "source", ".", "cd", "alias", "unalias", "conda", "activate",
		"deactivate", "nvm", "rbenv", "pyenv", "set", "unset", "module", "ml",
	}

	for _, keyword := range envKeywords {
		if strings.HasPrefix(command, keyword+" ") || command == keyword {
			return true
		}
	}

	// Check for variable assignments
	if strings.Contains(command, "=") && !strings.Contains(command, "==") {
		parts := strings.Split(command, "=")
		if len(parts) >= 2 {
			varName := strings.TrimSpace(parts[0])
			// Check if it looks like a variable name (starts with letter/underscore)
			if len(varName) > 0 && (varName[0] >= 'A' && varName[0] <= 'Z' || varName[0] == '_') {
				return true
			}
		}
	}

	return false
}

// isInstallCommand checks if the command is an installation command (these are OK to run in subshell)
func (ev *EnvironmentValidator) isInstallCommand(command string) bool {
	installPatterns := []string{
		"apt update", "apt upgrade", "apt install", "apt-get update", "apt-get upgrade", "apt-get install",
		"yum install", "yum update", "dnf install", "dnf update",
		"pacman -S", "pacman -Sy", "pacman -Syu",
		"brew install", "brew update", "brew upgrade",
		"pip install", "pip3 install",
		"npm install", "npm update",
		"gem install", "cargo install", "go install", "snap install",
	}

	// Remove sudo prefix for checking
	cleanCmd := command
	if strings.HasPrefix(cleanCmd, "sudo ") {
		cleanCmd = strings.TrimPrefix(cleanCmd, "sudo ")
	}
	cleanCmd = strings.TrimSpace(cleanCmd)

	for _, pattern := range installPatterns {
		if strings.HasPrefix(cleanCmd, pattern) {
			return true
		}
	}

	// Also check for common install command patterns with flags
	installWithFlagsPatterns := []string{
		"apt install -", "apt-get install -",
		"yum install -", "dnf install -",
		"brew install -",
	}

	for _, pattern := range installWithFlagsPatterns {
		if strings.HasPrefix(cleanCmd, pattern) {
			return true
		}
	}

	return false
}

// Specific detectors for different types of environment commands

func (ev *EnvironmentValidator) detectSourceCommand(coreCmd, fullCmd string) bool {
	sourcePatterns := []string{
		"source ",
		". ", // dot command is equivalent to source
	}

	for _, pattern := range sourcePatterns {
		if strings.HasPrefix(coreCmd, pattern) {
			// Make sure it's not just ". " followed by a command
			remaining := strings.TrimPrefix(coreCmd, pattern)
			remaining = strings.TrimSpace(remaining)

			// Check if it looks like a file (has extension or known config files)
			if ev.looksLikeSourceableFile(remaining) {
				return true
			}
		}
	}

	return false
}

func (ev *EnvironmentValidator) detectExportCommand(coreCmd, fullCmd string) bool {
	// Direct export commands
	if strings.HasPrefix(coreCmd, "export ") {
		return true
	}

	// Pattern: VAR=value (without export keyword) - check for uppercase variable names
	// Need to check against the original case since coreCmd is lowercased
	varAssignPattern := regexp.MustCompile(`^[A-Z_][A-Z0-9_]*=`)
	if varAssignPattern.MatchString(strings.TrimSpace(fullCmd)) {
		return true
	}

	// Check for PATH modifications in various forms
	pathPatterns := []string{
		"path=", "PATH=",
		"$path", "$PATH",
	}

	for _, pattern := range pathPatterns {
		if strings.Contains(fullCmd, pattern) {
			return true
		}
	}

	return false
}

func (ev *EnvironmentValidator) detectAliasCommand(coreCmd, fullCmd string) bool {
	return strings.HasPrefix(coreCmd, "alias ") || strings.HasPrefix(coreCmd, "unalias ")
}

func (ev *EnvironmentValidator) detectCdCommand(coreCmd, fullCmd string) bool {
	// Basic cd command
	if strings.HasPrefix(coreCmd, "cd ") || coreCmd == "cd" {
		return true
	}

	// pushd/popd commands
	if strings.HasPrefix(coreCmd, "pushd ") || strings.HasPrefix(coreCmd, "popd") {
		return true
	}

	return false
}

func (ev *EnvironmentValidator) detectVirtualEnvCommand(coreCmd, fullCmd string) bool {
	venvPatterns := []string{
		// Python virtual environments
		"activate",
		"deactivate",
		"workon ",
		"mkvirtualenv ",
		"rmvirtualenv ",
		"virtualenv",
		"python -m venv",
		"python3 -m venv",

		// Poetry
		"poetry shell",
		"poetry env",

		// Pipenv
		"pipenv shell",
		"pipenv activate",
	}

	for _, pattern := range venvPatterns {
		if strings.HasPrefix(coreCmd, pattern) || strings.Contains(coreCmd, pattern) {
			return true
		}
	}

	// Check for activation scripts
	if strings.Contains(fullCmd, "bin/activate") || strings.Contains(fullCmd, "Scripts/activate") {
		return true
	}

	return false
}

func (ev *EnvironmentValidator) detectShellFunctionCommand(coreCmd, fullCmd string) bool {
	// Function definition patterns
	functionPatterns := []string{
		"function ",
		"() {",
	}

	for _, pattern := range functionPatterns {
		if strings.Contains(fullCmd, pattern) {
			return true
		}
	}

	// Check for function calls that modify environment
	if strings.HasPrefix(coreCmd, "unset ") {
		return true
	}

	return false
}

func (ev *EnvironmentValidator) detectEnvironmentModuleCommand(coreCmd, fullCmd string) bool {
	modulePatterns := []string{
		"module load",
		"module unload",
		"module purge",
		"module swap",
		"ml ", // short form of module command
	}

	for _, pattern := range modulePatterns {
		if strings.HasPrefix(coreCmd, pattern) {
			return true
		}
	}

	return false
}

func (ev *EnvironmentValidator) detectPathModification(coreCmd, fullCmd string) bool {
	// Look for PATH modifications that aren't exports - check the full command
	pathModPatterns := []string{
		">> ~/.bashrc",
		">> ~/.zshrc",
		">> ~/.profile",
		">> ~/.bash_profile",
		">> $HOME/.bashrc",
		">> $HOME/.zshrc",
	}

	for _, pattern := range pathModPatterns {
		if strings.Contains(fullCmd, pattern) {
			// Also check that it contains PATH or export to be sure it's path modification
			if strings.Contains(fullCmd, "PATH") || strings.Contains(fullCmd, "export") {
				return true
			}
		}
	}

	return false
}

func (ev *EnvironmentValidator) detectShellOptions(coreCmd, fullCmd string) bool {
	shellOptPatterns := []string{
		"set -", "set +",
		"shopt -s", "shopt -u",
		"setopt", "unsetopt",
		"ulimit",
		"umask",
	}

	for _, pattern := range shellOptPatterns {
		if strings.HasPrefix(coreCmd, pattern) {
			return true
		}
	}

	return false
}

func (ev *EnvironmentValidator) detectCondaEnvironment(coreCmd, fullCmd string) bool {
	condaPatterns := []string{
		"conda activate",
		"conda deactivate",
		"conda env",
		"mamba activate",
		"mamba deactivate",
	}

	for _, pattern := range condaPatterns {
		if strings.HasPrefix(coreCmd, pattern) {
			return true
		}
	}

	return false
}

func (ev *EnvironmentValidator) detectDockerEnvironment(coreCmd, fullCmd string) bool {
	dockerEnvPatterns := []string{
		"eval $(docker-machine env",
		"docker-machine env",
		"$(aws ecr get-login",
		"eval $(aws ecr get-login",
	}

	for _, pattern := range dockerEnvPatterns {
		if strings.Contains(fullCmd, pattern) {
			return true
		}
	}

	return false
}

func (ev *EnvironmentValidator) detectVersionManagers(coreCmd, fullCmd string) bool {
	versionMgrPatterns := []string{
		// rbenv
		"rbenv shell",
		"rbenv local",
		"rbenv global",

		// pyenv
		"pyenv shell",
		"pyenv local",
		"pyenv global",

		// nvm
		"nvm use",
		"nvm alias",

		// nodenv
		"nodenv shell",
		"nodenv local",
		"nodenv global",

		// jenv
		"jenv shell",
		"jenv local",
		"jenv global",

		// tfenv
		"tfenv use",
	}

	for _, pattern := range versionMgrPatterns {
		if strings.HasPrefix(coreCmd, pattern) {
			return true
		}
	}

	return false
}

func (ev *EnvironmentValidator) looksLikeSourceableFile(filename string) bool {
	// Common sourceable file patterns
	sourceablePatterns := []string{
		".bashrc", ".zshrc", ".profile", ".bash_profile",
		".env", ".envrc",
		"activate", // virtualenv activation
		".sh", ".bash", ".zsh",
	}

	filename = strings.ToLower(filename)

	// Check for exact matches or file extensions
	for _, pattern := range sourceablePatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}

	// Check for environment variable files
	if strings.Contains(filename, "env") && (strings.HasSuffix(filename, ".txt") ||
		strings.HasSuffix(filename, ".conf") || !strings.Contains(filename, ".")) {
		return true
	}

	return false
}

// EnvironmentCommandError represents an error for environment-affecting commands
type EnvironmentCommandError struct {
	Command     string
	Reason      string
	Explanation string
}

func (e *EnvironmentCommandError) Error() string {
	return fmt.Sprintf("environment command detected: %s", e.Reason)
}

func (e *EnvironmentCommandError) GetKnightlyMessage() string {
	return fmt.Sprintf(`🏰 I cannot change the realm's environment for you, sire, as %s.
⚔️  However, here is the command you should execute in your own noble shell:

    %s
🛡️  Execute this command directly in your terminal to affect your current environment.`,
		e.Explanation, e.Command)
}
