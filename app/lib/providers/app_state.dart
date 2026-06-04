import 'package:flutter/foundation.dart';
import '../models/models.dart';
import '../services/api_service.dart';

class AppState extends ChangeNotifier {
  final ApiService api;
  AppState(this.api);

  bool loading = false;
  String? error;
  bool backendOk = false;

  Portfolio? portfolio;
  Quote? selectedQuote;
  List<StockBrief> marketList = [];
  List<TradeRecord> trades = [];
  String selectedCode = '600519';

  Future<void> init() async {
    await refreshAll();
  }

  Future<void> refreshAll() async {
    loading = true;
    error = null;
    notifyListeners();
    try {
      backendOk = await api.health();
      await Future.wait([
        loadPortfolio(),
        loadQuote(selectedCode),
        loadMarketList(),
        loadTrades(),
      ]);
    } catch (e) {
      error = e.toString();
      backendOk = false;
    } finally {
      loading = false;
      notifyListeners();
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
      await loadQuote(items.first.code);
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
