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
- Flexible configuration with Cobra/Viper

## Installation

```bash
go build -o execute-my-will cmd/execute-my-will/main.go
```

## Configuration

The application supports multiple configuration methods:

### 1. Command Line Flags
```bash
./execute-my-will --api-key "your-key" --provider gemini "list my files"
```

### 2. Config File (`~/.execute-my-will.yaml`)
```yaml
ai:
  provider: gemini
  api_key: your-api-key-here
  model: gemini-pro
  max_tokens: 1000
  temperature: 0.1
```

### 3. Environment Variables
```bash
export EXECUTE_MY_WILL_API_KEY="your-api-key"
export EXECUTE_MY_WILL_PROVIDER="gemini"
./execute-my-will "list my files"
```

### Configuration Priority
1. Command line flags (highest priority)
2. Environment variables
3. Config file
4. Default values (lowest priority)

## Usage

```bash
# Basic usage
./execute-my-will "list all the contents of my home directory"

# With flags
./execute-my-will --provider openai --api-key sk-xxx "add /python3/ to my path"

# Using custom config file
./execute-my-will --config /path/to/config.yaml "extract zip file"

# View help
./execute-my-will --help
```

## Available Flags

- `--config`: Custom config file path
- `--provider`: AI provider (gemini, openai, anthropic)
- `--api-key`: API key for the provider
- `--model`: Model to use (uses provider defaults if not specified)
- `--max-tokens`: Maximum tokens for response (default: 1000)
- `--temperature`: AI temperature setting (default: 0.1)

## Configuration Examples

### Gemini (Google)
```yaml
ai:
  provider: gemini
  api_key: your-gemini-key
  model: gemini-pro
```

### OpenAI
```yaml
ai:
  provider: openai
  api_key: sk-your-openai-key
  model: gpt-4
```

### Anthropic
```yaml
ai:
  provider: anthropic
  api_key: your-anthropic-key
  model: claude-3-sonnet-20240229
```

## Safety Features

- Command confirmation before execution
- Directory existence validation
- System capability analysis
- Safe command generation

Your faithful digital knight awaits your commands! ⚔️