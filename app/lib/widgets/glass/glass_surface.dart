import 'dart:ui';
import 'package:flutter/material.dart';
import '../../theme/glass_tokens.dart';

/// 磨砂玻璃表面 — 核心可复用组件
class GlassSurface extends StatelessWidget {
  final Widget? child;
  final double blur;
  final double radius;
  final EdgeInsetsGeometry? padding;
  final EdgeInsetsGeometry? margin;
  final Color? tint;
  final Color? borderColor;
  final double? width;
  final double? height;
  final Gradient? gradient;
  final VoidCallback? onTap;
  final bool interactive;

  const GlassSurface({
    super.key,
    this.child,
    this.blur = GlassTokens.blurMedium,
    this.radius = GlassTokens.radiusLg,
    this.padding,
    this.margin,
    this.tint,
    this.borderColor,
    this.width,
    this.height,
    this.gradient,
    this.onTap,
    this.interactive = false,
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
        padding: padding ?? const EdgeInsets.all(16),
        margin: margin,
        onTap: onTap,
        interactive: onTap != null,
        child: child,
      );

  factory GlassSurface.chip({
    Key? key,
    required Widget child,
    bool selected = false,
    VoidCallback? onTap,
  }) =>
      GlassSurface(
        key: key,
        blur: GlassTokens.blurLight,
        radius: GlassTokens.radiusMd,
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        tint: selected ? const Color(0xCC3B82F6) : GlassTokens.glassFillLight,
        borderColor: selected ? const Color(0x883B82F6) : GlassTokens.glassBorder,
        onTap: onTap,
        interactive: true,
        child: child,
      );

  @override
  Widget build(BuildContext context) {
    final content = ClipRRect(
      borderRadius: BorderRadius.circular(radius),
      child: BackdropFilter(
        filter: GlassTokens.blurFilter(blur),
        child: AnimatedContainer(
          duration: GlassTokens.quickDuration,
          curve: GlassTokens.iosEmphasized,
          width: width,
          height: height,
          padding: padding,
          decoration: gradient != null
              ? BoxDecoration(
                  borderRadius: BorderRadius.circular(radius),
                  gradient: gradient,
                  border: Border.all(color: borderColor ?? GlassTokens.glassBorder),
                )
              : GlassTokens.glassDecoration(
                  radius: radius,
                  fill: tint,
                  borderColor: borderColor,
                ),
          child: child,
        ),
      ),
    );

    Widget wrapped = content;
    if (onTap != null) {
      wrapped = Material(
        color: Colors.transparent,
        child: InkWell(onTap: onTap, borderRadius: BorderRadius.circular(radius), child: content),
      );
    }

    if (margin != null) {
      wrapped = Padding(padding: margin!, child: wrapped);
    }
    return wrapped;
  }
}
