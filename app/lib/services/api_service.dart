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
  ApiService({http.Client? client}) : _client = client ?? http.Client();

  Uri _uri(String path, [Map<String, String>? query]) =>
      Uri.parse('${ApiConfig.baseUrl}$path').replace(queryParameters: query);

  Future<Map<String, dynamic>> _get(String path, [Map<String, String>? q]) async {
    final res = await _client.get(_uri(path, q));
    return _decode(res);
  }

  Future<Map<String, dynamic>> _post(String path, Map<String, dynamic> body) async {
    final res = await _client.post(
      _uri(path),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode(body),
    );
    return _decode(res);
  }

  Map<String, dynamic> _decode(http.Response res) {
    final body = jsonDecode(res.body);
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

  Future<Quote> quote(String code) async {
    final j = await _get('/api/v1/market/quote', {'code': code});
    return Quote.fromJson(j);
  }

  Future<List<StockBrief>> listStocks({int page = 1, int size = 20}) async {
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
}
