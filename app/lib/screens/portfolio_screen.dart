import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';
import '../widgets/common.dart';

class PortfolioScreen extends StatelessWidget {
  const PortfolioScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    final p = state.portfolio;

    if (p == null) {
      return const Center(child: Text('暂无数据'));
    }

    return RefreshIndicator(
      onRefresh: state.refreshAll,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('账户总览', style: Theme.of(context).textTheme.titleLarge),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      SummaryTile(label: '总资产', value: fmtMoney(p.totalAssets)),
                      SummaryTile(label: '可用资金', value: fmtMoney(p.cash)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      SummaryTile(label: '持仓市值', value: fmtMoney(p.marketValue)),
                      SummaryTile(
                        label: '浮动盈亏',
                        value: fmtMoney(p.totalProfit),
                        valueColor: priceColor(p.totalProfit),
                      ),
                      SummaryTile(
                        label: '收益率',
                        value: fmtPct(p.totalProfitPct),
                        valueColor: priceColor(p.totalProfitPct),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 8),
          Text('持仓 (${p.positions.length})', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          if (p.positions.isEmpty)
            const Card(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Center(child: Text('暂无持仓，去「交易」页买入')),
              ),
            )
          else
            ...p.positions.map((pos) => Card(
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text('${pos.name} (${pos.code})',
                                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                            Chip(label: Text(pos.board), visualDensity: VisualDensity.compact),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Row(
                          children: [
                            SummaryTile(label: '持仓', value: '${pos.quantity}'),
                            SummaryTile(label: '可卖', value: '${pos.available}'),
                            SummaryTile(label: '成本', value: fmtMoney(pos.costPrice)),
                          ],
                        ),
                        const SizedBox(height: 4),
                        Row(
                          children: [
                            SummaryTile(label: '现价', value: fmtMoney(pos.marketPrice)),
                            SummaryTile(label: '市值', value: fmtMoney(pos.marketValue)),
                            SummaryTile(
                              label: '盈亏',
                              value: fmtMoney(pos.profit),
                              valueColor: priceColor(pos.profit),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                )),
        ],
      ),
    );
  }
}
