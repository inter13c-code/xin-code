package builtin

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestBashTool(t *testing.T) {
	bt := &BashTool{}
	input, _ := json.Marshal(bashInput{Command: "echo hello"})
	result, err := bt.Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Content, "hello") {
		t.Errorf("expected 'hello' in output, got: %s", result.Content)
	}
}

func TestBashToolError(t *testing.T) {
	bt := &BashTool{}
	input, _ := json.Marshal(bashInput{Command: "exit 1"})
	result, err := bt.Execute(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for exit 1")
	}
}
