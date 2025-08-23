# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based multi-agent system demonstrating LLM agents with tool capabilities. The project showcases:

1. **Multi-agent architecture**: Two specialized agents (coder and documentation) that can work independently or collaboratively
2. **Tool-based LLM interaction**: Agents use predefined tools to interact with the environment
3. **Concurrency patterns**: Leverages Go's goroutines and channels for parallel tool execution and inter-agent communication
4. **Network-based agent communication**: Agents can communicate over HTTP, enabling distributed deployment

## Architecture

### Core Components

- **Agent (`agent.go`)**: Base agent structure with HTTP server capabilities for network communication
- **Coder Agent (`coder_agent.go`)**: CLI-based agent with file system and command execution tools
- **Documentation Agent (`documentation_agent.go`)**: Network-based agent for searching Go documentation and RAG databases
- **Tools (`tools.go`)**: Tool definition framework with JSON schema generation for LLM function calling

### Agent Types

**Coder Agent (Port 8080)**:
- Tools: `read_file`, `execute_command`, `invoke_documentation_agent`
- Input/Output: CLI-based (stdin/stdout)
- Use case: Interactive coding assistance with access to local file system

**Documentation Agent (Port 8081)**:
- Tools: `search_go_documentation`, `search_rag_database`
- Input/Output: HTTP-based network communication
- Use case: Specialized documentation retrieval that can be invoked by other agents

## Development Commands

### Building and Running

```bash
# Build the project
go build -o react-go

# Run the coder agent (default)
./react-go -agent coder

# Run the documentation agent
./react-go -agent doc
```

### Multi-Agent Setup

For full functionality, run both agents simultaneously in separate terminals:

```bash
# Terminal 1: Start documentation agent
./react-go -agent doc

# Terminal 2: Start coder agent  
./react-go -agent coder
```

The coder agent can then invoke the documentation agent via the `invoke_documentation_agent` tool.

### Dependencies

```bash
# Install/update dependencies
go mod tidy

# Add new dependencies
go get <package>
```

## Key Patterns

### Tool Definition
Tools are defined with:
- JSON schema generation using struct tags
- Function signature: `func(input json.RawMessage) (string, error)`
- Anthropic SDK integration for LLM function calling

### Concurrency
- Parallel tool execution using goroutines and channels
- System prompt encourages LLM to use parallel tool calls for efficiency
- Network agents use channels for request/response handling

### Inter-Agent Communication
- HTTP POST requests with JSON payloads
- Agents can be deployed independently and communicate over network
- Port-based service discovery (8080 for coder, 8081 for documentation)

## Demo Use Cases

The project includes a demonstration where the coder agent can find secret passphrases in the `data/` directory by:
1. Using `execute_command` to list files (`ls data/`)
2. Using parallel `read_file` calls to examine multiple files simultaneously
3. Synthesizing information to locate the passphrase

## LLM Configuration

- Uses Anthropic's Claude Sonnet 4 model
- Configured for parallel tool execution
- System prompt encourages efficient concurrent operations
- Tool results are processed via channels to support parallel execution