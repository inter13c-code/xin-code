package tui

import (
	"strings"
	"testing"
)

func TestToolOutputFoldExpand(t *testing.T) {
	cv := NewChatView(80, 24)

	longOutput := strings.Repeat("line\n", 20)
	cv.messages = append(cv.messages, ChatMessage{
		ID:       "msg-1",
		Role:     "tool",
		ToolName: "Bash",
		Content:  longOutput,
	})

	// 默认：超阈值自动折叠
	cv.invalidateCache()
	cv.refreshContent(true)
	rendered := cv.viewport.View()
	if !strings.Contains(rendered, "+") || !strings.Contains(rendered, "行") {
		t.Error("超阈值输出应显示折叠提示 [+N 行]")
	}

	// 切换到展开模式
	cv.SetToolOutputExpanded(true)
	cv.invalidateCache()
	cv.refreshContent(true)
	rendered = cv.viewport.View()
	if strings.Contains(rendered, "[+") && strings.Contains(rendered, "行]") {
		t.Error("展开模式不应显示折叠提示")
	}

	// 切回折叠
	cv.SetToolOutputExpanded(false)
	cv.invalidateCache()
	cv.refreshContent(true)
	rendered = cv.viewport.View()
	if !strings.Contains(rendered, "+") {
		t.Error("折叠模式应恢复折叠提示")
	}
}

func TestToolOutputShortNoFold(t *testing.T) {
	cv := NewChatView(80, 24)

	cv.messages = append(cv.messages, ChatMessage{
		ID:       "msg-1",
		Role:     "tool",
		ToolName: "Read",
		Content:  "hello\nworld",
	})
	cv.invalidateCache()
	cv.refreshContent(true)
	rendered := cv.viewport.View()
	if strings.Contains(rendered, "[+") {
		t.Error("短输出不应折叠")
	}
}
