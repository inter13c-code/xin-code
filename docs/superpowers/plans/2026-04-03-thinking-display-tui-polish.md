# Extended Thinking + TUI Polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让用户看到 AI 推理过程（折叠/展开）+ 全面提升 TUI 交互细节

**Architecture:** 5 个独立任务，按依赖顺序排列。Task 1-2 改 thinking 数据流和渲染。Task 3 重做 slash 补全。Task 4-5 打磨消息显示和权限卡片。每个 task 可独立编译通过。

**Tech Stack:** Go 1.26, Bubbletea v1.3.10, Lipgloss, Glamour

---

### Task 1: Thinking 持久化 — processStream 收集 thinking blocks

**Files:**
- Modify: `agent.go:292-344`
- Modify: `main.go:467-501` (rebuildTranscript)
- Test: `internal/session/session_test.go` (新增 round-trip 验证)

- [ ] **Step 1: 修改 processStream 收集 thinking**

在 `agent.go` 的 `processStream` 函数中增加 thinking 累积：

```go
// agent.go processStream — 在 var 声明区增加
var thinkingContent string

// EventThinking case 改为：
case provider.EventThinking:
    if evt.Thinking != nil {
        a.send(tui.MsgThinking{Text: evt.Thinking.Text})
        thinkingContent += evt.Thinking.Text
    }

// 流结束后构建 blocks 时，在 textContent block 之前插入 thinking block：
// 替换当前 "构建 assistant 消息" 部分（agent.go:331-336）
var finalBlocks []provider.ContentBlock
if thinkingContent != "" {
    finalBlocks = append(finalBlocks, provider.ContentBlock{
        Type:     provider.BlockThinking,
        Thinking: thinkingContent,
    })
}
if textContent != "" {
    finalBlocks = append(finalBlocks, provider.ContentBlock{
        Type: provider.BlockText,
        Text: textContent,
    })
}
finalBlocks = append(finalBlocks, blocks...)

msg := provider.Message{
    Role:    provider.RoleAssistant,
    Content: finalBlocks,
}
```

- [ ] **Step 2: 修改 rebuildTranscript 映射 thinking blocks**

在 `main.go` 的 `rebuildTranscript` 函数中，`case provider.RoleAssistant:` 分支内，在处理 text 之前增加 thinking block 处理：

```go
// main.go rebuildTranscript, case provider.RoleAssistant: 内，text 处理之前
for _, block := range msg.Content {
    if block.Type == provider.BlockThinking && block.Thinking != "" {
        chatMsgs = append(chatMsgs, tui.ChatMessage{
            Role:    "thinking",
            Content: block.Thinking,
            Folded:  true,
        })
    }
}
// 然后继续现有的 text 和 toolCalls 处理
```

- [ ] **Step 3: 编译验证**

Run: `go build ./...`
Expected: 编译通过

- [ ] **Step 4: 提交**

```bash
git add agent.go main.go
git commit -m "feat: thinking 内容持久化到 session — processStream 收集 BlockThinking + rebuildTranscript 映射"
```

---

### Task 2: Thinking 折叠/展开渲染 + `t` 键 toggle

**Files:**
- Modify: `internal/tui/chat.go:497-503` (renderMessage thinking case)
- Modify: `internal/tui/chat.go:113-246` (Update — 增加 `t` 键处理)
- Modify: `internal/tui/app.go:830-837` (shouldRouteKeyToChat 增加 `t`)

- [ ] **Step 1: 改 thinking 渲染 — 支持折叠/展开**

替换 `chat.go` 的 `renderMessage` 中 `case "thinking":` 分支（当前行 497-503）：

```go
case "thinking":
    symStyle := StyleThinking
    if c.toolBlink {
        symStyle = lipgloss.NewStyle().Foreground(ColorBrand).Italic(true)
    }
    sym := symStyle.Render(SymThinking)

    if msg.Folded || msg.Content == "" {
        // 折叠态：符号 + 字数 + 首行摘要
        charCount := len([]rune(msg.Content))
        summary := msg.Content
        summaryRunes := []rune(summary)
        if len(summaryRunes) > 40 {
            summary = string(summaryRunes[:40]) + "…"
        }
        // 替换换行为空格
        summary = strings.ReplaceAll(summary, "\n", " ")
        countStr := StyleDim.Render(fmt.Sprintf("(%d 字)", charCount))
        previewStr := StyleDim.Render(summary)
        return sym + StyleThinking.Render(" Thinking ") + countStr + "  " + previewStr
    }

    // 展开态：符号 + 字数 + 完整内容（italic dim，用 ⎿ 包裹）
    charCount := len([]rune(msg.Content))
    countStr := StyleDim.Render(fmt.Sprintf("(%d 字)", charCount))
    header := sym + StyleThinking.Render(" Thinking ") + countStr
    body := lipgloss.NewStyle().Foreground(ColorTextDim).Italic(true).Render(msg.Content)
    return header + "\n" + wrapMessageResponse(body)
```

