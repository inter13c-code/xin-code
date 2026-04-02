# Xin Code

AI-powered terminal coding agent, built with Go.

## Tech Stack

- Go 1.23+
- Bubbletea (TUI framework, Elm architecture)
- Lipgloss (styling), Glamour (Markdown rendering)
- Anthropic SDK (`anthropics/anthropic-sdk-go`)
- OpenAI SDK (`openai/openai-go`)
- MCP (`mark3labs/mcp-go`)

## Build & Test

```bash
make build    # 构建二进制
make test     # 运行测试
make lint     # 代码检查
make clean    # 清理构建产物
```

## Project Structure

- `main.go` — 入口，初始化 Provider / TUI / Agent
- `agent.go` — Agent 循环引擎（消息 -> API -> 工具 -> 循环）
- `config.go` — 配置加载与合并（环境变量 > 项目配置 > 全局配置）
- `internal/provider/` — 多模型 Provider 抽象（Anthropic + OpenAI）
- `internal/tool/` — 工具系统（接口 + 注册表 + 并发执行器 + 权限）
- `internal/tool/builtin/` — 内置工具实现
- `internal/auth/` — 认证链（API Key + CC OAuth Keychain 读取）
- `internal/tui/` — 终端 UI（Bubbletea Model/Update/View）
- `internal/session/` — 会话管理（持久化 + 自动压缩）
- `internal/slash/` — 斜杠命令路由
- `internal/mcp/` — MCP 客户端
- `internal/skills/` — 技能系统
- `internal/plugins/` — 插件系统
- `internal/hooks/` — 钩子系统
- `internal/cost/` — 费用追踪
- `internal/context/` — 项目上下文 + System Prompt 组装

## Code Conventions

- 代码用英文，注释用中文
- 每个 internal 包职责单一，通过接口解耦
- 工具实现 `tool.Tool` 接口，Provider 实现 `provider.Provider` 接口
- 流式事件通过 channel 传递（`<-chan provider.Event`）
- TUI 和 Agent 通过 channel 通信，避免直接依赖

## Key Patterns

- **Provider 接口**: `Stream(ctx, req) (<-chan Event, error)` — 统一流式 API
- **工具执行**: 只读工具并发（max 10），写入工具串行
- **权限系统**: 五档模式（bypass / acceptEdits / default / plan / interactive）
- **钩子系统**: preToolUse 可阻止执行，postToolUse 仅通知
- **认证链**: 环境变量 > 配置文件 > CC OAuth Keychain

## Important Notes

- 不要修改 `version.go` 中的变量值，它们由 ldflags 在构建时注入
- 新增工具需要在 `builtin/register.go` 的 `RegisterAll` 中注册
- 新增 Provider 需要在 `provider/registry.go` 的 `NewProvider` 中注册
- MCP 工具通过 `mcp__<server>__<tool>` 命名空间注册
