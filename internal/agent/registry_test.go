package agent

import (
	"testing"
)

func TestSubAgentRegistry_RegisterAndGet(t *testing.T) {
	reg := NewSubAgentRegistry()

	entry := &SubAgentEntry{
		ID:          "sa-001",
		Name:        "test-agent",
		Description: "测试子 agent",
		InputCh:     make(chan string, 1),
	}
	reg.Register(entry)

	// 按 ID 查找
	got := reg.Get("sa-001")
	if got == nil {
		t.Fatal("按 ID 查找返回 nil")
	}
	if got.Name != "test-agent" {
		t.Errorf("Name = %q, want %q", got.Name, "test-agent")
	}

	// 按 Name 查找
	got = reg.Get("test-agent")
	if got == nil {
		t.Fatal("按 Name 查找返回 nil")
	}
	if got.ID != "sa-001" {
		t.Errorf("ID = %q, want %q", got.ID, "sa-001")
	}

	// 查找不存在的
	got = reg.Get("not-exist")
	if got != nil {
		t.Error("查找不存在的 agent 应返回 nil")
	}
}

func TestSubAgentRegistry_Remove(t *testing.T) {
	reg := NewSubAgentRegistry()

	entry := &SubAgentEntry{
		ID:   "sa-002",
		Name: "to-remove",
	}
	reg.Register(entry)

	reg.Remove("sa-002")
	got := reg.Get("sa-002")
	if got != nil {
		t.Error("移除后应返回 nil")
	}
}

func TestSubAgentRegistry_List(t *testing.T) {
	reg := NewSubAgentRegistry()

	// 空列表
	list := reg.List()
	if len(list) != 0 {
		t.Errorf("空注册表 List() 长度 = %d, want 0", len(list))
	}

	// 添加多个
	reg.Register(&SubAgentEntry{ID: "sa-a", Name: "agent-a"})
	reg.Register(&SubAgentEntry{ID: "sa-b", Name: "agent-b"})
	reg.Register(&SubAgentEntry{ID: "sa-c", Name: "agent-c"})

	list = reg.List()
	if len(list) != 3 {
		t.Errorf("List() 长度 = %d, want 3", len(list))
	}

	// 验证所有 ID 都在
	ids := make(map[string]bool)
	for _, e := range list {
		ids[e.ID] = true
	}
	for _, expected := range []string{"sa-a", "sa-b", "sa-c"} {
		if !ids[expected] {
			t.Errorf("List() 缺少 ID %q", expected)
		}
	}
}

func TestSubAgentRegistry_GetByNamePreference(t *testing.T) {
	reg := NewSubAgentRegistry()

	// ID 和另一个 entry 的 Name 相同时，ID 优先
	reg.Register(&SubAgentEntry{ID: "name-conflict", Name: "first"})
	reg.Register(&SubAgentEntry{ID: "other-id", Name: "name-conflict"})

	got := reg.Get("name-conflict")
	if got == nil {
		t.Fatal("查找返回 nil")
	}
	// ID 精确匹配优先
	if got.ID != "name-conflict" {
		t.Errorf("ID = %q, want %q（ID 优先于 Name）", got.ID, "name-conflict")
	}
}
