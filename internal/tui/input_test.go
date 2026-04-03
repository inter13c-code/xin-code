package tui

import (
	"strings"
	"testing"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		input, target string
		match         bool
	}{
		{"/co", "/commit", true},
		{"/co", "/compact", true},
		{"/cm", "/commit", true},
		{"/cm", "/help", false},
		{"/", "/help", true},
		{"/he", "/help", true},
		{"/hl", "/help", true},
		{"/xz", "/exit", false},
	}
	for _, tt := range tests {
		t.Run(tt.input+"→"+tt.target, func(t *testing.T) {
			got, _ := fuzzyMatchCommand(tt.input, tt.target)
			if got != tt.match {
				t.Errorf("fuzzyMatchCommand(%q, %q) = %v, want %v", tt.input, tt.target, got, tt.match)
			}
		})
	}
}

func TestFuzzyMatchScore(t *testing.T) {
	_, prefixScore := fuzzyMatchCommand("/co", "/cost")
	_, subseqScore := fuzzyMatchCommand("/ct", "/context")
	if prefixScore <= subseqScore {
		t.Errorf("prefix score (%d) should > subsequence score (%d)", prefixScore, subseqScore)
	}
}

func TestPasteReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantRef  bool
		wantText string
	}{
		{"短文本", "hello world", false, "hello world"},
		{"中等文本", strings.Repeat("x", 1025), true, "[Pasted text #1"},
		{"长文本含行数", "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\n" + strings.Repeat("line\n", 300), true, "+"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ib := NewInputBox(nil)
			ref, stored := ib.HandlePaste(tt.input)
			if tt.wantRef {
				if stored == "" {
					t.Error("应存储原文")
				}
				if !strings.Contains(ref, tt.wantText) {
					t.Errorf("引用文案应包含 %q，得到 %q", tt.wantText, ref)
				}
			} else {
				if ref != tt.input {
					t.Errorf("短文本应原样返回，得到 %q", ref)
				}
			}
		})
	}
}

func TestPasteReferenceExpansion(t *testing.T) {
	ib := NewInputBox(nil)
	// 模拟一次粘贴
	display, _ := ib.HandlePaste(strings.Repeat("code\n", 500))

	// 混入其他文本
	mixed := "请看这段代码 " + display + " 帮我修复"
	expanded := ib.expandPasteRefs(mixed)

	if strings.Contains(expanded, "[Pasted text") {
		t.Error("展开后不应包含占位符")
	}
	if !strings.Contains(expanded, "code\n") {
		t.Error("展开后应包含原始内容")
	}
}
