// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package system

// SystemAnalyzer defines the interface for system analysis operations
type SystemAnalyzer interface {
	AnalyzeSystem() (*Info, error)
}

// CommandExecutor defines the interface for command execution operations
type CommandExecutor interface {
	Execute(command string, shell string) error
	ExecuteScript(scriptContent string, shell string, showComments bool) error
}

// EnvironmentValidatorInterface defines the interface for environment validation
type EnvironmentValidatorInterface interface {
	ValidateEnvironmentCommand(command string) error
}

// IntentValidator defines the interface for intent validation
type IntentValidator interface {
	ValidateIntent(intent string) error
}

// Ensure our concrete types implement the interfaces at compile time
var (
	_ SystemAnalyzer                = (*Analyzer)(nil)
	_ CommandExecutor               = (*Executor)(nil)
	_ EnvironmentValidatorInterface = (*EnvironmentValidator)(nil)
	_ IntentValidator               = (*Validator)(nil)
)