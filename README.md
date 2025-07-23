# Execute My Will

A CLI application that interprets natural language intents and executes appropriate system commands with your permission.
(A digital knight that executes your will)

## Demo Videos

- [Windows](docs/videos/windows-demo-execute-my-will.mp4)
- [Linux](docs/videos/ubuntu-demo-execute-my-will.mp4)

## Features

- Natural language command interpretation
- AI-powered command generation (Gemini, OpenAI, Anthropic)
- **Two execution modes**: Monarch (streamlined) and Royal-Heir (educational)
- System analysis and validation with environment safety checks
- Safe command confirmation with detailed explanations (Royal-Heir mode)
- Alias and environment awareness
- Cross-platform support (Linux, macOS, Windows)
- Simple configuration management
- Comprehensive UI system with medieval knight theming
- Real-time output highlighting and pattern matching
- Multi-step script execution with progress indicators

## Installation

### Prerequisites
```bash
# Install required dependencies
go install golang.org/x/tools/cmd/goimports@latest
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms (Linux, macOS, Windows)
make build-all

# Install to GOPATH/bin
make install

# Manual build (alternative)
go build -o execute-my-will cmd/execute-my-will/main.go
```

## Quick Start

1. **Configure the application** (required on first run):
   ```bash
   ./execute-my-will configure
   ```

   This will start an interactive configuration session where you'll set:
   - AI Provider (gemini, openai, anthropic)
   - API Key (required)
   - **Execution Mode** (monarch or royal-heir)
   - Model (uses defaults if not specified)
   - Max Tokens (default: 1000)
   - Temperature (default: 0.1)

2. **Use the application**:
   ```bash
   ./execute-my-will "list all the contents of my home directory"
   ```

## Execution Modes

### Monarch Mode
**For experienced users who prefer efficiency**
- Streamlined execution with minimal explanations
- Shows only the generated command before confirmation
- Quick confirmation prompts
- Ideal for users comfortable with command-line operations

### Royal-Heir Mode
**For learning users who want to understand**
- Educational experience with detailed explanations
- Shows both the command AND a comprehensive explanation
- Breaks down what each part of the command does
- Helps users learn while accomplishing tasks
- Perfect for beginners or those wanting to expand their knowledge

## Configuration

### Interactive Configuration
Run the configure command without any flags for an interactive setup:

```bash
./execute-my-will configure
```

You'll be prompted for each setting with default values shown in brackets. Press Enter to accept defaults. The interactive mode includes detailed descriptions of each execution mode to help you choose.

### Non-Interactive Configuration
Set specific configuration values using flags:

```bash
# Set a single value
./execute-my-will configure --api-key "your-api-key-here"

# Set execution mode
./execute-my-will configure --mode royal-heir

# Set multiple values
./execute-my-will configure --provider openai --api-key "sk-xxx" --model "gpt-4" --mode monarch

# Update temperature setting
./execute-my-will configure --temperature 0.2
```

### Configuration File
The configuration is stored in `~/.config/execute-my-will/config.yaml`:

```yaml
ai:
  provider: gemini
  api_key: your-api-key-here
  model: gemini-pro
  max_tokens: 1000
  temperature: 0.1
  mode: royal-heir
```

### Runtime Mode Override
You can temporarily override your configured mode for a single command:

```bash
# Use monarch mode for this command only
./execute-my-will --mode monarch "install docker"

# Use royal-heir mode for this command only
./execute-my-will --mode royal-heir "setup nginx reverse proxy"
```

## Usage Examples

```bash
# Basic file operations
./execute-my-will "list my files"
./execute-my-will "copy file.txt to backup directory"

# System operations
./execute-my-will "add /usr/local/bin to my PATH permanently"
./execute-my-will "install docker"

# Archive operations
./execute-my-will "extract archive.zip to current directory"

# Learning mode examples (royal-heir will provide detailed explanations)
./execute-my-will --mode royal-heir "find all large files over 100MB"
./execute-my-will --mode royal-heir "create a secure SSH key"

# View current configuration
./execute-my-will configure --help
```

## Mode Comparison

| Feature | Monarch Mode | Royal-Heir Mode |
|---------|--------------|-----------------|
| Command Display | ✅ Yes | ✅ Yes |
| Detailed Explanation | ❌ No | ✅ Yes |
| Educational Breakdown | ❌ No | ✅ Yes |
| Quick Execution | ✅ Fast | ⚡ Moderate |
| Learning Value | ⭐ Low | ⭐⭐⭐ High |
| Best For | Experienced users | Beginners & learners |

## Safety Features

