import 'package:flutter/material.dart';
import 'ios_motion.dart';

/// iOS 风格页面转场
class GlassPageRoute<T> extends PageRouteBuilder<T> {
  GlassPageRoute({required Widget page, super.settings})
      : super(
          transitionDuration: IosMotion.slideIn,
          reverseTransitionDuration: IosMotion.fadeOut,
          pageBuilder: (context, animation, secondaryAnimation) => page,
          transitionsBuilder: (context, animation, secondaryAnimation, child) {
            final curved = CurvedAnimation(parent: animation, curve: IosMotion.enter);
            return FadeTransition(opacity: curved, child: child);
          },
        );
}

/// Tab 切换 — 轻量淡入，避免双页叠加重绘
class IosAnimatedSwitcher extends StatelessWidget {
  final Widget child;
  final int activeIndex;

  const IosAnimatedSwitcher({
    super.key,
    required this.child,
    required this.activeIndex,
  });

  @override
  Widget build(BuildContext context) {
    return AnimatedSwitcher(
      duration: IosMotion.tabSwitch,
      switchInCurve: IosMotion.enter,
      switchOutCurve: IosMotion.exit,
      transitionBuilder: (child, animation) => FadeTransition(opacity: animation, child: child),
      child: KeyedSubtree(
        key: ValueKey(activeIndex),
        child: child,
      ),
    );
  }
}

/// 列表项入场 — 直接渲染，避免滚动卡顿
class StaggeredFadeSlide extends StatelessWidget {
  final int index;
  final Widget child;
  final Duration delayStep;

  const StaggeredFadeSlide({
    super.key,
    required this.index,
    required this.child,
    this.delayStep = IosMotion.staggerStep,
  });

  @override
  Widget build(BuildContext context) => child;
}

/// 按压缩放 — 无弹簧，低开销
class IosTapScale extends StatefulWidget {
  final Widget child;
  final VoidCallback? onTap;
  final double pressedScale;

  const IosTapScale({
    super.key,
    required this.child,
    this.onTap,
    this.pressedScale = 0.97,
  });

  @override
  State<IosTapScale> createState() => _IosTapScaleState();
}

class _IosTapScaleState extends State<IosTapScale> {
  bool _pressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      behavior: HitTestBehavior.opaque,
      onTapDown: widget.onTap == null ? null : (_) => setState(() => _pressed = true),
      onTapUp: widget.onTap == null ? null : (_) => setState(() => _pressed = false),
      onTapCancel: widget.onTap == null ? null : () => setState(() => _pressed = false),
      onTap: widget.onTap,
      child: AnimatedScale(
        scale: _pressed ? widget.pressedScale : 1,
        duration: IosMotion.micro,
        curve: Curves.easeOut,
        child: widget.child,
      ),
    );
  }
}
