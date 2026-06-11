import 'dart:ui';
import 'package:flutter/material.dart';
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
    return ClipRect(
      child: BackdropFilter(
        filter: GlassTokens.blurFilter(GlassTokens.blurHeavy),
        child: Container(
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.72),
            border: Border(bottom: BorderSide(color: Colors.white.withValues(alpha: 0.5))),
          ),
          child: SafeArea(
            bottom: false,
            child: SizedBox(
              height: 56,
              child: NavigationToolbar(
                leading: leading,
                middle: Text(
                  title,
                  style: const TextStyle(
                    fontSize: 17,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF0F172A),
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
