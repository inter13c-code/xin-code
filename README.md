# Xin Code

```
 __  __ ___ _   _    ____ ___  ____  _____
 \ \/ /|_ _| \ | |  / ___/ _ \|  _ \| ____|
  \  /  | ||  \| | | |  | | | | | | |  _|
  /  \  | || |\  | | |__| |_| | |_| | |___
 /_/\_\|___|_| \_|  \____\___/|____/|_____|
```

**AI-powered terminal coding agent, built with Go.**

An open-source terminal AI agent that supports multiple model providers, real-time cost tracking, context visualization, and diff preview — all in a single binary with zero dependencies.

---

## Why Xin Code?

| Feature | Claude Code | Xin Code |
|---------|------------|----------|
| Multi-model support | Anthropic only | Anthropic + OpenAI + any compatible endpoint |
| CC OAuth reuse | N/A | Read Claude Code's Keychain token, zero cost for subscribers |
| Real-time cost panel | Token count (small text) | Status bar with live cost in CNY/USD |
| Context visualization | Text warning "almost full" | Progress bar with color gradient (green -> yellow -> red) |
| Diff preview | Text-only confirmation | Colored unified diff panel before writing |
| Startup time | ~500ms (Node.js) | ~50ms (Go binary) |
| Distribution | npm install + Node runtime | `brew install` / single binary / `go install` |

## Features

