package builtin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReadTool(t *testing.T) {
	// 创建临时文件
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644)

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
	rt := &ReadTool{}
	input, _ := json.Marshal(readInput{Path: "/nonexistent/file"})
	result, err := rt.Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent file")
	}
}
