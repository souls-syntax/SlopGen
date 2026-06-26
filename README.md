# Wrench

**A lightweight, tool-equipped LLM agent for local and remote models, written in Go.**

Wrench turns a local LLM (via [Ollama](https://ollama.com)) or any OpenAI-compatible endpoint into a small autonomous coding/automation agent. It can read files, write files, and run shell commands, looping with the model until the task is done.

![Go](https://img.shields.io/badge/Go-1.25%2B-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Tool use**: the model can call `Read`, `Write`, and `Execute` to interact with your filesystem and shell
- **Agent loop**: tool results are fed back to the model automatically until it produces a final answer
- **Local-first**: works out of the box with Ollama (`qwen2.5:7b` by default)
- **Remote-capable**: point it at any OpenAI-compatible API (e.g. via ngrok)
- **Safety check**: shell commands require an explicit `y` confirmation before running
- **Single binary, simple CLI**: no config files, just a prompt flag

## Prerequisites

- Go 1.25+
- [Ollama](https://ollama.com) running locally (recommended), or access to an OpenAI-compatible endpoint

## Installation

```bash
git clone https://github.com/souls-syntax/Wrench.git
cd Wrench
```

## Usage

### Local (default)

```bash
./your_program.sh -p "Create a Python script that prints 'Hello from Wrench!'"
```

### Remote endpoint

```bash
./your_program.sh -n "https://your-ngrok-url" -p "Analyze the files in this directory"
```

### Flags

| Flag | Required | Description |
|------|----------|--------------|
| `-p` | Yes | Initial prompt sent to the model |
| `-n` | No | Base URL for a remote OpenAI-compatible endpoint (defaults to local Ollama) |

## How It Works

1. Your prompt and the running conversation history are sent to the LLM.
2. The model can call one of three tools, `Read`, `Write`, or `Execute`, instead of (or alongside) a text reply.
3. Tool output is appended to the conversation and sent back to the model.
4. This repeats until the model returns a final response with no further tool calls.

This loop lets the agent explore a codebase, make edits, run commands to verify them, and iterate. Useful for small builds, refactors, or one-off automation tasks.

## Tools

| Tool | Description | Safety |
|------|--------------|--------|
| `Read` | Reads a file in the current working directory | Safe |
| `Write` | Creates or overwrites a file | Safe |
| `Execute` | Runs a shell command | Requires `y` confirmation |

## Project Structure

```
Wrench/
├── app/
│   ├── main.go             # Agent loop, CLI flags, model config
│   └── internal/model/     # Tool definitions (Read / Write / Execute)
├── your_program.sh         # Build + run wrapper
├── go.mod
└── README.md
```

## Configuration

By default, Wrench connects to:

- **Base URL:** `http://localhost:11434/v1` (Ollama)
- **Model:** `qwen2.5:7b`

Passing `-n` overrides the base URL with whatever endpoint you provide.

### Changing the default model

Edit `app/main.go`:

```go
if ngrok_url == "" {
    baseUrl = "http://localhost:11434/v1"
    modelName = "your-model:tag" // e.g. llama3.2, phi3, etc.
}
```

## Building from Source

```bash
go build -o wrench app/*.go
./wrench -p "Hello, who are you?"
```

## Use Cases

- Rapid prototyping of small scripts and tools
- Code maintenance: "refactor this function", "add tests"
- Interactive exploration of an unfamiliar codebase
- Simple file/batch automation

## Why "Wrench"?

Because that's what this is: a wrench you hand the model so it can actually do work instead of just describing it. Read, Write, and Execute are the three tools in the box.

## Contributing

Issues and pull requests are welcome.

## License

MIT