- [ ] **Step 2: chat.go Update 增加 `t` 键处理**

在 `chat.go` 的 `Update` 方法中，`case tea.KeyMsg:` 后（viewport 更新之前），增加 `t` 键处理。在 `c.viewport, cmd = c.viewport.Update(msg)` 行（当前 chat.go:245）之前插入：

```go
case tea.KeyMsg:
    if msg.String() == "t" {
        // toggle 最近的 thinking block 折叠状态
        for i := len(c.messages) - 1; i >= 0; i-- {
            if c.messages[i].Role == "thinking" {
                c.messages[i].Folded = !c.messages[i].Folded
                c.invalidateCache()
                c.refreshContent(false) // 不自动滚动，保持用户位置
                break
            }
        }
        return c, nil
    }
```

注意：这段需要插入在 `switch msg := msg.(type) {` 的主 case 区域之后、`c.viewport, cmd = c.viewport.Update(msg)` 之前。具体位置在现有 `case MsgError:` 块结束之后、viewport update 之前。

- [ ] **Step 3: app.go shouldRouteKeyToChat 增加 `t`**

修改 `app.go:830-837`：

```go
func (a *App) shouldRouteKeyToChat(msg tea.KeyMsg) bool {
    switch msg.String() {
    case "pgup", "pgdown", "home", "end", "t":
        return true
    default:
        return false
    }
}
```

- [ ] **Step 4: 编译验证**

Run: `go build ./...`
Expected: 编译通过

- [ ] **Step 5: 手动测试 thinking 展开**

Run: `go run .`
- 发送一条消息，等待 thinking 出现
- 按 `t` 键 — thinking 应展开显示完整内容
- 再按 `t` — 折回一行摘要

- [ ] **Step 6: 提交**

```bash
git add internal/tui/chat.go internal/tui/app.go
git commit -m "feat: thinking 折叠/展开 — t 键 toggle + 摘要/完整渲染"
```

---

### Task 3: Slash 命令模糊补全 + 方向键选择

**Files:**
- Modify: `internal/tui/input.go:22-30` (InputBox 结构体)
- Modify: `internal/tui/input.go:68-161` (Update 方法)
- Modify: `internal/tui/input.go:164-171` (View 方法)
- Modify: `internal/tui/input.go:202-241` (补全匹配和渲染)
- Test: `internal/tui/input_test.go` (新建)

- [ ] **Step 1: 新建 input_test.go — 模糊匹配测试**

创建 `internal/tui/input_test.go`：

```go
package tui

import "testing"

func TestFuzzyMatch(t *testing.T) {
    tests := []struct {
        input  string
        target string
        match  bool
    }{
        {"/co", "/commit", true},       // 前缀
        {"/co", "/compact", true},      // 前缀
        {"/co", "/config", true},       // 前缀
        {"/cm", "/commit", true},       // 子序列
        {"/cm", "/compact", true},      // 子序列
        {"/cm", "/help", false},        // 不匹配
        {"/", "/help", true},           // 只有 /
        {"/hep", "/help", false},       // 顺序不对… 等等，h-e-p 在 help 中是 h(0) e(1) p? help 没有 p
        {"/he", "/help", true},         // 前缀
        {"/hl", "/help", true},         // 子序列 h...l
        {"/xz", "/exit", false},        // 不匹配
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
    // 前缀匹配 > 子序列匹配
    _, prefixScore := fuzzyMatchCommand("/co", "/cost")
    _, subseqScore := fuzzyMatchCommand("/ct", "/context")

    if prefixScore <= subseqScore {
        t.Errorf("前缀匹配得分 (%d) 应高于子序列 (%d)", prefixScore, subseqScore)
    }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/tui/ -run TestFuzzyMatch -v`
Expected: FAIL — `fuzzyMatchCommand` 未定义

- [ ] **Step 3: 实现 fuzzyMatchCommand**

在 `input.go` 末尾（替换当前的 `commonPrefix` 函数之后）添加：

