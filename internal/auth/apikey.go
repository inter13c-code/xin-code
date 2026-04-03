// internal/auth/apikey.go
// 从配置文件读取 API Key
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// credentialsFile 凭据文件结构
type credentialsFile struct {
	APIKey string `json:"api_key"`
}

// SaveCredentials 将 API Key 写入 JSON 凭据文件
// 自动创建父目录
func SaveCredentials(path, apiKey string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create credentials dir: %w", err)
	}
	data, err := json.MarshalIndent(credentialsFile{APIKey: apiKey}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// ReadAPIKeyFromFile 从 JSON 文件读取 API Key（公开导出版本）
func ReadAPIKeyFromFile(path string) string {
	return readAPIKeyFromFile(path)
}

// readAPIKeyFromFile 从 JSON 文件读取 API Key
func readAPIKeyFromFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var creds credentialsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return ""
	}
	return creds.APIKey
}
