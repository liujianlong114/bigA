import 'package:flutter/material.dart';
import '../models/models.dart';
import '../theme/app_theme.dart';

/// 简易 K 线图（蜡烛图）
class KlineChart extends StatelessWidget {
  final List<KlineBar> bars;
  final double height;

  const KlineChart({super.key, required this.bars, this.height = 220});

  @override
  Widget build(BuildContext context) {
    if (bars.isEmpty) {
      return SizedBox(
        height: height,
        child: const Center(child: Text('暂无 K 线数据')),
      );
    }
    return SizedBox(
      height: height,
      child: CustomPaint(
        painter: _KlinePainter(bars: bars),
        size: Size.infinite,
      ),
    );
  }
}

class _KlinePainter extends CustomPainter {
  final List<KlineBar> bars;
  _KlinePainter({required this.bars});

  @override
  void paint(Canvas canvas, Size size) {
    if (bars.isEmpty) return;
    final n = bars.length;
    final pad = 8.0;
    final w = (size.width - pad * 2) / n;
    double minP = bars.first.low;
    double maxP = bars.first.high;
    for (final b in bars) {
      if (b.low < minP) minP = b.low;
      if (b.high > maxP) maxP = b.high;
    }
    final range = (maxP - minP).clamp(0.01, double.infinity);
    double yOf(double p) => size.height - pad - (p - minP) / range * (size.height - pad * 2);

    for (var i = 0; i < n; i++) {
      final b = bars[i];
      final x = pad + i * w + w / 2;
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
      canvas.drawRect(Rect.fromLTWH(x - w * 0.3, top, w * 0.6, bodyH), fill);
    }
  }

  @override
  bool shouldRepaint(covariant _KlinePainter old) => old.bars != bars;
}
