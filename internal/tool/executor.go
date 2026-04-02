package tool

import (
	"context"
	"sync"

	"github.com/xincode-ai/xin-code/internal/provider"
)

const maxConcurrentReadTools = 10

// AskPermissionFunc 权限询问回调（阻塞直到用户回答）
type AskPermissionFunc func(toolName string, input string) bool

// HookRunner 钩子执行接口（避免循环依赖）
type HookRunner interface {
	// RunPreToolUse 执行 preToolUse 钩子，返回 false 表示阻止
	RunPreToolUse(ctx context.Context, toolName, toolInput string) (bool, error)
	// RunPostToolUse 执行 postToolUse 钩子
	RunPostToolUse(ctx context.Context, toolName, toolOutput string, isError bool)
}

// ExecuteResult 单个工具执行结果
type ExecuteResult struct {
	ToolUseID string
	Result    *Result
}

// ExecuteBatch 批量执行工具调用
// 只读工具并发执行，写入工具顺序执行
func (r *Registry) ExecuteBatch(ctx context.Context, calls []*provider.ToolCall, checker PermissionChecker, askFn AskPermissionFunc, hooks ...HookRunner) []ExecuteResult {
	var hookRunner HookRunner
	if len(hooks) > 0 {
		hookRunner = hooks[0]
	}
	_ = hookRunner // 传递给 ExecuteWithPermission
	results := make([]ExecuteResult, len(calls))

	// 分组：只读 vs 写入
	type indexedCall struct {
		index int
		call  *provider.ToolCall
	}
	var readCalls, writeCalls []indexedCall

	for i, call := range calls {
		t, ok := r.Get(call.Name)
		if !ok || !t.IsReadOnly() {
			writeCalls = append(writeCalls, indexedCall{i, call})
		} else {
			readCalls = append(readCalls, indexedCall{i, call})
		}
	}

	// 只读工具并发执行
	if len(readCalls) > 0 {
		sem := make(chan struct{}, maxConcurrentReadTools)
		var wg sync.WaitGroup
		for _, ic := range readCalls {
			wg.Add(1)
			sem <- struct{}{}
			go func(ic indexedCall) {
				defer wg.Done()
				defer func() { <-sem }()
				result := r.ExecuteWithPermission(ctx, ic.call, checker, askFn, hookRunner)
				results[ic.index] = ExecuteResult{
					ToolUseID: ic.call.ID,
					Result:    result,
				}
			}(ic)
		}
		wg.Wait()
	}

	// 写入工具顺序执行
	for _, ic := range writeCalls {
		result := r.ExecuteWithPermission(ctx, ic.call, checker, askFn, hookRunner)
		results[ic.index] = ExecuteResult{
			ToolUseID: ic.call.ID,
			Result:    result,
		}
	}

	return results
}

// ExecuteWithPermission 执行单个工具调用（带权限检查 + 用户询问 + 钩子）
func (r *Registry) ExecuteWithPermission(ctx context.Context, call *provider.ToolCall, checker PermissionChecker, askFn AskPermissionFunc, hooks ...HookRunner) *Result {
	t, ok := r.Get(call.Name)
	if !ok {
		return &Result{Content: "unknown tool: " + call.Name, IsError: true}
	}

	// 权限检查
	if checker != nil {
		checkResult, reason := checker.Check(call.Name, t.IsReadOnly())
		switch checkResult {
		case ResultAllow:
			// 直接执行
		case ResultDeny:
			return &Result{Content: "permission denied: " + reason, IsError: true}
		case ResultNeedAsk:
			if askFn != nil {
				if !askFn(call.Name, call.Input) {
					return &Result{Content: "permission denied by user", IsError: true}
				}
			}
			// askFn 为 nil 时默认放行（向后兼容）
		}
	}

	// preToolUse 钩子
	var hookRunner HookRunner
	if len(hooks) > 0 {
		hookRunner = hooks[0]
	}
	if hookRunner != nil {
		allowed, _ := hookRunner.RunPreToolUse(ctx, call.Name, call.Input)
		if !allowed {
			return &Result{Content: "blocked by preToolUse hook", IsError: true}
		}
	}

	result, err := t.Execute(ctx, []byte(call.Input))
	if err != nil {
		return &Result{Content: "execution error: " + err.Error(), IsError: true}
	}

	// postToolUse 钩子
	if hookRunner != nil {
		hookRunner.RunPostToolUse(ctx, call.Name, result.Content, result.IsError)
	}

	return result
}
