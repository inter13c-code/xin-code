# P0 TUI 体验对齐实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐 4 项 P0 交互差距，让 Xin Code 的日常使用体验对齐 Claude Code

**Architecture:** 在现有五层布局和消息驱动架构上增量添加，不重构。每个 Task 独立可测试可提交。

**Tech Stack:** Go 1.26 + Bubbletea v1.3.10 + charmbracelet/bubbles + glamour

**范围决定：**
- ✅ Task 1: 工具结果展开/收起 (Ctrl+O)
- ✅ Task 2: Bracket Paste 智能粘贴
- ✅ Task 3: Ctrl+R 历史搜索
- ✅ Task 4: 文件路径 Tab 补全
- ❌ 延后: Shift+Enter（Bubbletea v1.3.10 不支持 Kitty 键盘协议，无法区分 Shift+Enter 和 Enter）
- ❌ 延后: 虚拟滚动（需重构 ChatView 的 viewport 架构，单独立项）

---

## 文件变更地图

```
internal/tui/
├── chat.go          — Task 1: 折叠/展开逻辑（renderToolMessage + Ctrl+O 响应）
├── chat_test.go     — Task 1: 折叠行为测试（新建）
├── app.go           — Task 1: Ctrl+O 快捷键路由
├── input.go         — Task 2/3/4: 粘贴处理 + 历史搜索 + 路径补全
├── input_test.go    — Task 2/3/4: 测试（已有，追加）
├── msg.go           — Task 2: MsgPasteContent 消息类型
└── composer.go      — Task 1: 状态栏 Ctrl+O 提示
```

---

## Task 1: 工具结果展开/收起 (Ctrl+O)

**目标:** 用户按 Ctrl+O 全局 toggle 工具输出的折叠/展开状态

**Files:**
- Modify: `internal/tui/chat.go:17-18,47-85,270-280,704-765`
- Modify: `internal/tui/app.go:32-103,308-333`
- Modify: `internal/tui/composer.go`
- Create: `internal/tui/chat_test.go`

### Step 1: 写测试 — 折叠/展开行为

- [ ] **创建 `internal/tui/chat_test.go`**

```go
package tui

import (
	"strings"
	"testing"
)

func TestToolOutputFoldExpand(t *testing.T) {
	cv := NewChatView(80, 24)

	// 添加一条超过 foldThreshold 的工具输出
	longOutput := strings.Repeat("line\n", 20)
	cv.messages = append(cv.messages, ChatMessage{
		ID:       "msg-1",
		Role:     "tool",
		ToolName: "Bash",
		Content:  longOutput,
	})

	// 默认状态：自动折叠（briefMode=false，但超阈值自动折叠）
	cv.invalidateCache()
	cv.refreshContent(true)
	rendered := cv.viewport.View()
	if !strings.Contains(rendered, "+") || !strings.Contains(rendered, "行") {
		t.Error("超阈值输出应显示折叠提示 [+N 行]")
	}

	// 切换到展开模式
	cv.SetBriefMode(false) // briefMode=false + expanded=true → 全展开
	cv.SetToolOutputExpanded(true)
	cv.invalidateCache()
	cv.refreshContent(true)
	rendered = cv.viewport.View()
	// 展开后不应有折叠提示
	if strings.Contains(rendered, "[+") && strings.Contains(rendered, "行]") {
		t.Error("展开模式不应显示折叠提示")
	}

	// 切回折叠模式
	cv.SetToolOutputExpanded(false)
	cv.invalidateCache()
	cv.refreshContent(true)
	rendered = cv.viewport.View()
	if !strings.Contains(rendered, "+") {
		t.Error("折叠模式应恢复折叠提示")
	}
}

func TestToolOutputShortNoFold(t *testing.T) {
	cv := NewChatView(80, 24)

	// 少于阈值的输出永远不折叠
	cv.messages = append(cv.messages, ChatMessage{
		ID:       "msg-1",
		Role:     "tool",
		ToolName: "Read",
		Content:  "hello\nworld",
	})
	cv.invalidateCache()
	cv.refreshContent(true)
	rendered := cv.viewport.View()
	if strings.Contains(rendered, "[+") {
		t.Error("短输出不应折叠")
	}
}
```

- [ ] **运行测试确认失败**