```go
// fuzzyMatchCommand 子序列匹配 slash 命令
// 返回 (是否匹配, 得分)。得分越高越好：前缀匹配加分，连续匹配加分。
func fuzzyMatchCommand(input, target string) (bool, int) {
    if input == "" || target == "" {
        return input == "", 0
    }
    // 都转小写比较
    inputLower := strings.ToLower(input)
    targetLower := strings.ToLower(target)

    inputRunes := []rune(inputLower)
    targetRunes := []rune(targetLower)

    ii := 0 // input index
    score := 0
    prevMatchIdx := -1

    for ti := 0; ti < len(targetRunes) && ii < len(inputRunes); ti++ {
        if targetRunes[ti] == inputRunes[ii] {
            // 匹配
            score += 10
            if ti == ii {
                score += 5 // 位置对齐加分（前缀特征）
            }
            if prevMatchIdx == ti-1 {
                score += 3 // 连续匹配加分
            }
            prevMatchIdx = ti
            ii++
        }
    }

    if ii < len(inputRunes) {
        return false, 0 // 没有匹配完所有输入字符
    }
    return true, score
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/tui/ -run TestFuzzyMatch -v`
Expected: PASS

- [ ] **Step 5: InputBox 增加补全状态字段**

修改 `input.go` InputBox 结构体（当前行 22-29），增加：

```go
type InputBox struct {
    textarea      textarea.Model
    history       []string
    histIdx       int
    current       string
    width         int
    slashCommands []CommandHint
    lastMouseTime time.Time

    // 补全状态
    completionIdx   int           // 选中索引，-1 = 无选中
    completionItems []CommandHint // 当前匹配结果
}
```

在 `NewInputBox` 中初始化：`completionIdx: -1`

- [ ] **Step 6: 重写 matchSlashCommands 为模糊匹配 + 排序**

替换当前 `matchSlashCommands`（input.go:228-241）：

```go
func (i InputBox) matchSlashCommands() []CommandHint {
    val := strings.TrimSpace(i.textarea.Value())
    if !strings.HasPrefix(val, "/") || strings.Contains(val, " ") {
        return nil
    }

    type scored struct {
        hint  CommandHint
        score int
    }
    var matches []scored
    for _, cmd := range i.slashCommands {
        if ok, sc := fuzzyMatchCommand(val, cmd.Name); ok {
            matches = append(matches, scored{cmd, sc})
        }
    }

    // 按得分降序，同分按名称升序
    sort.Slice(matches, func(a, b int) bool {
        if matches[a].score != matches[b].score {
            return matches[a].score > matches[b].score
        }
        return matches[a].hint.Name < matches[b].hint.Name
    })

    result := make([]CommandHint, 0, len(matches))
    for _, m := range matches {
        result = append(result, m.hint)
    }
    return result
}
```

需要在 `input.go` 的 import 中增加 `"sort"`。

- [ ] **Step 7: 重写 Update 按键路由 — ↑↓ Tab Enter 补全**

替换 `input.go` Update 方法中 `case tea.KeyTab:` 到 `case tea.KeyEnter:` 之间的逻辑（行 92-129）：

```go
        case tea.KeyTab:
            // Tab 补全选中项
            matches := i.matchSlashCommands()
            if len(matches) > 0 {
                idx := i.completionIdx
                if idx < 0 || idx >= len(matches) {
                    idx = 0
                }
                i.textarea.SetValue(matches[idx].Name + " ")
                i.completionIdx = -1
                i.completionItems = nil
                return i, nil
            }

        case tea.KeyEnter:
            // Alt+Enter 换行
            if msg.Alt {
                i.textarea.InsertString("\n")
                return i, nil
            }
            // 如果补全列表可见且有选中项，先补全再提交
            matches := i.matchSlashCommands()
            if len(matches) > 0 && i.completionIdx >= 0 && i.completionIdx < len(matches) {
                text := matches[i.completionIdx].Name
                i.completionIdx = -1
                i.completionItems = nil
                i.history = append(i.history, text)
                i.histIdx = -1
                i.textarea.Reset()
                return i, func() tea.Msg { return MsgSubmit{Text: text} }
            }
            text := strings.TrimSpace(i.textarea.Value())
            if text == "" {
                return i, nil
            }
            if containsANSI(text) {
                i.textarea.Reset()
                return i, nil
            }
            i.history = append(i.history, text)
            i.histIdx = -1
            i.textarea.Reset()
            i.completionIdx = -1
            i.completionItems = nil
            return i, func() tea.Msg { return MsgSubmit{Text: text} }

        case tea.KeyUp:
            // 补全列表可见时：向上导航
            if i.hasActiveCompletion() {
                if i.completionIdx > 0 {
                    i.completionIdx--
                }
                return i, nil
            }
            // 否则：历史导航
            if i.textarea.Line() == 0 && len(i.history) > 0 {
                if i.histIdx == -1 {
                    i.current = i.textarea.Value()
                    i.histIdx = len(i.history) - 1
                } else if i.histIdx > 0 {
                    i.histIdx--
                }
                i.textarea.SetValue(i.history[i.histIdx])
                return i, nil
            }

        case tea.KeyDown:
            // 补全列表可见时：向下导航
            if i.hasActiveCompletion() {
                matches := i.matchSlashCommands()
                maxIdx := len(matches) - 1
                if maxIdx > 7 {
                    maxIdx = 7 // 最多显示 8 项
                }
                if i.completionIdx < maxIdx {
                    i.completionIdx++
                }
                return i, nil
            }
            // 否则：历史导航
            if i.histIdx >= 0 {
                if i.histIdx < len(i.history)-1 {
                    i.histIdx++
                    i.textarea.SetValue(i.history[i.histIdx])
                } else {
                    i.histIdx = -1
                    i.textarea.SetValue(i.current)
                }
                return i, nil
            }
```

