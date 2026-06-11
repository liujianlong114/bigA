import 'package:flutter/material.dart';
import '../models/models.dart';
import '../theme/app_theme.dart';
import 'kline_indicators.dart';

/// K 线图：蜡烛 + MA + 成交量 / MACD / KDJ 副图，支持指标开关
class KlineChart extends StatefulWidget {
  final List<KlineBar> bars;
  final double height;

  const KlineChart({super.key, required this.bars, this.height = 420});

  @override
  State<KlineChart> createState() => _KlineChartState();
}

class _KlineChartState extends State<KlineChart> {
  bool showMA = true;
  bool showVolume = true;
  bool showMacd = true;
  bool showKdj = false;

  @override
  Widget build(BuildContext context) {
    if (widget.bars.isEmpty) {
      return SizedBox(
        height: widget.height,
        child: const Center(child: Text('暂无 K 线数据')),
      );
    }

    final subCount = (showVolume ? 1 : 0) + (showMacd ? 1 : 0) + (showKdj ? 1 : 0);
    final mainH = widget.height * (subCount == 0 ? 1.0 : 0.55);
    final subH = subCount > 0 ? (widget.height - mainH - 36) / subCount : 0.0;

    final ma5 = calcMA(widget.bars, 5);
    final ma10 = calcMA(widget.bars, 10);
    final ma20 = calcMA(widget.bars, 20);
    final ma60 = calcMA(widget.bars, 60);
    final macd = calcMACD(widget.bars);
    final kdj = calcKDJ(widget.bars);
    final maxVol = maxVolume(widget.bars);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          spacing: 6,
          runSpacing: 4,
          children: [
            _chip('MA', showMA, (v) => setState(() => showMA = v)),
            _chip('成交量', showVolume, (v) => setState(() => showVolume = v)),
            _chip('MACD', showMacd, (v) => setState(() => showMacd = v)),
            _chip('KDJ', showKdj, (v) => setState(() => showKdj = v)),
          ],
        ),
        const SizedBox(height: 6),
        SizedBox(
          height: mainH,
          width: double.infinity,
          child: CustomPaint(
            painter: _CandlePainter(
              bars: widget.bars,
              ma5: showMA ? ma5 : null,
              ma10: showMA ? ma10 : null,
              ma20: showMA ? ma20 : null,
              ma60: showMA ? ma60 : null,
            ),
          ),
        ),
        if (showVolume)
          SizedBox(
            height: subH,
            width: double.infinity,
            child: CustomPaint(
              painter: _VolumePainter(bars: widget.bars, maxVol: maxVol),
            ),
          ),
        if (showMacd)
          SizedBox(
            height: subH,
            width: double.infinity,
            child: CustomPaint(painter: _MacdPainter(points: macd)),
          ),
        if (showKdj)
          SizedBox(
            height: subH,
            width: double.infinity,
            child: CustomPaint(painter: _KdjPainter(points: kdj)),
          ),
      ],
    );
  }

  Widget _chip(String label, bool selected, ValueChanged<bool> onChanged) {
    return FilterChip(
      label: Text(label, style: const TextStyle(fontSize: 12)),
      selected: selected,
      onSelected: onChanged,
      visualDensity: VisualDensity.compact,
      padding: EdgeInsets.zero,
    );
  }
}

class _ChartScale {
  final double pad;
  final double barW;
  _ChartScale(Size size, int count) : pad = 8, barW = count > 0 ? (size.width - 16) / count : 0;
  double xOf(int i) => pad + i * barW + barW / 2;
}

class _CandlePainter extends CustomPainter {
  final List<KlineBar> bars;
  final List<double?>? ma5;
  final List<double?>? ma10;
  final List<double?>? ma20;
  final List<double?>? ma60;

  _CandlePainter({required this.bars, this.ma5, this.ma10, this.ma20, this.ma60});

  @override
  void paint(Canvas canvas, Size size) {
    if (bars.isEmpty) return;
    final scale = _ChartScale(size, bars.length);
    var minP = bars.first.low;
    var maxP = bars.first.high;
    for (final b in bars) {
      if (b.low < minP) minP = b.low;
      if (b.high > maxP) maxP = b.high;
    }
    for (final series in [ma5, ma10, ma20, ma60]) {
      if (series == null) continue;
      for (final v in series) {
        if (v != null) {
          if (v < minP) minP = v;
          if (v > maxP) maxP = v;
        }
      }
    }
    final range = (maxP - minP).clamp(0.01, double.infinity);
    double yOf(double p) => size.height - scale.pad - (p - minP) / range * (size.height - scale.pad * 2);

    void drawMA(List<double?>? data, Color color) {
      if (data == null) return;
      final path = Path();
      var started = false;
      for (var i = 0; i < data.length; i++) {
        final v = data[i];
        if (v == null) continue;
        final pt = Offset(scale.xOf(i), yOf(v));
        if (!started) {
          path.moveTo(pt.dx, pt.dy);
          started = true;
        } else {
          path.lineTo(pt.dx, pt.dy);
        }
      }
      canvas.drawPath(path, Paint()..color = color..strokeWidth = 1.2..style = PaintingStyle.stroke);
    }

    drawMA(ma5, Colors.orange);
    drawMA(ma10, Colors.blue);
    drawMA(ma20, Colors.purple);
    drawMA(ma60, Colors.teal);

    for (var i = 0; i < bars.length; i++) {
      final b = bars[i];
      final x = scale.xOf(i);
      final up = b.close >= b.open;
      final color = up ? AppColors.up : AppColors.down;
      final paint = Paint()
        ..color = color
        ..strokeWidth = 1.2
        ..style = PaintingStyle.stroke;
      final fill = Paint()..color = color;

      canvas.drawLine(Offset(x, yOf(b.high)), Offset(x, yOf(b.low)), paint);
      final top = yOf(up ? b.close : b.open);
      final bot = yOf(up ? b.open : b.close);
      final bodyH = (bot - top).clamp(1.0, size.height);
      canvas.drawRect(Rect.fromLTWH(x - scale.barW * 0.3, top, scale.barW * 0.6, bodyH), fill);
    }
  }

