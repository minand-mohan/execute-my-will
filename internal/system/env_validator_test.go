// File: internal/system/env_validator_test.go
package system

import (
	"strings"
	"testing"
)

func TestEnvironmentValidator(t *testing.T) {
	// Mock system info
	sysInfo := &Info{
		OS:    "linux",
		Shell: "bash",
	}

	validator := NewEnvironmentValidator(sysInfo)

	testCases := []struct {
		name           string
		command        string
		shouldError    bool
		expectedReason string
	}{
		// Source commands
		{
			name:           "source bashrc",
			command:        "source ~/.bashrc",
			shouldError:    true,
			expectedReason: "source",
		},
		{
			name:           "dot source command",
			command:        ". /etc/environment",
			shouldError:    true,
			expectedReason: "source",
		},
		{
			name:           "source with sudo",
			command:        "sudo source ~/.zshrc",
			shouldError:    true,
			expectedReason: "source",
		},

		// Export commands
		{
			name:           "export variable",
			command:        "export PATH=$PATH:/usr/local/bin",
			shouldError:    true,
			expectedReason: "export",
		},
		{
			name:           "variable assignment",
			command:        "JAVA_HOME=/usr/lib/jvm/java-11",
			shouldError:    true,
			expectedReason: "export",
		},
		{
			name:           "path modification",
			command:        "echo 'export PATH=$PATH:/new/path' >> ~/.bashrc",
			shouldError:    true,
			expectedReason: "path_modification",
		},
		{
			name:           "export with sudo",
			command:        "which python && export PYTHON_ENV=$(which python)",
			shouldError:    true,
			expectedReason: "export",
		},

		// CD commands
		{
			name:           "change directory",
			command:        "cd /home/user/projects",
			shouldError:    true,
			expectedReason: "cd",
		},
		{
			name:           "pushd command",
			command:        "pushd /tmp",
			shouldError:    true,
			expectedReason: "cd",
		},

		// Virtual environment commands
		{
			name:           "activate virtualenv",
			command:        "source venv/bin/activate",
			shouldError:    true,
			expectedReason: "source", // Will be caught by source detector
		},
		{
			name:           "conda activate",
			command:        "conda activate myenv",
			shouldError:    true,
			expectedReason: "conda_env",
		},
		{
			name:           "poetry shell",
			command:        "poetry shell",
			shouldError:    true,
			expectedReason: "virtual_env",
		},
		{
			name:           "pipenv shell",
			command:        "pipenv shell",
			shouldError:    true,
			expectedReason: "virtual_env",
		},

		// Version managers
		{
			name:           "nvm use",
			command:        "nvm use 16",
			shouldError:    true,
			expectedReason: "rbenv_pyenv",
		},
		{
			name:           "rbenv shell",
			command:        "rbenv shell 3.0.0",
			shouldError:    true,
			expectedReason: "rbenv_pyenv",
		},
		{
			name:           "pyenv local",
			command:        "pyenv local 3.9.0",
			shouldError:    true,
			expectedReason: "rbenv_pyenv",
		},

		// Alias commands
		{
			name:           "create alias",
			command:        "alias ll='ls -la'",
			shouldError:    true,
			expectedReason: "alias",
		},
		{
			name:           "remove alias",
			command:        "unalias ll",
			shouldError:    true,
			expectedReason: "alias",
		},

		// Shell options
		{
			name:           "set option",
			command:        "set -e",
			shouldError:    true,
			expectedReason: "shell_options",
		},
		{
			name:           "shell option",
			command:        "shopt -s histappend",
			shouldError:    true,
			expectedReason: "shell_options",
		},
		{
			name:           "ulimit",
			command:        "ulimit -n 4096",
			shouldError:    true,
			expectedReason: "shell_options",
		},

		// Function definitions
		{
			name:           "function definition",
			command:        "function myFunc() { echo 'hello'; }",
			shouldError:    true,
			expectedReason: "shell_function",
		},
		{
			name:           "unset variable",
			command:        "unset JAVA_HOME",
			shouldError:    true,
			expectedReason: "shell_function",
		},

		// Docker environment
		{
			name:           "docker machine env",
			command:        "eval $(docker-machine env default)",
			shouldError:    true,
			expectedReason: "docker_env",
		},
		{
			name:           "aws ecr login",
			command:        "eval $(aws ecr get-login --no-include-email)",
			shouldError:    true,
			expectedReason: "docker_env",
		},

		// Environment modules
		{
			name:           "module load",
			command:        "module load gcc/9.2.0",
			shouldError:    true,
			expectedReason: "environment_module",
		},
		{
			name:           "ml command",
			command:        "ml python/3.8",
			shouldError:    true,
			expectedReason: "environment_module",
		},

		// Commands that should NOT error
		{
			name:        "list files",
			command:     "ls -la",
			shouldError: false,
		},
		{
			name:        "install package",
			command:     "sudo apt-get install vim",
			shouldError: false,
		},
		{
			name:        "copy files",
			command:     "cp file1.txt file2.txt",
			shouldError: false,
		},
		{
			name:        "create directory",
			command:     "mkdir -p /tmp/test",
			shouldError: false,
		},
		{
			name:        "grep search",
			command:     "grep -r 'pattern' /var/log/",
			shouldError: false,
		},
		{
			name:        "install then use tool",
			command:     "sudo apt install unzip && unzip archive.zip",
			shouldError: false,
		},
		{
			name:        "pip install",
			command:     "pip install requests",
			shouldError: false,
		},
		{
			name:        "echo command",
			command:     "echo 'Hello World'",
			shouldError: false,
		},
		{
			name:        "systemctl command",
			command:     "sudo systemctl restart nginx",
			shouldError: false,
		},
		{
			name:        "docker run",
			command:     "docker run -it ubuntu:latest bash",
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateEnvironmentCommand(tc.command)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for command '%s', but got none", tc.command)
					return
				}

				envErr, ok := err.(*EnvironmentCommandError)
				if !ok {
					t.Errorf("Expected EnvironmentCommandError, got %T", err)
					return
				}

				if tc.expectedReason != "" && envErr.Reason != tc.expectedReason {
					t.Errorf("Expected reason '%s', got '%s' for command '%s'",
						tc.expectedReason, envErr.Reason, tc.command)
				}

				// Test that the knightly message is generated
				msg := envErr.GetKnightlyMessage()
				if msg == "" {
					t.Errorf("Expected non-empty knightly message for command '%s'", tc.command)
				}

				// Check that the original command is included in the message
				if !strings.Contains(msg, tc.command) {
					t.Errorf("Expected knightly message to contain original command '%s'", tc.command)
				}

			} else {
				if err != nil {
					t.Errorf("Expected no error for command '%s', but got: %v", tc.command, err)
				}
			}
		})
	}
}

