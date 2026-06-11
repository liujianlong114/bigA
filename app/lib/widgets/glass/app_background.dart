import 'package:flutter/material.dart';
import '../../theme/glass_tokens.dart';

/// 全局渐变网格背景
class AppBackground extends StatelessWidget {
  final Widget child;
  final bool showAccent;

  const AppBackground({super.key, required this.child, this.showAccent = true});

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: const BoxDecoration(gradient: AppGradients.background),
      child: Stack(
        fit: StackFit.expand,
        children: [
          if (showAccent)
            const DecoratedBox(decoration: BoxDecoration(gradient: AppGradients.meshAccent)),
          child,
        ],
      ),
    );
  }
}
