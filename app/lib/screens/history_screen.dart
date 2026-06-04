import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';

class HistoryScreen extends StatelessWidget {
  const HistoryScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();

    return RefreshIndicator(
      onRefresh: state.refreshAll,
      child: state.trades.isEmpty
          ? ListView(
              children: const [
                SizedBox(height: 120),
                Center(child: Text('暂无成交记录')),
              ],
            )
          : ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: state.trades.length,
              separatorBuilder: (_, __) => const SizedBox(height: 8),
              itemBuilder: (_, i) {
                final t = state.trades[i];
                final isBuy = t.side == 'buy';
                return Card(
                  child: ListTile(
                    leading: CircleAvatar(
                      backgroundColor: isBuy ? AppColors.up.withValues(alpha: 0.15) : AppColors.down.withValues(alpha: 0.15),
                      child: Icon(isBuy ? Icons.arrow_upward : Icons.arrow_downward,
                          color: isBuy ? AppColors.up : AppColors.down),
                    ),
                    title: Text('${t.name} (${t.code})'),
                    subtitle: Text('${t.tradeDate} · ${isBuy ? "买入" : "卖出"} ${t.quantity}股 @ ${fmtMoney(t.price)}'),
                    trailing: Text(
                      fmtMoney(t.amount),
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: isBuy ? AppColors.up : AppColors.down,
                      ),
                    ),
                  ),
                );
              },
            ),
    );
  }
}
