// File: README.md
# Execute My Will

A CLI application that interprets natural language intents and executes appropriate system commands with your permission.
(A digital knight that executes your will)

## Features

- Natural language command interpretation
- AI-powered command generation (Gemini, OpenAI, Anthropic)
- System analysis and validation
- Safe command confirmation
- Alias and environment awareness
- Cross-platform support (Linux focus)
- Simple configuration management

## Installation

```bash
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
   - Model (uses defaults if not specified)
   - Max Tokens (default: 1000)
   - Temperature (default: 0.1)

2. **Use the application**:
   ```bash
   ./execute-my-will "list all the contents of my home directory"
   ```

## Configuration

### Interactive Configuration
Run the configure command without any flags for an interactive setup:

```bash
./execute-my-will configure
```

You'll be prompted for each setting with default values shown in brackets. Press Enter to accept defaults.

### Non-Interactive Configuration
Set specific configuration values using flags:

```bash
# Set a single value
./execute-my-will configure --api-key "your-api-key-here"

# Set multiple values
./execute-my-will configure --provider openai --api-key "sk-xxx" --model "gpt-4"

# Update temperature setting
./execute-my-will configure --temperature 0.2
```

### Configuration File
The configuration is stored in `~/.execute-my-will.yaml`:

```yaml
ai:
  provider: gemini
  api_key: your-api-key-here
  model: gemini-pro
  max_tokens: 1000
  temperature: 0.1
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

# View current configuration
./execute-my-will configure --help
```

## Safety Features

- **Configuration validation**: Ensures all required settings are present
- **Command confirmation**: Always asks before executing commands
- **Directory validation**: Checks that referenced directories exist
- **System analysis**: Understands your shell, aliases, and available commands
- **Safe command generation**: AI is instructed to generate safe, non-destructive commands

## Configuration Commands

| Command | Description |
|---------|-------------|
| `configure` | Interactive configuration setup |
| `configure --api-key KEY` | Set API key |
| `configure --provider PROVIDER` | Set AI provider (gemini/openai/anthropic) |
| `configure --model MODEL` | Set model name |
| `configure --max-tokens N` | Set maximum tokens |
| `configure --temperature N` | Set temperature (0.0-1.0) |

## Supported AI Providers

### Gemini (Google)
```bash
./execute-my-will configure --provider gemini --api-key your-gemini-key
```
Default model: `gemini-pro`

### OpenAI (Coming Soon)
```bash
./execute-my-will configure --provider openai --api-key sk-your-openai-key
```
Default model: `gpt-3.5-turbo`

### Anthropic (Coming Soon)  
```bash
./execute-my-will configure --provider anthropic --api-key your-anthropic-key
```
Default model: `claude-3-sonnet-20240229`

Your faithful digital knight awaits your commands! ⚔️

## Troubleshooting

**Configuration not found**: If you see a message about configuration not found, run `./execute-my-will configure` to set up your configuration.

**API key issues**: Make sure your API key is valid and has the necessary permissions for your chosen AI provider.

**Command validation errors**: The application validates directory references and command safety. Make sure referenced paths exist and are accessible.