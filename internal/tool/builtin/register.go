package builtin

import "github.com/xincode-ai/xin-code/internal/tool"

// RegisterAll 注册所有内置工具
func RegisterAll(reg *tool.Registry) {
	reg.Register(&ReadTool{})
	reg.Register(&BashTool{})
	reg.Register(&GlobTool{})
	reg.Register(&GrepTool{})
}