- [ ] **Step 8: 增加 hasActiveCompletion 辅助方法**

在 `input.go` 末尾添加：

```go
// hasActiveCompletion 是否有活跃的补全列表
func (i InputBox) hasActiveCompletion() bool {
    val := strings.TrimSpace(i.textarea.Value())
    if !strings.HasPrefix(val, "/") || strings.Contains(val, " ") {
        return false
    }
    matches := i.matchSlashCommands()
    return len(matches) > 0
}
```

- [ ] **Step 9: 在 textarea 值变化时重置 completionIdx**

在 Update 方法末尾，`i.textarea, cmd = i.textarea.Update(msg)` 之后，添加补全状态同步：

```go
    var cmd tea.Cmd
    i.textarea, cmd = i.textarea.Update(msg)

    // 输入变化时刷新补全状态
    newMatches := i.matchSlashCommands()
    i.completionItems = newMatches
    if len(newMatches) > 0 {
        if i.completionIdx < 0 {
            i.completionIdx = 0
        }
        if i.completionIdx >= len(newMatches) {
            i.completionIdx = len(newMatches) - 1
        }
    } else {
        i.completionIdx = -1
    }

    return i, cmd
```

- [ ] **Step 10: 重写 renderSlashHint — 圆角边框 + 选中高亮**

替换 `input.go` 的 `renderSlashHint`（当前行 202-226）：

```go
func (i InputBox) renderSlashHint() string {
    matches := i.completionItems
    if len(matches) == 0 {
        return ""
    }

    boxWidth := min(72, max(34, i.width-2))
    var lines []string

    limit := min(8, len(matches))
    for idx, cmd := range matches[:limit] {
        name := lipgloss.NewStyle().Foreground(ColorText).Bold(true).Render(
            truncateText(cmd.Name, 16))
        desc := StyleDim.Render(truncateText(cmd.Description, max(12, boxWidth-24)))
        line := fmt.Sprintf("  %s  %s", name, desc)

        if idx == i.completionIdx {
            // 选中行：品牌色高亮
            line = lipgloss.NewStyle().Foreground(ColorBrand).Bold(true).Render(
                truncateText(cmd.Name, 16)) +
                "  " + lipgloss.NewStyle().Foreground(ColorText).Render(
                truncateText(cmd.Description, max(12, boxWidth-24)))
            line = "❯ " + line
        } else {
            line = "  " + line
        }

        lines = append(lines, line)
    }

    if len(matches) > limit {
        lines = append(lines, StyleDim.Render(
            fmt.Sprintf("  还有 %d 个命令", len(matches)-limit)))
    }
    lines = append(lines, StyleDim.Render("  Tab 补全 · Enter 执行 · ↑↓ 选择"))

    return lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(ColorInputBorder).
        PaddingLeft(1).
        Width(boxWidth).
        Render(strings.Join(lines, "\n"))
}
```

- [ ] **Step 11: 编译 + 测试**

Run: `go build ./... && go test ./internal/tui/ -v`
Expected: 编译通过，所有测试 PASS

- [ ] **Step 12: 提交**

