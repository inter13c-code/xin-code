package context

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/xincode-ai/xin-code/internal/provider"
)

// BuildSystemPrompt 组装系统提示词
func BuildSystemPrompt(tools []provider.ToolDef, projectInstructions string) string {
	var sb strings.Builder

	// 1. 基础身份
	sb.WriteString("你是 Xin Code，一个 AI 驱动的终端编程助手。你可以读写文件、执行 shell 命令、搜索代码来帮助用户完成编程任务。\n\n")

	// 2. 工具说明
	sb.WriteString("# 可用工具\n\n")
	sb.WriteString("你可以使用以下工具：\n")
	for _, t := range tools {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", t.Name, t.Description))
	}
	sb.WriteString("\n")

	// 3. 行为规范
	sb.WriteString("# 行为规范\n\n")
	sb.WriteString("- 直接给出结论，不要冗余解释\n")
	sb.WriteString("- 修改文件前先读取文件内容\n")
	sb.WriteString("- 不要把 API Key、密码等写入代码\n")
	sb.WriteString("- 使用工具时优先用专用工具（Read 而非 cat，Glob 而非 find）\n\n")

	// 4. 项目指令（XINCODE.md）
	if projectInstructions != "" {
		sb.WriteString("# 项目指令\n\n")
		sb.WriteString(projectInstructions)
		sb.WriteString("\n\n")
	}

	// 5. 环境信息
	sb.WriteString("# 环境信息\n\n")
	cwd, _ := os.Getwd()
	sb.WriteString(fmt.Sprintf("- 工作目录: %s\n", cwd))
	sb.WriteString(fmt.Sprintf("- 操作系统: %s/%s\n", runtime.GOOS, runtime.GOARCH))
	sb.WriteString(fmt.Sprintf("- 当前日期: %s\n", time.Now().Format("2006-01-02")))

	return sb.String()
}