  @override
  bool shouldRepaint(covariant _CandlePainter old) =>
      old.bars != bars || old.ma5 != ma5 || old.ma10 != ma10;
}

class _VolumePainter extends CustomPainter {
  final List<KlineBar> bars;
  final double maxVol;
  _VolumePainter({required this.bars, required this.maxVol});

  @override
  void paint(Canvas canvas, Size size) {
    if (bars.isEmpty) return;
    final scale = _ChartScale(size, bars.length);
    final h = size.height - scale.pad * 2;
    for (var i = 0; i < bars.length; i++) {
      final b = bars[i];
      final x = scale.xOf(i);
      final up = b.close >= b.open;
      final color = (up ? AppColors.up : AppColors.down).withValues(alpha: 0.65);
      final barH = b.volume / maxVol * h;
      canvas.drawRect(
        Rect.fromLTWH(x - scale.barW * 0.35, size.height - scale.pad - barH, scale.barW * 0.7, barH),
        Paint()..color = color,
      );
    }
  }

  @override
  bool shouldRepaint(covariant _VolumePainter old) => old.bars != bars || old.maxVol != maxVol;
}

class _MacdPainter extends CustomPainter {
  final List<MacdPoint> points;
  _MacdPainter({required this.points});

  @override
  void paint(Canvas canvas, Size size) {
    if (points.isEmpty) return;
    final scale = _ChartScale(size, points.length);
    var minV = 0.0;
    var maxV = 0.0;
    var has = false;
    for (final p in points) {
      for (final v in [p.dif, p.dea, p.hist]) {
        if (v == null) continue;
        if (!has) {
          minV = maxV = v;
          has = true;
        } else {
          if (v < minV) minV = v;
          if (v > maxV) maxV = v;
        }
      }
    }
    if (!has) return;
    final range = (maxV - minV).clamp(0.001, double.infinity);
    final midY = size.height / 2;
    double yOf(double v) => midY - (v / range) * (size.height * 0.4);

    canvas.drawLine(Offset(0, midY), Offset(size.width, midY), Paint()..color = Colors.grey.shade300);

    for (var i = 0; i < points.length; i++) {
      final h = points[i].hist;
      if (h == null) continue;
      final x = scale.xOf(i);
      final color = h >= 0 ? AppColors.up.withValues(alpha: 0.7) : AppColors.down.withValues(alpha: 0.7);
      final top = h >= 0 ? yOf(h) : midY;
      final bot = h >= 0 ? midY : yOf(h);
      canvas.drawRect(
        Rect.fromLTWH(x - scale.barW * 0.3, top, scale.barW * 0.6, (bot - top).abs().clamp(1, size.height)),
        Paint()..color = color,
      );
    }

    void drawLine(double? Function(MacdPoint) pick, Color color) {
      final path = Path();
      var started = false;
      for (var i = 0; i < points.length; i++) {
        final v = pick(points[i]);
        if (v == null) continue;
        final pt = Offset(scale.xOf(i), yOf(v));
        if (!started) {
          path.moveTo(pt.dx, pt.dy);
          started = true;
        } else {
          path.lineTo(pt.dx, pt.dy);
        }
      }
      canvas.drawPath(path, Paint()..color = color..strokeWidth = 1..style = PaintingStyle.stroke);
    }

    drawLine((p) => p.dif, Colors.orange);
    drawLine((p) => p.dea, Colors.blue);
  }

  @override
  bool shouldRepaint(covariant _MacdPainter old) => old.points != points;
}

class _KdjPainter extends CustomPainter {
  final List<KdjPoint> points;
  _KdjPainter({required this.points});

  @override
  void paint(Canvas canvas, Size size) {
    if (points.isEmpty) return;
    final scale = _ChartScale(size, points.length);
    const minV = 0.0;
    const maxV = 100.0;
    double yOf(double v) => size.height - scale.pad - (v - minV) / (maxV - minV) * (size.height - scale.pad * 2);

    for (final entry in [
      (Colors.orange, (KdjPoint p) => p.k),
      (Colors.blue, (KdjPoint p) => p.d),
      (Colors.purple, (KdjPoint p) => p.j),
    ]) {
      final path = Path();
      var started = false;
      for (var i = 0; i < points.length; i++) {
        final v = entry.$2(points[i]);
        if (v == null) continue;
        final pt = Offset(scale.xOf(i), yOf(v));
        if (!started) {
          path.moveTo(pt.dx, pt.dy);
          started = true;
        } else {
          path.lineTo(pt.dx, pt.dy);
        }
      }
      canvas.drawPath(path, Paint()..color = entry.$1..strokeWidth = 1..style = PaintingStyle.stroke);
    }
  }

  @override
  bool shouldRepaint(covariant _KdjPainter old) => old.points != points;
}
