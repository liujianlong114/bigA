export 'stock_list_tile.dart' show PriceText;
import 'package:flutter/material.dart';

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
          Text(label, style: TextStyle(color: Colors.black.withValues(alpha: 0.45), fontSize: 12)),
          const SizedBox(height: 4),
          Text(
            value,
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: valueColor ?? const Color(0xFF0F172A),
              letterSpacing: -0.2,
            ),
          ),
        ],
      ),
    );
  }
}
