package main

import (
	"context"
	"fmt"
	"io"

	xcontext "github.com/xincode-ai/xin-code/internal/context"
	"github.com/xincode-ai/xin-code/internal/provider"
	"github.com/xincode-ai/xin-code/internal/tool"
)

// Agent 核心引擎
type Agent struct {
	provider   provider.Provider
	tools      *tool.Registry
	permission tool.PermissionChecker
	config     *Config
	messages   []provider.Message
	output     io.Writer // 输出目标（Phase 1: os.Stdout, Phase 2: TUI）
}

// NewAgent 创建 Agent 实例
func NewAgent(p provider.Provider, tools *tool.Registry, cfg *Config, w io.Writer) *Agent {
	return &Agent{
		provider:   p,
		tools:      tools,
		permission: &tool.SimplePermissionChecker{Mode: tool.PermissionMode(cfg.Permission.Mode)},
		config:     cfg,
		messages:   make([]provider.Message, 0),
		output:     w,
	}
}

// Run 执行一轮 Agent 循环（用户消息 -> API -> 工具 -> 循环）
func (a *Agent) Run(ctx context.Context, userMessage string) error {
	// 追加用户消息
	a.messages = append(a.messages, provider.NewTextMessage(provider.RoleUser, userMessage))

	// 组装 system prompt
	projectInstructions := xcontext.LoadProjectInstructions()
	systemPrompt := xcontext.BuildSystemPrompt(a.tools.ToolDefs(), projectInstructions)

	turns := 0
	for {
		turns++
		if turns > a.config.MaxTurns {
			fmt.Fprintln(a.output, "\n[达到最大轮次限制]")
			break
		}

		// 构建请求
		req := &provider.Request{
			Model:     a.config.Model,
			System:    systemPrompt,
			Messages:  a.messages,
			Tools:     a.tools.ToolDefs(),
			MaxTokens: a.config.MaxTokens,
		}

		// 流式调用 API
		events, err := a.provider.Stream(ctx, req)
		if err != nil {
			return fmt.Errorf("API error: %w", err)
		}

		// 处理流式事件
		assistantMsg, toolCalls, err := a.processStream(events)
		if err != nil {
			return err
		}

		// 追加 assistant 消息到历史
		a.messages = append(a.messages, assistantMsg)

		// 没有工具调用 -> 本轮结束
		if len(toolCalls) == 0 {
			break
		}

		// 执行工具
		results := a.tools.ExecuteBatch(ctx, toolCalls, a.permission)

		// 工具结果追加到消息历史
		for _, r := range results {
			a.messages = append(a.messages,
				provider.NewToolResultMessage(r.ToolUseID, r.Result.Content, r.Result.IsError))
		}
	}

	return nil
}

// processStream 处理流式事件，收集 assistant 消息和工具调用
func (a *Agent) processStream(events <-chan provider.Event) (provider.Message, []*provider.ToolCall, error) {
	var textContent string
	var toolCalls []*provider.ToolCall
	var blocks []provider.ContentBlock

	for evt := range events {
		switch evt.Type {
		case provider.EventTextDelta:
			fmt.Fprint(a.output, evt.Text)
			textContent += evt.Text

		case provider.EventThinking:
			if evt.Thinking != nil {
				fmt.Fprintf(a.output, "\n[thinking] %s\n", evt.Thinking.Text)
			}

		case provider.EventToolUse:
			if evt.ToolCall != nil {
				fmt.Fprintf(a.output, "\n⚙ %s\n", evt.ToolCall.Name)
				toolCalls = append(toolCalls, evt.ToolCall)
				blocks = append(blocks, provider.ContentBlock{
					Type:     provider.BlockToolUse,
					ToolCall: evt.ToolCall,
				})
			}

		case provider.EventUsage:
			// Phase 2: 更新 cost tracker + statusbar
			if evt.Usage != nil {
				fmt.Fprintf(a.output, "\n[tokens: in=%d out=%d]\n",
					evt.Usage.InputTokens, evt.Usage.OutputTokens)
			}

		case provider.EventError:
			return provider.Message{}, nil, fmt.Errorf("stream error: %w", evt.Error)

		case provider.EventDone:
			// 流结束
		}
	}

	// 构建 assistant 消息
	if textContent != "" {
		blocks = append([]provider.ContentBlock{
			{Type: provider.BlockText, Text: textContent},
		}, blocks...)
	}

	msg := provider.Message{
		Role:    provider.RoleAssistant,
		Content: blocks,
	}

	fmt.Fprintln(a.output) // 换行

	return msg, toolCalls, nil
}
