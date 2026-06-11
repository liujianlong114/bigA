import 'package:flutter/material.dart';
import '../../theme/glass_tokens.dart';
import 'glass_surface.dart';

class GlassChip extends StatelessWidget {
  final String label;
  final bool selected;
  final VoidCallback? onPressed;
  final Color? selectedColor;
  final Widget? trailing;

  const GlassChip({
    super.key,
    required this.label,
    this.selected = false,
    this.onPressed,
    this.selectedColor,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    final accent = selectedColor ?? const Color(0xFF2563EB);
    return GlassSurface(
      blur: GlassTokens.blurLight,
      radius: GlassTokens.radiusMd,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      tint: selected ? accent.withValues(alpha: 0.85) : GlassTokens.glassFillLight,
      borderColor: selected ? accent.withValues(alpha: 0.5) : GlassTokens.glassBorder,
      onTap: onPressed,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            label,
            style: TextStyle(
              fontSize: 13,
              fontWeight: selected ? FontWeight.w600 : FontWeight.w500,
              color: selected ? Colors.white : const Color(0xFF334155),
            ),
          ),
          if (trailing != null) ...[const SizedBox(width: 4), trailing!],
        ],
      ),
    );
  }
}

class GlassChipRow extends StatelessWidget {
  final List<GlassChip> chips;
  final double spacing;

  const GlassChipRow({super.key, required this.chips, this.spacing = 8});

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(children: [
        for (var i = 0; i < chips.length; i++) ...[
          if (i > 0) SizedBox(width: spacing),
          chips[i],
        ],
      ]),
    );
  }
}
