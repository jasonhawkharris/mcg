# McGraph Development Notes

## Key Commands

### Build
```bash
go build -o mcg ./cmd/mcg
```

### Test
```bash
# Test typing prevention
./mcg --test-typing-prevention
```

### Run
```bash
# Show help
./mcg help

# Show version
./mcg version

# Start chat (interactive mode)
./mcg chat

# Continue a conversation
./mcg chat --continue <conversation_id>

# Ask a single question
./mcg ask "How do I create a goroutine in Go?"
```

## Recent Changes

1. **LLM Persistence**: Verified that LLM selection persists between sessions
   - LLM selection is stored in `~/.mcgraph/config`
   - When starting a new chat, the previously selected LLM is used
   - When continuing a chat, the LLM used in that conversation is automatically selected

2. **Typing Prevention**: Added check to prevent users from sending messages while the LLM is typing
   - Modified `KeyEnter` handling in `chat_model.go` to check both `!m.waitingForResp` and `!m.typingActive`

3. **Extension System**: Implemented plugin architecture for custom extensions
   - Built-in "system" extension for filesystem access
   - Support for custom Go plugin extensions (.so files)
   - Command interface through `/extension_name command_name args` syntax

4. **Database Integration**: PostgreSQL for conversation persistence
   - Tables for conversations and messages
   - Support for continuing past conversations
   - Automatic title generation from first message

## Code Style

- Use camelCase for variable names
- Use Go standard formatting (run `go fmt ./...` before commits)
- Add meaningful comments for complex logic

## Architecture Notes

- `/cmd/mcg` - Command-line interface
- `/internal/db` - Database operations
- `/internal/extensions` - Extension system
- `/internal/llm` - LLM client implementations
- `/internal/tui` - Terminal user interface