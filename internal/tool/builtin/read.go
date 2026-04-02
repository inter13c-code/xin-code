package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xincode-ai/xin-code/internal/tool"
)

// ReadTool 文件读取工具
type ReadTool struct{}

type readInput struct {
	Path   string `json:"path"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func (t *ReadTool) Name() string        { return "Read" }
func (t *ReadTool) Description() string { return "读取文件内容。支持通过 offset 和 limit 读取部分内容。" }
func (t *ReadTool) IsReadOnly() bool    { return true }
func (t *ReadTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":   map[string]any{"type": "string", "description": "文件的绝对路径"},
			"offset": map[string]any{"type": "integer", "description": "起始行号（从 0 开始）"},
			"limit":  map[string]any{"type": "integer", "description": "读取的行数"},
		},
		"required": []string{"path"},
	}
}

func (t *ReadTool) Execute(_ context.Context, input json.RawMessage) (*tool.Result, error) {
	var in readInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	data, err := os.ReadFile(in.Path)
	if err != nil {
		return &tool.Result{Content: fmt.Sprintf("error reading file: %s", err), IsError: true}, nil
	}
	content := string(data)

	// 按行切分并应用 offset/limit
	if in.Offset > 0 || in.Limit > 0 {
		lines := strings.Split(content, "\n")
		start := in.Offset
		if start >= len(lines) {
			return &tool.Result{Content: ""}, nil
		}
		end := len(lines)
		if in.Limit > 0 && start+in.Limit < end {
			end = start + in.Limit
		}
		// 添加行号
		var numbered []string
		for i := start; i < end; i++ {
			numbered = append(numbered, fmt.Sprintf("%d\t%s", i+1, lines[i]))
		}
		content = strings.Join(numbered, "\n")
	}

	return &tool.Result{Content: content}, nil
}