func TestExtractCoreCommand(t *testing.T) {
	validator := NewEnvironmentValidator(&Info{})

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple command",
			input:    "ls -la",
			expected: "ls -la",
		},
		{
			name:     "sudo command",
			input:    "sudo source ~/.bashrc",
			expected: "source ~/.bashrc",
		},
		{
			name:     "command chain with install",
			input:    "sudo apt install curl && source ~/.bashrc",
			expected: "source ~/.bashrc",
		},
		{
			name:     "command with pipes",
			input:    "export PATH=$PATH:/usr/local/bin",
			expected: "export PATH=$PATH:/usr/local/bin",
		},
		{
			name:     "complex chain",
			input:    "sudo apt update && sudo apt install -y nodejs && source ~/.nvm/nvm.sh",
			expected: "source ~/.nvm/nvm.sh",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.extractCoreCommand(strings.ToLower(tc.input))
			if result != strings.ToLower(tc.expected) {
				t.Errorf("Expected '%s', got '%s'", strings.ToLower(tc.expected), result)
			}
		})
	}
}

// Benchmark test for performance
func BenchmarkValidateEnvironmentCommand(b *testing.B) {
	validator := NewEnvironmentValidator(&Info{OS: "linux", Shell: "bash"})
	commands := []string{
		"ls -la",
		"source ~/.bashrc",
		"export PATH=$PATH:/usr/local/bin",
		"sudo apt install vim",
		"cd /home/user",
		"conda activate myenv",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cmd := range commands {
			validator.ValidateEnvironmentCommand(cmd)
		}
	}
}
