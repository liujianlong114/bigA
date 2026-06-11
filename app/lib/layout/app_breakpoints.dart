import 'package:flutter/material.dart';
import '../theme/glass_tokens.dart';

enum AppLayoutMode { mobile, desktop }

/// 宽屏 → 桌面布局；竖屏或窄屏 → 手机布局
class AppBreakpoints {
  const AppBreakpoints._();

  static AppLayoutMode layoutOf(BuildContext context) {
    final size = MediaQuery.sizeOf(context);
    final isWide = size.width >= GlassTokens.desktopMinWidth;
    final isLandscape = size.width > size.height;
    if (isWide && (isLandscape || size.width >= 1200)) {
      return AppLayoutMode.desktop;
    }
    return AppLayoutMode.mobile;
  }

  static bool isDesktop(BuildContext context) =>
      layoutOf(context) == AppLayoutMode.desktop;

  static bool isMobile(BuildContext context) => !isDesktop(context);

  static double contentMaxWidth(BuildContext context) {
    if (isDesktop(context)) return 1440;
    return double.infinity;
  }

  static EdgeInsets pagePadding(BuildContext context) {
    if (isDesktop(context)) {
      return const EdgeInsets.symmetric(horizontal: 24, vertical: 16);
    }
    return const EdgeInsets.symmetric(horizontal: 16, vertical: 12);
  }
}

class AdaptiveLayoutBuilder extends StatelessWidget {
  final Widget Function(BuildContext context, AppLayoutMode mode) builder;

  const AdaptiveLayoutBuilder({super.key, required this.builder});

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, _) => builder(context, AppBreakpoints.layoutOf(context)),
    );
  }
}
