# bigA 启动与部署文档

大 A 股模拟盘：Go 后端 + MySQL + Redis + Flutter 多端客户端，含全市场 WebSocket 实时行情。

---

## 1. 环境要求

| 组件 | 版本建议 | 说明 |
|------|----------|------|
| Go | 1.25+ | `go version` |
| MySQL | 8.0+ | 本地或远程均可 |
| Redis | 6.0+ | 行情缓存、实时 Hash、交易锁 |
| Flutter | 3.2+ | 仅跑客户端时需要 |

```bash
# macOS 示例
brew install go mysql redis
brew install --cask flutter   # 可选

# 验证
go version
mysql -u root -e "SELECT 1"
redis-cli ping                # 应返回 PONG
flutter --version             # 可选
```

---

## 2. 获取代码与依赖

```bash
cd /path/to/bigA
cp .env.example .env          # 按需修改，见下文

# 国内网络建议
export GOPROXY=https://goproxy.cn,direct
go mod tidy
go build -o bin/bigA ./cmd/server
```

---

## 3. 配置（`.env`）

复制 `.env.example` 为 `.env`：

```bash
cp .env.example .env
```

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MYSQL_DSN` | `root@tcp(127.0.0.1:3306)/biga?...` | MySQL 连接串，库名 `biga` 自动创建 |
| `REDIS_ADDR` | `127.0.0.1:6379` | Redis 地址 |
| `REDIS_PASSWORD` | 空 | Redis 密码 |
| `REDIS_DB` | `0` | Redis DB 编号 |
| `HTTP_ADDR` | `:8080` | 后端监听地址 |
| `SIM_INITIAL_CASH` | `1000000` | 模拟账户初始资金 |
| `SIM_RELAX_HOURS` | `true` | `true`=开发模式 24h 拉行情；`false`=仅交易时段 |

**生产环境务必设置：**

```env
SIM_RELAX_HOURS=false
```

---

## 4. 数据库

首次启动后端时会自动执行迁移建表，无需手动导入。

如需手动初始化：

```bash
mysql -u root -e "CREATE DATABASE IF NOT EXISTS biga DEFAULT CHARSET utf8mb4;"
mysql -u root biga < migrations/001_schema.sql
```

主要表：`sim_account`、`sim_position`、`sim_order`、`sim_trade`、`market_quote_log`、`market_kline`、`ai_action_log`、`stock_info`。

---

## 5. 首次启动：股票列表缓存（重要）

全市场实时行情需要先加载约 **5500** 只 A 股代码。新浪列表接口有频率限制，**首次部署必须先执行种子脚本**：

```bash
go run ./cmd/seed-universe
```

成功输出示例：

```
universe: 在线加载 5500 只
完成：5500 只股票已缓存到 Redis key biga:universe:v1
```

- 耗时约 **1–2 分钟**
- 数据缓存 7 天，之后启动会从 Redis 秒读
- 若种子脚本报 `拒绝访问` 或 JSON 解析失败，等待几分钟后重试

---

## 6. 启动后端

### 6.1 开发模式

```bash
# 终端 1
go run ./cmd/server

# 或编译后运行
go build -o bin/bigA ./cmd/server
./bin/bigA
```

启动日志应包含：

```
stream: 从 Redis 加载 5500 只
stream: 实时行情引擎已启动 interval=1s
实时行情: WS /api/v1/ws/market | REST /api/v1/market/live
bigA 大A模拟盘启动 :8080
```

### 6.2 一键前后端（开发）

```bash
./scripts/dev.sh
```

- 自动释放 `:8080`，启动 Go 后端
- 启动 Flutter Web（`http://localhost:3000`）

### 6.3 验证后端

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/market/live/meta
curl "http://localhost:8080/api/v1/market/live?code=600519"
redis-cli HLEN biga:live:quotes    # 交易时段内应 > 4000
```

---

## 7. 启动 Flutter 客户端

```bash
cd app
flutter pub get
flutter run -d chrome              # Web
flutter run -d macos             # macOS 桌面
flutter run                      # 已连接的设备/模拟器
```

### API 地址配置

默认连 `http://localhost:8080`。其他环境编译时传入：

```bash
# Android 模拟器
flutter run -d android --dart-define=API_BASE=http://10.0.2.2:8080

# 真机（改成电脑局域网 IP）
flutter run --dart-define=API_BASE=http://192.168.1.100:8080
```

WebSocket 地址由 `API_BASE` 自动推导（`http` → `ws`，`https` → `wss`）。

---

## 8. 生产部署

### 8.1 编译

```bash
# Linux 服务器
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/bigA ./cmd/server

# 本机
go build -o bin/bigA ./cmd/server
```

### 8.2 systemd 示例

```ini
# /etc/systemd/system/biga.service
[Unit]
Description=bigA stock sim server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=biga
WorkingDirectory=/opt/bigA
EnvironmentFile=/opt/bigA/.env
ExecStart=/opt/bigA/bin/bigA
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now biga
sudo systemctl status biga
journalctl -u biga -f
```

### 8.3 Nginx 反向代理（含 WebSocket）

