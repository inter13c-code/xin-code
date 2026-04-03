#!/bin/sh
set -e

# Xin Code installer
# Usage: curl -fsSL https://raw.githubusercontent.com/xincode-ai/xin-code/main/install.sh | sh

REPO="xincode-ai/xin-code"
INSTALL_DIR="/usr/local/bin"

# 检测系统和架构
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "不支持的系统: $OS (请使用 Linux 或 macOS)"; exit 1 ;;
esac

# 获取最新版本
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "获取最新版本失败"
  exit 1
fi

echo "安装 xin-code v${LATEST} (${OS}/${ARCH})..."

# 下载并解压
TMP=$(mktemp -d)
URL="https://github.com/${REPO}/releases/download/v${LATEST}/xin-code_${OS}_${ARCH}.tar.gz"
curl -fsSL "$URL" -o "${TMP}/xin-code.tar.gz"
tar xzf "${TMP}/xin-code.tar.gz" -C "$TMP"

# 安装
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP}/xin-code" "${INSTALL_DIR}/xin-code"
else
  echo "需要 sudo 权限安装到 ${INSTALL_DIR}"
  sudo mv "${TMP}/xin-code" "${INSTALL_DIR}/xin-code"
fi

rm -rf "$TMP"

echo "✓ xin-code v${LATEST} 已安装到 ${INSTALL_DIR}/xin-code"
echo "  运行 'xin-code' 开始使用"
