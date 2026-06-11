import 'package:flutter/material.dart';
import '../theme/glass_tokens.dart';

enum AppLayoutMode { mobile, desktop }

/// 宽屏 → 桌面布局；窄屏 → 手机布局（不再强制横屏）
class AppBreakpoints {
  const AppBreakpoints._();

  static const desktopMinWidth = GlassTokens.desktopMinWidth;
  static const contentMax = 1280.0;
  static const formMax = 480.0;
  static const narrowFormMax = 360.0;

  static AppLayoutMode layoutOf(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    return width >= desktopMinWidth ? AppLayoutMode.desktop : AppLayoutMode.mobile;
  }

  static bool isDesktop(BuildContext context) => layoutOf(context) == AppLayoutMode.desktop;
  static bool isMobile(BuildContext context) => !isDesktop(context);

  static EdgeInsets pagePadding(BuildContext context) {
    if (isDesktop(context)) {
      return const EdgeInsets.symmetric(horizontal: 28, vertical: 20);
    }
    return const EdgeInsets.symmetric(horizontal: 16, vertical: 12);
  }

  static EdgeInsets desktopContentPadding(BuildContext context) {
    return isDesktop(context)
        ? const EdgeInsets.fromLTRB(8, 0, 28, 20)
        : pagePadding(context);
  }

  static bool shouldExpandButtons(BuildContext context) => isMobile(context);

  static int gridColumns(BuildContext context, {int desktop = 2, int mobile = 1}) =>
      isDesktop(context) ? desktop : mobile;
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
