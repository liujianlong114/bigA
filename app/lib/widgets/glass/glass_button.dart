import 'package:flutter/material.dart';
import '../../animations/glass_transitions.dart';
import '../../layout/app_breakpoints.dart';
import '../../theme/app_theme.dart';
import '../../theme/glass_theme.dart';
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
  final double? minWidth;

  const GlassButton({
    super.key,
    required this.label,
    this.onPressed,
    this.variant = GlassButtonVariant.primary,
    this.icon,
    this.loading = false,
    this.expanded = false,
    this.padding,
    this.minWidth,
  });

  @override
  Widget build(BuildContext context) {
    final g = GlassTheme.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final primary = isDark ? AppColors.primaryDark : AppColors.primary;
    final fullWidth = expanded && AppBreakpoints.shouldExpandButtons(context);

    final (fill, textColor, border) = switch (variant) {
      GlassButtonVariant.primary => (
          primary.withValues(alpha: 0.88),
          Colors.white,
          primary.withValues(alpha: 0.2),
        ),
      GlassButtonVariant.danger => (
          AppColors.up.withValues(alpha: 0.88),
          Colors.white,
          AppColors.up.withValues(alpha: 0.2),
        ),
      GlassButtonVariant.secondary => (g.glassFillFlat, g.labelPrimary, g.glassBorder),
      GlassButtonVariant.ghost => (Colors.transparent, g.labelPrimary, Colors.transparent),
    };

    final useFrosted = variant == GlassButtonVariant.primary || variant == GlassButtonVariant.danger;

    final surface = GlassSurface(
      frosted: useFrosted,
      blur: useFrosted ? GlassTokens.blurLight : 0,
      radius: GlassTokens.radiusMd,
      tint: fill,
      borderColor: border,
      padding: padding ?? const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
      onTap: loading ? null : onPressed,
      child: Row(
        mainAxisSize: fullWidth ? MainAxisSize.max : MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          if (loading)
            SizedBox(
              width: 18,
              height: 18,
              child: CircularProgressIndicator(strokeWidth: 2, color: textColor),
            )
          else ...[
            if (icon != null) ...[Icon(icon, size: 18, color: textColor), const SizedBox(width: 8)],
            Text(label, style: TextStyle(color: textColor, fontWeight: FontWeight.w600, fontSize: 15)),
          ],
        ],
      ),
    );

    Widget child = minWidth != null
        ? ConstrainedBox(constraints: BoxConstraints(minWidth: minWidth!), child: surface)
        : surface;

    if (fullWidth) {
      child = SizedBox(width: double.infinity, child: child);
    }

    return IosTapScale(onTap: loading ? null : onPressed, child: child);
  }
}
