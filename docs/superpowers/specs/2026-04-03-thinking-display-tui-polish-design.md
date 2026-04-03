# Extended Thinking Display + TUI Interaction Polish

**Date**: 2026-04-03
**Status**: Approved
**Scope**: 2 themes, 5 deliverables

---

## Theme A: Extended Thinking 可折叠展示

### 目标

让用户能看到 AI 的推理过程，默认折叠不占空间，需要时展开查看。Thinking 内容纳入持久化，resume 后可回看。

### A1. 数据模型变更

**provider/message.go** — `ContentBlock` 已有 `BlockThinking` 类型和 `Thinking string` 字段，无需改动。

**agent.go `processStream()`** — 当前只发 `MsgThinking` 事件，不收集到 assistant message blocks。改为：
- 新增 `thinkingBuilder strings.Builder` 累积 thinking 文本
- `EventThinking` 时：发 `MsgThinking` 给 TUI + 追加到 builder
- 流结束时：如果 builder 非空，在 assistant message 的 blocks 中插入 `ContentBlock{Type: BlockThinking, Thinking: thinkingBuilder.String()}`，位于 text block 之前
- 这样 `session.AddMessage(assistantMsg)` 自然持久化 thinking

**tui/chat.go `ChatMessage`** — 已有 `Role: "thinking"` 和 `Folded bool`，复用。Thinking 消息创建时 `Folded: true`。

**main.go `rebuildTranscript()`** — 增加 `BlockThinking` → `ChatMessage{Role: "thinking", Content: text, Folded: true}` 的映射。

### A2. 渲染

**折叠态**：
```
∴ Thinking (1247 字)  让我分析一下这个函数的逻辑…
```
- `∴` 符号用 brand/dim 色交替闪烁（已有）
- 字数统计 + 首行摘要（前 40 字 + `…`），dim 色
- 视觉占 1 行

**展开态**：
```
∴ Thinking (1247 字)
  ⎿ 让我分析一下这个函数的逻辑。首先看 session.go 中的…
    第二段推理内容继续展开…
    …完整 thinking 文本
```
- 完整文本用 `wrapMessageResponse()` 包裹（⎿ 前缀 + 缩进）
- 字体样式：italic + dim，和 assistant 正文区分

### A3. 交互

- **`t` 键**：toggle 离视口底部最近的 thinking block 的折叠状态
- 实现：`shouldRouteKeyToChat` 增加 `"t"` 路由到 chat
- ChatView 处理 `t` 键：从 messages 末尾向前搜索第一个 `Role == "thinking"` 的消息，翻转其 `Folded`，`invalidateCache()` + `refreshContent()`
- 折叠态下 thinking 内容不参与 viewport 行数计算（性能保证）

### A4. 持久化 round-trip

- 保存：thinking 作为 `ContentBlock{Type: BlockThinking}` 存入 assistant message → session JSON
- 加载：`rebuildTranscript` 遍历 assistant message blocks，遇到 `BlockThinking` 生成 `ChatMessage{Role: "thinking", Folded: true}`
- ContentBlock JSON 序列化：`Type` 是 int（`BlockThinking = 1`），`Thinking` 字段是 string，标准 JSON round-trip 即可
- 旧会话无 thinking block：兼容，不影响

---

## Theme B: TUI 交互打磨

### B1. Slash 命令补全重做

**文件**：`internal/tui/input.go`

#### 匹配算法

替换当前 `matchSlashCommands()` 的前缀匹配为子序列匹配：

```go
func fuzzyMatch(input, target string) (bool, int) {
    // input 的每个字符必须按序出现在 target 中
    // score: 连续匹配加分，首字符匹配加分，前缀匹配最高分
    // 返回 (是否匹配, 得分)
}
```

排序规则：
1. 精确前缀匹配（`/co` → `/cost` `/compact` `/commit`）得分最高
2. 连续子串匹配次之
3. 分散子序列最低
4. 同分按字母序

#### 状态

InputBox 新增字段：
```go
completionIdx    int            // 补全列表选中索引，-1 = 无选中
completionItems  []CommandHint  // 当前匹配结果（缓存，输入变化时刷新）
```

#### 按键路由

```
输入 "/" 后：
  ↑↓ → completionIdx 导航（优先于历史导航）
  Tab → 补全选中项到输入框
  Enter → 补全选中项 + 立即提交
  Esc → 关闭补全列表
  其他字符 → 更新匹配列表，completionIdx 重置为 0
无 "/" 前缀时：
  ↑↓ → 历史导航（现有逻辑不变）
```

判断条件：`strings.HasPrefix(value, "/") && len(completionItems) > 0`

