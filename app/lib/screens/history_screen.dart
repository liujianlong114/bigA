import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../animations/glass_transitions.dart';
import '../layout/responsive_page.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';
import '../widgets/components.dart';

class HistoryScreen extends StatelessWidget {
  const HistoryScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();

    if (state.trades.isEmpty) {
      return RefreshIndicator(
        onRefresh: state.refreshAll,
        color: AppColors.primary,
        child: ResponsivePage(
          child: const SizedBox(
            height: 200,
            child: GlassCard(
              margin: EdgeInsets.zero,
              child: Center(child: Text('暂无成交记录')),
            ),
          ),
        ),
      );
    }

    final tiles = state.trades.asMap().entries.map((e) {
      final i = e.key;
      final t = e.value;
      final isBuy = t.side == 'buy';
      return StaggeredFadeSlide(
        index: i,
        child: GlassCard(
          margin: EdgeInsets.zero,
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          child: ListTile(
            leading: CircleAvatar(
              backgroundColor: (isBuy ? AppColors.up : AppColors.down).withValues(alpha: 0.12),
              child: Icon(isBuy ? Icons.arrow_upward_rounded : Icons.arrow_downward_rounded,
                  color: isBuy ? AppColors.up : AppColors.down, size: 20),
            ),
            title: Text('${t.name} (${t.code})', style: const TextStyle(fontWeight: FontWeight.w600)),
            subtitle: Text('${t.tradeDate} · ${isBuy ? "买入" : "卖出"} ${t.quantity}股 @ ${fmtMoney(t.price)}'),
            trailing: Text(
              fmtMoney(t.amount),
              style: TextStyle(fontWeight: FontWeight.w600, color: isBuy ? AppColors.up : AppColors.down),
            ),
          ),
        ),
      );
    }).toList();

    return RefreshIndicator(
      onRefresh: state.refreshAll,
      color: AppColors.primary,
      child: ResponsivePage(
        scrollable: true,
        child: ResponsiveGrid(desktopColumns: 2, children: tiles),
      ),
    );
  }
}
