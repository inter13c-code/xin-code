package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/xincode-ai/xin-code/internal/tool"
)

// BashTool Shell 命令执行工具
type BashTool struct{}

type bashInput struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"` // 毫秒
}

func (t *BashTool) Name() string        { return "Bash" }
func (t *BashTool) Description() string { return "执行 shell 命令并返回输出。" }
func (t *BashTool) IsReadOnly() bool    { return false }
func (t *BashTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{"type": "string", "description": "要执行的 shell 命令"},
			"timeout": map[string]any{"type": "integer", "description": "超时时间（毫秒），默认 120000"},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) Execute(ctx context.Context, input json.RawMessage) (*tool.Result, error) {
	var in bashInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	timeout := 120 * time.Second
	if in.Timeout > 0 {
		timeout = time.Duration(in.Timeout) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", in.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var result strings.Builder
	if stdout.Len() > 0 {
		result.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString("STDERR:\n")
		result.WriteString(stderr.String())
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &tool.Result{Content: "command timed out", IsError: true}, nil
		}
		exitErr := ""
		if result.Len() > 0 {
			exitErr = result.String() + "\n"
		}
		return &tool.Result{
			Content: fmt.Sprintf("%sexit status: %s", exitErr, err),
			IsError: true,
		}, nil
	}

	output := result.String()
	if output == "" {
		output = "(no output)"
	}

	// 限制输出大小，和 Grep 工具保持一致
	const maxOutput = 50 * 1024
	if len(output) > maxOutput {
		output = output[:maxOutput] + "\n... (output truncated)"
	}

	return &tool.Result{Content: output}, nil
}
