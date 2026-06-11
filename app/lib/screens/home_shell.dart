import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../animations/glass_transitions.dart';
import '../layout/app_breakpoints.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';
import '../widgets/components.dart';
import 'analysis_screen.dart';
import 'history_screen.dart';
import 'market_screen.dart';
import 'portfolio_screen.dart';
import 'trade_screen.dart';

const _navItems = [
  GlassNavItem(icon: Icons.show_chart_outlined, selectedIcon: Icons.show_chart, label: '行情'),
  GlassNavItem(icon: Icons.analytics_outlined, selectedIcon: Icons.analytics, label: '分析'),
  GlassNavItem(icon: Icons.swap_horiz_rounded, selectedIcon: Icons.swap_horiz, label: '交易'),
  GlassNavItem(icon: Icons.account_balance_wallet_outlined, selectedIcon: Icons.account_balance_wallet, label: '持仓'),
  GlassNavItem(icon: Icons.receipt_long_outlined, selectedIcon: Icons.receipt_long, label: '记录'),
];

class HomeShell extends StatefulWidget {
  const HomeShell({super.key});

  @override
  State<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends State<HomeShell> {
  int _index = 0;

  static const _pages = [
    MarketScreen(),
    AnalysisScreen(),
    TradeScreen(),
    PortfolioScreen(),
    HistoryScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();

    return AppBackground(
      child: AdaptiveLayoutBuilder(
        builder: (context, mode) {
          if (mode == AppLayoutMode.desktop) {
            return _DesktopShell(
              index: _index,
              onIndexChanged: (i) => setState(() => _index = i),
              state: state,
              child: IosAnimatedSwitcher(
                child: KeyedSubtree(key: ValueKey(_index), child: _pages[_index]),
              ),
            );
          }
          return _MobileShell(
            index: _index,
            onIndexChanged: (i) => setState(() => _index = i),
            state: state,
            child: IosAnimatedSwitcher(
              child: KeyedSubtree(key: ValueKey(_index), child: _pages[_index]),
            ),
          );
        },
      ),
    );
  }
}

class _MobileShell extends StatelessWidget {
  final int index;
  final ValueChanged<int> onIndexChanged;
  final AppState state;
  final Widget child;

  const _MobileShell({
    required this.index,
    required this.onIndexChanged,
    required this.state,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      extendBody: true,
      appBar: _buildAppBar(context, state),
      body: GlassLoadingOverlay(show: state.loading, child: child),
      bottomNavigationBar: GlassBottomNav(
        selectedIndex: index,
        onSelected: onIndexChanged,
        items: _navItems,
      ),
    );
  }
}

class _DesktopShell extends StatelessWidget {
  final int index;
  final ValueChanged<int> onIndexChanged;
  final AppState state;
  final Widget child;

  const _DesktopShell({
    required this.index,
    required this.onIndexChanged,
    required this.state,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      body: SafeArea(
        child: Row(
          children: [
            SizedBox(
              width: 240,
              child: GlassSidebar(
                selectedIndex: index,
                onSelected: onIndexChanged,
                items: _navItems,
                header: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'bigA',
                      style: TextStyle(fontSize: 22, fontWeight: FontWeight.w700, letterSpacing: -0.5),
                    ),
                    Text('模拟盘', style: TextStyle(color: Colors.black.withValues(alpha: 0.45), fontSize: 13)),
                  ],
                ),
                footer: _UserFooter(state: state),
              ),
            ),
            Expanded(
              child: Column(
                children: [
                  _DesktopTopBar(state: state, title: _navItems[index].label),
                  Expanded(
                    child: GlassLoadingOverlay(show: state.loading, child: child),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

PreferredSizeWidget _buildAppBar(BuildContext context, AppState state) {
  return GlassAppBar(
    title: 'bigA 模拟盘',
    actions: _AppActions(state: state).build(context),
  );
}

class _DesktopTopBar extends StatelessWidget {
  final AppState state;
  final String title;
  const _DesktopTopBar({required this.state, required this.title});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(0, 8, 16, 8),
      child: Row(
        children: [
          Text(title, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w600)),
          const Spacer(),
          ..._AppActions(state: state).build(context),
        ],
      ),
    );
  }
}

class _AppActions {
  final AppState state;
  const _AppActions({required this.state});

  List<Widget> build(BuildContext context) {
    return [
      if (state.liveConnected) StatusBadge.live(open: state.marketOpen),
      const SizedBox(width: 8),
      Icon(
        state.backendOk ? Icons.cloud_done_rounded : Icons.cloud_off_rounded,
        color: state.backendOk ? AppColors.down : const Color(0xFFFB923C),
        size: 20,
      ),
      IconButton(
        icon: const Icon(Icons.refresh_rounded),
        onPressed: state.loading ? null : () => state.refreshAll(),
      ),
      PopupMenuButton<String>(
        icon: const Icon(Icons.person_outline_rounded),
        onSelected: (v) async {
          if (v == 'logout') await state.logout();
        },
        itemBuilder: (_) => [
          if (state.isLoggedIn || state.isGuest)
            const PopupMenuItem(value: 'logout', child: Text('退出登录')),
        ],
      ),
      const SizedBox(width: 4),
    ];
  }
}

class _UserFooter extends StatelessWidget {
  final AppState state;
  const _UserFooter({required this.state});

  @override
  Widget build(BuildContext context) {
    final name = state.currentUser != null
        ? (state.currentUser!.nickname.isNotEmpty ? state.currentUser!.nickname : state.currentUser!.username)
        : (state.isGuest ? '游客' : '');
    if (name.isEmpty) return const SizedBox.shrink();
    return GlassSurface(
      radius: 14,
      padding: const EdgeInsets.all(12),
      child: Row(
        children: [
          CircleAvatar(
            radius: 16,
            backgroundColor: AppColors.primary.withValues(alpha: 0.15),
            child: Text(name.characters.first, style: const TextStyle(color: AppColors.primary, fontWeight: FontWeight.w600)),
          ),
          const SizedBox(width: 10),
          Expanded(child: Text(name, style: const TextStyle(fontWeight: FontWeight.w500), overflow: TextOverflow.ellipsis)),
        ],
      ),
    );
  }
}
