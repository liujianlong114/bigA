import 'dart:ui';
import 'package:flutter/material.dart';
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
    return Stack(
      children: [
        child,
        if (show)
          Positioned.fill(
            child: ClipRect(
              child: BackdropFilter(
                filter: GlassTokens.blurFilter(GlassTokens.blurLight),
                child: Container(
                  color: Colors.white.withValues(alpha: 0.35),
                  child: Center(
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 28, vertical: 24),
                      decoration: GlassTokens.glassDecoration(radius: GlassTokens.radiusLg),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const SizedBox(
                            width: 32,
                            height: 32,
                            child: CircularProgressIndicator(strokeWidth: 2.5, color: Color(0xFF2563EB)),
                          ),
                          if (message != null) ...[
                            const SizedBox(height: 12),
                            Text(message!, style: const TextStyle(color: Color(0xFF475569), fontSize: 14)),
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
