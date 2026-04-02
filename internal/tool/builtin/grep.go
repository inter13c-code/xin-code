package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/xincode-ai/xin-code/internal/tool"
)

// GrepTool 文件内容搜索工具
type GrepTool struct{}

type grepInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
	Glob    string `json:"glob,omitempty"`
}

func (t *GrepTool) Name() string        { return "Grep" }
func (t *GrepTool) Description() string { return "搜索文件内容，支持正则表达式。底层使用 grep -rn。" }
func (t *GrepTool) IsReadOnly() bool    { return true }
func (t *GrepTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{"type": "string", "description": "搜索模式（正则表达式）"},
			"path":    map[string]any{"type": "string", "description": "搜索路径，默认当前目录"},
			"glob":    map[string]any{"type": "string", "description": "文件过滤，如 *.go"},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepTool) Execute(ctx context.Context, input json.RawMessage) (*tool.Result, error) {
	var in grepInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	path := in.Path
	if path == "" {
		path = "."
	}

	// 优先用 ripgrep，回退到 grep
	args := []string{"-rn", "--color=never"}
	bin := "grep"
	if _, err := exec.LookPath("rg"); err == nil {
		bin = "rg"
		args = []string{"-n", "--no-heading", "--color=never"}
		if in.Glob != "" {
			args = append(args, "--glob", in.Glob)
		}
	} else if in.Glob != "" {
		args = append(args, "--include="+in.Glob)
	}
	args = append(args, in.Pattern, path)

	cmd := exec.CommandContext(ctx, bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	// grep 退出码 1 表示没有匹配，不是错误
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return &tool.Result{Content: "no matches found"}, nil
		}
		return &tool.Result{
			Content: fmt.Sprintf("grep error: %s\n%s", err, stderr.String()),
			IsError: true,
		}, nil
	}

	if output == "" {
		return &tool.Result{Content: "no matches found"}, nil
	}

	// 限制输出大小
	const maxOutput = 50 * 1024
	if len(output) > maxOutput {
		output = output[:maxOutput] + "\n... (output truncated)"
	}

	return &tool.Result{Content: output}, nil
}
