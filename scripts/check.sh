#!/usr/bin/env bash
# 提交前本地检查（与 GoLand Commit checks 对齐）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> gofmt"
gofmt -w $(find . -name '*.go' -not -path './.tools/*')

echo "==> go vet"
go vet ./...

echo "==> go build"
go build -o /dev/null ./cmd/server

echo "==> flutter analyze"
cd app
flutter pub get
flutter analyze

echo "==> OK"
