import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../providers/app_state.dart';
import '../../theme/glass_theme.dart';

/// 主题切换：亮色 / 暗色 / 跟随系统
class ThemeModeSwitch extends StatelessWidget {
  final bool compact;

  const ThemeModeSwitch({super.key, this.compact = false});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    final g = GlassTheme.of(context);

    if (compact) {
      return IconButton(
        tooltip: state.themeModeLabel,
        icon: Icon(_icon(state.themeMode), color: g.labelPrimary, size: 22),
        onPressed: () => state.cycleThemeMode(),
      );
    }

    return PopupMenuButton<ThemeMode>(
      tooltip: '外观：${state.themeModeLabel}',
      icon: Icon(_icon(state.themeMode), color: g.labelPrimary, size: 22),
      onSelected: state.setThemeMode,
      itemBuilder: (_) => [
        _item(ThemeMode.light, '亮色', Icons.light_mode_rounded, state.themeMode),
        _item(ThemeMode.dark, '暗色', Icons.dark_mode_rounded, state.themeMode),
        _item(ThemeMode.system, '跟随系统', Icons.brightness_auto_rounded, state.themeMode),
      ],
    );
  }

  IconData _icon(ThemeMode mode) => switch (mode) {
        ThemeMode.light => Icons.light_mode_rounded,
        ThemeMode.dark => Icons.dark_mode_rounded,
        ThemeMode.system => Icons.brightness_auto_rounded,
      };

  PopupMenuItem<ThemeMode> _item(ThemeMode mode, String label, IconData icon, ThemeMode current) {
    return PopupMenuItem(
      value: mode,
      child: Row(
        children: [
          Icon(icon, size: 20),
          const SizedBox(width: 10),
          Expanded(child: Text(label)),
          if (mode == current) const Icon(Icons.check_rounded, size: 18),
        ],
      ),
    );
  }
}
