import 'package:flutter/material.dart';

/// 明暗模式下的磨砂玻璃主题扩展
class GlassTheme extends ThemeExtension<GlassTheme> {
  final Color glassFillTop;
  final Color glassFillBottom;
  /// 无 BackdropFilter 的轻量表面（列表行、芯片等）
  final Color glassFillFlat;
  final Color glassBorder;
  final Color glassBorderHighlight;
  final Color glassShadow;
  final Color labelPrimary;
  final Color labelSecondary;
  final Color navSelected;
  final Color navUnselected;
  final Color navSelectedBg;
  final LinearGradient backgroundGradient;
  final List<Color> orbColors;

  const GlassTheme({
    required this.glassFillTop,
    required this.glassFillBottom,
    required this.glassFillFlat,
    required this.glassBorder,
    required this.glassBorderHighlight,
    required this.glassShadow,
    required this.labelPrimary,
    required this.labelSecondary,
    required this.navSelected,
    required this.navUnselected,
    required this.navSelectedBg,
    required this.backgroundGradient,
    required this.orbColors,
  });

  static const light = GlassTheme(
    // 磨砂层：极低不透明度 + 背景模糊，透出壁纸色
    glassFillTop: Color(0x28FFFFFF),
    glassFillBottom: Color(0x28FFFFFF),
    glassFillFlat: Color(0x45FFFFFF),
    glassBorder: Color(0x59FFFFFF),
    glassBorderHighlight: Color(0x80FFFFFF),
    glassShadow: Color(0x00000000),
    labelPrimary: Color(0xFF1C1C1E),
    labelSecondary: Color(0x993C3C43),
    navSelected: Color(0xFF007AFF),
    navUnselected: Color(0xFF8E8E93),
    navSelectedBg: Color(0x22007AFF),
    backgroundGradient: LinearGradient(
      begin: Alignment.topLeft,
      end: Alignment.bottomRight,
      colors: [
        Color(0xFFC8D8FF),
        Color(0xFFE4D4FF),
        Color(0xFFD0F0EE),
        Color(0xFFEDF0F8),
      ],
      stops: [0.0, 0.38, 0.72, 1.0],
    ),
    orbColors: [
      Color(0x70007AFF),
      Color(0x608585FF),
      Color(0x50FF6482),
      Color(0x4540C4AA),
    ],
  );

  static const dark = GlassTheme(
    // 与亮色一致：低不透明度 + 背景模糊，透出暗色壁纸
    glassFillTop: Color(0x1FFFFFFF),
    glassFillBottom: Color(0x1FFFFFFF),
    glassFillFlat: Color(0x33FFFFFF),
    glassBorder: Color(0x33FFFFFF),
    glassBorderHighlight: Color(0x44FFFFFF),
    glassShadow: Color(0x00000000),
    labelPrimary: Color(0xFFFFFFFF),
    labelSecondary: Color(0x99EBEBF5),
    navSelected: Color(0xFF0A84FF),
    navUnselected: Color(0xFF8E8E93),
    navSelectedBg: Color(0x220A84FF),
    backgroundGradient: LinearGradient(
      begin: Alignment.topLeft,
      end: Alignment.bottomRight,
      colors: [
        Color(0xFF0C1224),
        Color(0xFF141028),
        Color(0xFF101A2E),
        Color(0xFF080810),
      ],
      stops: [0.0, 0.38, 0.72, 1.0],
    ),
    orbColors: [
      Color(0x700A84FF),
      Color(0x605856D6),
      Color(0x50FF375F),
      Color(0x4530D158),
    ],
  );

  static GlassTheme of(BuildContext context) {
    return Theme.of(context).extension<GlassTheme>() ?? light;
  }

  @override
  GlassTheme copyWith({
    Color? glassFillTop,
    Color? glassFillBottom,
    Color? glassFillFlat,
    Color? glassBorder,
    Color? glassBorderHighlight,
    Color? glassShadow,
    Color? labelPrimary,
    Color? labelSecondary,
    Color? navSelected,
    Color? navUnselected,
    Color? navSelectedBg,
    LinearGradient? backgroundGradient,
    List<Color>? orbColors,
  }) {
    return GlassTheme(
      glassFillTop: glassFillTop ?? this.glassFillTop,
      glassFillBottom: glassFillBottom ?? this.glassFillBottom,
      glassFillFlat: glassFillFlat ?? this.glassFillFlat,
      glassBorder: glassBorder ?? this.glassBorder,
      glassBorderHighlight: glassBorderHighlight ?? this.glassBorderHighlight,
      glassShadow: glassShadow ?? this.glassShadow,
      labelPrimary: labelPrimary ?? this.labelPrimary,
      labelSecondary: labelSecondary ?? this.labelSecondary,
      navSelected: navSelected ?? this.navSelected,
      navUnselected: navUnselected ?? this.navUnselected,
      navSelectedBg: navSelectedBg ?? this.navSelectedBg,
      backgroundGradient: backgroundGradient ?? this.backgroundGradient,
      orbColors: orbColors ?? this.orbColors,
    );
  }

  @override
  GlassTheme lerp(ThemeExtension<GlassTheme>? other, double t) {
    if (other is! GlassTheme) return this;
    return GlassTheme(
      glassFillTop: Color.lerp(glassFillTop, other.glassFillTop, t)!,
      glassFillBottom: Color.lerp(glassFillBottom, other.glassFillBottom, t)!,
      glassFillFlat: Color.lerp(glassFillFlat, other.glassFillFlat, t)!,
      glassBorder: Color.lerp(glassBorder, other.glassBorder, t)!,
      glassBorderHighlight: Color.lerp(glassBorderHighlight, other.glassBorderHighlight, t)!,
      glassShadow: Color.lerp(glassShadow, other.glassShadow, t)!,
      labelPrimary: Color.lerp(labelPrimary, other.labelPrimary, t)!,
      labelSecondary: Color.lerp(labelSecondary, other.labelSecondary, t)!,
      navSelected: Color.lerp(navSelected, other.navSelected, t)!,
      navUnselected: Color.lerp(navUnselected, other.navUnselected, t)!,
      navSelectedBg: Color.lerp(navSelectedBg, other.navSelectedBg, t)!,
      backgroundGradient: LinearGradient.lerp(backgroundGradient, other.backgroundGradient, t)!,
      orbColors: orbColors,
    );
  }
}
