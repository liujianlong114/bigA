import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../models/models.dart';
import '../services/api_service.dart';
import '../services/live_market_service.dart';

class AppState extends ChangeNotifier {
  static const _tokenKey = 'biga_auth_token';
  static const _themeKey = 'biga_theme_mode';

  final ApiService api;
  final LiveMarketService live = LiveMarketService();
  AppState(this.api);

  ThemeMode themeMode = ThemeMode.system;

  bool authReady = false;
  bool isGuest = false;
  SimUser? currentUser;
  bool get isLoggedIn => currentUser != null;

  bool loading = false;
  bool loadingMore = false;
  String? error;
  String? marketWarning;
  bool backendOk = false;
  bool liveConnected = false;
  bool marketOpen = false;

  Portfolio? portfolio;
  Quote? selectedQuote;
  KlineData? klineData;
  List<StockBrief> stockList = [];
  int stockPage = 1;
  int stockTotal = 0;
  bool stockHasMore = true;
  String stockBoard = 'all';
  String stockSort = 'change_pct';
  String? stockSearchQ;
  String? activeSectorCode;
  String? activeSectorName;
  Map<String, String> boardNames = {'all': '全部'};
  Map<String, int> boardCounts = {};
  List<SectorBrief> hotSectors = [];
  String hotSectorsSource = '';
  List<ConditionalOrder> conditionalOrders = [];
  PerformanceReport? performance;
  List<TradeRecord> trades = [];
  String selectedCode = '600519';
  String klinePeriod = 'day';

  static const int stockPageSize = 50;

  @override
  void dispose() {
    live.disconnect();
    super.dispose();
  }

  Future<void> init() async {
    await Future.wait([_restoreAuth(), _restoreTheme()]);
    authReady = true;
    notifyListeners();
    if (isLoggedIn || isGuest) {
      await refreshAll();
      _startLive();
    }
  }

  String get themeModeLabel => switch (themeMode) {
        ThemeMode.light => '亮色',
        ThemeMode.dark => '暗色',
        ThemeMode.system => '跟随系统',
      };

  Future<void> _restoreTheme() async {
    final prefs = await SharedPreferences.getInstance();
    themeMode = _parseThemeMode(prefs.getString(_themeKey));
  }

  ThemeMode _parseThemeMode(String? raw) => switch (raw) {
        'light' => ThemeMode.light,
        'dark' => ThemeMode.dark,
        _ => ThemeMode.system,
      };

