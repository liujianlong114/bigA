import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

class PriceText extends StatelessWidget {
  final double value;
  final double? fontSize;
  final bool isPercent;
  final FontWeight? weight;

  const PriceText({
    super.key,
    required this.value,
    this.fontSize,
    this.isPercent = false,
    this.weight,
  });

  @override
  Widget build(BuildContext context) {
    final text = isPercent ? fmtPct(value) : fmtMoney(value);
    return Text(
      text,
      style: TextStyle(
        color: priceColor(value),
        fontSize: fontSize ?? 16,
        fontWeight: weight ?? FontWeight.w600,
      ),
    );
  }
}

class SummaryTile extends StatelessWidget {
  final String label;
  final String value;
  final Color? valueColor;

  const SummaryTile({
    super.key,
    required this.label,
    required this.value,
    this.valueColor,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: TextStyle(color: Colors.grey[600], fontSize: 12)),
          const SizedBox(height: 4),
          Text(
            value,
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: valueColor ?? Colors.black87,
            ),
          ),
        ],
      ),
    );
  }
}

class LoadingOverlay extends StatelessWidget {
  final bool show;
  final Widget child;
  const LoadingOverlay({super.key, required this.show, required this.child});

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        child,
        if (show)
          const Positioned.fill(
            child: ColoredBox(
              color: Color(0x33FFFFFF),
              child: Center(child: CircularProgressIndicator()),
            ),
          ),
      ],
    );
  }
}
