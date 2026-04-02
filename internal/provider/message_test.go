package provider

import "testing"

func TestNewTextMessage(t *testing.T) {
	msg := NewTextMessage(RoleUser, "hello")
	if msg.Role != RoleUser {
		t.Errorf("expected RoleUser, got %v", msg.Role)
	}
	if msg.TextContent() != "hello" {
		t.Errorf("expected 'hello', got '%s'", msg.TextContent())
	}
}

func TestNewToolResultMessage(t *testing.T) {
	msg := NewToolResultMessage("tool-1", "result text", false)
	if msg.Role != RoleUser {
		t.Errorf("expected RoleUser, got %v", msg.Role)
	}
	if len(msg.Content) != 1 || msg.Content[0].Type != BlockToolResult {
		t.Fatal("expected one ToolResult block")
	}
	if msg.Content[0].ToolResult.ToolUseID != "tool-1" {
		t.Errorf("expected tool-1, got %s", msg.Content[0].ToolResult.ToolUseID)
	}
}

func TestToolCalls(t *testing.T) {
	msg := Message{
		Role: RoleAssistant,
		Content: []ContentBlock{
			{Type: BlockText, Text: "Let me read that file"},
			{Type: BlockToolUse, ToolCall: &ToolCall{ID: "tc-1", Name: "Read", Input: `{"path":"main.go"}`}},
		},
	}
	calls := msg.ToolCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	if calls[0].Name != "Read" {
		t.Errorf("expected Read, got %s", calls[0].Name)
	}
}
