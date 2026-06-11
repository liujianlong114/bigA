import 'dart:convert';
import 'package:http/http.dart' as http;
import '../config/api_config.dart';
import '../models/models.dart';

class ApiException implements Exception {
  final String message;
  ApiException(this.message);
  @override
  String toString() => message;
}

class ApiService {
  final http.Client _client;
  String? _token;

  ApiService({http.Client? client}) : _client = client ?? http.Client();

  String? get token => _token;

  void setToken(String? token) {
    _token = token?.isNotEmpty == true ? token : null;
  }

  Map<String, String> _headers({bool jsonBody = false}) {
    final h = <String, String>{};
    if (jsonBody) h['Content-Type'] = 'application/json';
    if (_token != null) h['Authorization'] = 'Bearer $_token';
    return h;
  }

  Uri _uri(String path, [Map<String, String>? query]) =>
      Uri.parse('${ApiConfig.baseUrl}$path').replace(queryParameters: query);

  Future<Map<String, dynamic>> _get(String path, [Map<String, String>? q]) async {
    final res = await _client.get(_uri(path, q), headers: _headers());
    return _decode(res);
  }

  Future<Map<String, dynamic>> _post(String path, Map<String, dynamic> body) async {
    final res = await _client.post(
      _uri(path),
      headers: _headers(jsonBody: true),
      body: jsonEncode(body),
    );
    return _decode(res);
  }

  Map<String, dynamic> _decode(http.Response res) {
    final raw = res.body.trim();
    if (raw.isEmpty) {
      throw ApiException('服务器返回空响应 (${res.statusCode})');
    }
    if (raw.startsWith('<')) {
      throw ApiException('接口异常 (${res.statusCode})，请确认后端 :8080 已启动');
    }
    final body = jsonDecode(raw);
    if (res.statusCode >= 400) {
      final msg = body is Map ? (body['error'] ?? body['reject_reason'] ?? res.body) : res.body;
      throw ApiException(msg.toString());
    }
    return body as Map<String, dynamic>;
  }

  Future<bool> health() async {
    final j = await _get('/health');
    return j['mysql'] == true;
  }

  Future<AuthResult> register(String username, String password, {String? nickname}) async {
    final j = await _post('/api/v1/auth/register', {
      'username': username,
      'password': password,
      if (nickname != null && nickname.isNotEmpty) 'nickname': nickname,
    });
    final auth = AuthResult.fromJson(j);
    setToken(auth.token);
    return auth;
  }

  Future<AuthResult> login(String username, String password) async {
    final j = await _post('/api/v1/auth/login', {
      'username': username,
      'password': password,
    });
    final auth = AuthResult.fromJson(j);
    setToken(auth.token);
    return auth;
  }

  Future<SimUser> me() async {
    final j = await _get('/api/v1/auth/me');
    return SimUser.fromJson(j['user'] as Map<String, dynamic>? ?? {});
  }

  Future<Quote> quote(String code) async {
    final j = await _get('/api/v1/market/quote', {'code': code});
    return Quote.fromJson(j);
  }

