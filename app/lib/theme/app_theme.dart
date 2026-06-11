import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'glass_theme.dart';
import 'glass_tokens.dart';

class AppColors {
  static const up = Color(0xFFFF453A);
  static const down = Color(0xFF30D158);
  static const primary = Color(0xFF007AFF);
  static const primaryDark = Color(0xFF0A84FF);
  static const secondary = Color(0xFF5856D6);
  static const label = Color(0xFF3C3C43);
  static const labelSecondary = Color(0x993C3C43);
}

ThemeData buildAppTheme({Brightness brightness = Brightness.light}) {
  final isDark = brightness == Brightness.dark;
  final glass = isDark ? GlassTheme.dark : GlassTheme.light;
  final primary = isDark ? AppColors.primaryDark : AppColors.primary;

  final baseText = TextTheme(
    headlineMedium: TextStyle(color: glass.labelPrimary, fontWeight: FontWeight.w700, fontSize: 28),
    titleLarge: TextStyle(color: glass.labelPrimary, fontWeight: FontWeight.w600, fontSize: 22),
    titleMedium: TextStyle(color: glass.labelPrimary, fontWeight: FontWeight.w600, fontSize: 17),
    titleSmall: TextStyle(color: glass.labelPrimary, fontWeight: FontWeight.w600, fontSize: 14),
    bodyLarge: TextStyle(color: glass.labelPrimary, fontSize: 16),
    bodyMedium: TextStyle(color: glass.labelPrimary, fontSize: 14),
    bodySmall: TextStyle(color: glass.labelSecondary, fontSize: 12),
    labelLarge: TextStyle(color: glass.labelPrimary, fontSize: 14, fontWeight: FontWeight.w600),
  );

  return ThemeData(
    useMaterial3: true,
    brightness: brightness,
    scaffoldBackgroundColor: Colors.transparent,
    extensions: [glass],
    colorScheme: ColorScheme(
      brightness: brightness,
      primary: primary,
      onPrimary: Colors.white,
      secondary: AppColors.secondary,
      onSecondary: Colors.white,
      error: AppColors.up,
      onError: Colors.white,
      surface: isDark ? const Color(0x1FFFFFFF) : const Color(0xFFF2F2F7),
      onSurface: glass.labelPrimary,
    ),
    fontFamily: '.AppleSystemUIFont',
    pageTransitionsTheme: const PageTransitionsTheme(
      builders: {
        TargetPlatform.android: CupertinoPageTransitionsBuilder(),
        TargetPlatform.iOS: CupertinoPageTransitionsBuilder(),
        TargetPlatform.macOS: CupertinoPageTransitionsBuilder(),
        TargetPlatform.linux: CupertinoPageTransitionsBuilder(),
        TargetPlatform.windows: CupertinoPageTransitionsBuilder(),
      },
    ),
    appBarTheme: AppBarTheme(
      elevation: 0,
      scrolledUnderElevation: 0,
      backgroundColor: Colors.transparent,
      foregroundColor: glass.labelPrimary,
      systemOverlayStyle: isDark ? SystemUiOverlayStyle.light : SystemUiOverlayStyle.dark,
    ),
    textTheme: baseText,
    listTileTheme: ListTileThemeData(
      titleTextStyle: TextStyle(color: glass.labelPrimary, fontSize: 16, fontWeight: FontWeight.w600),
      subtitleTextStyle: TextStyle(color: glass.labelSecondary, fontSize: 13),
      iconColor: glass.labelSecondary,
    ),
    cardTheme: CardThemeData(
      elevation: 0,
      color: Colors.transparent,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(GlassTokens.radiusLg)),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: false,
      border: InputBorder.none,
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      hintStyle: TextStyle(color: glass.labelSecondary),
      labelStyle: TextStyle(color: glass.labelSecondary),
    ),
    tabBarTheme: TabBarThemeData(
      labelColor: primary,
      unselectedLabelColor: glass.labelSecondary,
      indicatorColor: primary,
      dividerColor: Colors.transparent,
    ),
    dropdownMenuTheme: DropdownMenuThemeData(textStyle: TextStyle(color: glass.labelPrimary, fontSize: 14)),
    popupMenuTheme: PopupMenuThemeData(
      color: isDark ? const Color(0xCC1C1C1E) : Colors.white,
      textStyle: TextStyle(color: glass.labelPrimary),
    ),
    expansionTileTheme: ExpansionTileThemeData(
      textColor: glass.labelPrimary,
      collapsedTextColor: glass.labelPrimary,
      iconColor: glass.labelSecondary,
      collapsedIconColor: glass.labelSecondary,
    ),
    textButtonTheme: TextButtonThemeData(
      style: TextButton.styleFrom(foregroundColor: primary),
    ),
    snackBarTheme: SnackBarThemeData(
      behavior: SnackBarBehavior.floating,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(GlassTokens.radiusMd)),
      backgroundColor: isDark ? const Color(0xE61C1C1E) : const Color(0xE6232A36),
      contentTextStyle: TextStyle(color: isDark ? Colors.white : Colors.white),
    ),
    dividerColor: isDark ? Colors.white.withValues(alpha: 0.12) : Colors.black.withValues(alpha: 0.08),
    iconTheme: IconThemeData(color: glass.labelPrimary),
  );
}

ThemeData buildAppLightTheme() => buildAppTheme(brightness: Brightness.light);

ThemeData buildAppDarkTheme() => buildAppTheme(brightness: Brightness.dark);

Color priceColor(double v) => v >= 0 ? AppColors.up : AppColors.down;

String fmtMoney(double v) => v.toStringAsFixed(2);

String fmtPct(double v) => '${v >= 0 ? '+' : ''}${v.toStringAsFixed(2)}%';
