---
AIGC:
    Label: "1"
    ContentProducer: 001191440300708461136T1XGW3
    ProduceID: a35043cbf736614f204d734ba1b75696_6ab9180e65a611f195cd525400d9a7a1
    ReservedCode1: wUYU1/ls32MIA782Cj/o6JRBhy2KF9uTIV/HFPTxSIgCrdzf11cjvALaSEGhanNO10SAqUw7xXeLEj1NTH/zCr3Yi7xwCS/rP4BY4vVEV2GLl5zdYrwXz240FNFVXwZ9Zfoc0ymiWXGoTDRzbtxJIwDBp0gPpwk7PkoOznbUCA45wPzKW5eEIz5eosY=
    ContentPropagator: 001191440300708461136T1XGW3
    PropagateID: a35043cbf736614f204d734ba1b75696_6ab9180e65a611f195cd525400d9a7a1
    ReservedCode2: wUYU1/ls32MIA782Cj/o6JRBhy2KF9uTIV/HFPTxSIgCrdzf11cjvALaSEGhanNO10SAqUw7xXeLEj1NTH/zCr3Yi7xwCS/rP4BY4vVEV2GLl5zdYrwXz240FNFVXwZ9Zfoc0ymiWXGoTDRzbtxJIwDBp0gPpwk7PkoOznbUCA45wPzKW5eEIz5eosY=
---

# BigA 后续开发 Prompt

> 本文档供 AI 阅读，描述 BigA 模拟盘项目当前完成度及下一步待开发任务。
> 项目根目录：`/Users/lijianjun/GolandProjects/bigA`

---

## 一、项目当前架构概览

### 技术栈
- **后端**：Go 1.22+, chi 路由, MySQL 8.0, Redis 7, gorilla/websocket
- **客户端**：Flutter 3.x (Android/iOS/macOS/Web)
- **部署**：systemd + Nginx 反向代理

### 已有代码分层
```
cmd/server/main.go          — 启动入口，组装依赖链
cmd/seed-universe/main.go   — 一次性拉取全A股列表写入 Redis

internal/
  api/          — HTTP Handler（16端点）+ WebSocket 桥接
  cache/        — Redis 操作（行情缓存/实时Hash/代码表/交易锁）
  config/       — 环境变量加载（MYSQL_DSN, REDIS_ADDR, SIM_INITIAL_CASH等）
  db/           — MySQL 自动建库建表（9张表，嵌入式SQL迁移）
  market/       — 行情系统（新浪+东财双源、K线、板块、全市场批量、交易时段）
  sim/          — 模拟撮合引擎（限价/市价、T+1、涨跌停、费率）
  store/        — MySQL 全表 CRUD + K线读写
  stream/       — 每秒实时行情引擎 + WebSocket 广播中心

app/            — Flutter 客户端（行情/交易/持仓/成交，含简易蜡烛图）
```

### 数据流
```
新浪API / 东方财富API
        │
        ▼
   market.Service (多源聚合 + 5s Redis缓存)
        │
        ├──► REST API (/api/v1/market/quote, /kline, ...)
        │
        ├──► MySQL (market_quote_log 全量落库)
        │
        └──► stream.Engine (每秒批量拉取 → Redis Hash → WebSocket分片推送)
                    │
                    ▼
              Flutter WebSocket 客户端 (实时行情展示)
```

### 模拟交易流
```
POST /api/v1/order
        │
        ▼
   sim.Engine.PlaceOrder()
        ├── Redis 交易锁 (SetNX)
        ├── sim.Rules.ValidateOrder() — T+1/涨跌停/手数/时段
        ├── sim.Rules.MatchPrice() — 限价/市价撮合
        ├── sim.Fees — 佣金/印花税/过户费
        └── store.Repo 事务写入 — order/trade/position/account
```

---

## 二、已完成清单（无需重复开发）

- [x] 双源行情（新浪+东财）多级容错，GBK 解码
- [x] 全市场 5500+ 股票代码表（Redis 7天缓存，东财→新浪三级降级）
- [x] 每秒批量全市场行情（并发4路，200只/批），Redis Hash + WebSocket 分片推送
- [x] 板块自动检测（主板10%/双创20%/ST5%/北交所30%），涨跌停价格计算
- [x] K 线数据（日/周/月/分钟9种周期，东财+新浪双源），MySQL 惟一键去重落库
- [x] 交易时段判断（9:30-11:30, 13:00-15:00），开发模式 24h（SIM_RELAX_HOURS）
- [x] 模拟撮合引擎：市价/限价、T+1、涨跌停、100股整数倍、费率（佣金万三/印花税千一/过户费万0.2）
- [x] MySQL 9 表自动迁移 + 全表 CRUD
- [x] AI 接口：`GET /api/v1/ai/state`（账户/持仓/订单/盈亏）+ `POST /api/v1/ai/order`
- [x] Flutter 4 页面客户端（行情主页/交易下单/持仓盈亏/成交记录）+ 简易蜡烛图
- [x] WebSocket 实时行情推送
- [x] 部署方案（systemd + Nginx + 生产检查清单）
- [x] 全市场种子脚本

---