- **Environment validation**: Blocks commands that would change shell environment (exports, cd, source) since they won't persist
- **Intent validation**: Additional safety layer checking for directory operations and unsafe commands
- **Configuration validation**: Ensures all required settings are present including execution mode
- **Command confirmation**: Always asks before executing commands with clear explanations
- **Educational explanations**: Royal-heir mode provides detailed breakdowns to help users understand commands
- **Directory validation**: Checks that referenced directories exist before command generation
- **System analysis**: Understands your shell, aliases, and available commands for context-aware generation
- **Safe command generation**: AI is instructed to generate safe, non-destructive commands
- **Mode validation**: Ensures only valid execution modes are accepted
- **Cross-platform safety**: Different validation rules for Unix vs Windows environments

## Configuration Commands

| Command | Description |
|---------|-------------|
| `configure` | Interactive configuration setup |
| `configure --api-key KEY` | Set API key |
| `configure --provider PROVIDER` | Set AI provider (gemini/openai/anthropic) |
| `configure --mode MODE` | Set execution mode (monarch/royal-heir) |
| `configure --model MODEL` | Set model name |
| `configure --max-tokens N` | Set maximum tokens |
| `configure --temperature N` | Set temperature (0.0-1.0) |

## Supported AI Providers

### Gemini (Google)
```bash
./execute-my-will configure --provider gemini --api-key your-gemini-key
```
Default model: `gemini-pro`

### OpenAI
```bash
./execute-my-will configure --provider openai --api-key sk-your-openai-key
```
Default model: `gpt-3.5-turbo`

### Anthropic
```bash
./execute-my-will configure --provider anthropic --api-key your-anthropic-key
```
Default model: `claude-3-sonnet-20240229`

## Development

### Development Commands

```bash
# Download and tidy dependencies
make deps

# Format code (includes goimports)
make fmt

# Run linter
make vet

# Run all tests with coverage (generates coverage.html)
make test

# Run tests only
go test -v -race ./test/

# Run specific test
go test ./test -run TestEnvironmentValidator

# Run all checks (test + vet)
make check

# Run in development mode
make dev ARGS="configure"

# Clean build artifacts
make clean

# Run full CI pipeline locally
make ci

# Check if ready for release
make release-check

# Update version (e.g., make update-version VERSION=1.2.3)
make update-version VERSION=x.y.z

# Show all available commands
make help
```

### Testing Strategy

The project includes comprehensive unit tests located in the `/test` directory:
- `env_validator_test.go` - Environment validator functionality
- `intent_validator_test.go` - Intent validation for directory operations
- `ai_client_test.go` - AI provider integration tests
- `command_executor_test.go` - Command execution logic
- `system_analyzer_test.go` - System analysis functionality
- `config_test.go` - Configuration management
- `cli_configure_test.go` - CLI configuration tests
- `error_handling_test.go` - Error handling scenarios
- `mocks.go` - Test mocks and utilities

All tests run with race detection and generate coverage reports as `coverage.html`.

## UI System

The application features a comprehensive UI system with medieval knight theming:

### Key UI Components
- **Status Boxes**: Structured status displays with color coding
- **Command Boxes**: Proposed command display with syntax highlighting
- **Script Boxes**: Multi-step script display with mode awareness
- **Execution Headers**: Command/script execution progress indicators
- **Phase Headers**: Process phase transitions with themed icons
- **Configuration Display**: Structured configuration tables

### UI Features
- **Mode Awareness**: Royal-heir mode shows detailed explanations, monarch mode shows streamlined output
- **Real-time Highlighting**: Pattern matching for errors, warnings, success indicators
- **Cross-platform Consistency**: Unified experience across Unix and Windows
- **Text Wrapping**: Proper handling of long content with emoji-aware width calculations
- **Medieval Knight Theme**: Consistent theming with appropriate emojis and terminology

Your faithful digital knight awaits your commands! ⚔️

Choose your path:
- **Monarch**: Swift and efficient execution for the experienced
- **Royal-Heir**: Patient guidance and education for the learning

## Troubleshooting

**Configuration not found**: If you see a message about configuration not found, run `./execute-my-will configure` to set up your configuration including the execution mode.

**Mode not configured**: If you see a message about mode not being set, run `./execute-my-will configure --mode [monarch|royal-heir]` or use the interactive configuration.

**API key issues**: Make sure your API key is valid and has the necessary permissions for your chosen AI provider.

**Command validation errors**: The application validates directory references and command safety. Make sure referenced paths exist and are accessible.

**Mode selection guidance**:
- Choose **monarch** if you're comfortable with command-line operations and prefer quick execution
- Choose **royal-heir** if you're learning or want to understand what commands do before executing them
