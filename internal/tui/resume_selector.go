package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ResumeEntry 历史会话条目
type ResumeEntry struct {
	ID    string
	Model string
	Turns int
	Cost  string
}

// ResumeSelector 会话恢复选择器
type ResumeSelector struct {
	visible  bool
	entries  []ResumeEntry
	selected int
	width    int
	height   int
}

// NewResumeSelector 创建会话恢复选择器
func NewResumeSelector() ResumeSelector {
	return ResumeSelector{}
}

// Show 显示选择器并填充条目
func (r *ResumeSelector) Show(entries []ResumeEntry) {
	r.visible = true
	r.entries = entries
	r.selected = 0
}

// Hide 隐藏选择器
func (r *ResumeSelector) Hide() {
	r.visible = false
}

// IsVisible 是否可见
func (r ResumeSelector) IsVisible() bool {
	return r.visible
}

// SelectedEntry 返回当前选中的条目
func (r ResumeSelector) SelectedEntry() *ResumeEntry {
	if r.selected < len(r.entries) {
		return &r.entries[r.selected]
	}
	return nil
}

// Update 处理键盘事件
func (r ResumeSelector) Update(msg tea.Msg) (ResumeSelector, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if r.selected > 0 {
				r.selected--
			}
		case tea.KeyDown:
			if r.selected < len(r.entries)-1 {
				r.selected++
			}
		case tea.KeyEnter:
			r.visible = false
			// 选中后由外部处理
		case tea.KeyEsc:
			r.visible = false
			r.entries = nil // Esc 取消，清空条目
		}
	}
	return r, nil
}

// View 渲染选择器列表
func (r ResumeSelector) View() string {
	if !r.visible || len(r.entries) == 0 {
		return ""
	}

	bold := lipgloss.NewStyle().Bold(true).Foreground(ColorText)
	dim := StyleDim
	hl := lipgloss.NewStyle().Foreground(ColorBrand).Bold(true)

	var lines []string
	lines = append(lines, bold.Render("  恢复会话")+"  "+dim.Render("↑/↓ 选择  Enter 恢复  Esc 取消"))
	lines = append(lines, "")

	limit := min(10, len(r.entries))
	for i := 0; i < limit; i++ {
		e := r.entries[i]
		cursor := "  "
		style := dim
		if i == r.selected {
			cursor = hl.Render("❯ ")
			style = lipgloss.NewStyle().Foreground(ColorText)
		}
		// 截取 ID 前 8 位
		idShort := e.ID
		if len(idShort) > 8 {
			idShort = idShort[:8]
		}
		info := fmt.Sprintf("%s  %s  %d 轮  %s", idShort, e.Model, e.Turns, e.Cost)
		lines = append(lines, "  "+cursor+style.Render(info))
	}
	if len(r.entries) > limit {
		lines = append(lines, dim.Render(fmt.Sprintf("  ... 还有 %d 个会话", len(r.entries)-limit)))
	}

	return strings.Join(lines, "\n")
}
