package agent

import "sync"

// SubAgentEntry 运行中的子 agent 信息
type SubAgentEntry struct {
	ID          string
	Name        string
	Description string
	InputCh     chan string // 接收外部消息
	Done        bool
	Result      string
}

// SubAgentRegistry 子 agent 注册表，追踪所有活跃的子 agent
type SubAgentRegistry struct {
	mu     sync.RWMutex
	agents map[string]*SubAgentEntry
}

// NewSubAgentRegistry 创建子 agent 注册表
func NewSubAgentRegistry() *SubAgentRegistry {
	return &SubAgentRegistry{
		agents: make(map[string]*SubAgentEntry),
	}
}

// Register 注册一个子 agent
func (r *SubAgentRegistry) Register(entry *SubAgentEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[entry.ID] = entry
}

// Get 根据 ID 或名称查找子 agent
func (r *SubAgentRegistry) Get(nameOrID string) *SubAgentEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 先按 ID 精确匹配
	if entry, ok := r.agents[nameOrID]; ok {
		return entry
	}

	// 再按 Name 模糊匹配
	for _, entry := range r.agents {
		if entry.Name == nameOrID {
			return entry
		}
	}
	return nil
}

// Remove 移除子 agent
func (r *SubAgentRegistry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, id)
}

// List 列出所有子 agent
func (r *SubAgentRegistry) List() []*SubAgentEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*SubAgentEntry, 0, len(r.agents))
	for _, entry := range r.agents {
		result = append(result, entry)
	}
	return result
}