  Future<List<StockBrief>> listStocks({int page = 1, int size = 20}) async {
    final j = await _get('/api/v1/market/stocks', {
      'page': '$page',
      'size': '$size',
    });
    return (j['items'] as List? ?? [])
        .map((e) => StockBrief.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<StockListPage> listLiveStocks({
    int page = 1,
    int size = 50,
    String board = 'all',
    String sort = 'change_pct',
    String? q,
  }) async {
    final qmap = {
      'page': '$page',
      'size': '$size',
      'board': board,
      'sort': sort,
    };
    if (q != null && q.isNotEmpty) qmap['q'] = q;
    final j = await _get('/api/v1/market/stocks', qmap);
    return StockListPage.fromJson(j);
  }

  Future<Map<String, dynamic>> boardStats() async => _get('/api/v1/market/boards');

  Future<SectorListPage> listSectors({int page = 1, int size = 12}) async {
    final j = await _get('/api/v1/market/sectors', {
      'page': '$page',
      'size': '$size',
      'type': 'industry',
    });
    return SectorListPage.fromJson(j);
  }

  Future<SectorStocksPage> listSectorStocks(String sectorCode, {int page = 1, int size = 50}) async {
    final j = await _get('/api/v1/market/sector/$sectorCode/stocks', {
      'page': '$page',
      'size': '$size',
    });
    return SectorStocksPage.fromJson(j);
  }

  Future<PerformanceReport> performance() async {
    final j = await _get('/api/v1/performance');
    return PerformanceReport.fromJson(j);
  }

  Future<List<ConditionalOrder>> listConditionalOrders({int limit = 20}) async {
    final j = await _get('/api/v1/conditional-orders', {'limit': '$limit'});
    return (j['items'] as List? ?? [])
        .map((e) => ConditionalOrder.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<ConditionalOrder> createConditionalOrder({
    required String code,
    required String conditionType,
    required double triggerValue,
    required String side,
    required String orderType,
    required int quantity,
    double? price,
    String? remark,
  }) async {
    final body = {
      'code': code,
      'condition_type': conditionType,
      'trigger_value': triggerValue,
      'side': side,
      'order_type': orderType,
      'quantity': quantity,
      'source': 'flutter',
      if (price != null) 'price': price,
      if (remark != null && remark.isNotEmpty) 'remark': remark,
    };
    final j = await _post('/api/v1/conditional-order', body);
    return ConditionalOrder.fromJson(j);
  }

  Future<void> cancelConditionalOrder(int id) async {
    final res = await _client.delete(
      _uri('/api/v1/conditional-order/$id'),
      headers: _headers(),
    );
    if (res.statusCode >= 400) {
      _decode(res);
    }
  }

  Future<List<StockBrief>> listStocksLegacy({int page = 1, int size = 20}) async {
    final j = await _get('/api/v1/market/list', {
      'page': '$page',
      'size': '$size',
    });
    return (j['items'] as List? ?? [])
        .map((e) => StockBrief.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<StockBrief>> search(String q) async {
    final j = await _get('/api/v1/market/search', {'q': q});
    return (j['items'] as List? ?? [])
        .map((e) => StockBrief.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Portfolio> portfolio() async {
    final j = await _get('/api/v1/portfolio');
    return Portfolio.fromJson(j);
  }

  Future<Account> account() async {
    final j = await _get('/api/v1/account');
    return Account.fromJson(j);
  }

  Future<List<TradeRecord>> trades({int limit = 30}) async {
    final j = await _get('/api/v1/trades', {'limit': '$limit'});
    return (j['items'] as List? ?? [])
        .map((e) => TradeRecord.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<OrderResult> placeOrder({
    required String code,
    required String side,
    required String orderType,
    required int quantity,
    double? price,
  }) async {
    final body = {
      'code': code,
      'side': side,
      'order_type': orderType,
      'quantity': quantity,
      'source': 'flutter',
    };
    if (price != null) body['price'] = price;
    final j = await _post('/api/v1/ai/order', body);
    return OrderResult.fromJson(j);
  }

  Future<KlineData> kline(String code, {String period = 'day', int limit = 120}) async {
    final j = await _get('/api/v1/market/kline', {
      'code': code,
      'period': period,
      'limit': '$limit',
    });
    return KlineData.fromJson(j);
  }

  Future<void> settle() async {
    await _post('/api/v1/trade/settle', {});
  }

  Future<Map<String, dynamic>> marketAnalysis() async =>
      _get('/api/v1/analysis/market');

  Future<Map<String, dynamic>> portfolioAnalysis() async =>
      _get('/api/v1/analysis/portfolio');

  Future<Map<String, dynamic>> anomalies() async =>
      _get('/api/v1/analysis/anomalies');

  Future<Map<String, dynamic>> dailyReport() async =>
      _get('/api/v1/analysis/daily-report');

  Future<Map<String, dynamic>> predict(String code) async =>
      _get('/api/v1/analysis/predict', {'code': code});
}
