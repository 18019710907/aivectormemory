#!/usr/bin/env bash
# 桌面端开发模式启动（热重载，无需打包）
set -e
cd "$(dirname "$0")/desktop"
wails dev
