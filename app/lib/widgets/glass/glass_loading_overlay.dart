import 'dart:ui';
import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/glass_theme.dart';
import '../../theme/glass_tokens.dart';

class GlassLoadingOverlay extends StatelessWidget {
  final bool show;
  final Widget child;
  final String? message;

  const GlassLoadingOverlay({
    super.key,
    required this.show,
    required this.child,
    this.message,
  });

  @override
  Widget build(BuildContext context) {
    final g = GlassTheme.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Stack(
      children: [
        child,
        if (show)
          Positioned.fill(
            child: ClipRect(
              child: BackdropFilter(
                filter: GlassTokens.blurFilter(GlassTokens.blurMedium),
                child: Container(
                  color: Colors.white.withValues(alpha: isDark ? 0.08 : 0.25),
                  child: Center(
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 28, vertical: 24),
                      decoration: GlassTokens.glassDecoration(context, radius: GlassTokens.radiusLg),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          SizedBox(
                            width: 32,
                            height: 32,
                            child: CircularProgressIndicator(
                              strokeWidth: 2.5,
                              color: isDark ? AppColors.primaryDark : AppColors.primary,
                            ),
                          ),
                          if (message != null) ...[
                            const SizedBox(height: 12),
                            Text(message!, style: TextStyle(color: g.labelSecondary, fontSize: 14)),
                          ],
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }
}