Run: `go test ./internal/tui/ -run TestToolOutput -v`
Expected: FAIL — `SetBriefMode` 和 `SetToolOutputExpanded` 方法不存在

### Step 2: 实现 ChatView 的展开控制

- [ ] **在 `chat.go` 的 ChatView 结构体中添加字段（约第 85 行前）**

```go
	// 工具输出展开状态：true = 所有输出全展开，false = 超阈值自动折叠
	toolOutputExpanded bool
```

- [ ] **添加 setter 方法（在 NewChatView 之后）**

```go
// SetToolOutputExpanded 设置工具输出展开状态
func (c *ChatView) SetToolOutputExpanded(expanded bool) {
	c.toolOutputExpanded = expanded
}

// IsToolOutputExpanded 返回工具输出展开状态
func (c ChatView) IsToolOutputExpanded() bool {
	return c.toolOutputExpanded
}

// SetBriefMode 占位，兼容测试（当前无独立 brief 逻辑）
func (c *ChatView) SetBriefMode(brief bool) {}
```

### Step 3: 修改 renderToolMessage 使用展开状态

- [ ] **修改 `chat.go:747` 的折叠判断条件**

将：
```go
		if msg.Folded || lineCount > foldThreshold {
```

改为：
```go
		if !c.toolOutputExpanded && (msg.Folded || lineCount > foldThreshold) {
```

这样当 `toolOutputExpanded=true` 时，所有工具输出全展开。

### Step 4: 在 ChatView.Update 中响应 toggle 按键

- [ ] **在 `chat.go:270`（`t` 键 toggle thinking 的位置）之后添加**

```go
	// Ctrl+O toggle 工具输出展开/折叠
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyCtrlO {
		c.toolOutputExpanded = !c.toolOutputExpanded
		c.invalidateCache()
		c.refreshContent(false)
		return c, nil
	}
```

### Step 5: app.go 路由 Ctrl+O 到 chat

- [ ] **在 `app.go:333`（`case tea.KeyEsc` 块之后）添加 Ctrl+O 路由**

注意：不需要在 app.go 中添加额外处理——Ctrl+O 作为普通 KeyMsg 已经会被转发到 `chat.Update()`（通过 `shouldRouteKeyToChat`）。只需确认 `shouldRouteKeyToChat` 不拦截 Ctrl+O。

检查 `shouldRouteKeyToChat` 方法：如果它对 StateQuery/StateToolExec 也转发按键，Ctrl+O 在 AI 回答中也能用。

### Step 6: 状态栏添加 Ctrl+O 提示

- [ ] **在 `composer.go` 的右侧状态栏中添加 Ctrl+O 提示**

在渲染 `^Y` 鼠标模式提示的同一区域（右侧），添加：

```go
// 在已有的右侧元素之前（如 ctxStr 之前）
expandHint := StyleDim.Render("^O 展开")
if a.chat.IsToolOutputExpanded() {
	expandHint = StyleDim.Render("^O 折叠")
}
```

### Step 7: 运行测试确认通过

- [ ] Run: `go test ./internal/tui/ -run TestToolOutput -v`
Expected: PASS

### Step 8: 提交

- [ ] ```bash
git add internal/tui/chat.go internal/tui/chat_test.go internal/tui/app.go internal/tui/composer.go
git commit -m "feat: Ctrl+O 全局切换工具输出展开/折叠"
```

---

## Task 2: Bracket Paste 智能粘贴

**目标:** 大段粘贴文本自动引用化，避免输入框被长文本淹没

**原理:** Bubbletea v1.3.10 已内置 bracket paste 支持。粘贴内容到达时 `KeyMsg.Paste=true`，所有粘贴文字在 `KeyMsg.Runes` 中一次性到达。

**Files:**
- Modify: `internal/tui/input.go:20-55,71-214`
- Modify: `internal/tui/input_test.go`

### Step 1: 写测试 — 粘贴引用化

- [ ] **追加到 `internal/tui/input_test.go`**

```go
func TestPasteReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantRef  bool     // 是否生成引用
		wantText string   // 期望文案（子串匹配）
	}{
		{"短文本", "hello world", false, "hello world"},
		{"中等文本", strings.Repeat("x", 1025), true, "[Pasted text #1"},
		{"长文本含行数", "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\n" + strings.Repeat("line\n", 100), true, "+"},
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
```

