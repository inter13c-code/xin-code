# Contributing to Xin Code

## Development Setup

### Prerequisites

- Go 1.23+
- golangci-lint (for linting)

### Build

```bash
# Build binary
make build

# Run
./xin-code --version

# Install to $GOPATH/bin
make install
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

## Code Style

- **Language**: Go code in English, comments in Chinese (中文注释)
- **Formatting**: `gofmt` (enforced by CI)
- **Linting**: `golangci-lint` with project config (`.golangci.yml`)

## Project Structure

```
xin-code/
├── main.go              # Minimal entry point
├── agent.go             # Agent loop engine
├── config.go            # Config loading and merging
├── version.go           # Version info (ldflags)
└── internal/
    ├── provider/        # Multi-model provider abstraction
    ├── tool/            # Tool system + executor
    │   └── builtin/     # Built-in tools (Read, Write, Bash, etc.)
    ├── auth/            # Authentication chain (API key, CC OAuth)
    ├── mcp/             # MCP client
    ├── session/         # Session management
    ├── context/         # Project context + system prompt
    ├── slash/           # Slash commands
    ├── skills/          # Skills system
    ├── plugins/         # Plugins system
    ├── hooks/           # Hooks system
    ├── tui/             # Terminal UI (Bubbletea)
    └── cost/            # Cost tracking
```

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/your-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Commit with a clear message
7. Open a Pull Request

## Commit Messages

- Use Chinese for commit messages
- Keep it concise and descriptive
- Examples:
  - `feat: 添加 OpenAI Provider`
  - `fix: 修复流式响应中断问题`
  - `refactor: 重构权限检查逻辑`

## Adding a New Tool

1. Create `internal/tool/builtin/yourtool.go`
2. Implement the `tool.Tool` interface
3. Register in `internal/tool/builtin/register.go`
4. Add tests

## Adding a New Provider

1. Create `internal/provider/yourprovider.go`
2. Implement the `provider.Provider` interface
3. Register in `internal/provider/registry.go`

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
