// internal/provider/message.go
package provider

// Role 消息角色
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// BlockType 内容块类型
type BlockType int

const (
	BlockText BlockType = iota
	BlockThinking
	BlockImage
	BlockToolUse
	BlockToolResult
)

// ContentBlock 消息内容块
type ContentBlock struct {
	Type       BlockType
	Text       string
	Thinking   string
	ImageURL   string
	ToolCall   *ToolCall
	ToolResult *ToolResult
}

// ToolResult 工具执行结果
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// Message 统一消息格式
type Message struct {
	Role    Role
	Content []ContentBlock
}

// NewTextMessage 创建纯文本消息
func NewTextMessage(role Role, text string) Message {
	return Message{
		Role: role,
		Content: []ContentBlock{
			{Type: BlockText, Text: text},
		},
	}
}

// NewToolResultMessage 创建工具结果消息
func NewToolResultMessage(toolUseID, content string, isError bool) Message {
	return Message{
		Role: RoleUser,
		Content: []ContentBlock{
			{
				Type: BlockToolResult,
				ToolResult: &ToolResult{
					ToolUseID: toolUseID,
					Content:   content,
					IsError:   isError,
				},
			},
		},
	}
}

// TextContent 提取消息中的纯文本内容
func (m Message) TextContent() string {
	var text string
	for _, block := range m.Content {
		if block.Type == BlockText {
			text += block.Text
		}
	}
	return text
}

// ToolCalls 提取消息中的工具调用
func (m Message) ToolCalls() []*ToolCall {
	var calls []*ToolCall
	for _, block := range m.Content {
		if block.Type == BlockToolUse && block.ToolCall != nil {
			calls = append(calls, block.ToolCall)
		}
	}
	return calls
}