- [ ] **运行测试确认失败**

Run: `go test ./internal/tui/ -run TestPaste -v`
Expected: FAIL — `HandlePaste` 方法不存在

### Step 2: 实现粘贴处理

- [ ] **在 `input.go` InputBox 结构体中添加字段（约第 30 行）**

```go
	// 粘贴内容存储
	pasteStore   map[int]string // pasteID → 原始内容
	nextPasteID  int
```

- [ ] **在 NewInputBox 中初始化**

```go
	pasteStore: make(map[int]string),
```

- [ ] **添加 HandlePaste 方法**

```go
const pasteRefThreshold = 1024 // 超过此长度生成引用

// HandlePaste 处理粘贴内容，返回 (显示文本, 存储的原文)
// 短文本原样返回；长文本生成引用并存储原文
func (i *InputBox) HandlePaste(content string) (display string, stored string) {
	if len(content) <= pasteRefThreshold {
		return content, ""
	}

	i.nextPasteID++
	id := i.nextPasteID
	i.pasteStore[id] = content

	lines := strings.Count(content, "\n")
	if lines == 0 {
		display = fmt.Sprintf("[Pasted text #%d]", id)
	} else {
		display = fmt.Sprintf("[Pasted text #%d +%d lines]", id, lines)
	}
	return display, content
}

// GetPastedContent 根据 ID 获取粘贴原文
func (i InputBox) GetPastedContent(id int) string {
	return i.pasteStore[id]
}
```

### Step 3: 在 Update 中拦截粘贴事件

- [ ] **在 `input.go:82`（KeyMsg switch 开头）添加粘贴检测**

在 `case tea.KeyMsg:` 块内、SGR 过滤之后、`switch msg.Type` 之前：

```go
		// 粘贴事件：Paste=true 时内容在 Runes 中一次性到达
		if msg.Paste {
			content := string(msg.Runes)
			display, _ := i.HandlePaste(content)
			i.textarea.InsertString(display)
			i.syncHeight()
			return i, nil
		}
```

### Step 4: 运行测试确认通过

- [ ] Run: `go test ./internal/tui/ -run TestPaste -v`
Expected: PASS

### Step 5: 全量测试

- [ ] Run: `go test ./... -v`
Expected: ALL PASS

### Step 6: 提交

- [ ] ```bash
git add internal/tui/input.go internal/tui/input_test.go
git commit -m "feat: bracket paste 智能粘贴 — 长文本自动引用化"
```

---

## Task 3: Ctrl+R 历史搜索

**目标:** Ctrl+R 进入历史搜索模式，输入关键词实时过滤历史记录

**Files:**
- Modify: `internal/tui/input.go:20-55,71-214,216-260`
- Modify: `internal/tui/input_test.go`

### Step 1: 写测试 — 历史搜索

- [ ] **追加到 `internal/tui/input_test.go`**

```go
func TestHistorySearch(t *testing.T) {
	ib := NewInputBox(nil)
	ib.history = []string{
		"/help",
		"写一个 HTTP 服务器",
		"/model claude-3-opus",
		"修复那个 bug",
		"/compact",
	}

	// 搜索 "model"
	results := ib.SearchHistory("model")
	if len(results) != 1 || results[0] != "/model claude-3-opus" {
		t.Errorf("搜索 'model' 应返回 1 条，得到 %v", results)
	}

	// 搜索 "/"
	results = ib.SearchHistory("/")
	if len(results) != 3 {
		t.Errorf("搜索 '/' 应返回 3 条，得到 %d", len(results))
	}

	// 空搜索返回全部（倒序）
	results = ib.SearchHistory("")
	if len(results) != 5 {
		t.Errorf("空搜索应返回 5 条，得到 %d", len(results))
	}

	// 最近的在前
	if results[0] != "/compact" {
		t.Errorf("空搜索第一条应是最近的 '/compact'，得到 %q", results[0])
	}
}
```

- [ ] **运行测试确认失败**

Run: `go test ./internal/tui/ -run TestHistorySearch -v`
Expected: FAIL — `SearchHistory` 方法不存在

### Step 2: 添加搜索状态字段

- [ ] **在 InputBox 结构体中添加（约第 30 行）**

