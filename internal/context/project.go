package context

import "os"

// LoadProjectInstructions 读取 XINCODE.md
func LoadProjectInstructions() string {
	data, err := os.ReadFile("XINCODE.md")
	if err != nil {
		return ""
	}
	return string(data)
}