#### 视觉

```
╭─────────────────────────────────────╮
│  /commit        生成 commit 并提交    │  ← 选中行：品牌色背景
│  /compact       压缩上下文            │
│  /config        显示当前配置          │
│  /cost          费用详情              │
│                                      │
│  Tab 补全 · Enter 执行 · ↑↓ 选择     │
╰─────────────────────────────────────╯
```

- 圆角边框（`lipgloss.RoundedBorder()`），边框色 `ColorInputBorder`
- 选中行：命令名 bold + 品牌色
- 匹配字符高亮：在命令名中对匹配到的字符用品牌色标记
- 描述右对齐，dim 色
- 最多显示 8 项
- 底部操作提示 dim 色

### B2. 权限卡片升级

**文件**：`internal/tui/permission.go`

变更：
1. `Card()` 方法中 `BorderLeft(true)` → `BorderStyle(lipgloss.RoundedBorder())`
2. 边框色保持 `ColorPerm`
3. 按键提示简化为：`Y 允许  N 拒绝  A 总是  E 从不`
4. 按下反馈：
   - PermissionDialog 新增 `feedbackKey string` + `feedbackAge int`
   - 按下 Y/N/A/E 时设置 `feedbackKey`，在 `MsgSpinnerTick` 中递增 `feedbackAge`
   - age < 2（~160ms）时，对应按键文字用高亮色渲染
   - age >= 2 时清除 feedback

### B3. 消息显示细节

**文件**：`internal/tui/chat.go`

1. **错误消息**（`renderMessage` case "error"）：
   - 当前：`wrapMessageResponse(StyleError.Render(msg.Content))`
   - 改为：`⚠ ` 前缀 + 左边框 `ColorError` + 内容
   ```go
   lipgloss.NewStyle().
       BorderLeft(true).
       BorderForeground(ColorError).
       PaddingLeft(1).
       Render("⚠ " + StyleError.Render(msg.Content))
   ```

2. **工具折叠行**：
   - 当前：`StyleDim.Render(fmt.Sprintf("… +%d lines", ...))`
   - 改为：`lipgloss.NewStyle().Foreground(ColorBrand).Render(fmt.Sprintf("[+%d 行]", ...))`

3. **欢迎屏**（`renderWelcomeBanner`）：
   - 在现有 info 行下方增加一行 dim 提示：
   ```
   输入 / 查看命令 · Ctrl+C 退出 · Ctrl+Y 切换鼠标
   ```

### B4. 滚动提示增强

**文件**：`internal/tui/chat.go`

1. **"有新消息" 提示**（`ViewWithHint()`）：
   - 当前：`StyleDim.Render("  ↓ 有新消息，按 End 跳到最新")`
   - 改为：计数未读消息数，用品牌色渲染
   ```go
   count := c.countNewMessages()
   hint := lipgloss.NewStyle().Foreground(ColorBrand).Render(
       fmt.Sprintf("  ↓ %d 条新消息，按 End 跳到最新", count))
   ```
   - `countNewMessages()`: 从 `unreadDividerIdx` 到末尾的非 system 消息数

2. **未读分隔线**（`rebuildCommittedCache` 中的分隔线渲染）：
   - 当前：`StyleDim.Render(strings.Repeat("─", w) + " 新消息")`
   - 改为：`lipgloss.NewStyle().Foreground(ColorWarning).Render("━━━ 新消息 ━━━")`
   - 粗线居中，两侧用 `━` 填充到 `dividerWidth`

---

## 不在本次范围

- 非交互 CLI 模式（Phase 1 后续）
- 消息时间戳（低优先级）
- Modal 背景 dimming（复杂度高，收益低）
- 状态栏重新设计（当前可用）
- Spinner 动画更换（当前风格已有特色）

---

## 影响的文件

| 文件 | 改动 |
|---|---|
| `agent.go` | processStream 收集 thinking blocks |
| `internal/tui/chat.go` | thinking 渲染、`t` 键 toggle、错误样式、折叠行样式、滚动提示 |
| `internal/tui/input.go` | 模糊匹配、补全导航、补全面板渲染 |
| `internal/tui/app.go` | `t` 键路由、欢迎屏提示 |
| `internal/tui/permission.go` | 圆角边框、按键反馈 |
| `main.go` | rebuildTranscript 增加 thinking block 映射 |

## 测试要求

- 模糊匹配算法单测（精确前缀 > 子串 > 子序列排序）
- thinking 持久化 round-trip 测试
- thinking 折叠/展开渲染测试
- IsReadOnlySafe 保持通过（不改 slash handler）
