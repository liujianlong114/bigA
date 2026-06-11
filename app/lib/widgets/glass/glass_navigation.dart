import 'package:flutter/material.dart';
import '../../animations/glass_transitions.dart';
import '../../theme/glass_theme.dart';
import '../../theme/glass_tokens.dart';
import 'glass_surface.dart';

class GlassNavItem {
  final IconData icon;
  final IconData? selectedIcon;
  final String label;

  const GlassNavItem({required this.icon, this.selectedIcon, required this.label});
}

/// 手机底部导航 — 磨砂玻璃
class GlassBottomNav extends StatelessWidget {
  final int selectedIndex;
  final ValueChanged<int> onSelected;
  final List<GlassNavItem> items;

  const GlassBottomNav({
    super.key,
    required this.selectedIndex,
    required this.onSelected,
    required this.items,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
      child: GlassSurface(
        blur: GlassTokens.blurHeavy,
        radius: GlassTokens.radiusXl,
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
        child: Row(
          children: [
            for (var i = 0; i < items.length; i++)
              Expanded(
                child: _NavButton(
                  item: items[i],
                  selected: selectedIndex == i,
                  onTap: () => onSelected(i),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _NavButton extends StatelessWidget {
  final GlassNavItem item;
  final bool selected;
  final VoidCallback onTap;

  const _NavButton({required this.item, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final g = GlassTheme.of(context);
    return IosTapScale(
      onTap: onTap,
      child: AnimatedContainer(
        duration: GlassTokens.quickDuration,
        curve: GlassTokens.iosEmphasized,
        padding: const EdgeInsets.symmetric(vertical: 8),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(GlassTokens.radiusMd),
          color: selected ? g.navSelectedBg : Colors.transparent,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              selected ? (item.selectedIcon ?? item.icon) : item.icon,
              size: 22,
              color: selected ? g.navSelected : g.navUnselected,
            ),
            const SizedBox(height: 2),
            Text(
              item.label,
              style: TextStyle(
                fontSize: 10,
                fontWeight: selected ? FontWeight.w600 : FontWeight.w500,
                color: selected ? g.navSelected : g.navUnselected,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// 桌面侧边栏
class GlassSidebar extends StatelessWidget {
  final int selectedIndex;
  final ValueChanged<int> onSelected;
  final List<GlassNavItem> items;
  final Widget? header;
  final Widget? footer;

  const GlassSidebar({
    super.key,
    required this.selectedIndex,
    required this.onSelected,
    required this.items,
    this.header,
    this.footer,
  });

  @override
  Widget build(BuildContext context) {
    return GlassSurface(
      blur: GlassTokens.blurHeavy,
      radius: GlassTokens.radiusXl,
      margin: const EdgeInsets.all(16),
      padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (header != null) ...[header!, const SizedBox(height: 16)],
          Expanded(
            child: ListView.builder(
              itemCount: items.length,
              itemBuilder: (context, i) {
                final item = items[i];
                final selected = selectedIndex == i;
                return Padding(
                  padding: const EdgeInsets.only(bottom: 6),
                  child: IosTapScale(
                    onTap: () => onSelected(i),
                    child: AnimatedContainer(
                      duration: GlassTokens.quickDuration,
                      curve: GlassTokens.iosEmphasized,
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(GlassTokens.radiusMd),
                        color: selected ? GlassTheme.of(context).navSelectedBg : Colors.transparent,
                        border: selected ? Border.all(color: GlassTheme.of(context).navSelected.withValues(alpha: 0.25)) : null,
                      ),
                      child: Row(
                        children: [
                          Icon(
                            selected ? (item.selectedIcon ?? item.icon) : item.icon,
                            color: selected ? GlassTheme.of(context).navSelected : GlassTheme.of(context).navUnselected,
                          ),
                          const SizedBox(width: 12),
                          Text(
                            item.label,
                            style: TextStyle(
                              fontWeight: selected ? FontWeight.w600 : FontWeight.w500,
                              color: selected ? GlassTheme.of(context).navSelected : GlassTheme.of(context).labelPrimary,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
          if (footer != null) footer!,
        ],
      ),
    );
  }
}
