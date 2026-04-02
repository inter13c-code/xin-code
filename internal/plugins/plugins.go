// internal/plugins/plugins.go
// 插件系统：发现和加载 plugin.json
package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Plugin 插件定义
type Plugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Path        string `json:"-"` // 插件目录路径
	Source      string `json:"-"` // 来源：user / project
	Loaded      bool   `json:"-"` // 是否成功加载
	Error       string `json:"-"` // 加载错误信息
}

// Registry 插件注册表
type Registry struct {
	plugins map[string]*Plugin
}

// NewRegistry 创建插件注册表
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*Plugin),
	}
}

// Discover 发现并加载插件
// 扫描 ~/.xincode/plugins/ 和 .xincode/plugins/ 目录
func (r *Registry) Discover(configDir string) {
	// 用户级插件
	userPluginsDir := filepath.Join(configDir, "plugins")
	r.scanDir(userPluginsDir, "user")

	// 项目级插件
	r.scanDir(filepath.Join(".xincode", "plugins"), "project")
}

// scanDir 扫描目录中的插件
func (r *Registry) scanDir(dir, source string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // 目录不存在则跳过
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(dir, entry.Name())
		jsonPath := filepath.Join(pluginPath, "plugin.json")

		plugin, err := loadPlugin(jsonPath)
		if err != nil {
			// 加载失败跳过并记录警告
			r.plugins[entry.Name()] = &Plugin{
				Name:   entry.Name(),
				Path:   pluginPath,
				Source: source,
				Loaded: false,
				Error:  err.Error(),
			}
			continue
		}

		plugin.Path = pluginPath
		plugin.Source = source
		plugin.Loaded = true

		// 使用目录名作为 key（如果 plugin.json 没指定 name）
		if plugin.Name == "" {
			plugin.Name = entry.Name()
		}
		r.plugins[plugin.Name] = plugin
	}
}

// loadPlugin 从 plugin.json 加载插件元数据
func loadPlugin(path string) (*Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 plugin.json 失败: %w", err)
	}

	var plugin Plugin
	if err := json.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("解析 plugin.json 失败: %w", err)
	}

	return &plugin, nil
}

// Get 获取插件
func (r *Registry) Get(name string) (*Plugin, bool) {
	p, ok := r.plugins[name]
	return p, ok
}

// All 返回所有插件
func (r *Registry) All() []*Plugin {
	result := make([]*Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		result = append(result, p)
	}
	return result
}

// ListString 返回插件列表的格式化字符串
func (r *Registry) ListString() string {
	all := r.All()
	if len(all) == 0 {
		return "未发现任何插件\n\n插件目录:\n  ~/.xincode/plugins/<name>/plugin.json （用户级）\n  .xincode/plugins/<name>/plugin.json  （项目级）"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔌 已发现 %d 个插件\n\n", len(all)))
	for _, p := range all {
		status := "✓"
		if !p.Loaded {
			status = "✗"
		}
		sb.WriteString(fmt.Sprintf("  %s %-20s [%s]", status, p.Name, p.Source))
		if p.Version != "" {
			sb.WriteString(fmt.Sprintf(" v%s", p.Version))
		}
		if p.Description != "" {
			sb.WriteString(fmt.Sprintf(" - %s", p.Description))
		}
		if !p.Loaded {
			sb.WriteString(fmt.Sprintf(" (错误: %s)", p.Error))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
