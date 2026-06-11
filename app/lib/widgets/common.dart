export 'stock_list_tile.dart' show PriceText;
import 'package:flutter/material.dart';
import '../theme/glass_theme.dart';

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
    final g = GlassTheme.of(context);
    return Expanded(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: TextStyle(color: g.labelSecondary, fontSize: 12)),
          const SizedBox(height: 4),
          Text(
            value,
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w600,
              color: valueColor ?? g.labelPrimary,
              letterSpacing: -0.2,
            ),
          ),
        ],
      ),
    );
  }
}
