// internal/skills/skills.go
// 技能系统：发现和加载 SKILL.md 文件
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Skill 技能定义
type Skill struct {
	Name        string // 技能名称
	Description string // 描述（给 AI 判断是否调用）
	WhenToUse   string // 触发条件描述
	Content     string // 技能内容（Markdown）
	Path        string // SKILL.md 文件路径
	Source      string // 来源：user / project
}

// Registry 技能注册表
type Registry struct {
	skills map[string]*Skill
}

// NewRegistry 创建技能注册表
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]*Skill),
	}
}

// Discover 发现并加载技能
// 扫描 ~/.xincode/skills/ 和 .xincode/skills/ 目录
func (r *Registry) Discover(configDir string) {
	// 用户级技能
	userSkillsDir := filepath.Join(configDir, "skills")
	r.scanDir(userSkillsDir, "user")

	// 项目级技能
	r.scanDir(filepath.Join(".xincode", "skills"), "project")
}

// scanDir 扫描目录中的技能
func (r *Registry) scanDir(dir, source string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // 目录不存在则跳过
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(dir, entry.Name(), "SKILL.md")
		skill, err := loadSkill(skillPath, source)
		if err != nil {
			continue // 加载失败跳过
		}

		// 使用目录名作为技能名（如果 frontmatter 没指定）
		if skill.Name == "" {
			skill.Name = entry.Name()
		}
		r.skills[skill.Name] = skill
	}
}

// loadSkill 从 SKILL.md 加载技能
func loadSkill(path, source string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	skill := &Skill{
		Path:   path,
		Source: source,
	}

	// 解析 frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			parseFrontmatter(parts[1], skill)
			skill.Content = strings.TrimSpace(parts[2])
		} else {
			skill.Content = content
		}
	} else {
		skill.Content = content
	}

	return skill, nil
}

// parseFrontmatter 解析 YAML frontmatter（简易实现，不引入 yaml 库）
func parseFrontmatter(fm string, skill *Skill) {
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "name":
			skill.Name = val
		case "description":
			skill.Description = val
		case "whenToUse":
			skill.WhenToUse = val
		}
	}
}

// Get 获取技能
func (r *Registry) Get(name string) (*Skill, bool) {
	s, ok := r.skills[name]
	return s, ok
}

// All 返回所有技能
func (r *Registry) All() []*Skill {
	result := make([]*Skill, 0, len(r.skills))
	for _, s := range r.skills {
		result = append(result, s)
	}
	return result
}

// ListString 返回技能列表的格式化字符串
func (r *Registry) ListString() string {
	all := r.All()
	if len(all) == 0 {
		return "未发现任何技能\n\n技能目录:\n  ~/.xincode/skills/<name>/SKILL.md （用户级）\n  .xincode/skills/<name>/SKILL.md  （项目级）"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔧 已发现 %d 个技能\n\n", len(all)))
	for _, s := range all {
		sb.WriteString(fmt.Sprintf("  %-20s [%s] %s\n", s.Name, s.Source, s.Description))
	}
	return sb.String()
}
