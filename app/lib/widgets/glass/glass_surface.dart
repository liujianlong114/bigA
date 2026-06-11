import 'dart:ui';
import 'package:flutter/material.dart';
import '../../theme/glass_theme.dart';
import '../../theme/glass_tokens.dart';

/// 磨砂玻璃表面 — 核心可复用组件
///
/// [frosted] 为 true 时使用 BackdropFilter（适合导航栏、大卡片，数量应少）；
/// 为 false 时使用纯色半透明（适合列表行、芯片等高频组件）。
class GlassSurface extends StatelessWidget {
  final Widget? child;
  final double blur;
  final double radius;
  final EdgeInsetsGeometry? padding;
  final EdgeInsetsGeometry? margin;
  final Color? tint;
  final Color? tintTop;
  final Color? tintBottom;
  final Color? borderColor;
  final double? width;
  final double? height;
  final VoidCallback? onTap;
  final bool frosted;

  const GlassSurface({
    super.key,
    this.child,
    this.blur = GlassTokens.blurMedium,
    this.radius = GlassTokens.radiusLg,
    this.padding,
    this.margin,
    this.tint,
    this.tintTop,
    this.tintBottom,
    this.borderColor,
    this.width,
    this.height,
    this.onTap,
    this.frosted = true,
  });

  factory GlassSurface.card({
    Key? key,
    required Widget child,
    EdgeInsetsGeometry? padding,
    EdgeInsetsGeometry? margin,
    VoidCallback? onTap,
  }) =>
      GlassSurface(
        key: key,
        frosted: true,
        blur: GlassTokens.blurMedium,
        padding: padding ?? const EdgeInsets.all(16),
        margin: margin,
        onTap: onTap,
        child: child,
      );

  factory GlassSurface.flat({
    Key? key,
    required Widget child,
    EdgeInsetsGeometry? padding,
    EdgeInsetsGeometry? margin,
    double radius = GlassTokens.radiusMd,
    VoidCallback? onTap,
  }) =>
      GlassSurface(
        key: key,
        frosted: false,
        blur: 0,
        radius: radius,
        padding: padding,
        margin: margin,
        onTap: onTap,
        child: child,
      );

  factory GlassSurface.chip({
    Key? key,
    required Widget child,
    bool selected = false,
    VoidCallback? onTap,
    required BuildContext context,
  }) {
    final g = GlassTheme.of(context);
    return GlassSurface(
      key: key,
      frosted: false,
      blur: 0,
      radius: GlassTokens.radiusMd,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      tint: selected ? g.navSelected.withValues(alpha: 0.22) : null,
      borderColor: selected ? g.navSelected.withValues(alpha: 0.35) : null,
      onTap: onTap,
      child: child,
    );
  }

  @override
  Widget build(BuildContext context) {
    final decoration = GlassTokens.glassDecoration(
      context,
      radius: radius,
      fill: tint ?? tintTop,
      fillTop: tintTop,
      fillBottom: tintBottom,
      borderColor: borderColor,
      frosted: frosted,
    );

    Widget panel = Container(
      width: width,
      height: height,
      padding: padding,
      decoration: decoration,
      child: child,
    );

    if (frosted && blur > 0) {
      panel = ClipRRect(
        borderRadius: BorderRadius.circular(radius),
        child: BackdropFilter(
          filter: GlassTokens.blurFilter(blur),
          child: panel,
        ),
      );
    } else {
      panel = ClipRRect(
        borderRadius: BorderRadius.circular(radius),
        child: panel,
      );
    }

    Widget wrapped = panel;
    if (onTap != null) {
      wrapped = Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(radius),
          splashColor: Colors.white24,
          highlightColor: Colors.white10,
          child: panel,
        ),
      );
    }

    if (margin != null) {
      wrapped = Padding(padding: margin!, child: wrapped);
    }
    return wrapped;
  }
}
