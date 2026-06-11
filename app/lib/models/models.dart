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
  final String board;
  final String boardName;

  StockBrief({
    required this.code,
    required this.name,
    required this.price,
    required this.changePct,
    this.board = '',
    this.boardName = '',
  });

  factory StockBrief.fromJson(Map<String, dynamic> j) => StockBrief(
        code: j['code'] ?? '',
        name: j['name'] ?? '',
        price: _d(j['price']),
        changePct: _d(j['change_pct']),
        board: j['board']?.toString() ?? '',
        boardName: j['board_name']?.toString() ?? '',
      );

  StockBrief copyWith({double? price, double? changePct}) => StockBrief(
        code: code,
        name: name,
        price: price ?? this.price,
        changePct: changePct ?? this.changePct,
        board: board,
        boardName: boardName,
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

class SimUser {
  final int id;
  final String username;
  final String nickname;

  SimUser({required this.id, required this.username, required this.nickname});

  factory SimUser.fromJson(Map<String, dynamic> j) => SimUser(
        id: (j['id'] as num?)?.toInt() ?? 0,
        username: j['username']?.toString() ?? '',
        nickname: j['nickname']?.toString() ?? '',
      );
}

class AuthResult {
  final String token;
  final SimUser user;
  final Account account;

  AuthResult({required this.token, required this.user, required this.account});

  factory AuthResult.fromJson(Map<String, dynamic> j) => AuthResult(
        token: j['token']?.toString() ?? '',
        user: SimUser.fromJson(j['user'] as Map<String, dynamic>? ?? {}),
        account: Account.fromJson(j['account'] as Map<String, dynamic>? ?? {}),
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

class SectorListPage {
  final String source;
  final List<SectorBrief> items;

  SectorListPage({required this.source, required this.items});

  factory SectorListPage.fromJson(Map<String, dynamic> j) => SectorListPage(
        source: j['source']?.toString() ?? '',
        items: (j['items'] as List? ?? [])
            .map((e) => SectorBrief.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class SectorBrief {
  final String code;
  final String name;
  final double changePct;

  SectorBrief({required this.code, required this.name, required this.changePct});

  factory SectorBrief.fromJson(Map<String, dynamic> j) => SectorBrief(
        code: j['code'] ?? '',
        name: j['name'] ?? '',
        changePct: _d(j['change_pct']),
      );
}

class SectorStocksPage {
  final String sectorCode;
  final String sectorName;
  final int total;
  final List<StockBrief> items;

  SectorStocksPage({
    required this.sectorCode,
    required this.sectorName,
    required this.total,
    required this.items,
  });

  factory SectorStocksPage.fromJson(Map<String, dynamic> j) => SectorStocksPage(
        sectorCode: j['sector_code']?.toString() ?? '',
        sectorName: j['sector_name']?.toString() ?? '',
        total: j['total'] ?? 0,
        items: (j['items'] as List? ?? [])
            .map((e) => StockBrief.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class ConditionalOrder {
  final int id;
  final String code;
  final String name;
  final String conditionType;
  final double triggerValue;
  final String side;
  final String orderType;
  final int quantity;
  final String status;
  final String? remark;

  ConditionalOrder({
    required this.id,
    required this.code,
    required this.name,
    required this.conditionType,
    required this.triggerValue,
    required this.side,
    required this.orderType,
    required this.quantity,
    required this.status,
    this.remark,
  });

  factory ConditionalOrder.fromJson(Map<String, dynamic> j) => ConditionalOrder(
        id: (j['id'] as num?)?.toInt() ?? 0,
        code: j['code']?.toString() ?? '',
        name: j['name']?.toString() ?? '',
        conditionType: j['condition_type']?.toString() ?? '',
        triggerValue: _d(j['trigger_value']),
        side: j['side']?.toString() ?? '',
        orderType: j['order_type']?.toString() ?? '',
        quantity: (j['quantity'] as num?)?.toInt() ?? 0,
        status: j['status']?.toString() ?? '',
        remark: j['remark']?.toString(),
      );

  String get conditionLabel {
    switch (conditionType) {
      case 'price_gte':
        return '价格 ≥ ${triggerValue.toStringAsFixed(2)}';
      case 'price_lte':
        return '价格 ≤ ${triggerValue.toStringAsFixed(2)}';
      case 'change_pct_gte':
        return '涨幅 ≥ ${triggerValue.toStringAsFixed(2)}%';
      case 'change_pct_lte':
        return '跌幅 ≤ ${triggerValue.toStringAsFixed(2)}%';
      default:
        return conditionType;
    }
  }
}

class PerformanceReport {
  final double totalReturnPct;
  final double benchmarkReturnPct;
  final double excessReturnPct;
  final double maxDrawdownPct;
  final double sharpeRatio;
  final double winRate;
  final int closedTrades;

  PerformanceReport({
    required this.totalReturnPct,
    required this.benchmarkReturnPct,
    required this.excessReturnPct,
    required this.maxDrawdownPct,
    required this.sharpeRatio,
    required this.winRate,
    required this.closedTrades,
  });

  factory PerformanceReport.fromJson(Map<String, dynamic> j) => PerformanceReport(
        totalReturnPct: _d(j['total_return_pct']),
        benchmarkReturnPct: _d(j['benchmark_return_pct']),
        excessReturnPct: _d(j['excess_return_pct']),
        maxDrawdownPct: _d(j['max_drawdown_pct']),
        sharpeRatio: _d(j['sharpe_ratio']),
        winRate: _d(j['win_rate']),
        closedTrades: (j['closed_trades'] as num?)?.toInt() ?? 0,
      );
}

class StockListPage {
  final int total;
  final int page;
  final int size;
  final String source;
  final List<StockBrief> items;

  StockListPage({
    required this.total,
    required this.page,
    required this.size,
    required this.source,
    required this.items,
  });

  factory StockListPage.fromJson(Map<String, dynamic> j) => StockListPage(
        total: j['total'] ?? 0,
        page: j['page'] ?? 1,
        size: j['size'] ?? 50,
        source: j['source'] ?? '',
        items: (j['items'] as List? ?? [])
            .map((e) => StockBrief.fromJson(e as Map<String, dynamic>))
            .toList(),
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
