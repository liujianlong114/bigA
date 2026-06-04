#!/usr/bin/env bash
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> 启动 Go 后端..."
lsof -ti :8080 | xargs kill -9 2>/dev/null || true
cd "$ROOT"
go run ./cmd/server &
BACK_PID=$!
sleep 2

echo "==> 启动 Flutter Web..."
cd "$ROOT/app"
if ! command -v flutter &>/dev/null; then
  echo "未找到 flutter，请先安装: brew install --cask flutter"
  echo "后端已在 :8080 运行 (pid $BACK_PID)"
  exit 1
fi
flutter pub get
flutter create . --platforms=web 2>/dev/null || true
flutter run -d web-server --web-port=3000
