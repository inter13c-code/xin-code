package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xincode-ai/xin-code/internal/tool"
)

// GlobTool 文件模式匹配工具
type GlobTool struct{}

type globInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

func (t *GlobTool) Name() string        { return "Glob" }
func (t *GlobTool) Description() string { return "按 glob 模式匹配文件路径。" }
func (t *GlobTool) IsReadOnly() bool    { return true }
func (t *GlobTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{"type": "string", "description": "Glob 模式，如 **/*.go"},
			"path":    map[string]any{"type": "string", "description": "搜索根目录，默认当前目录"},
		},
		"required": []string{"pattern"},
	}
}

func (t *GlobTool) Execute(_ context.Context, input json.RawMessage) (*tool.Result, error) {
	var in globInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	root := in.Path
	if root == "" {
		root = "."
	}

	pattern := filepath.Join(root, in.Pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return &tool.Result{Content: fmt.Sprintf("glob error: %s", err), IsError: true}, nil
	}

	sort.Strings(matches)
	if len(matches) == 0 {
		return &tool.Result{Content: "no matches found"}, nil
	}
	return &tool.Result{Content: strings.Join(matches, "\n")}, nil
}
