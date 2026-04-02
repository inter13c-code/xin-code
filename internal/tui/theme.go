package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

func init() {
	// 预设暗色背景，避免 Lipgloss 在运行时查询终端背景色（OSC 11），
	// 查询的响应会被 Bubbletea 的 stdin reader 捕获并泄漏到输入框。
	lipgloss.SetHasDarkBackground(true)

	// 同时通过环境变量强制 Lipgloss 跳过终端检测
	// （某些版本的 lipgloss 不支持 SetHasDarkBackground）
	os.Setenv("CLICOLOR_FORCE", "1")
}

// 品牌色系
var (
	// 主色调
	ColorBrand   = lipgloss.Color("#7C3AED") // 紫色品牌色
	ColorAccent  = lipgloss.Color("#06B6D4") // 青色点缀

	// 语义色
	ColorSuccess = lipgloss.Color("#22C55E") // 绿色
	ColorWarning = lipgloss.Color("#EAB308") // 黄色
	ColorError   = lipgloss.Color("#EF4444") // 红色
	ColorInfo    = lipgloss.Color("#3B82F6") // 蓝色

	// 文本色
	ColorText     = lipgloss.Color("#E2E8F0") // 浅灰文本
	ColorTextDim  = lipgloss.Color("#64748B") // 暗灰辅助文本
	ColorTextBold = lipgloss.Color("#F8FAFC") // 高亮文本

	// 背景色
	ColorBg       = lipgloss.Color("#0F172A") // 深色背景
	ColorBgAlt    = lipgloss.Color("#1E293B") // 交替背景

	// 上下文进度条颜色
	ColorCtxLow  = lipgloss.Color("#22C55E") // <60% 绿色
	ColorCtxMid  = lipgloss.Color("#EAB308") // 60-80% 黄色
	ColorCtxHigh = lipgloss.Color("#EF4444") // >80% 红色
)

// 样式定义
var (
	StyleBrand = lipgloss.NewStyle().
		Foreground(ColorBrand).
		Bold(true)

	StyleModel = lipgloss.NewStyle().
		Foreground(ColorAccent)

	StyleCost = lipgloss.NewStyle().
		Foreground(ColorWarning)

	// 对话区域
	StyleUserMsg = lipgloss.NewStyle().
		Foreground(ColorTextBold).
		Bold(true)

	StyleUserPrefix = lipgloss.NewStyle().
		Foreground(ColorBrand).
		Bold(true)

	StyleAIMsg = lipgloss.NewStyle().
		Foreground(ColorText)

	StyleThinking = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Italic(true)

	// 工具相关
	StyleToolName = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	StyleToolRunning = lipgloss.NewStyle().
		Foreground(ColorInfo)

	StyleToolSuccess = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	StyleToolError = lipgloss.NewStyle().
		Foreground(ColorError)

	StyleToolOutput = lipgloss.NewStyle().
		Foreground(ColorTextDim)

	// 权限对话框
	StylePermBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(0, 1)

	StylePermTitle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	// Diff 预览
	StyleDiffAdd = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	StyleDiffDel = lipgloss.NewStyle().
		Foreground(ColorError)

	StyleDiffCtx = lipgloss.NewStyle().
		Foreground(ColorTextDim)

	StyleDiffHeader = lipgloss.NewStyle().
		Foreground(ColorInfo).
		Bold(true)

	// 输入框
	StyleInputBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBrand).
		Padding(0, 1)

	StyleInputPrompt = lipgloss.NewStyle().
		Foreground(ColorBrand).
		Bold(true)

	// 辅助文本
	StyleTextDim = lipgloss.NewStyle().
		Foreground(ColorTextDim)

	// 提示文本
	StyleHint = lipgloss.NewStyle().
		Foreground(ColorTextDim)

	// 错误消息
	StyleErrorMsg = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)

	// ========== 边框与分隔样式 ==========

	// 主边框（区域分隔）
	StyleBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBgAlt).
		Foreground(ColorText)

	// 标题边框
	StyleTitleBorder = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorBrand).
		BorderTop(true).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false)

	// 分隔线（水平）
	StyleSeparator = lipgloss.NewStyle().
		Foreground(ColorBgAlt)

	// 分割线字符
	SeparatorChar = "─"
	SeparatorDouble = "═"

	// 消息块边框
	StyleUserBlock = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBrand).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(true).
		BorderRight(false).
		Padding(0, 0, 0, 1)

	StyleAssistantBlock = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(true).
		BorderRight(false).
		Padding(0, 0, 0, 1)

	// 工具调用块
	StyleToolBlock = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorBgAlt).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(true).
		BorderRight(false).
		Padding(0, 0, 0, 1)

	// 状态栏样式（带边框）
	StyleStatusBar = lipgloss.NewStyle().
		Background(ColorBgAlt).
		Foreground(ColorText).
		Border(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(ColorBgAlt).
		Padding(0, 1)

	// 输入区边框
	StyleInputFrame = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBrand).
		BorderTop(true).
		BorderBottom(false).
		BorderLeft(true).
		BorderRight(true).
		Padding(1, 1, 0, 1)

	// 角标（左上角）
	StyleCorner = lipgloss.NewStyle().
		Foreground(ColorTextDim)

	// 标签（用于小标签）
	StyleTag = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Background(ColorBgAlt).
		Padding(0, 1)

	StyleTagBrand = lipgloss.NewStyle().
		Foreground(ColorBrand).
		Background(ColorBgAlt).
		Padding(0, 1)
)

// ContextColor 根据上下文使用百分比返回对应颜色
func ContextColor(percent float64) lipgloss.Color {
	switch {
	case percent >= 80:
		return ColorCtxHigh
	case percent >= 60:
		return ColorCtxMid
	default:
		return ColorCtxLow
	}
}
