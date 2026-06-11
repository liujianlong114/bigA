import 'package:flutter/material.dart';
import '../theme/glass_tokens.dart';

/// iOS 风格动效常量与曲线
abstract final class IosMotion {
  static const fadeIn = Duration(milliseconds: 200);
  static const fadeOut = Duration(milliseconds: 160);
  static const tabSwitch = Duration(milliseconds: 220);
  static const slideIn = Duration(milliseconds: 280);
  static const switchGap = Duration.zero;
  static const micro = Duration(milliseconds: 80);
  static const staggerStep = Duration(milliseconds: 16);

  static const emphasized = GlassTokens.iosEmphasized;
  static const decelerate = GlassTokens.iosDecelerate;
  static const spring = Curves.easeOutCubic;
  static const enter = Curves.easeOutCubic;
  static const exit = Curves.easeInCubic;
}
