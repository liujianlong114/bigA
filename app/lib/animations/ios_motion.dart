import 'package:flutter/material.dart';
import '../theme/glass_tokens.dart';

/// iOS 风格动效常量与曲线
abstract final class IosMotion {
  static const fadeIn = Duration(milliseconds: 320);
  static const slideIn = Duration(milliseconds: 420);
  static const tabSwitch = Duration(milliseconds: 380);
  static const micro = Duration(milliseconds: 180);

  static const emphasized = GlassTokens.iosEmphasized;
  static const decelerate = GlassTokens.iosDecelerate;
  static const standard = Curves.easeInOutCubic;

  static Animation<double> fade(AnimationController c) =>
      CurvedAnimation(parent: c, curve: emphasized);

  static Animation<Offset> slideUp(AnimationController c, {double dy = 0.08}) =>
      Tween<Offset>(begin: Offset(0, dy), end: Offset.zero).animate(
        CurvedAnimation(parent: c, curve: emphasized),
      );

  static Animation<double> scaleIn(AnimationController c) =>
      Tween<double>(begin: 0.96, end: 1).animate(
        CurvedAnimation(parent: c, curve: emphasized),
      );
}
