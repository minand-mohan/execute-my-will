# Execute My Will

A CLI application that interprets natural language intents and executes appropriate system commands with your permission.
(A digital knight that executes your will)

## Features

- Natural language command interpretation
- AI-powered command generation (Gemini, OpenAI, Anthropic)
- **Two execution modes**: Monarch (streamlined) and Royal-Heir (educational)
- System analysis and validation
- Safe command confirmation with detailed explanations (Royal-Heir mode)
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

- **Configuration validation**: Ensures all required settings are present including execution mode
- **Command confirmation**: Always asks before executing commands
- **Educational explanations**: Royal-heir mode provides detailed breakdowns to help users understand commands
- **Directory validation**: Checks that referenced directories exist
- **System analysis**: Understands your shell, aliases, and available commands
- **Safe command generation**: AI is instructed to generate safe, non-destructive commands
- **Mode validation**: Ensures only valid execution modes are accepted

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