```bash
git add internal/tui/input.go internal/tui/input_test.go
git commit -m "feat: slash 命令模糊补全 + ↑↓ 选择 + 圆角面板"
```

---

### Task 4: 消息显示细节 — 错误样式 + 折叠行 + 欢迎屏 + 滚动提示

**Files:**
- Modify: `internal/tui/chat.go:532-535` (error 渲染)
- Modify: `internal/tui/chat.go:692-693` (折叠行样式)
- Modify: `internal/tui/chat.go:348-354` (未读分隔线)
- Modify: `internal/tui/chat.go:417-423` (ViewWithHint)
- Modify: `internal/tui/app.go:869-898` (欢迎屏)

- [ ] **Step 1: 错误消息样式升级**

替换 `chat.go` `renderMessage` 中 `case "error":`（当前行 532-534）：

```go
    case "error":
        return lipgloss.NewStyle().
            BorderLeft(true).
            BorderForeground(ColorError).
            PaddingLeft(1).
            Render("⚠ " + StyleError.Render(msg.Content))
```

- [ ] **Step 2: 工具折叠行样式**

替换 `chat.go` 行 693：

```go
    // 当前：
    // outputBody += "\n" + StyleDim.Render(fmt.Sprintf("… +%d lines", lineCount-len(displayLines)))
    // 改为：
    outputBody += "\n" + lipgloss.NewStyle().Foreground(ColorBrand).Render(
        fmt.Sprintf("[+%d 行]", lineCount-len(displayLines)))
```

- [ ] **Step 3: 未读分隔线增强**

替换 `chat.go` 行 354：

```go
    // 当前：
    // sb.WriteString(StyleDim.Render(strings.Repeat("─", dividerWidth) + " 新消息"))
    // 改为：
    divStyle := lipgloss.NewStyle().Foreground(ColorWarning)
    labelWidth := lipgloss.Width("━━━ 新消息 ━━━")
    sideWidth := (dividerWidth - labelWidth) / 2
    if sideWidth < 2 {
        sideWidth = 2
    }
    sb.WriteString(divStyle.Render(
        strings.Repeat("━", sideWidth) + " 新消息 " + strings.Repeat("━", sideWidth)))
```

- [ ] **Step 4: 滚动提示增强 — 消息计数 + 品牌色**

替换 `chat.go` `ViewWithHint()`（当前行 417-424）：

```go
func (c ChatView) ViewWithHint() string {
    view := c.viewport.View()
    if c.hasNewMessages && !c.userAtBottom {
        count := c.countNewMessages()
        hint := lipgloss.NewStyle().Foreground(ColorBrand).Render(
            fmt.Sprintf("  ↓ %d 条新消息，按 End 跳到最新", count))
        return view + "\n" + hint
    }
    return view
}

// countNewMessages 统计未读消息数
func (c ChatView) countNewMessages() int {
    if c.unreadDividerIdx < 0 {
        return 0
    }
    count := 0
    for i := c.unreadDividerIdx; i < len(c.messages); i++ {
        if c.messages[i].Role != "system" && c.messages[i].Role != "thinking" {
            count++
        }
    }
    if count < 1 {
        count = 1
    }
    return count
}
```

- [ ] **Step 5: 欢迎屏增加操作提示**

修改 `app.go` `renderWelcomeBanner`，在 `return strings.Join(lines, "\n")` 之前（当前行 898）插入：

```go
    // 操作提示行
    lines = append(lines, "")
    lines = append(lines, dim.Render("  输入 / 查看命令 · Ctrl+C 退出 · Ctrl+Y 切换鼠标"))
```

- [ ] **Step 6: 编译验证**

Run: `go build ./...`
Expected: 编译通过

- [ ] **Step 7: 提交**

```bash
git add internal/tui/chat.go internal/tui/app.go
git commit -m "feat: 消息显示打磨 — 错误样式/折叠行/未读分隔线/欢迎提示"
```

---

### Task 5: 权限卡片升级 — 圆角边框 + 按键反馈

**Files:**
- Modify: `internal/tui/permission.go:10-17` (结构体)
- Modify: `internal/tui/permission.go:71-96` (Card 方法)
- Modify: `internal/tui/permission.go:27-59` (Update 方法)

- [ ] **Step 1: 结构体增加反馈字段**

修改 `permission.go` PermissionDialog 结构体（当前行 11-17）：

```go
type PermissionDialog struct {
    visible     bool
    toolName    string
    input       string
    response    chan PermissionResponse
    width       int
    height      int
    feedbackKey string // 刚按下的键（"y"/"n"/"a"/"e"），用于高亮反馈
    feedbackAge int    // tick 计数，>= 2 时清除
}
```

