import 'package:flutter/material.dart';
import 'glass_surface.dart';

class GlassCard extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry? padding;
  final EdgeInsetsGeometry? margin;
  final VoidCallback? onTap;
  final double radius;

  const GlassCard({
    super.key,
    required this.child,
    this.padding,
    this.margin,
    this.onTap,
    this.radius = 20,
  });

  @override
  Widget build(BuildContext context) {
    return GlassSurface.card(
      padding: padding ?? const EdgeInsets.all(16),
      margin: margin ?? const EdgeInsets.only(bottom: 12),
      onTap: onTap,
      child: child,
    );
  }
}
