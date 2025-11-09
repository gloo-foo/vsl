# VSL (Vessel) - Modern Container Runner

A modern Go application for running Linux containers with intelligent directory mounting and git repository awareness.

## Features

- 🚀 **Smart Mounting**: Automatic mounting of current working directory
- 📦 **Git Integration**: Discovers and mounts git repositories automatically
- 📜 **Script Support**: Execute UP script files with shebang-style invocation
- 🔧 **Interactive Mode**: TTY support for interactive container sessions
- 🗂️ **Custom Volumes**: Flexible volume mounting with read-only support
- 🌐 **Network Modes**: Configure container networking (bridge, host, none)
- ⚡ **Modern Architecture**: Built following Go best practices with clean separation of concerns

## Installation

```bash
# Build from source
make build

# Install to $GOPATH/bin
go install ./cmd/vsl
```

## Usage

### Basic Container Execution

```bash
# Run a command in Ubuntu container
vsl run --image ubuntu:latest -- echo "Hello World"

# Interactive shell
vsl run --image alpine:latest --interactive

# With custom working directory
vsl run --image node:latest --working-dir /app -- npm test
```

### Git Repository Integration

By default, `vsl` automatically discovers git repositories and mounts them:

```bash
# Automatically mounts git repo
vsl run --image golang:latest -- go test ./...

# Disable git discovery
vsl run --image node:latest --no-git -- npm test
```

### Custom Volumes

```bash
# Mount additional volumes
vsl run --image postgres:latest \
  --volume /data:/var/lib/postgresql/data \
  --volume ~/.config:/root/.config:ro

# Multiple environment variables
vsl run --image redis:latest \
  --env REDIS_PORT=6379 \
  --env REDIS_PASSWORD=secret
```

### UP Script Files

Create executable UP script files:

```up
#!/usr/bin/env vsl

image ubuntu:latest
command ["echo", "Hello from script"]
working_dir /workspace
environment {
  MY_VAR value1
  ANOTHER another_value
}
volumes [
  "~/.config:/root/.config:ro"
]
stdin_open true
tty true
```

Make it executable and run:

```bash
chmod +x my-script.up
./my-script.up arg1 arg2
```

## Architecture

This project follows modern Go application architecture patterns:

```
cmd/vsl/              # Application entry point
├── main.go           # Minimal main with signal handling

internal/
├── app/              # Application framework
│   ├── action.go     # Generic action handlers
│   ├── flags.go      # Reusable flag helpers
│   ├── output.go     # Output formatting
│   ├── types.go      # Common types
│   ├── log/          # Logging configuration
│   └── commands/     # CLI command structure
│       └── run/      # Run command implementation
│
├── container/        # Container domain
│   ├── types.go      # Domain types (strongly typed)
│   └── run/          # Run business logic
│       ├── config.go # Configuration struct
│       └── run.go    # Implementation
│
├── git/              # Git utilities
│   └── discovery.go  # Repository discovery
│
├── mount/            # Mount utilities
│   └── parser.go     # Volume parsing
│
└── script/           # Script parsing
    └── parser.go     # UP file parser
```

### Key Design Principles

1. **Strong Typing**: Domain concepts use custom types (e.g., `container.Image`, `container.WorkingDir`)
2. **Separation of Concerns**: Clear boundaries between CLI, business logic, and utilities
3. **Generic Actions**: Reusable action handlers with generics for type safety
4. **Dependency Injection**: Functions accept dependencies for easy testing
5. **Structured Logging**: slog-based structured logging throughout
6. **Environment Variables**: All flags support environment variable configuration

## Development

### Build

```bash
make build          # Build all binaries
make build-release  # Build optimized release binaries
make build-debug    # Build with debugging symbols
```

### Testing

```bash
make test              # Run unit tests
make test-all          # Run all tests
make test-integration  # Run integration tests
```

### Code Quality

```bash
make lint       # Run linter
make check      # Run tests, linters, and API compatibility checks
make gorelease  # Check API compatibility
```

### Tools

All development tools are managed via `go.mod` tool directives:

- `golangci-lint` - Comprehensive linter
- `gotestsum` - Enhanced test runner
- `gorelease` - API compatibility checker
- `gqlgen` - GraphQL code generator
- `buf` - Protocol buffer tooling
- `mga` - Code generation

## Configuration

### Environment Variables

All configuration can be set via environment variables with `VSL_` prefix:

```bash
# Logging
export VSL_LOG_LEVEL=debug
export VSL_LOG_FORMAT=json

# Run command
export VSL_RUN_IMAGE=ubuntu:latest
export VSL_RUN_NO_GIT=true
export VSL_RUN_INTERACTIVE=false
```

### Flags

Command-line flags override environment variables:

```bash
vsl --log-level debug run --image alpine:latest --no-git
```

## License

MIT License - see LICENSE file for details

## Credits

Architecture inspired by [modern-go-application](https://github.com/gomatic/modern-go-application) standards.