  Future<void> setThemeMode(ThemeMode mode) async {
    if (themeMode == mode) return;
    themeMode = mode;
    notifyListeners();
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_themeKey, switch (mode) {
      ThemeMode.light => 'light',
      ThemeMode.dark => 'dark',
      ThemeMode.system => 'system',
    });
  }

  Future<void> cycleThemeMode() async {
    final next = switch (themeMode) {
      ThemeMode.system => ThemeMode.light,
      ThemeMode.light => ThemeMode.dark,
      ThemeMode.dark => ThemeMode.system,
    };
    await setThemeMode(next);
  }

  Future<void> _restoreAuth() async {
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString(_tokenKey);
    if (token == null || token.isEmpty) return;
    api.setToken(token);
    try {
      currentUser = await api.me();
    } catch (_) {
      api.setToken(null);
      await prefs.remove(_tokenKey);
      currentUser = null;
    }
  }

  Future<String?> login(String username, String password) async {
    try {
      final auth = await api.login(username, password);
      currentUser = auth.user;
      isGuest = false;
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_tokenKey, auth.token);
      notifyListeners();
      await refreshAll();
      _startLive();
      return null;
    } catch (e) {
      return e.toString();
    }
  }

  Future<String?> register(String username, String password, {String? nickname}) async {
    try {
      final auth = await api.register(username, password, nickname: nickname);
      currentUser = auth.user;
      isGuest = false;
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_tokenKey, auth.token);
      notifyListeners();
      await refreshAll();
      _startLive();
      return null;
    } catch (e) {
      return e.toString();
    }
  }

  Future<void> enterGuest() async {
    api.setToken(null);
    currentUser = null;
    isGuest = true;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
    notifyListeners();
    await refreshAll();
    _startLive();
  }

  Future<void> logout() async {
    live.disconnect();
    api.setToken(null);
    currentUser = null;
    isGuest = false;
    portfolio = null;
    trades = [];
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
    notifyListeners();
  }

  void _startLive() {
    if (!backendOk) return;
    live.connect((quotes, open) {
      marketOpen = open;
      liveConnected = live.connected;
      if (quotes.containsKey(selectedCode)) {
        final s = quotes[selectedCode]!;
        selectedQuote = Quote(
          code: s.code,
          name: s.name,
          price: s.price,
          changePct: s.changePct,
          prevClose: selectedQuote?.prevClose ?? s.price,
          limitUp: selectedQuote?.limitUp ?? 0,
          limitDown: selectedQuote?.limitDown ?? 0,
          board: selectedQuote?.board ?? s.board,
          bid1: selectedQuote?.bid1 ?? 0,
          ask1: selectedQuote?.ask1 ?? 0,
          source: 'live_ws',
        );
      }
      stockList = stockList.map((item) {
        final live = quotes[item.code];
        if (live == null) return item;
        return item.copyWith(price: live.price, changePct: live.changePct);
      }).toList();
      notifyListeners();
    });
  }

  Future<void> refreshAll() async {
    loading = true;
    error = null;
    marketWarning = null;
    notifyListeners();
    try {
      backendOk = await api.health();
      if (!backendOk) {
        error = 'Go 后端未连接，请确认 :8080 已启动';
        return;
      }
      await Future.wait([
        _safe(() => loadPortfolio()),
        _safe(() => loadQuote(selectedCode)),
        _safe(() => loadKline()),
        _safe(() => loadBoardStats()),
        _safe(() => loadHotSectors()),
        _safe(() => loadStockList(reset: true)),
        _safe(() => loadTrades()),
        _safe(() => loadPerformance()),
        _safe(() => loadConditionalOrders()),
      ]);
      if (!live.connected) _startLive();
    } catch (e) {
      error = e.toString();
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  Future<void> _safe(Future<void> Function() fn) async {
    try {
      await fn();
    } catch (e) {
      marketWarning ??= e.toString();
    }
  }

  Future<void> loadPortfolio() async {
    portfolio = await api.portfolio();
    notifyListeners();
  }

  Future<void> loadQuote(String code) async {
    selectedCode = code;
    selectedQuote = await api.quote(code);
    notifyListeners();
  }

  Future<void> loadKline() async {
    klineData = await api.kline(selectedCode, period: klinePeriod, limit: 120);
    notifyListeners();
  }

  Future<void> setKlinePeriod(String period) async {
    klinePeriod = period;
    await loadKline();
  }

  Future<void> loadBoardStats() async {
    final j = await api.boardStats();
    boardCounts = (j['counts'] as Map?)?.map((k, v) => MapEntry(k.toString(), (v as num).toInt())) ?? {};
    boardNames = (j['names'] as Map?)?.map((k, v) => MapEntry(k.toString(), v.toString())) ?? {'all': '全部'};
    notifyListeners();
  }

  Future<void> loadHotSectors() async {
    final page = await api.listSectors(page: 1, size: 10);
    hotSectors = page.items;
    hotSectorsSource = page.source;
    notifyListeners();
  }

  Future<void> loadStockList({bool reset = false}) async {
    if (reset) {
      stockPage = 1;
      stockHasMore = true;
      stockList = [];
    }
    if (activeSectorCode != null) {
      final page = await api.listSectorStocks(activeSectorCode!, page: stockPage, size: stockPageSize);
      stockTotal = page.total;
      if (reset) {
        stockList = page.items;
      } else {
        stockList = [...stockList, ...page.items];
      }
      stockHasMore = stockList.length < stockTotal;
      notifyListeners();
      return;
    }
    final page = await api.listLiveStocks(
      page: stockPage,
      size: stockPageSize,
      board: stockBoard,
      sort: stockSort,
      q: stockSearchQ,
    );
    stockTotal = page.total;
    if (reset) {
      stockList = page.items;
    } else {
      stockList = [...stockList, ...page.items];
    }
    stockHasMore = stockList.length < stockTotal;
    notifyListeners();
  }

  Future<void> loadMoreStocks() async {
    if (loadingMore || !stockHasMore) return;
    loadingMore = true;
    notifyListeners();
    try {
      stockPage++;
      await loadStockList(reset: false);
    } finally {
      loadingMore = false;
      notifyListeners();
    }
  }

  Future<void> setStockBoard(String board) async {
    activeSectorCode = null;
    activeSectorName = null;
    stockBoard = board;
    await loadStockList(reset: true);
  }

  Future<void> setStockSort(String sort) async {
    stockSort = sort;
    await loadStockList(reset: true);
  }

  Future<void> searchStockList(String q) async {
    activeSectorCode = null;
    activeSectorName = null;
    stockSearchQ = q.trim().isEmpty ? null : q.trim();
    await loadStockList(reset: true);
  }

  Future<void> openSector(SectorBrief sector) async {
    activeSectorCode = sector.code;
    activeSectorName = sector.name;
    stockSearchQ = null;
    stockBoard = 'all';
    await loadStockList(reset: true);
  }

  Future<void> clearSectorView() async {
    activeSectorCode = null;
    activeSectorName = null;
    await loadStockList(reset: true);
  }

  Future<void> loadPerformance() async {
    performance = await api.performance();
    notifyListeners();
  }

  Future<void> loadConditionalOrders() async {
    conditionalOrders = await api.listConditionalOrders();
    notifyListeners();
  }

  Future<String?> createConditionalOrder({
    required String code,
    required String conditionType,
    required double triggerValue,
    required String side,
    required String orderType,
    required int quantity,
    double? price,
    String? remark,
  }) async {
    try {
      await api.createConditionalOrder(
        code: code,
        conditionType: conditionType,
        triggerValue: triggerValue,
        side: side,
        orderType: orderType,
        quantity: quantity,
        price: price,
        remark: remark,
      );
      await loadConditionalOrders();
      return null;
    } catch (e) {
      return e.toString();
    }
  }

  Future<String?> cancelConditionalOrder(int id) async {
    try {
      await api.cancelConditionalOrder(id);
      await loadConditionalOrders();
      return null;
    } catch (e) {
      return e.toString();
    }
  }

  Future<void> loadTrades() async {
    trades = await api.trades();
    notifyListeners();
  }

  Future<String?> searchAndSelect(String q) async {
    try {
      final items = await api.search(q.trim());
      if (items.isEmpty) return '未找到 $q';
      selectedCode = items.first.code;
      await loadQuote(selectedCode);
      await loadKline();
      return null;
    } catch (e) {
      return e.toString();
    }
  }

  Future<String?> submitOrder({
    required String side,
    required String orderType,
    required int quantity,
    double? price,
  }) async {
    try {
      final res = await api.placeOrder(
        code: selectedCode,
        side: side,
        orderType: orderType,
        quantity: quantity,
        price: price,
      );
      if (res.status == 'rejected') {
        return res.rejectReason ?? '委托被拒绝';
      }
      await refreshAll();
      return null;
    } catch (e) {
      return e.toString();
    }
  }

  Future<String?> doSettle() async {
    try {
      await api.settle();
      await refreshAll();
      return null;
    } catch (e) {
      return e.toString();
    }
  }
}
