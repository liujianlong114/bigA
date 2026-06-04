# bigA Flutter 客户端

多端 UI（Web / iOS / Android），对接 Go 后端 `http://localhost:8080`。

## 功能

- **行情**：搜索、实时报价、涨幅榜
- **交易**：市价/限价 买入卖出、T+1 交割
- **持仓**：总资产、浮动盈亏
- **记录**：历史成交

## 前置

1. Go 后端已启动：`cd .. && go run ./cmd/server`
2. 已安装 [Flutter SDK](https://docs.flutter.dev/get-started/install)

## 首次初始化

```bash
cd app
flutter create . --platforms=web,android,ios
flutter pub get
```

## 运行

### Web（浏览器，最快体验）

```bash
flutter run -d chrome
# 或
flutter run -d web-server --web-port=3000
```

浏览器打开提示的地址（通常 `http://localhost:3000`）。

### macOS 桌面

```bash
flutter run -d macos
```

### Android 模拟器

模拟器里 `localhost` 指向自身，需指定 API 地址：

```bash
flutter run --dart-define=API_BASE=http://10.0.2.2:8080
```

### 真机

把 `API_BASE` 改成你电脑的局域网 IP：

```bash
flutter run --dart-define=API_BASE=http://192.168.x.x:8080
```

## 配置

默认 API：`http://localhost:8080`  
修改见 `lib/config/api_config.dart` 或 `--dart-define=API_BASE=...`

## 目录

```
lib/
  config/     API 地址
  models/     数据模型
  services/   HTTP 客户端
  providers/  状态管理
  screens/    页面
  theme/      主题（A 股红涨绿跌）
```
