# mcgraph

A multi-LLM coding assistant CLI tool with conversation history persistence.

## Description

mcgraph is a command-line interface (CLI) tool that provides coding assistance using multiple Large Language Models (LLMs). It allows developers to interact with different AI models to get help with coding tasks and saves conversation history for later reference.

## Installation

```bash
# Clone the repository
git clone https://github.com/hawk/mcgraph.git

# Navigate to the project directory
cd mcgraph

# Build the project
go build -o mcg ./cmd/mcg
```

## Database Setup

McGraph uses PostgreSQL to store conversation history. Before using the history features, you need to set up a PostgreSQL database:

```bash
# Create PostgreSQL database
createdb mcgraph

# Set environment variables for the database connection
export MCGRAPH_DB_HOST=localhost
export MCGRAPH_DB_PORT=5432
export MCGRAPH_DB_USER=postgres
export MCGRAPH_DB_PASSWORD=postgres
export MCGRAPH_DB_NAME=mcgraph
```

You can add these environment variables to your shell profile for persistence.

## Usage

Basic usage:

```bash
# Show help
./mcg help

# Check version
./mcg version

# Pick which LLM to use
./mcg pick openai    # Use OpenAI models
./mcg pick claude    # Use Anthropic Claude models
./mcg pick deepseek  # Use DeepSeek Coder models
./mcg pick gemini    # Use Google's Gemini models

# Ask a one-off coding question
export OPENAI_API_KEY=your_api_key_here      # If using OpenAI
# OR
export ANTHROPIC_API_KEY=your_api_key_here   # If using Claude
# OR
export DEEPSEEK_API_KEY=your_api_key_here    # If using DeepSeek
# OR
export GEMINI_API_KEY=your_api_key_here      # If using Gemini

./mcg ask "How do I create a goroutine in Go?"
./mcg ask --no-save "How do I create a goroutine in Go?"  # Don't save to history

# Start an interactive chat session (TUI)
./mcg chat
# OR
./mcg ask -i

# List saved conversations
./mcg list

# View conversation history
./mcg history
./mcg history show <conversation_id>
./mcg history delete <conversation_id>

# Continue a previous conversation
./mcg chat --continue <conversation_id>
```

## Interactive Mode

The interactive chat mode provides a rich text user interface (TUI) for having multi-turn conversations with McGraph. Features include:

- Full-screen terminal interface
- Message history with timestamps
- Animated typing effect for AI responses
- Animated "Thinking..." indicator with cycling dots
- Syntax highlighting for code blocks
- Code language auto-detection
- Waiting indicators during response generation
- Keyboard navigation
- Automatic conversation saving

To exit the chat, press Ctrl+C or Esc.

## Conversation History

McGraph saves all conversations to a PostgreSQL database for later reference:

- Each conversation is automatically saved with a unique ID
- Titles are generated automatically from the first user message
- View past conversations with `mcg history` or `mcg list`
- Continue previous conversations with `mcg chat --continue <id>`
- Delete conversations with `mcg history delete <id>`

All history commands work with shortened IDs (first 8 characters) for convenience.

## Environment Variables

### LLM API Keys
- `OPENAI_API_KEY`: Required for API access to OpenAI's models.
- `ANTHROPIC_API_KEY`: Required for API access to Anthropic's Claude models.
- `DEEPSEEK_API_KEY`: Required for API access to DeepSeek's models.
- `GEMINI_API_KEY`: Required for API access to Google's Gemini models.

### Database Configuration
- `MCGRAPH_DB_HOST`: PostgreSQL host (default: localhost)
- `MCGRAPH_DB_PORT`: PostgreSQL port (default: 5432)
- `MCGRAPH_DB_USER`: PostgreSQL username (default: postgres)
- `MCGRAPH_DB_PASSWORD`: PostgreSQL password (default: postgres)
- `MCGRAPH_DB_NAME`: PostgreSQL database name (default: mcgraph)

## Features

- Multi-LLM support:
  - OpenAI GPT models
  - Anthropic Claude models
  - DeepSeek Coder models
  - Google's Gemini models
- Interactive chat mode (TUI)
- Conversation history persistence
- Simple CLI interface
- Ask coding questions
- List available LLMs
- Pick different LLMs
- Version command

## Configuration

The tool stores your LLM preference in `~/.mcgraph/config`. You can modify this directly or use the `pick` command.