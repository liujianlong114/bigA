import 'package:flutter/material.dart';

class AppColors {
  static const up = Color(0xFFE53935);
  static const down = Color(0xFF43A047);
  static const bg = Color(0xFFF5F6FA);
  static const card = Colors.white;
  static const primary = Color(0xFF1565C0);
}

ThemeData buildAppTheme() {
  final base = ThemeData(
    useMaterial3: true,
    colorScheme: ColorScheme.fromSeed(seedColor: AppColors.primary, brightness: Brightness.light),
    scaffoldBackgroundColor: AppColors.bg,
  );
  return base.copyWith(
    appBarTheme: const AppBarTheme(
      centerTitle: true,
      elevation: 0,
      backgroundColor: AppColors.primary,
      foregroundColor: Colors.white,
    ),
    cardTheme: CardThemeData(
      color: AppColors.card,
      elevation: 1,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: Colors.white,
      border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
    ),
  );
}

Color priceColor(double v) => v >= 0 ? AppColors.up : AppColors.down;

String fmtMoney(double v) => v.toStringAsFixed(2);

String fmtPct(double v) => '${v >= 0 ? '+' : ''}${v.toStringAsFixed(2)}%';
