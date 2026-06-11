import 'dart:ui';
import 'package:flutter/material.dart';
import 'glass_theme.dart';

/// iOS 风格磨砂玻璃设计令牌
abstract final class GlassTokens {
  static const desktopMinWidth = 960.0;
  static const tabletMinWidth = 600.0;

  /// Web 上 BackdropFilter 开销大，模糊值不宜过高
  static const blurLight = 18.0;
  static const blurMedium = 24.0;
  static const blurHeavy = 32.0;

  static const radiusSm = 12.0;
  static const radiusMd = 16.0;
  static const radiusLg = 22.0;
  static const radiusXl = 28.0;

  static const springDuration = Duration(milliseconds: 420);
  static const quickDuration = Duration(milliseconds: 280);
  static const slowDuration = Duration(milliseconds: 560);

  static const springCurve = Curves.easeOutCubic;
  static const iosEmphasized = Cubic(0.2, 0.9, 0.3, 1.0);
  static const iosDecelerate = Cubic(0.0, 0.0, 0.2, 1.0);

  static ImageFilter blurFilter([double sigma = blurMedium]) =>
      ImageFilter.blur(sigmaX: sigma, sigmaY: sigma);

  static BoxDecoration glassDecoration(
    BuildContext context, {
    double radius = radiusLg,
    Color? fill,
    Color? fillTop,
    Color? fillBottom,
    Color? borderColor,
    double borderWidth = 1,
    bool frosted = true,
  }) {
    final g = GlassTheme.of(context);
    final effectiveBorder = borderColor ?? g.glassBorder;
    final surfaceFill = fill ?? fillTop ?? (frosted ? g.glassFillTop : g.glassFillFlat);

    return BoxDecoration(
      borderRadius: BorderRadius.circular(radius),
      border: Border.all(color: effectiveBorder, width: 0.5),
      color: surfaceFill,
    );
  }
}