```go
	// 历史搜索模式
	searchMode    bool
	searchQuery   string
	searchResults []string // 匹配的历史记录（最近优先）
	searchIdx     int      // 当前选中的搜索结果 (-1 = 无)
```

### Step 3: 实现搜索逻辑

- [ ] **添加 SearchHistory 方法**

```go
// SearchHistory 搜索历史记录，返回匹配项（最近优先）
func (i InputBox) SearchHistory(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	var results []string
	// 从最近到最早遍历
	for j := len(i.history) - 1; j >= 0; j-- {
		if query == "" || strings.Contains(strings.ToLower(i.history[j]), query) {
			results = append(results, i.history[j])
		}
	}
	return results
}
```

### Step 4: 添加搜索模式按键处理

- [ ] **在 `input.go` Update 方法的 `case tea.KeyMsg:` 块中，`switch msg.Type` 之前添加搜索模式处理**

```go
		// 搜索模式下的按键处理
		if i.searchMode {
			switch msg.Type {
			case tea.KeyEsc:
				i.searchMode = false
				i.searchQuery = ""
				i.searchResults = nil
				i.searchIdx = -1
				return i, nil
			case tea.KeyEnter:
				// 选中搜索结果
				if i.searchIdx >= 0 && i.searchIdx < len(i.searchResults) {
					i.textarea.SetValue(i.searchResults[i.searchIdx])
				}
				i.searchMode = false
				i.searchQuery = ""
				i.searchResults = nil
				i.searchIdx = -1
				return i, nil
			case tea.KeyUp:
				if i.searchIdx > 0 {
					i.searchIdx--
				}
				return i, nil
			case tea.KeyDown:
				if i.searchIdx < len(i.searchResults)-1 {
					i.searchIdx++
				}
				return i, nil
			case tea.KeyBackspace:
				if len(i.searchQuery) > 0 {
					i.searchQuery = i.searchQuery[:len(i.searchQuery)-1]
					i.searchResults = i.SearchHistory(i.searchQuery)
					i.searchIdx = 0
				}
				return i, nil
			case tea.KeyRunes:
				i.searchQuery += string(msg.Runes)
				i.searchResults = i.SearchHistory(i.searchQuery)
				i.searchIdx = 0
				return i, nil
			}
			return i, nil
		}
```

- [ ] **在 `switch msg.Type` 中添加 Ctrl+R 触发（在 `case tea.KeyEsc` 之后）**

注意：Bubbletea 中 Ctrl+R 是 `tea.KeyCtrlR`。

```go
		case tea.KeyCtrlR:
			if len(i.history) > 0 {
				i.searchMode = true
				i.searchQuery = ""
				i.searchResults = i.SearchHistory("")
				i.searchIdx = 0
			}
			return i, nil
```

### Step 5: 渲染搜索面板

- [ ] **在 `input.go` 的 View 方法（约第 216 行）中，在 textarea 上方添加搜索面板**

```go
func (i InputBox) View() string {
	var sections []string

	// 搜索模式面板（优先于补全列表）
	if i.searchMode {
		sections = append(sections, i.renderSearchPanel())
	} else if hint := i.renderSlashHint(); hint != "" {
		sections = append(sections, hint)
	}

	sections = append(sections, i.textarea.View())
	return strings.Join(sections, "\n")
}
```

- [ ] **添加 renderSearchPanel 方法**

```go
// renderSearchPanel 渲染历史搜索面板
func (i InputBox) renderSearchPanel() string {
	width := min(72, max(34, i.width-2))
	prompt := StyleBrand.Render("搜索历史") + StyleDim.Render(": "+i.searchQuery+"█")

	var lines []string
	lines = append(lines, prompt)

	// 最多显示 8 条搜索结果
	maxShow := 8
	start := 0
	if i.searchIdx >= maxShow {
		start = i.searchIdx - maxShow + 1
	}
	end := start + maxShow
	if end > len(i.searchResults) {
		end = len(i.searchResults)
	}

	if len(i.searchResults) == 0 && i.searchQuery != "" {
		lines = append(lines, StyleDim.Render("  无匹配"))
	}

	for idx := start; idx < end; idx++ {
		entry := i.searchResults[idx]
		if len(entry) > width-4 {
			entry = entry[:width-7] + "..."
		}
		if idx == i.searchIdx {
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorBrand).Bold(true).Render("❯ "+entry))
		} else {
			lines = append(lines, StyleDim.Render("  "+entry))
		}
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBrand).
		Padding(0, 1).
		Width(width).
		Render(strings.Join(lines, "\n"))
}
```