```nginx
upstream biga_backend {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name your.domain.com;

    location /api/v1/ws/ {
        proxy_pass http://biga_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 3600s;
    }

    location / {
        proxy_pass http://biga_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 8.4 生产检查清单

- [ ] MySQL、Redis 已启动且可连
- [ ] `.env` 中 `SIM_RELAX_HOURS=false`
- [ ] 已执行 `go run ./cmd/seed-universe`
- [ ] 防火墙放行 `8080`（或仅 Nginx 80/443）
- [ ] `curl /health` 返回 `mysql:true, redis:true`
- [ ] 交易时段 `redis-cli HLEN biga:live:quotes` 有数据

### 8.5 Flutter 生产构建

```bash
cd app
flutter build web --dart-define=API_BASE=https://your.domain.com
flutter build apk --dart-define=API_BASE=https://your.domain.com
flutter build macos --dart-define=API_BASE=https://your.domain.com
```

Web 产物在 `app/build/web`，可挂到 Nginx 静态目录或与 API 同域部署。

---

## 9. API 速查

### 健康与实时行情

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 服务状态、live 元信息、表行数 |
| WS | `/api/v1/ws/market` | 全市场实时推送（分片） |
| GET | `/api/v1/market/live/meta` | 最新 tick 序号、数量、时间 |
| GET | `/api/v1/market/live?code=600519` | 单只股票 live 缓存 |
| GET | `/api/v1/market/live?cursor=0` | 分页读取全量 live |

### 行情

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/market/quote?code=` | 单只行情（落库） |
| GET | `/api/v1/market/list?page=&size=` | 涨幅榜列表 |
| GET | `/api/v1/market/search?q=` | 搜索 |
| GET | `/api/v1/market/kline?code=&period=&limit=` | K 线（day/week/month/m5…） |
| GET | `/api/v1/market/quotes/history?code=` | MySQL 历史行情 |

### 模拟盘

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/account` | 账户资金 |
| GET | `/api/v1/portfolio` | 持仓盈亏 |
| GET | `/api/v1/orders` | 委托 |
| GET | `/api/v1/trades` | 成交 |
| POST | `/api/v1/trade/settle` | T+1 交割 |
| POST | `/api/v1/ai/order` | AI/策略下单 |
| GET | `/api/v1/ai/state` | AI 完整状态快照 |

### WebSocket 消息格式

```json
{"type":"meta","seq":1,"updated_at":1710000000000,"market_open":true,"count":5184}
{"type":"quotes","seq":1,"part":1,"total":11,"market_open":true,"data":[{...}]}
```

`data` 中每条字段：`code`、`name`、`price`、`change_pct`、`volume`、`bid1`、`ask1` 等。

---

## 10. Redis 键说明

| Key | 说明 | TTL |
|-----|------|-----|
| `biga:universe:v1` | 全 A 股代码表 JSON | 7 天 |
| `biga:live:quotes` | 全市场实时行情 Hash | 120s |
| `biga:live:meta` | 最新 tick 元信息 | 120s |
| `biga:quote:{code}` | 单只行情短缓存 | 5s |
| `biga:lock:account:{id}` | 下单并发锁 | 请求级 |

---

## 11. 常见问题

### 端口 8080 被占用 / 进程 exit 137

```bash
lsof -ti:8080 | xargs kill -9
go run ./cmd/server
```

exit 137 表示进程被 SIGKILL（重复启动、手动 kill 或 OOM）。

### `universe: 拒绝访问` / 列表加载失败

新浪限流。等待 2–5 分钟后重试：

```bash
go run ./cmd/seed-universe
```

### `live_quotes: 0` / 无实时数据

1. 确认 `SIM_RELAX_HOURS=true`（开发）或当前在交易时段（生产）
2. 确认 universe 已缓存：`redis-cli EXISTS biga:universe:v1`
3. 查看日志是否有 `stream: tick seq=...`

### Flutter 显示「后端未连接」

1. 确认 `go run ./cmd/server` 在跑
2. Web 用 `localhost:8080`；Android 模拟器用 `10.0.2.2:8080`
3. `curl http://localhost:8080/health`

### `go get` 超时

```bash
export GOPROXY=https://goproxy.cn,direct
go mod tidy
```

### 行情数据源说明

- **新浪**：列表、批量实时行情（主力）
- **东方财富 push2his**：K 线
- 部分网络 `push2.eastmoney.com` 不可用，后端已自动降级

---

## 12. 日常运维命令

```bash
# 健康检查
./scripts/check.sh

# 重载股票列表（列表接口恢复后）
go run ./cmd/seed-universe

# 查看实时行情数量
redis-cli HLEN biga:live:quotes
redis-cli GET biga:live:meta

# 查看日志（systemd）
journalctl -u biga -f --since "10 min ago"
```

---

## 13. 目录结构

```
bigA/
├── cmd/
│   ├── server/          # 主服务入口
│   └── seed-universe/   # 股票列表种子脚本
├── internal/
│   ├── api/             # HTTP + WebSocket 路由
│   ├── market/          # 行情源、批量拉取、universe
│   ├── stream/          # 每秒实时引擎 + WS 广播
│   ├── sim/             # 模拟撮合引擎
│   ├── store/           # MySQL
│   └── cache/           # Redis
├── app/                 # Flutter 客户端
├── migrations/          # SQL  schema
├── scripts/
│   ├── dev.sh           # 开发一键启动
│   └── check.sh         # 提交前检查
├── .env.example
└── DEPLOY.md            # 本文档
```

---

## 14. 最小启动流程（复制即用）

```bash
# 1. 依赖
redis-cli ping && mysql -u root -e "SELECT 1"

# 2. 配置
cp .env.example .env

# 3. 依赖与编译
export GOPROXY=https://goproxy.cn,direct
go mod tidy

# 4. 首次：缓存股票列表（只需一次，或缓存过期后重做）
go run ./cmd/seed-universe

# 5. 启动后端
go run ./cmd/server

# 6. 另开终端：Flutter
cd app && flutter pub get && flutter run -d chrome
```

默认访问：后端 `http://localhost:8080`，Flutter Web `http://localhost:3000`（`dev.sh`）或 Chrome 自动打开端口。