- [ ] **Step 2: Update 记录按键反馈 + spinner tick 老化**

修改 `permission.go` Update 方法中 Y/N/A/E 的 case，在 `p.respond()` 之前记录 feedback：

```go
        case "y", "Y":
            p.feedbackKey = "y"
            p.respond(PermAllow)
            return p, nil
        case "n", "N":
            p.feedbackKey = "n"
            p.respond(PermDeny)
            return p, nil
        case "a", "A":
            p.feedbackKey = "a"
            p.respond(PermAlways)
            return p, nil
        case "e", "E":
            p.feedbackKey = "e"
            p.respond(PermNever)
            return p, nil
```

在 `case tea.WindowSizeMsg:` 之后增加 spinner tick 处理：

```go
    case MsgSpinnerTick:
        if p.feedbackKey != "" {
            p.feedbackAge++
            if p.feedbackAge >= 2 {
                p.feedbackKey = ""
                p.feedbackAge = 0
            }
        }
```

注意：MsgSpinnerTick 需要在 app.go 中转发给 permission。当前 app.go 的 MsgSpinnerTick handler 没有转发。需要在 `app.go` 的 `case MsgSpinnerTick:` 中增加：

```go
    if a.permission.IsVisible() {
        a.permission, _ = a.permission.Update(msg)
    }
```

- [ ] **Step 3: Card 方法 — 圆角边框 + 按键高亮**

替换 `permission.go` Card 方法（当前行 72-96）：

```go
func (p PermissionDialog) Card(width int) string {
    if width < 40 {
        width = 40
    }

    // 工具名（品牌色加粗）+ 摘要（dim 色）
    summary := toolInputSummary(p.toolName, p.input)
    nameStyle := lipgloss.NewStyle().Foreground(ColorPerm).Bold(true)
    maxSummaryWidth := width - lipgloss.Width(p.toolName) - 8
    if maxSummaryWidth < 20 {
        maxSummaryWidth = 20
    }
    summaryText := truncateText(summary, maxSummaryWidth)
    header := nameStyle.Render(p.toolName) + "  " + StyleDim.Render(summaryText)

    // 按键提示（简化 + 反馈高亮）
    keys := p.renderKeys()

    return lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(ColorPerm).
        Padding(0, 1).
        Width(width).
        Render(header + "\n" + keys)
}

// renderKeys 渲染按键提示，刚按下的键高亮
func (p PermissionDialog) renderKeys() string {
    type keyDef struct {
        key   string
        label string
    }
    defs := []keyDef{
        {"y", "Y 允许"},
        {"n", "N 拒绝"},
        {"a", "A 总是"},
        {"e", "E 从不"},
    }

    var parts []string
    for _, d := range defs {
        style := StyleDim
        if p.feedbackKey == d.key {
            style = lipgloss.NewStyle().Foreground(ColorPerm).Bold(true)
        }
        parts = append(parts, style.Render(d.label))
    }
    return strings.Join(parts, "  ")
}
```

- [ ] **Step 4: 编译验证**

Run: `go build ./...`
Expected: 编译通过

- [ ] **Step 5: 提交**

```bash
git add internal/tui/permission.go internal/tui/app.go
git commit -m "feat: 权限卡片升级 — 圆角边框 + 简化按键 + 按下高亮反馈"
```

---

### Task 6: 全量验证

- [ ] **Step 1: 完整测试**

```bash
go build ./... && go vet ./... && go test ./... -count=1
```

Expected: 全部通过

- [ ] **Step 2: 手动验证矩阵**

| 场景 | 预期 |
|---|---|
| 输入 `/co` | 显示 /commit /compact /config /cost /context |
| ↓↓ 键 | 高亮移动到第 3 项 |
| Tab | 补全选中项到输入框 |
| Enter（补全列表可见时）| 直接提交选中命令 |
| Esc | 关闭补全列表 |
| Thinking 出现 | 显示 `∴ Thinking (N 字) 摘要…` |
| 按 `t` | 展开完整 thinking |
| 再按 `t` | 折回 |
| 权限弹窗 | 圆角边框，按 Y 时 "Y 允许" 短暂高亮 |
| 错误消息 | ⚠ 前缀 + 红色左边框 |
| 滚动看旧消息 | 新消息到来时显示 `↓ N 条新消息` 品牌色 |
| 欢迎屏 | 底部显示操作提示行 |
