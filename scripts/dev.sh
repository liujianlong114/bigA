#!/usr/bin/env bash
# 本地一键启动：Go 后端 :8080 + Flutter Web :3000
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOG="/tmp/biga-server.log"
PIDFILE="/tmp/biga-server.pid"

stop_port() {
  local p="$1"
  if lsof -ti ":$p" >/dev/null 2>&1; then
    echo "==> 停止占用 :$p 的进程..."
    lsof -ti ":$p" | xargs kill -9 2>/dev/null || true
    sleep 1
  fi
}

echo "==> 编译 Go 后端..."
cd "$ROOT"
go build -o "$ROOT/server" ./cmd/server

stop_port 8080
echo "==> 启动 Go 后端 :8080 (日志 $LOG)..."
nohup "$ROOT/server" >>"$LOG" 2>&1 &
echo $! >"$PIDFILE"
for i in $(seq 1 30); do
  if curl -sf http://127.0.0.1:8080/health >/dev/null 2>&1; then
    echo "    后端就绪 pid=$(cat "$PIDFILE")"
    break
  fi
  if ! kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
    echo "    后端启动失败，最近日志："
    tail -20 "$LOG"
    exit 1
  fi
  sleep 1
done

if ! curl -sf http://127.0.0.1:8080/health >/dev/null 2>&1; then
  echo "    后端超时未就绪"
  tail -20 "$LOG"
  exit 1
fi

bash "$ROOT/scripts/smoke-test.sh" || true

echo ""
echo "==> 后端: http://127.0.0.1:8080/health"
echo "==> 启动 Flutter Web :3000 ..."
cd "$ROOT/app"
if ! command -v flutter &>/dev/null; then
  echo "未找到 flutter。后端已在运行，请手动: cd app && flutter run -d web-server --web-port=3000 --web-hostname=127.0.0.1"
  exit 0
fi
flutter pub get
stop_port 3000
flutter run -d web-server --web-port=3000 --web-hostname=127.0.0.1
