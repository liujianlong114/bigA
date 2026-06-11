import 'dart:ui';
import 'package:flutter/material.dart';

/// iOS 风格磨砂玻璃设计令牌
abstract final class GlassTokens {
  static const desktopMinWidth = 960.0;
  static const tabletMinWidth = 600.0;

  static const blurLight = 24.0;
  static const blurMedium = 32.0;
  static const blurHeavy = 48.0;

  static const radiusSm = 12.0;
  static const radiusMd = 16.0;
  static const radiusLg = 22.0;
  static const radiusXl = 28.0;

  static const glassFillLight = Color(0xBFFFFFFF);
  static const glassFillDark = Color(0x99232A36);
  static const glassBorder = Color(0x66FFFFFF);
  static const glassHighlight = Color(0x33FFFFFF);

  static const springDuration = Duration(milliseconds: 420);
  static const quickDuration = Duration(milliseconds: 280);
  static const slowDuration = Duration(milliseconds: 560);

  static const springCurve = Curves.easeOutCubic;
  static const iosEmphasized = Cubic(0.2, 0.9, 0.3, 1.0);
  static const iosDecelerate = Cubic(0.0, 0.0, 0.2, 1.0);

  static ImageFilter blurFilter([double sigma = blurMedium]) =>
      ImageFilter.blur(sigmaX: sigma, sigmaY: sigma);

  static BoxDecoration glassDecoration({
    double radius = radiusLg,
    Color? fill,
    Color? borderColor,
    double borderWidth = 1,
    List<BoxShadow>? shadows,
  }) {
    return BoxDecoration(
      borderRadius: BorderRadius.circular(radius),
      border: Border.all(color: borderColor ?? glassBorder, width: borderWidth),
      gradient: LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [
          (fill ?? glassFillLight).withValues(alpha: 0.92),
          (fill ?? glassFillLight).withValues(alpha: 0.72),
        ],
      ),
      boxShadow: shadows ??
          [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.06),
              blurRadius: 24,
              offset: const Offset(0, 8),
            ),
          ],
    );
  }
}

abstract final class AppGradients {
  static const background = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [
      Color(0xFFE8F0FE),
      Color(0xFFF3E8FF),
      Color(0xFFE0F7FA),
      Color(0xFFF5F7FA),
    ],
    stops: [0.0, 0.35, 0.7, 1.0],
  );

  static const meshAccent = RadialGradient(
    center: Alignment(0.8, -0.6),
    radius: 1.2,
    colors: [Color(0x403B82F6), Color(0x00000000)],
  );
}
