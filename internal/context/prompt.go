package context

import (
	"fmt"
	"os"
	"os/exec"
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

	// Git 信息
	if branch, err := getGitOutput("rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		sb.WriteString(fmt.Sprintf("- Git 分支: %s\n", branch))
	}
	if status, err := getGitOutput("status", "--short"); err == nil && status != "" {
		// 只显示摘要：多少文件有变更
		lines := strings.Split(strings.TrimSpace(status), "\n")
		sb.WriteString(fmt.Sprintf("- Git 状态: %d 个文件有变更\n", len(lines)))
	}

	return sb.String()
}

// getGitOutput 执行 git 命令并返回输出
func getGitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
