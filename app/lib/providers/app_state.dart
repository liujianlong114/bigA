import 'package:flutter/foundation.dart';
import '../models/models.dart';
import '../services/api_service.dart';
import '../services/live_market_service.dart';

class AppState extends ChangeNotifier {
  final ApiService api;
  final LiveMarketService live = LiveMarketService();
  AppState(this.api);

  bool loading = false;
  String? error;
  String? marketWarning;
  bool backendOk = false;
  bool liveConnected = false;
  bool marketOpen = false;

  Portfolio? portfolio;
  Quote? selectedQuote;
  KlineData? klineData;
  List<StockBrief> marketList = [];
  List<TradeRecord> trades = [];
  String selectedCode = '600519';
  String klinePeriod = 'day';

  @override
  void dispose() {
    live.disconnect();
    super.dispose();
  }

  Future<void> init() async {
    await refreshAll();
    _startLive();
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
          board: selectedQuote?.board ?? 'main',
          bid1: selectedQuote?.bid1 ?? 0,
          ask1: selectedQuote?.ask1 ?? 0,
          source: 'live_ws',
        );
      }
      final sorted = quotes.values.toList()
        ..sort((a, b) => b.changePct.compareTo(a.changePct));
      marketList = sorted.take(15).toList();
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
        _safe(() => loadMarketList()),
        _safe(() => loadTrades()),
      ]);
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

  Future<void> loadMarketList() async {
    marketList = await api.listStocks(page: 1, size: 15);
    notifyListeners();
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
