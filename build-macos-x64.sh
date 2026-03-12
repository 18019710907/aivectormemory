#!/usr/bin/env bash
# 打 macOS x86 (darwin/amd64) 包，产物：AIVectorMemory-<版本>-macos-x64.zip
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$SCRIPT_DIR"
WAILS="${WAILS_CMD:-$HOME/go/bin/wails}"
SQLITE_VEC_VERSION="0.1.6"

echo "==> 项目根目录: $ROOT"
cd "$ROOT"

# 1. 前端
echo "==> [1/4] 构建前端..."
cd "$ROOT/desktop/frontend"
npx vue-tsc --noEmit && npx vite build
cd "$ROOT"

# 2. Wails 桌面端 (darwin/amd64)
if [[ ! -x "$WAILS" ]]; then
  echo "错误: 未找到 wails，请先执行: GOBIN=\$HOME/go/bin go install github.com/wailsapp/wails/v2/cmd/wails@latest"
  exit 1
fi
echo "==> [2/4] 构建桌面端 (darwin/amd64)..."
cd "$ROOT/desktop"
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 "$WAILS" build
cd "$ROOT"

# 3. 下载 sqlite-vec macos x86_64
echo "==> [3/4] 下载 sqlite-vec (macos x86_64)..."
VEC_ARCHIVE="vec_archive.tar.gz"
VEC_EXTRACT="vec_extract"
curl -fSL -o "$VEC_ARCHIVE" \
  "https://github.com/asg017/sqlite-vec/releases/download/v${SQLITE_VEC_VERSION}/sqlite-vec-${SQLITE_VEC_VERSION}-loadable-macos-x86_64.tar.gz"
rm -rf "$VEC_EXTRACT" && mkdir -p "$VEC_EXTRACT" && tar xzf "$VEC_ARCHIVE" -C "$VEC_EXTRACT"
find "$VEC_EXTRACT" -name "vec0.dylib" -exec cp {} "$ROOT/" \;
rm -rf "$VEC_EXTRACT" "$VEC_ARCHIVE"

# 4. 打包 zip
VERSION=$(grep '^version' "$ROOT/pyproject.toml" | sed 's/.*= *"\(.*\)"/\1/')
PKG_NAME="AIVectorMemory-${VERSION}-macos-x64"
echo "==> [4/4] 打包: ${PKG_NAME}.zip"
rm -rf "$PKG_NAME"
mkdir -p "$PKG_NAME"
cp -R "$ROOT/desktop/build/bin/AIVectorMemory.app" "$PKG_NAME/"
cp "$ROOT/vec0.dylib" "$PKG_NAME/"
cp "$ROOT/scripts/install.sh" "$ROOT/README.md" "$PKG_NAME/"
zip -r "${PKG_NAME}.zip" "$PKG_NAME"
rm -rf "$PKG_NAME" "$ROOT/vec0.dylib"

echo "==> 完成: $ROOT/${PKG_NAME}.zip"
