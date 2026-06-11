class Quote {
  final String code;
  final String name;
  final double price;
  final double changePct;
  final double prevClose;
  final double limitUp;
  final double limitDown;
  final String board;
  final double bid1;
  final double ask1;
  final String source;

  Quote({
    required this.code,
    required this.name,
    required this.price,
    required this.changePct,
    required this.prevClose,
    required this.limitUp,
    required this.limitDown,
    required this.board,
    required this.bid1,
    required this.ask1,
    required this.source,
  });

  factory Quote.fromJson(Map<String, dynamic> j) => Quote(
        code: j['code'] ?? '',
        name: j['name'] ?? '',
        price: _d(j['price']),
        changePct: _d(j['change_pct']),
        prevClose: _d(j['prev_close']),
        limitUp: _d(j['limit_up']),
        limitDown: _d(j['limit_down']),
        board: j['board'] ?? 'main',
        bid1: _d(j['bid1']),
        ask1: _d(j['ask1']),
        source: j['source'] ?? '',
      );

  bool get isUp => changePct >= 0;
}

class StockBrief {
  final String code;
  final String name;
  final double price;
  final double changePct;

  StockBrief({
    required this.code,
    required this.name,
    required this.price,
    required this.changePct,
  });

  factory StockBrief.fromJson(Map<String, dynamic> j) => StockBrief(
        code: j['code'] ?? '',
        name: j['name'] ?? '',
        price: _d(j['price']),
        changePct: _d(j['change_pct']),
      );
}

class Account {
  final double cash;
  final double frozenCash;

  Account({required this.cash, required this.frozenCash});

  factory Account.fromJson(Map<String, dynamic> j) => Account(
        cash: _d(j['cash']),
        frozenCash: _d(j['frozen_cash']),
      );
}

class Position {
  final String code;
  final String name;
  final String board;
  final int quantity;
  final int available;
  final double costPrice;
  final double marketPrice;
  final double marketValue;
  final double profit;
  final double profitPct;
  final double changePct;

  Position({
    required this.code,
    required this.name,
    required this.board,
    required this.quantity,
    required this.available,
    required this.costPrice,
    required this.marketPrice,
    required this.marketValue,
    required this.profit,
    required this.profitPct,
    required this.changePct,
  });

  factory Position.fromJson(Map<String, dynamic> j) => Position(
        code: j['code'] ?? '',
        name: j['name'] ?? '',
        board: j['board'] ?? '',
        quantity: j['quantity'] ?? 0,
        available: j['available'] ?? 0,
        costPrice: _d(j['cost_price']),
        marketPrice: _d(j['market_price']),
        marketValue: _d(j['market_value']),
        profit: _d(j['profit']),
        profitPct: _d(j['profit_pct']),
        changePct: _d(j['change_pct']),
      );
}

class Portfolio {
  final double cash;
  final double marketValue;
  final double totalAssets;
  final double totalProfit;
  final double totalProfitPct;
  final List<Position> positions;

  Portfolio({
    required this.cash,
    required this.marketValue,
    required this.totalAssets,
    required this.totalProfit,
    required this.totalProfitPct,
    required this.positions,
  });

  factory Portfolio.fromJson(Map<String, dynamic> j) => Portfolio(
        cash: _d(j['cash']),
        marketValue: _d(j['market_value']),
        totalAssets: _d(j['total_assets']),
        totalProfit: _d(j['total_profit']),
        totalProfitPct: _d(j['total_profit_pct']),
        positions: (j['positions'] as List? ?? [])
            .map((e) => Position.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class OrderResult {
  final int orderId;
  final int? tradeId;
  final String code;
  final String name;
  final String side;
  final String status;
  final double? fillPrice;
  final int quantity;
  final String? rejectReason;

  OrderResult({
    required this.orderId,
    this.tradeId,
    required this.code,
    required this.name,
    required this.side,
    required this.status,
    this.fillPrice,
    required this.quantity,
    this.rejectReason,
  });

  factory OrderResult.fromJson(Map<String, dynamic> j) => OrderResult(
        orderId: j['order_id'] ?? 0,
        tradeId: j['trade_id'],
        code: j['code'] ?? '',
        name: j['name'] ?? '',
        side: j['side'] ?? '',
        status: j['status'] ?? '',
        fillPrice: j['fill_price'] != null ? _d(j['fill_price']) : null,
        quantity: j['quantity'] ?? 0,
        rejectReason: j['reject_reason'],
      );
}

class KlineBar {
  final String date;
  final double open;
  final double high;
  final double low;
  final double close;
  final int volume;

  KlineBar({
    required this.date,
    required this.open,
    required this.high,
    required this.low,
    required this.close,
    required this.volume,
  });

  factory KlineBar.fromJson(Map<String, dynamic> j) => KlineBar(
        date: j['date'] ?? '',
        open: _d(j['open']),
        high: _d(j['high']),
        low: _d(j['low']),
        close: _d(j['close']),
        volume: (j['volume'] as num?)?.toInt() ?? 0,
      );

  bool get isUp => close >= open;
}

class KlineData {
  final String code;
  final String name;
  final String period;
  final String source;
  final List<KlineBar> bars;

  KlineData({
    required this.code,
    required this.name,
    required this.period,
    required this.source,
    required this.bars,
  });

  factory KlineData.fromJson(Map<String, dynamic> j) => KlineData(
        code: j['code'] ?? '',
        name: j['name'] ?? '',
        period: j['period'] ?? 'day',
        source: j['source'] ?? '',
        bars: (j['bars'] as List? ?? [])
            .map((e) => KlineBar.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class TradeRecord {
  final int id;
  final String code;
  final String name;
  final String side;
  final double price;
  final int quantity;
  final double amount;
  final String tradeDate;

  TradeRecord({
    required this.id,
    required this.code,
    required this.name,
    required this.side,
    required this.price,
    required this.quantity,
    required this.amount,
    required this.tradeDate,
  });

  factory TradeRecord.fromJson(Map<String, dynamic> j) => TradeRecord(
        id: j['id'] ?? 0,
        code: j['code'] ?? '',
        name: j['name'] ?? '',
        side: j['side'] ?? '',
        price: _d(j['price']),
        quantity: j['quantity'] ?? 0,
        amount: _d(j['amount']),
        tradeDate: j['trade_date'] ?? '',
      );
}

double _d(dynamic v) {
  if (v == null) return 0;
  if (v is num) return v.toDouble();
  return double.tryParse(v.toString()) ?? 0;
}
