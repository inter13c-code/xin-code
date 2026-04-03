package tui

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PermissionDialog 权限确认对话框
type PermissionDialog struct {
	visible  bool
	toolName string
	input    string
	response chan PermissionResponse
	width    int
	height   int
}

// NewPermissionDialog 创建权限对话框
func NewPermissionDialog() PermissionDialog {
	return PermissionDialog{}
}

func (p PermissionDialog) Init() tea.Cmd { return nil }

func (p PermissionDialog) Update(msg tea.Msg) (PermissionDialog, tea.Cmd) {
	if !p.visible {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			p.respond(PermAllow)
			return p, nil
		case "n", "N":
			p.respond(PermDeny)
			return p, nil
		case "a", "A":
			p.respond(PermAlways)
			return p, nil
		case "e", "E":
			// 使用 E 表示始终拒绝
			p.respond(PermNever)
			return p, nil
		case "esc", "ctrl+c":
			p.respond(PermDeny)
			return p, nil
		}

	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
	}

	return p, nil
}

func (p PermissionDialog) View() string {
	if !p.visible {
		return ""
	}

	cardWidth := min(78, max(52, p.width-6))
	box := p.Card(cardWidth)
	return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Bottom, box)
}

// Card 渲染权限确认卡片（紧凑模式：工具名+摘要在一行，快捷键在第二行）
func (p PermissionDialog) Card(width int) string {
	if width < 40 {
		width = 40
	}

	// 工具名（品牌色加粗）+ 摘要（dim 色）
	summary := toolInputSummary(p.toolName, p.input)
	nameStyle := lipgloss.NewStyle().Foreground(ColorPerm).Bold(true)
	maxSummaryWidth := width - lipgloss.Width(p.toolName) - 6
	if maxSummaryWidth < 20 {
		maxSummaryWidth = 20
	}
	summaryText := truncateText(summary, maxSummaryWidth)
	header := nameStyle.Render(p.toolName) + "  " + StyleDim.Render(summaryText)

	// 快捷键提示
	keys := StyleDim.Render("y 允许 · n 拒绝 · a 总是允许 · e 总是拒绝")

	return lipgloss.NewStyle().
		BorderLeft(true).
		BorderForeground(ColorPerm).
		PaddingLeft(1).
		Width(width).
		Render(header + "\n" + keys)
}

// toolInputSummary 根据工具名提取紧凑摘要
func toolInputSummary(toolName, rawInput string) string {
	if rawInput == "" {
		return "无参数"
	}

	// 尝试解析 JSON
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(rawInput), &input); err != nil {
		if len(rawInput) > 80 {
			return rawInput[:80] + "..."
		}
		return rawInput
	}

	switch toolName {
	case "Bash":
		if cmd, ok := input["command"].(string); ok {
			if len(cmd) > 80 {
				return cmd[:80] + "..."
			}
			return cmd
		}
	case "Edit", "Write", "Read":
		if fp, ok := input["file_path"].(string); ok {
			return fp
		}
		// 兼容旧字段名
		if fp, ok := input["path"].(string); ok {
			return fp
		}
	case "Glob":
		if p, ok := input["pattern"].(string); ok {
			return p
		}
	case "Grep":
		if p, ok := input["pattern"].(string); ok {
			path := ""
			if pp, ok := input["path"].(string); ok {
				path = " in " + pp
			}
			return p + path
		}
	}

	// 兜底：JSON 前 80 字符
	if len(rawInput) > 80 {
		return rawInput[:80] + "..."
	}
	return rawInput
}

// Show 显示权限对话框
func (p *PermissionDialog) Show(toolName, input string, responseCh chan PermissionResponse) {
	p.visible = true
	p.toolName = toolName
	p.input = input
	p.response = responseCh
}

// Hide 隐藏对话框
func (p *PermissionDialog) Hide() {
	p.visible = false
	p.toolName = ""
	p.input = ""
	p.response = nil
}

// IsVisible 是否可见
func (p PermissionDialog) IsVisible() bool {
	return p.visible
}

func (p *PermissionDialog) respond(r PermissionResponse) {
	if p.response != nil {
		p.response <- r
		close(p.response)
	}
	p.Hide()
}