### Step 6: 运行测试确认通过

- [ ] Run: `go test ./internal/tui/ -run TestHistory -v`
Expected: PASS

### Step 7: 全量测试

- [ ] Run: `go test ./...`
Expected: ALL PASS

### Step 8: 提交

- [ ] ```bash
git add internal/tui/input.go internal/tui/input_test.go
git commit -m "feat: Ctrl+R 历史搜索 — 实时过滤 + ↑↓ 选择"
```

---

## Task 4: 文件路径 Tab 补全

**目标:** 输入非 slash 命令时，对路径字符串按 Tab 触发文件系统补全

**Files:**
- Modify: `internal/tui/input.go:20-55,71-214,216-260`
- Modify: `internal/tui/input_test.go`

### Step 1: 写测试 — 路径检测

- [ ] **追加到 `internal/tui/input_test.go`**

```go
func TestExtractPathPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"hello", ""},
		{"./src/", "./src/"},
		{"read ../config", "../config"},
		{"/usr/local/bin/", "/usr/local/bin/"},
		{"看看 ~/Desktop/", "~/Desktop/"},
		{"internal/tui/app", "internal/tui/app"},
		{"/help", ""}, // slash 命令不算
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractPathPrefix(tt.input)
			if got != tt.want {
				t.Errorf("extractPathPrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
```

- [ ] **运行测试确认失败**

Run: `go test ./internal/tui/ -run TestExtractPath -v`
Expected: FAIL — `extractPathPrefix` 不存在

### Step 2: 实现路径提取

- [ ] **添加 `extractPathPrefix` 函数**

```go
// extractPathPrefix 从输入文本中提取最后一个路径前缀
// 路径特征：以 ./ ../ / ~/ 开头，或包含 / 的连续非空字符
func extractPathPrefix(input string) string {
	if strings.HasPrefix(input, "/") && !strings.Contains(input, " ") {
		// 可能是 slash 命令
		return ""
	}

	// 找到最后一个空格后的 token
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]

	// 判断是否像路径
	if strings.HasPrefix(last, "./") ||
		strings.HasPrefix(last, "../") ||
		strings.HasPrefix(last, "/") ||
		strings.HasPrefix(last, "~/") ||
		strings.Contains(last, "/") {
		return last
	}
	return ""
}
```

### Step 3: 添加路径补全状态

- [ ] **在 InputBox 结构体中添加字段**

```go
	// 路径补全
	pathCompletions []string // 当前路径匹配结果
	pathCompIdx     int      // 当前选中的路径补全 (-1 = 无)
```

### Step 4: 实现路径补全逻辑

- [ ] **添加 `completeFilePath` 方法**

```go
// completeFilePath 对输入中的路径前缀进行文件系统补全
func (i *InputBox) completeFilePath() bool {
	text := i.textarea.Value()
	prefix := extractPathPrefix(text)
	if prefix == "" {
		return false
	}

	// 展开 ~
	expandedPrefix := prefix
	if strings.HasPrefix(prefix, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			expandedPrefix = home + prefix[1:]
		}
	}

	// glob 匹配
	pattern := expandedPrefix + "*"
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return false
	}

	// 排序
	sort.Strings(matches)

	// 转回用户路径格式（如果原始用 ~/ 开头）
	var completions []string
	for _, m := range matches {
		if strings.HasPrefix(prefix, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				m = "~/" + strings.TrimPrefix(m, home+"/")
			}
		}
		// 目录加 /
		if info, err := os.Stat(m); err == nil && info.IsDir() {
			m += "/"
		}
		completions = append(completions, m)
	}

	if len(completions) == 1 {
		// 唯一匹配：直接替换
		newText := strings.TrimSuffix(text, prefix) + completions[0]
		i.textarea.SetValue(newText)
		return true
	}

	// 多个匹配：存储到补全列表，用 ↑↓ 选择
	i.pathCompletions = completions
	i.pathCompIdx = 0
	return true
}
```

