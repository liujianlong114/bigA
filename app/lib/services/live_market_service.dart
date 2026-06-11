import 'dart:async';
import 'dart:convert';

import 'package:web_socket_channel/web_socket_channel.dart';

import '../config/api_config.dart';
import '../models/models.dart';

typedef LiveQuotesHandler = void Function(Map<String, StockBrief> quotes, bool marketOpen);

/// 订阅后端全市场实时行情 WebSocket
class LiveMarketService {
  WebSocketChannel? _channel;
  StreamSubscription? _sub;
  final Map<String, StockBrief> _quotes = {};
  bool marketOpen = false;
  bool connected = false;

  Map<String, StockBrief> get quotes => Map.unmodifiable(_quotes);

  Future<void> connect(LiveQuotesHandler onUpdate) async {
    await disconnect();
    try {
      _channel = WebSocketChannel.connect(Uri.parse(ApiConfig.wsUrl));
      connected = true;
      _sub = _channel!.stream.listen(
        (raw) {
          final msg = jsonDecode(raw as String) as Map<String, dynamic>;
          final type = msg['type'] as String? ?? '';
          if (type == 'meta' || type == 'snapshot_meta') {
            marketOpen = msg['market_open'] == true;
            return;
          }
          if (type != 'quotes') return;
          marketOpen = msg['market_open'] == true;
          final data = msg['data'];
          if (data is! List) return;
          for (final item in data) {
            if (item is! Map<String, dynamic>) continue;
            final code = item['code']?.toString() ?? '';
            if (code.isEmpty) continue;
            _quotes[code] = StockBrief(
              code: code,
              name: item['name']?.toString() ?? '',
              price: _d(item['price']),
              changePct: _d(item['change_pct']),
            );
          }
          onUpdate(_quotes, marketOpen);
        },
        onError: (_) {
          connected = false;
        },
        onDone: () {
          connected = false;
        },
        cancelOnError: true,
      );
    } catch (_) {
      connected = false;
    }
  }

  Future<void> disconnect() async {
    await _sub?.cancel();
    _sub = null;
    await _channel?.sink.close();
    _channel = null;
    connected = false;
  }

  double _d(dynamic v) {
    if (v == null) return 0;
    if (v is num) return v.toDouble();
    return double.tryParse(v.toString()) ?? 0;
  }
}