- **Agent Loop** — Full agentic coding with tool use, permission system, auto-compact, and interrupt recovery
- **Multi-Provider** — Native Anthropic + OpenAI providers, plus any OpenAI-compatible endpoint via `BASE_URL` (OpenRouter, etc.)
- **CC OAuth Reuse** — Automatically reads Claude Code's OAuth token from macOS Keychain
- **20+ Built-in Tools** — Read, Write, Edit, Bash, Glob, Grep, WebFetch, WebSearch, Agent, MCP, Task, and more
- **36 Slash Commands** — `/commit`, `/pr`, `/review`, `/cost`, `/compact`, `/skills`, and more
- **MCP Integration** — Connect to any MCP server (stdio/SSE/HTTP)
- **Skills & Plugins** — Extend with custom skills (`SKILL.md`) and plugins (`plugin.json`)
- **Hooks** — `preToolUse` / `postToolUse` hooks to intercept or react to tool calls
- **Interactive TUI** — Built with [Charm](https://charm.sh) (Bubbletea + Lipgloss + Glamour)
- **Session Management** — Auto-save, resume, export to Markdown
- **XINCODE.md** — Project-level instructions file (like `.cursorrules`)

## Installation

### Homebrew (macOS / Linux)

```bash
brew install xincode-ai/tap/xin-code
```

### Go Install

```bash
go install github.com/xincode-ai/xin-code@latest
```

### Download Binary

```bash
# macOS / Linux
curl -fsSL https://github.com/xincode-ai/xin-code/releases/latest/download/xin-code_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv xin-code /usr/local/bin/

# Verify
xin-code --version
```

## Quick Start

### 1. Configure API Key

```bash
# Option A: Anthropic API Key
export ANTHROPIC_API_KEY=sk-ant-...

# Option B: OpenAI API Key
export OPENAI_API_KEY=sk-...

# Option C: Universal key
export XINCODE_API_KEY=your-key

# Option D: If you have Claude Code installed and logged in,
# Xin Code will automatically reuse your OAuth token (macOS only)
```

### 2. Launch

```bash
# Start in current directory
xin-code

# Check version
xin-code --version
```

### 3. Start Coding

```
> Help me refactor this function to use error wrapping

> /commit     # Auto-generate commit message
> /review     # Code review current changes
> /cost       # Check session cost
> /compact    # Compress context when running low
```

## Configuration

### Settings File

Global config: `~/.xincode/settings.json`
Project config: `.xincode/settings.json`

```json
{
  "model": "claude-sonnet-4-6-20250514",
  "provider": "anthropic",
  "max_tokens": 16384,
  "permissions": {
    "mode": "default",
    "rules": [
      { "tool": "Bash(rm -rf *)", "behavior": "deny" },
      { "tool": "Bash(sudo *)",   "behavior": "deny" }
    ]
  },
  "cost": {
    "currency": "CNY",
    "budget": 10.0,
    "budget_action": "warn"
  }
}
```

### Environment Variables

```bash
XINCODE_API_KEY          # Universal API key
ANTHROPIC_API_KEY        # Anthropic-specific
OPENAI_API_KEY           # OpenAI-specific
XINCODE_MODEL            # Default model
XINCODE_BASE_URL         # Custom API endpoint (OpenRouter, etc.)
XINCODE_PERMISSION_MODE  # bypass / acceptEdits / default / plan / interactive
```

## Multi-Model Support

### Providers

| Provider | Models | Config |
|----------|--------|--------|
| Anthropic | Claude Sonnet, Opus, Haiku | `ANTHROPIC_API_KEY` |
| OpenAI | GPT-4o, o1, o3, o4-mini | `OPENAI_API_KEY` |
| OpenRouter | 200+ models | `OPENAI_API_KEY` + `XINCODE_BASE_URL=https://openrouter.ai/api/v1` |
| Any compatible | Via BASE_URL | `XINCODE_API_KEY` + `XINCODE_BASE_URL` |

### Auto-Routing

Model names are automatically routed to the correct provider:

- `claude-*`, `sonnet-*`, `opus-*`, `haiku-*` -> Anthropic
- `gpt-*`, `o1-*`, `o3-*`, `o4-*` -> OpenAI
- Everything else -> OpenAI (compatible mode)

## Built-in Tools

| Tool | Type | Description |
|------|------|-------------|
| Read | Read-only | Read file contents (images, PDFs supported) |
| Write | Write | Create or overwrite files |
| Edit | Write | Edit files with diff preview |
| Bash | Write | Execute shell commands (streaming output) |
| Glob | Read-only | File pattern matching |
| Grep | Read-only | Content search (ripgrep syntax) |
| WebFetch | Read-only | Fetch web page content |
| WebSearch | Read-only | Web search |
| Agent | Write | Spawn sub-agents with isolated context |
| MCP | Varies | Bridge to MCP server tools |
| AskUser | Read-only | Ask the user a question |
| Task | Write | Task management (create/get/list/update/stop) |

## Slash Commands

| Category | Commands |
|----------|----------|
| Session | `/help` `/session` `/resume` `/compact` `/clear` `/export` `/quit` |
| Model & Config | `/model` `/provider` `/config` `/login` `/logout` `/permissions` `/cost` `/status` |
| Dev Workflow | `/commit` `/pr` `/review` `/diff` `/plan` `/test` `/init` `/branch` `/refactor` |
| Extensions | `/mcp` `/skills` `/plugins` `/hooks` `/agents` `/team` |
| System | `/env` `/version` `/context` `/tips` `/upgrade` `/memory` |

## MCP Integration

Configure MCP servers in `settings.json`:

```json
{
  "mcp_servers": [
    {
      "name": "filesystem",
      "command": "mcp-server-filesystem",
      "args": ["/path/to/dir"]
    }
  ]
}
```

MCP tools are automatically registered as `mcp__<server>__<tool>` and follow the same permission system as built-in tools.

## XINCODE.md

Create a `XINCODE.md` file in your project root to provide project-specific instructions to the AI agent. This is similar to `.cursorrules` or `CLAUDE.md`.

```markdown
# My Project

## Overview
A web application built with Next.js and TypeScript.

## Code Style
- Use functional components with hooks
- Comments in Chinese
- Use pnpm as package manager

## Build & Test
- `pnpm dev` - Start dev server
- `pnpm test` - Run tests
```

Use `/init` to auto-generate one based on your project structure.

## Skills

Create custom skills in `~/.xincode/skills/<name>/SKILL.md` or `.xincode/skills/<name>/SKILL.md`:

```markdown
---
name: my-skill
description: Description for AI to decide when to use
whenToUse: When the user asks about X
---

Skill content here (Markdown, injected as context when needed)
```

## Hooks

Configure hooks in `settings.json`:

```json
{
  "hooks": {
    "preToolUse": [
      {
        "match": "Bash",
        "command": "echo 'Running: $TOOL_INPUT'"
      }
    ],
    "postToolUse": [
      {
        "match": "",
        "command": "echo 'Tool $TOOL_NAME completed'"
      }
    ]
  }
}
```

- `preToolUse` hooks that exit non-zero will block tool execution
- `postToolUse` hooks run after tool completion
- Hooks timeout after 10 seconds

## Permission Modes

| Mode | Behavior |
|------|----------|
| `bypass` | All tools auto-approved |
| `acceptEdits` | File read/write auto-approved, Bash requires confirmation |
| `default` | Read-only auto-approved, write tools require confirmation (recommended) |
| `plan` | Read-only allowed, write operations blocked |
| `interactive` | All tools require confirmation |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

[MIT](LICENSE)
