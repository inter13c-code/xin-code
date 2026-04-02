package tool

// PermissionChecker 权限检查接口
type PermissionChecker interface {
	Check(toolName string, isReadOnly bool) (allowed bool, reason string)
}

// PermissionMode 权限模式
type PermissionMode string

const (
	ModeBypass      PermissionMode = "bypass"
	ModeAcceptEdits PermissionMode = "acceptEdits"
	ModeDefault     PermissionMode = "default"
	ModePlan        PermissionMode = "plan"
	ModeInteractive PermissionMode = "interactive"
)

// SimplePermissionChecker 基于模式的简单权限检查器
// Phase 1 使用简单实现，Phase 2 加入规则系统和用户交互
type SimplePermissionChecker struct {
	Mode PermissionMode
}

func (c *SimplePermissionChecker) Check(toolName string, isReadOnly bool) (bool, string) {
	switch c.Mode {
	case ModeBypass:
		return true, ""
	case ModeAcceptEdits:
		// 文件操作自动放行，Bash 等需要检查
		if isReadOnly || toolName == "Write" || toolName == "Edit" {
			return true, ""
		}
		// Phase 2: 这里会弹出 TUI 确认对话框
		// Phase 1: 暂时自动放行所有工具，不做权限拦截。
		// WARNING: 这意味着 Bash 等危险工具无需用户确认即可执行。
		// Phase 2 必须实现 TUI 确认机制后才能移除此放行逻辑。
		return true, ""
	case ModeDefault:
		if isReadOnly {
			return true, ""
		}
		// Phase 1: 暂时自动放行所有写入工具，不做权限拦截。
		// WARNING: 生产环境中，非只读工具必须经过用户确认。
		// Phase 2 必须实现 TUI 确认机制后才能移除此放行逻辑。
		return true, ""
	case ModePlan:
		if isReadOnly {
			return true, ""
		}
		return false, "plan mode: write operations are not allowed"
	case ModeInteractive:
		// Phase 2: 所有工具都需要弹出 TUI 确认框
		// Phase 1: 暂时自动放行所有工具，不做权限拦截。
		// WARNING: interactive 模式的本意是每个工具调用都需用户确认。
		// Phase 2 必须实现 TUI 确认机制后才能移除此放行逻辑。
		return true, ""
	default:
		return true, ""
	}
}
