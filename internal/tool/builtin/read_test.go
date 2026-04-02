package builtin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTool(t *testing.T) {
	// 创建临时文件
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644)

	// 切换到临时目录以通过路径校验
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	rt := &ReadTool{}
	input, _ := json.Marshal(readInput{Path: path})
	result, err := rt.Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if result.Content != "line1\nline2\nline3\n" {
		t.Errorf("unexpected content: %q", result.Content)
	}
}

func TestReadToolNotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	rt := &ReadTool{}
	// 路径在 cwd 内但文件不存在
	input, _ := json.Marshal(readInput{Path: filepath.Join(dir, "nonexistent")})
	result, err := rt.Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadToolPathTraversal(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	rt := &ReadTool{}

	tests := []struct {
		name string
		path string
	}{
		{"绝对路径 /etc/passwd", "/etc/passwd"},
		{"相对路径遍历 ../", filepath.Join(dir, "..", "something")},
		{"根目录文件", "/tmp/some-secret-file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, _ := json.Marshal(readInput{Path: tt.path})
			result, err := rt.Execute(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			if !result.IsError {
				t.Error("expected access denied error")
			}
			if !strings.Contains(result.Content, "access denied") {
				t.Errorf("expected 'access denied' in error, got: %s", result.Content)
			}
		})
	}
}

func TestReadToolAllowXinCodeConfig(t *testing.T) {
	// 测试允许读取 ~/.xincode/ 下的文件
	homeDir, _ := os.UserHomeDir()
	xincodeDir := filepath.Join(homeDir, ".xincode")

	// 创建临时配置文件用于测试
	os.MkdirAll(xincodeDir, 0755)
	testFile := filepath.Join(xincodeDir, "test-read-permission.tmp")
	os.WriteFile(testFile, []byte("config content"), 0644)
	defer os.Remove(testFile)

	// cwd 设置为其他目录
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	rt := &ReadTool{}
	input, _ := json.Marshal(readInput{Path: testFile})
	result, err := rt.Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("should allow reading ~/.xincode/ files, got error: %s", result.Content)
	}
	if result.Content != "config content" {
		t.Errorf("unexpected content: %q", result.Content)
	}
}

func TestReadToolFileSizeLimit(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// 创建一个超过 10MB 的文件
	bigFile := filepath.Join(dir, "big.txt")
	f, err := os.Create(bigFile)
	if err != nil {
		t.Fatal(err)
	}
	// 写入 11MB 的数据
	data := make([]byte, 11*1024*1024)
	f.Write(data)
	f.Close()

	rt := &ReadTool{}
	input, _ := json.Marshal(readInput{Path: bigFile})
	result, err := rt.Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for file exceeding 10MB limit")
	}
	if !strings.Contains(result.Content, "file too large") {
		t.Errorf("expected 'file too large' in error, got: %s", result.Content)
	}
}

func TestReadToolWithOffsetLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")
	os.WriteFile(path, []byte("a\nb\nc\nd\ne"), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	rt := &ReadTool{}
	input, _ := json.Marshal(readInput{Path: path, Offset: 1, Limit: 2})
	result, err := rt.Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	// 应该返回第 2-3 行（offset=1, limit=2），带行号
	expected := "2\tb\n3\tc"
	if result.Content != expected {
		t.Errorf("expected %q, got %q", expected, result.Content)
	}
}
