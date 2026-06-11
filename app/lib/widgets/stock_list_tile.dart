import 'package:flutter/material.dart';
import '../animations/glass_transitions.dart';
import '../theme/app_theme.dart';
import 'glass/glass_surface.dart';

class StockListTile extends StatelessWidget {
  final String code;
  final String name;
  final String subtitle;
  final double changePct;
  final double price;
  final VoidCallback? onTap;
  final int index;
  final bool compact;

  const StockListTile({
    super.key,
    required this.code,
    required this.name,
    required this.subtitle,
    required this.changePct,
    required this.price,
    this.onTap,
    this.index = 0,
    this.compact = false,
  });

  @override
  Widget build(BuildContext context) {
    return StaggeredFadeSlide(
      index: index,
      child: Padding(
        padding: EdgeInsets.only(bottom: compact ? 6 : 8),
        child: IosTapScale(
          onTap: onTap,
          child: GlassSurface(
            blur: 16,
            radius: compact ? 14 : 16,
            padding: EdgeInsets.symmetric(horizontal: compact ? 12 : 14, vertical: compact ? 10 : 12),
            onTap: onTap,
            child: Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        name,
                        style: TextStyle(
                          fontWeight: FontWeight.w600,
                          fontSize: compact ? 14 : 15,
                          color: const Color(0xFF0F172A),
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 2),
                      Text(
                        '$code · $subtitle · ${fmtMoney(price)}',
                        style: TextStyle(fontSize: 12, color: Colors.black.withValues(alpha: 0.45)),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
                PriceText(value: changePct, isPercent: true, fontSize: compact ? 14 : 15),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

// Re-export PriceText from common for tile
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
