# bigA — 大 A 股模拟盘（AI 可操作）

完整模仿 A 股规则的**模拟盘**：真实免费行情 + **MySQL 全量落库** + **Redis 缓存/锁**，供 AI 或策略程序调用。

## 架构

```
行情(东方财富/新浪) → Redis 缓存(5s) → MySQL market_quote_log（每次拉取都记）
AI/策略 → POST /api/v1/ai/order → 撮合引擎 → MySQL sim_order / sim_trade / sim_position
```

| 组件 | 用途 |
|------|------|
| **MySQL `biga`** | 账户、持仓、委托、成交、行情历史、T+1 结算、AI 操作审计 |
| **Redis** | 行情短缓存、`biga:lock:account:*` 防并发下单 |

## 模拟规则（贴近大 A）

- 100 股整数倍、佣金万三（最低 5 元）、卖出印花税千一、过户费
- **T+1**：当日买入 `available=0`，需 `POST /api/v1/trade/settle` 模拟下一交易日
- **涨跌停**：主板 10% / 创业板·科创板 20% / ST 5%
- **市价单**：买按卖一、卖按买一；**限价单**需价格能成交
- **交易时段**：默认 `SIM_RELAX_HOURS=true` 方便开发；生产可设 `false` 强制 9:30–15:00

## 快速开始

```bash
# 确保本地 MySQL、Redis 已启动
redis-cli ping   # PONG
mysql -u root -e "SELECT 1"

cd /Users/lijianjun/GolandProjects/bigA
cp .env.example .env   # 按需改 MYSQL_DSN
GOPROXY=https://goproxy.cn,direct go mod tidy
go run ./cmd/server
```

默认：`http://localhost:8080`，库 `biga` 自动建表，账户 `default` 初始资金 100 万。

## AI 专用接口

```bash
# 完整状态（账户+持仓+委托+成交+实时盈亏）
curl http://localhost:8080/api/v1/ai/state

# 下单（市价/限价）
curl -X POST http://localhost:8080/api/v1/ai/order \
  -H "Content-Type: application/json" \
  -d '{"code":"600519","side":"buy","order_type":"market","quantity":100}'

curl -X POST http://localhost:8080/api/v1/ai/order \
  -H "Content-Type: application/json" \
  -d '{"code":"600519","side":"sell","order_type":"limit","price":1310,"quantity":100}'
```

## 常用 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | MySQL/Redis/各表行数 |
| GET | `/api/v1/market/quote?code=` | 实时行情（落库） |
| GET | `/api/v1/market/quotes/history?code=` | MySQL 历史行情 |
| GET | `/api/v1/portfolio` | 持仓盈亏 |
| GET | `/api/v1/orders` | 委托 |
| GET | `/api/v1/trades` | 成交 |
| POST | `/api/v1/trade/settle` | T+1 交割 |

## MySQL 表

- `sim_account` / `sim_position` / `sim_order` / `sim_trade`
- `market_quote_log` — 每次行情写入
- `ai_action_log` — AI 每次下单审计
- `sim_settlement_log` — 日终/T+1 记录
- `stock_info` — 股票板块缓存

## 环境变量

见 [.env.example](.env.example)：`MYSQL_DSN`、`REDIS_ADDR`、`SIM_INITIAL_CASH`、`SIM_RELAX_HOURS`。

## Flutter 客户端

```bash
# 终端 1：后端
go run ./cmd/server

# 终端 2：前端（需先安装 Flutter）
cd app
flutter create . --platforms=web,android,ios
flutter pub get
flutter run -d chrome
```

详见 [app/README.md](app/README.md)。

## 提交前检查（GoLand）

若 Commit 提示几百个错误，常见原因：

1. **索引了 `.tools/flutter`**（已删除并加入排除）— 在 GoLand：`右键 .tools → Mark Directory as → Excluded`
2. **误开了 jetbra 的 `*.vmoptions`** — 不是 bigA 代码，不要和本项目一起提交
3. 仅提交 **`/Users/lijianjun/GolandProjects/bigA`** 目录

本地一键检查：

```bash
./scripts/check.sh
```

通过后再 Commit（可只勾选 **Go fmt**，Analyze 范围设为 Project Files Only）。

## 说明

- 行情为公开免费源，仅供学习研究。
- 这是**模拟盘**，不接券商实盘；实盘需另接 OpenAPI。
