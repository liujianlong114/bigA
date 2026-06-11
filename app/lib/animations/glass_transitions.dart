import 'dart:ui';
import 'package:flutter/material.dart';
import 'ios_motion.dart';

/// iOS 风格页面转场
class GlassPageRoute<T> extends PageRouteBuilder<T> {
  GlassPageRoute({required Widget page, RouteSettings? settings})
      : super(
          settings: settings,
          transitionDuration: IosMotion.slideIn,
          reverseTransitionDuration: IosMotion.fadeIn,
          pageBuilder: (context, animation, secondaryAnimation) => page,
          transitionsBuilder: (context, animation, secondaryAnimation, child) {
            final curved = CurvedAnimation(parent: animation, curve: IosMotion.emphasized);
            final fade = Tween<double>(begin: 0, end: 1).animate(curved);
            final slide = Tween<Offset>(begin: const Offset(0.04, 0), end: Offset.zero).animate(curved);
            final scale = Tween<double>(begin: 0.98, end: 1).animate(curved);
            return FadeTransition(
              opacity: fade,
              child: SlideTransition(
                position: slide,
                child: ScaleTransition(scale: scale, child: child),
              ),
            );
          },
        );
}

/// 内容切换时的 iOS 弹簧过渡
class IosAnimatedSwitcher extends StatelessWidget {
  final Widget child;
  final Duration duration;
  final Alignment alignment;

  const IosAnimatedSwitcher({
    super.key,
    required this.child,
    this.duration = IosMotion.tabSwitch,
    this.alignment = Alignment.center,
  });

  @override
  Widget build(BuildContext context) {
    return AnimatedSwitcher(
      duration: duration,
      switchInCurve: IosMotion.emphasized,
      switchOutCurve: IosMotion.decelerate,
      transitionBuilder: (child, animation) {
        final fade = CurvedAnimation(parent: animation, curve: IosMotion.emphasized);
        return FadeTransition(
          opacity: fade,
          child: SlideTransition(
            position: Tween<Offset>(begin: const Offset(0, 0.02), end: Offset.zero).animate(fade),
            child: ScaleTransition(
              scale: Tween<double>(begin: 0.98, end: 1).animate(fade),
              child: child,
            ),
          ),
        );
      },
      layoutBuilder: (current, previous) => Stack(
        alignment: alignment,
        children: [...previous, if (current != null) current],
      ),
      child: child,
    );
  }
}

/// 列表项交错入场
class StaggeredFadeSlide extends StatefulWidget {
  final int index;
  final Widget child;
  final Duration delayStep;

  const StaggeredFadeSlide({
    super.key,
    required this.index,
    required this.child,
    this.delayStep = const Duration(milliseconds: 40),
  });

  @override
  State<StaggeredFadeSlide> createState() => _StaggeredFadeSlideState();
}

class _StaggeredFadeSlideState extends State<StaggeredFadeSlide> with SingleTickerProviderStateMixin {
  late final AnimationController _c;
  late final Animation<double> _fade;
  late final Animation<Offset> _slide;

  @override
  void initState() {
    super.initState();
    _c = AnimationController(vsync: this, duration: IosMotion.slideIn);
    _fade = CurvedAnimation(parent: _c, curve: IosMotion.emphasized);
    _slide = Tween<Offset>(begin: const Offset(0, 0.04), end: Offset.zero).animate(_fade);
    Future.delayed(widget.delayStep * widget.index, () {
      if (mounted) _c.forward();
    });
  }

  @override
  void dispose() {
    _c.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FadeTransition(
      opacity: _fade,
      child: SlideTransition(position: _slide, child: widget.child),
    );
  }
}

/// 按压缩放反馈
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
      onTapDown: widget.onTap == null ? null : (_) => setState(() => _pressed = true),
      onTapUp: widget.onTap == null ? null : (_) => setState(() => _pressed = false),
      onTapCancel: widget.onTap == null ? null : () => setState(() => _pressed = false),
      onTap: widget.onTap,
      child: AnimatedScale(
        scale: _pressed ? widget.pressedScale : 1,
        duration: IosMotion.micro,
        curve: IosMotion.emphasized,
        child: widget.child,
      ),
    );
  }
}