### Step 5: 修改 Tab 按键处理

- [ ] **修改 `input.go:96-107` 的 Tab 处理**

```go
		case tea.KeyTab:
			// 优先：slash 命令补全
			if len(i.completionItems) > 0 {
				idx := i.completionIdx
				if idx < 0 {
					idx = 0
				}
				i.textarea.SetValue(i.completionItems[idx].Name + " ")
				i.completionIdx = -1
				i.completionItems = nil
				return i, nil
			}
			// 路径补全列表导航
			if len(i.pathCompletions) > 0 {
				if i.pathCompIdx >= 0 && i.pathCompIdx < len(i.pathCompletions) {
					text := i.textarea.Value()
					prefix := extractPathPrefix(text)
					newText := strings.TrimSuffix(text, prefix) + i.pathCompletions[i.pathCompIdx]
					i.textarea.SetValue(newText)
					i.pathCompIdx++
					if i.pathCompIdx >= len(i.pathCompletions) {
						i.pathCompIdx = 0
					}
				}
				return i, nil
			}
			// 尝试路径补全
			if i.completeFilePath() {
				return i, nil
			}
```

### Step 6: 渲染路径补全列表

- [ ] **在 View 方法中添加路径补全列表渲染**

在 `renderSlashHint` 调用的同一位置，当无 slash 补全但有路径补全时渲染：

```go
	// 路径补全列表
	if !i.searchMode && len(i.pathCompletions) > 0 && len(i.completionItems) == 0 {
		sections = append(sections, i.renderPathHint())
	}
```

- [ ] **添加 `renderPathHint` 方法**

```go
func (i InputBox) renderPathHint() string {
	width := min(72, max(34, i.width-2))
	var lines []string
	maxShow := 8
	start := 0
	if i.pathCompIdx >= maxShow {
		start = i.pathCompIdx - maxShow + 1
	}
	end := start + maxShow
	if end > len(i.pathCompletions) {
		end = len(i.pathCompletions)
	}

	for idx := start; idx < end; idx++ {
		entry := i.pathCompletions[idx]
		if len(entry) > width-4 {
			entry = "..." + entry[len(entry)-width+7:]
		}
		if idx == i.pathCompIdx {
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorBrand).Bold(true).Render("❯ "+entry))
		} else {
			lines = append(lines, StyleDim.Render("  "+entry))
		}
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorTextDim).
		Padding(0, 1).
		Width(width).
		Render(strings.Join(lines, "\n"))
}
```

### Step 7: 清除路径补全状态

- [ ] **在 Update 方法末尾（`matchSlashCommands` 之后）添加路径补全状态管理**

```go
	// 非 Tab 键时清除路径补全状态
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type != tea.KeyTab {
		i.pathCompletions = nil
		i.pathCompIdx = -1
	}
```

### Step 8: 添加 import

- [ ] **在 `input.go` 的 import 中添加**

```go
	"os"
	"path/filepath"
```

### Step 9: 运行测试

- [ ] Run: `go test ./internal/tui/ -run TestExtractPath -v`
Expected: PASS

- [ ] Run: `go test ./...`
Expected: ALL PASS

### Step 10: 提交

- [ ] ```bash
git add internal/tui/input.go internal/tui/input_test.go
git commit -m "feat: Tab 文件路径补全 — glob 匹配 + 循环选择"
```

---

## 延后项说明

### Shift+Enter 多行输入
Bubbletea v1.3.10 不支持 Kitty 键盘协议（`CSI u`），无法区分 Shift+Enter 和普通 Enter。需要等 Bubbletea 升级支持增强键盘模式，或自行在 `input driver` 层做 CSI u 解析。当前 Alt+Enter 方案跨终端兼容。

### 虚拟滚动
当前 ChatView 用 `viewport.SetContent(全量字符串)` 驱动滚动。改为虚拟滚动需要：
1. 消息级高度估算和缓存
2. 二分查找可见范围
3. spacer 占位（上/下空行）
4. 滚动时动态重建内容字符串
5. sticky scroll 粘性判断

建议单独立项，参考 CC 的 `useVirtualScroll.ts` 设计。初期可先做简化版：限制 `committedRendered` 只包含最近 N 条消息（如 200 条），避免极长会话性能问题。
