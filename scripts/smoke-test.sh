#!/usr/bin/env bash
# API 冒烟测试 — 编排器也会调用同等检查
set -euo pipefail
BASE="${MARVIS_BASE_URL:-http://localhost:8080}"

pass=0
fail=0

check() {
  local name="$1" path="$2"
  local code
  code=$(curl -sf -o /dev/null -w "%{http_code}" "${BASE}${path}" 2>/dev/null || echo "000")
  if [[ "$code" =~ ^2 ]]; then
    echo "✅ $name ($code)"
    pass=$((pass + 1))
  elif [[ "$name" == "live_meta" && "$code" == "404" ]]; then
    echo "⚠️  $name (404, stream 未就绪)"
    pass=$((pass + 1))
  else
    echo "❌ $name ($code) $path"
    fail=$((fail + 1))
  fi
}

echo "==> bigA smoke test @ $BASE"
check health /health
check quote /api/v1/market/quote?code=600519
check list /api/v1/market/list?page=1&size=5
check kline /api/v1/market/kline?code=600519&period=day&limit=10
check portfolio /api/v1/portfolio
check ai_state /api/v1/ai/state
check analysis_market /api/v1/analysis/market
check analysis_anomalies /api/v1/analysis/anomalies
check analysis_portfolio /api/v1/analysis/portfolio
check analysis_report /api/v1/analysis/daily-report
check analysis_predict /api/v1/analysis/predict?code=600519
check sectors /api/v1/market/sectors?page=1&size=5
check sector_stocks /api/v1/market/sector/new_hghy/stocks?page=1&size=5
check conditional_orders /api/v1/conditional-orders
check performance /api/v1/performance
check live_meta /api/v1/market/live/meta

echo "==> $pass passed, $fail failed"
[[ "$fail" -eq 0 ]]
