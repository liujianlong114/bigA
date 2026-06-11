import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../layout/app_breakpoints.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';
import '../widgets/components.dart';

class PortfolioScreen extends StatelessWidget {
  const PortfolioScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    final p = state.portfolio;
    final perf = state.performance;

    if (p == null) {
      return const Center(child: Text('暂无数据', style: TextStyle(color: Color(0xFF64748B))));
    }

    return RefreshIndicator(
      onRefresh: state.refreshAll,
      color: AppColors.primary,
      child: ListView(
        padding: AppBreakpoints.pagePadding(context),
        children: [
          GlassCard(
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
                    SummaryTile(label: '浮动盈亏', value: fmtMoney(p.totalProfit), valueColor: priceColor(p.totalProfit)),
                    SummaryTile(label: '收益率', value: fmtPct(p.totalProfitPct), valueColor: priceColor(p.totalProfitPct)),
                  ],
                ),
              ],
            ),
          ),
          if (perf != null)
            GlassCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('绩效分析（vs 沪深300）', style: Theme.of(context).textTheme.titleMedium),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      SummaryTile(label: '总收益', value: fmtPct(perf.totalReturnPct), valueColor: priceColor(perf.totalReturnPct)),
                      SummaryTile(label: '超额收益', value: fmtPct(perf.excessReturnPct), valueColor: priceColor(perf.excessReturnPct)),
                      SummaryTile(label: '最大回撤', value: fmtPct(-perf.maxDrawdownPct)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      SummaryTile(label: '夏普', value: perf.sharpeRatio.toStringAsFixed(2)),
                      SummaryTile(label: '胜率', value: fmtPct(perf.winRate)),
                      SummaryTile(label: '平仓笔数', value: '${perf.closedTrades}'),
                    ],
                  ),
                ],
              ),
            ),
          SectionHeader(title: '持仓 (${p.positions.length})'),
          if (p.positions.isEmpty)
            const GlassCard(child: Center(child: Padding(padding: EdgeInsets.all(24), child: Text('暂无持仓，去「交易」页买入'))))
          else
            ...p.positions.map((pos) => GlassCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text('${pos.name} (${pos.code})', style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
                          StatusBadge(label: pos.board, color: AppColors.secondary),
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
                          SummaryTile(label: '盈亏', value: fmtMoney(pos.profit), valueColor: priceColor(pos.profit)),
                        ],
                      ),
                    ],
                  ),
                )),
        ],
      ),
    );
  }
}
