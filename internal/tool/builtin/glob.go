package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
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

	pattern := in.Pattern

	// 如果包含 **，使用 WalkDir 实现递归匹配
	if strings.Contains(pattern, "**") {
		matches, err := globRecursive(root, pattern)
		if err != nil {
			return &tool.Result{Content: fmt.Sprintf("glob error: %s", err), IsError: true}, nil
		}
		sort.Strings(matches)
		if len(matches) == 0 {
			return &tool.Result{Content: "no matches found"}, nil
		}
		return &tool.Result{Content: strings.Join(matches, "\n")}, nil
	}

	// 不含 **，使用标准 filepath.Glob
	fullPattern := filepath.Join(root, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return &tool.Result{Content: fmt.Sprintf("glob error: %s", err), IsError: true}, nil
	}

	sort.Strings(matches)
	if len(matches) == 0 {
		return &tool.Result{Content: "no matches found"}, nil
	}
	return &tool.Result{Content: strings.Join(matches, "\n")}, nil
}

// globRecursive 使用 WalkDir 实现 ** 递归匹配
// 支持模式如 **/*.go、src/**/*.ts、**/*_test.go
func globRecursive(root, pattern string) ([]string, error) {
	// 将 ** 模式拆成各段
	// 例如 "**/*.go" -> 匹配任意深度下以 .go 结尾的文件
	// 例如 "src/**/*.ts" -> src 目录下任意深度的 .ts 文件
	parts := strings.Split(filepath.ToSlash(pattern), "/")

	var matches []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 跳过无权限的目录
		}
		// 跳过隐藏目录（.git 等）
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") && path != root {
			return fs.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		// 获取相对于 root 的路径
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		if matchDoubleStarPattern(filepath.ToSlash(relPath), parts) {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil && !os.IsPermission(err) {
		return nil, err
	}
	return matches, nil
}

// matchDoubleStarPattern 匹配包含 ** 的模式
func matchDoubleStarPattern(path string, patternParts []string) bool {
	pathParts := strings.Split(path, "/")
	return matchParts(pathParts, patternParts)
}

func matchParts(pathParts, patternParts []string) bool {
	if len(patternParts) == 0 {
		return len(pathParts) == 0
	}

	if patternParts[0] == "**" {
		rest := patternParts[1:]
		// ** 可以匹配零个或多个目录层级
		for i := 0; i <= len(pathParts); i++ {
			if matchParts(pathParts[i:], rest) {
				return true
			}
		}
		return false
	}

	if len(pathParts) == 0 {
		return false
	}

	matched, err := filepath.Match(patternParts[0], pathParts[0])
	if err != nil || !matched {
		return false
	}
	return matchParts(pathParts[1:], patternParts[1:])
}