## 三、待开发任务（按优先级排列）

### P0 — 核心分析能力（用户明确要求的差异化功能）

#### 任务 1：持仓分析模块 `internal/analysis/portfolio.go`
**目标**：对用户持仓进行多维度分析
**具体需求**：
- 行业分布统计：接入申万行业分类数据，计算持仓在各行业的占比，识别单一行业超 40% 的集中风险
- 持仓质量评分：如果可获取 PE/PB/ROE/毛利率等基本面数据，对每只持仓股打分（A/B/C/D）
- 持仓-Beta 计算：计算持仓组合相对沪深 300 的 Beta 值
- 最大回撤估算：基于历史 K 线模拟当前持仓在过去 30/90/180 日的最大回撤
- 止盈止损参考价：基于 20 日均线 / 布林带上轨下轨 / ATR 给出参考价位
- **接口设计**：`GET /api/v1/analysis/portfolio` 返回 JSON

#### 任务 2：大盘解析模块 `internal/analysis/market_breadth.go`
**目标**：实时监控大盘整体状态
**具体需求**：
- 主要指数实时行情：上证/深证/创业板/科创50/沪深300/中证500/中证1000（复用现有批量拉取，新增指数代码）
- 市场宽度：上涨/下跌/平盘家数、涨跌比、涨幅>5%/<5%家数
- 情绪指标：涨停/跌停家数、炸板率、连板高度、成交额相对 20 日均量比值
- 资金流向：北向资金（沪股通+深股通）实时净流入（可从东财接口获取）
- **接口设计**：`GET /api/v1/analysis/market` 返回 JSON

#### 任务 3：异动检测 + 原因汇总 `internal/analysis/anomaly.go`
**目标**：自动检测异动股票并归因
**具体需求**：
- 异动检测规则：
  - 涨跌幅 > ±7%（主板）/ ±15%（双创）
  - 成交量 > 20 日均量 3 倍
  - 换手率 > 20%
  - 日内振幅 > 10%
- 原因自动归因：调用新闻 API / 爬取财联社快讯，将异动股票与近期新闻关联
- 生成异动汇总列表（代码/名称/涨跌幅/异动类型/可能原因）
- **接口设计**：`GET /api/v1/analysis/anomalies` 返回 JSON

#### 任务 4：AI 每日复盘报告生成 `internal/analysis/report.go`
**目标**：盘后自动生成结构化复盘报告
**具体需求**：
- 大盘综述：各指数涨跌幅、市场宽度总结、成交量分析
- 板块排名：涨幅前 5 和后 5 的申万一级行业
- 异动个股：涨跌幅榜 + 异动原因
- 明日关注：连续放量、形态突破、即将解禁/财报的个股
- 输出格式：结构化的 JSON（便于 AI 消费）+ 可选 Markdown 文本
- **接口设计**：`GET /api/v1/analysis/daily-report`，支持 `?date=2026-06-11` 参数

### P1 — 高价值增强

#### 任务 5：涨跌推测引擎 `internal/analysis/prediction.go`
**目标**：多因子评分，给出短期涨跌概率
**具体需求**：
- 技术因子：5/10/20/60 日均线趋势、MACD 金叉死叉、RSI 超买超卖、布林带位置、成交量放量缩量
- 每只股票计算综合评分（0-100），映射为涨跌概率
- 经典形态识别：双底/头肩顶/三角形/旗形 等（可选，可后续迭代）
- **接口设计**：`GET /api/v1/analysis/predict?code=600519` 返回评分+概率
- 注意：本任务可利用已有 K 线数据（market_kline 表）

#### 任务 6：板块/行业行情 `internal/market/sector.go`
**目标**：提供申万行业和概念板块行情
**具体需求**：
- 接入东方财富行业板块接口获取申万一级/二级行业实时涨跌幅和成分股
- 概念板块（如"ChatGPT""新能源车"）行情
- 板块资金流向
- **接口设计**：`GET /api/v1/market/sectors`、`GET /api/v1/market/sector/{code}/stocks`

#### 任务 7：条件单系统 `internal/sim/conditional.go`
**目标**：支持止盈止损等条件单
**具体需求**：
- 条件类型：价格触及（>=X 或 <=X）、涨跌幅触发（>=X%）
- 动作类型：市价买入/卖出、限价买入/卖出
- 条件单生命周期：创建 → 待触发 → 已触发/已过期（当日有效）
- 在 stream.Engine 的每秒 tick 中检查条件单触发
- **接口设计**：`POST /api/v1/conditional-order`、`GET /api/v1/conditional-orders`、`DELETE /api/v1/conditional-order/{id}`

#### 任务 8：模拟绩效分析 `internal/sim/performance.go`
**目标**：计算并展示账户绩效指标
**具体需求**：
- 日收益率序列、累计收益率曲线
- 最大回撤、夏普比率、Calmar 比率
- 胜率、盈亏比、平均持仓天数
- 相对沪深 300 的超额收益
- **接口设计**：`GET /api/v1/performance` 返回 JSON（含每日净值数据）

