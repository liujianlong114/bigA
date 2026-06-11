import 'package:flutter/material.dart';
import '../../theme/glass_theme.dart';

/// 全局渐变 + 彩色光斑背景，让磨砂玻璃有东西可模糊
class AppBackground extends StatelessWidget {
  final Widget child;
  final bool showOrbs;

  const AppBackground({super.key, required this.child, this.showOrbs = true});

  @override
  Widget build(BuildContext context) {
    final g = GlassTheme.of(context);
    return DecoratedBox(
      decoration: BoxDecoration(gradient: g.backgroundGradient),
      child: Stack(
        fit: StackFit.expand,
        children: [
          if (showOrbs) ..._buildOrbs(g.orbColors),
          child,
        ],
      ),
    );
  }

  List<Widget> _buildOrbs(List<Color> colors) {
    if (colors.length < 4) return const [];
    return [
      _Orb(color: colors[0], size: 420, top: -120, left: -100),
      _Orb(color: colors[1], size: 360, top: 160, right: -120),
      _Orb(color: colors[2], size: 300, bottom: 80, left: 0),
      _Orb(color: colors[3], size: 260, bottom: -80, right: 20),
    ];
  }
}

class _Orb extends StatelessWidget {
  final Color color;
  final double size;
  final double? top;
  final double? left;
  final double? right;
  final double? bottom;

  const _Orb({
    required this.color,
    required this.size,
    this.top,
    this.left,
    this.right,
    this.bottom,
  });

  @override
  Widget build(BuildContext context) {
    return Positioned(
      top: top,
      left: left,
      right: right,
      bottom: bottom,
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          gradient: RadialGradient(
            colors: [color, color.withValues(alpha: 0)],
          ),
        ),
      ),
    );
  }
}
