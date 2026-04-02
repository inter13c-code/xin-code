// internal/hooks/hooks.go
// 钩子系统：preToolUse / postToolUse 事件
package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// EventType 钩子事件类型
type EventType string

const (
	EventPreToolUse  EventType = "preToolUse"
	EventPostToolUse EventType = "postToolUse"
)

// hookTimeout 钩子执行超时时间
const hookTimeout = 10 * time.Second

// HookDef 单个钩子定义
type HookDef struct {
	Match   string `json:"match"`   // 匹配的工具名（空表示匹配所有）
	Command string `json:"command"` // 要执行的 shell 命令
}

// HooksConfig 钩子配置（从 settings.json 读取）
type HooksConfig struct {
	PreToolUse  []HookDef `json:"preToolUse,omitempty"`
	PostToolUse []HookDef `json:"postToolUse,omitempty"`
}

// Manager 钩子管理器
type Manager struct {
	config HooksConfig
}

// NewManager 创建钩子管理器
func NewManager(config HooksConfig) *Manager {
	return &Manager{config: config}
}

// LoadConfig 从 settings.json 加载钩子配置
func LoadConfig(settingsPath string) HooksConfig {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return HooksConfig{}
	}

	// 从 settings.json 的 hooks 字段解析
	var settings struct {
		Hooks HooksConfig `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return HooksConfig{}
	}

	return settings.Hooks
}

// RunPreToolUse 执行 preToolUse 钩子
// 返回 false 表示应阻止工具执行
func (m *Manager) RunPreToolUse(ctx context.Context, toolName, toolInput string) (bool, error) {
	hooks := m.matchHooks(EventPreToolUse, toolName)
	if len(hooks) == 0 {
		return true, nil
	}

	for _, hook := range hooks {
		env := map[string]string{
			"TOOL_NAME":  toolName,
			"TOOL_INPUT": toolInput,
		}
		exitCode, err := runHook(ctx, hook.Command, env)
		if err != nil {
			return true, err // 执行出错不阻止
		}
		if exitCode != 0 {
			return false, nil // 非 0 退出码阻止工具执行
		}
	}

	return true, nil
}

// RunPostToolUse 执行 postToolUse 钩子
func (m *Manager) RunPostToolUse(ctx context.Context, toolName, toolOutput string, isError bool) {
	hooks := m.matchHooks(EventPostToolUse, toolName)
	if len(hooks) == 0 {
		return
	}

	errStr := "false"
	if isError {
		errStr = "true"
	}

	for _, hook := range hooks {
		env := map[string]string{
			"TOOL_NAME":   toolName,
			"TOOL_OUTPUT": toolOutput,
			"TOOL_ERROR":  errStr,
		}
		_, _ = runHook(ctx, hook.Command, env)
	}
}

// matchHooks 找出匹配的钩子
func (m *Manager) matchHooks(event EventType, toolName string) []HookDef {
	var source []HookDef
	switch event {
	case EventPreToolUse:
		source = m.config.PreToolUse
	case EventPostToolUse:
		source = m.config.PostToolUse
	}

	var matched []HookDef
	for _, hook := range source {
		if hook.Match == "" || hook.Match == toolName || matchWildcard(hook.Match, toolName) {
			matched = append(matched, hook)
		}
	}
	return matched
}

// matchWildcard 简易通配符匹配（支持 * 前缀/后缀）
func matchWildcard(pattern, name string) bool {
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(name, prefix)
	}
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(name, suffix)
	}
	return pattern == name
}

// runHook 执行 shell 命令
func runHook(ctx context.Context, command string, env map[string]string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, hookTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	// 继承当前环境变量并添加钩子专用变量
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return -1, err
	}
	return 0, nil
}

// HasHooks 检查是否配置了任何钩子
func (m *Manager) HasHooks() bool {
	return len(m.config.PreToolUse) > 0 || len(m.config.PostToolUse) > 0
}

// ListString 返回钩子列表的格式化字符串
func (m *Manager) ListString() string {
	if !m.HasHooks() {
		return "未配置任何钩子\n\n在 settings.json 中配置:\n  {\n    \"hooks\": {\n      \"preToolUse\": [{\"match\": \"Bash\", \"command\": \"echo $TOOL_INPUT\"}]\n    }\n  }"
	}

	var sb strings.Builder
	sb.WriteString("🪝 已配置的钩子\n\n")

	if len(m.config.PreToolUse) > 0 {
		sb.WriteString("  preToolUse:\n")
		for _, h := range m.config.PreToolUse {
			match := h.Match
			if match == "" {
				match = "*"
			}
			sb.WriteString(fmt.Sprintf("    [%s] %s\n", match, h.Command))
		}
	}

	if len(m.config.PostToolUse) > 0 {
		sb.WriteString("  postToolUse:\n")
		for _, h := range m.config.PostToolUse {
			match := h.Match
			if match == "" {
				match = "*"
			}
			sb.WriteString(fmt.Sprintf("    [%s] %s\n", match, h.Command))
		}
	}

	return sb.String()
}