#### 任务 9：高级 K 线图增强（Flutter 客户端）
**目标**：在现有蜡烛图上叠加技术指标
**具体需求**：
- 在 `app/lib/widgets/kline_chart.dart` 基础上增加：
  - 成交量柱状图（底部副图）
  - MACD 指标（DIF/DEA/柱状，副图）
  - KDJ 指标（K/D/J 线，副图）
  - 移动平均线（MA5/MA10/MA20/MA60，主图叠加）
- 支持指标显示/隐藏切换

### P2 — 实用增强

#### 任务 10：节假日与停牌日历 `internal/market/calendar_v2.go`
**目标**：替换现有的硬编码交易时段判断
**具体需求**：
- 维护中国法定节假日列表（从公开 API 或配置文件获取）
- 支持临时停牌股票标记（当日不可交易）
- 除权除息日管理（影响前收盘价计算）
- 集合竞价时段标记（9:15-9:25）

#### 任务 11：Docker 容器化部署
**目标**：一键启动开发环境
**具体需求**：
- 编写 `Dockerfile`（Go 后端多阶段构建）
- 编写 `docker-compose.yml`（MySQL + Redis + bigA）
- 编写 `app/Dockerfile`（Flutter Web 构建 + Nginx 托管）
- 环境变量通过 `.env` 文件注入

#### 任务 12：用户认证系统 `internal/api/auth.go`
**目标**：支持多用户隔离
**具体需求**：
- JWT 登录/注册（用户名+密码）
- 中间件验证，每个用户独立账户/持仓/订单
- 新增 `sim_user` 表，existing sim_account 表关联 user_id

#### 任务 13：单元测试覆盖
**目标**：核心模块测试率 > 60%
**需要测试的模块**：
- `internal/sim/rules.go` — 下单规则校验（T+1/涨跌停/手数/撮价）
- `internal/sim/fees.go` — 费用计算（各场景验证）
- `internal/sim/engine.go` — 撮合流程（mock DB）
- `internal/market/secid.go` — 代码转换
- `internal/market/board.go` — 板块检测
- `internal/market/flex_float.go` — JSON 容错解析

### P3 — 锦上添花

#### 任务 14：Level-2 数据模拟
**目标**：基于 Level-1 数据模拟 Level-2 效果
- 十档买卖盘口
- 大单统计（>50 万 / >100 万）
- 逐笔成交方向判断（内盘/外盘）

#### 任务 15：多策略回测系统
**目标**：支持历史数据回测
- 策略接口定义（Go interface）
- 回测引擎（逐日模拟交易 + 绩效统计）
- 结果对比：收益曲线/回撤/夏普 vs 基准

#### 任务 16：全球市场联动
**目标**：展示外围市场作为参考
- 标普 500 / 纳斯达克
- 恒生指数 / 国企指数
- 富时 A50 期货
- 汇率（USD/CNY）

#### 任务 17：监控告警体系
**目标**：Prometheus + Grafana
- 行情拉取延迟/成功率
- Redis/MySQL 连接健康
- API 请求 QPS/延迟
- 模拟订单量

---

## 四、开发约束与注意事项

1. **新代码路径**：新增的分析模块统一放在 `internal/analysis/` 包下
2. **数据库迁移**：新增表需要在 `internal/db/migrations.sql` 和 `migrations/001_schema.sql` 中同步添加 DDL
3. **API 风格**：新接口沿用 `/api/v1/analysis/xxx` 前缀，在 `internal/api/handlers.go` 中注册
4. **行情数据复用**：分析模块优先从 Redis 缓存和 MySQL 历史数据获取，避免重复调用外部 API
5. **错误处理**：使用 Go 标准 error wrapping，HTTP 层统一返回 `{"error": "message"}` 格式
6. **日志规范**：关键操作使用 `log.Printf`，后续统一迁移到结构化日志
7. **配置管理**：新功能如需配置项，在 `internal/config/config.go` 的 `Config` 结构体中新增字段，并在 `.env.example` 中添加示例
8. **数据源限制**：新浪/东财为免费接口，有频率限制，批量操作注意限速
9. **交易时段**：分析类接口在非交易时段也应可用（使用最新缓存数据）

---

## 五、当前项目可直接运行

```bash
# 1. 启动 MySQL 和 Redis
# 2. 配置 .env 文件（参考 .env.example）
# 3. 初始化全市场代码表
go run ./cmd/seed-universe
# 4. 启动服务
go run ./cmd/server
# 5. 启动 Flutter 客户端
cd app && flutter run
```

当前已有 18 个 API 端点完整可用，模拟交易全流程可走通。

---

## 六、建议开发顺序

```
第一阶段（P0）
  任务1 持仓分析 ──┬── 任务2 大盘解析 ──┬── 任务4 每日复盘报告
                    │                    │
                    └── 任务3 异动检测 ──┘
第二阶段（P1）
  任务5 涨跌推测 ── 任务6 板块行情 ── 任务7 条件单 ── 任务8 绩效分析 ── 任务9 K线增强
第三阶段（P2）
  任务10 节假日日历 ── 任务11 Docker ── 任务12 用户认证 ── 任务13 单元测试
第四阶段（P3）
  任务14~17 按需开发
```
*（内容由AI生成，仅供参考）*
