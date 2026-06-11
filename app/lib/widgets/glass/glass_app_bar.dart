import 'dart:ui';
import 'package:flutter/material.dart';
import '../../theme/glass_theme.dart';
import '../../theme/glass_tokens.dart';

class GlassAppBar extends StatelessWidget implements PreferredSizeWidget {
  final String title;
  final List<Widget>? actions;
  final Widget? leading;
  final bool centerTitle;

  const GlassAppBar({
    super.key,
    required this.title,
    this.actions,
    this.leading,
    this.centerTitle = true,
  });

  @override
  Size get preferredSize => const Size.fromHeight(56);

  @override
  Widget build(BuildContext context) {
    final g = GlassTheme.of(context);
    return ClipRect(
      child: BackdropFilter(
        filter: GlassTokens.blurFilter(GlassTokens.blurMedium),
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: g.glassFillTop,
            border: Border(bottom: BorderSide(color: g.glassBorder.withValues(alpha: 0.5))),
          ),
          child: SafeArea(
            bottom: false,
            child: SizedBox(
              height: 56,
              child: NavigationToolbar(
                leading: leading,
                middle: Text(
                  title,
                  style: TextStyle(
                    fontSize: 17,
                    fontWeight: FontWeight.w600,
                    color: g.labelPrimary,
                    letterSpacing: -0.2,
                  ),
                ),
                centerMiddle: centerTitle,
                trailing: actions != null
                    ? Row(mainAxisSize: MainAxisSize.min, children: actions!)
                    : null,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
