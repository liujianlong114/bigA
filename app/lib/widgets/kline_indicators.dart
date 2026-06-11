import '../models/models.dart';

/// MA 简单移动平均
List<double?> calcMA(List<KlineBar> bars, int period) {
  final out = List<double?>.filled(bars.length, null);
  if (bars.length < period) return out;
  for (var i = period - 1; i < bars.length; i++) {
    var sum = 0.0;
    for (var j = i - period + 1; j <= i; j++) {
      sum += bars[j].close;
    }
    out[i] = sum / period;
  }
  return out;
}

/// MACD: DIF, DEA, histogram
class MacdPoint {
  final double? dif;
  final double? dea;
  final double? hist;
  MacdPoint({this.dif, this.dea, this.hist});
}

List<MacdPoint> calcMACD(List<KlineBar> bars, {int short = 12, int long = 26, int signal = 9}) {
  final closes = bars.map((b) => b.close).toList();
  final emaShort = _ema(closes, short);
  final emaLong = _ema(closes, long);
  final dif = List<double?>.filled(bars.length, null);
  for (var i = 0; i < bars.length; i++) {
    if (emaShort[i] != null && emaLong[i] != null) {
      dif[i] = emaShort[i]! - emaLong[i]!;
    }
  }
  final difValues = dif.map((e) => e ?? 0.0).toList();
  final deaRaw = _ema(difValues, signal);
  final out = <MacdPoint>[];
  for (var i = 0; i < bars.length; i++) {
    if (dif[i] == null || deaRaw[i] == null) {
      out.add(MacdPoint());
    } else {
      final h = (dif[i]! - deaRaw[i]!) * 2;
      out.add(MacdPoint(dif: dif[i], dea: deaRaw[i], hist: h));
    }
  }
  return out;
}

/// KDJ
class KdjPoint {
  final double? k;
  final double? d;
  final double? j;
  KdjPoint({this.k, this.d, this.j});
}

List<KdjPoint> calcKDJ(List<KlineBar> bars, {int period = 9}) {
  final out = <KdjPoint>[];
  double k = 50, d = 50;
  for (var i = 0; i < bars.length; i++) {
    if (i + 1 < period) {
      out.add(KdjPoint());
      continue;
    }
    var low = double.infinity;
    var high = 0.0;
    for (var j = i - period + 1; j <= i; j++) {
      if (bars[j].low < low) low = bars[j].low;
      if (bars[j].high > high) high = bars[j].high;
    }
    final rsv = high == low ? 50.0 : (bars[i].close - low) / (high - low) * 100;
    k = k * 2 / 3 + rsv / 3;
    d = d * 2 / 3 + k / 3;
    final j = 3 * k - 2 * d;
    out.add(KdjPoint(k: k, d: d, j: j));
  }
  return out;
}

List<double?> _ema(List<double> values, int period) {
  final out = List<double?>.filled(values.length, null);
  if (values.isEmpty || period <= 0) return out;
  final k = 2.0 / (period + 1);
  double? prev;
  for (var i = 0; i < values.length; i++) {
    if (i < period - 1) continue;
    if (prev == null) {
      var sum = 0.0;
      for (var j = 0; j < period; j++) {
        sum += values[j];
      }
      prev = sum / period;
    } else {
      prev = values[i] * k + prev * (1 - k);
    }
    out[i] = prev;
  }
  return out;
}

double maxVolume(List<KlineBar> bars) {
  if (bars.isEmpty) return 1;
  return bars.map((b) => b.volume.toDouble()).reduce((a, b) => a > b ? a : b).clamp(1, double.infinity);
}
