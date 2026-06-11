import 'package:flutter/material.dart';
import '../../animations/glass_transitions.dart';
import '../../theme/glass_tokens.dart';
import 'glass_surface.dart';

enum GlassButtonVariant { primary, secondary, ghost, danger }

class GlassButton extends StatelessWidget {
  final String label;
  final VoidCallback? onPressed;
  final GlassButtonVariant variant;
  final IconData? icon;
  final bool loading;
  final bool expanded;
  final EdgeInsetsGeometry? padding;

  const GlassButton({
    super.key,
    required this.label,
    this.onPressed,
    this.variant = GlassButtonVariant.primary,
    this.icon,
    this.loading = false,
    this.expanded = false,
    this.padding,
  });

  Color get _tint => switch (variant) {
        GlassButtonVariant.primary => const Color(0xCC2563EB),
        GlassButtonVariant.secondary => GlassTokens.glassFillLight,
        GlassButtonVariant.ghost => Colors.transparent,
        GlassButtonVariant.danger => const Color(0xCCDC2626),
      };

  Color get _textColor => switch (variant) {
        GlassButtonVariant.primary => Colors.white,
        GlassButtonVariant.danger => Colors.white,
        _ => const Color(0xFF1E293B),
      };

  @override
  Widget build(BuildContext context) {
    final child = GlassSurface(
      radius: GlassTokens.radiusMd,
      blur: GlassTokens.blurLight,
      tint: _tint,
      borderColor: variant == GlassButtonVariant.ghost ? Colors.transparent : null,
      padding: padding ?? const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
      onTap: loading ? null : onPressed,
      child: Row(
        mainAxisSize: expanded ? MainAxisSize.max : MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          if (loading)
            SizedBox(
              width: 18,
              height: 18,
              child: CircularProgressIndicator(strokeWidth: 2, color: _textColor),
            )
          else ...[
            if (icon != null) ...[Icon(icon, size: 18, color: _textColor), const SizedBox(width: 8)],
            Text(label, style: TextStyle(color: _textColor, fontWeight: FontWeight.w600, fontSize: 15)),
          ],
        ],
      ),
    );

    return IosTapScale(
      onTap: loading ? null : onPressed,
      child: expanded ? SizedBox(width: double.infinity, child: child) : child,
    );
  }
}
