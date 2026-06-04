import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';
import '../widgets/common.dart';
import 'trade_screen.dart';
import 'portfolio_screen.dart';
import 'history_screen.dart';

class HomeShell extends StatefulWidget {
  const HomeShell({super.key});

  @override
  State<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends State<HomeShell> {
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    final pages = [
      const MarketScreen(),
      const TradeScreen(),
      const PortfolioScreen(),
      const HistoryScreen(),
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('bigA 模拟盘'),
        actions: [
          Icon(
            state.backendOk ? Icons.cloud_done : Icons.cloud_off,
            color: state.backendOk ? Colors.lightGreenAccent : Colors.orangeAccent,
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: state.loading ? null : () => state.refreshAll(),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: LoadingOverlay(
        show: state.loading,
        child: pages[_index],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: const [
          NavigationDestination(icon: Icon(Icons.show_chart), label: '行情'),
          NavigationDestination(icon: Icon(Icons.swap_horiz), label: '交易'),
          NavigationDestination(icon: Icon(Icons.account_balance_wallet), label: '持仓'),
          NavigationDestination(icon: Icon(Icons.receipt_long), label: '记录'),
        ],
      ),
    );
  }
}

class MarketScreen extends StatefulWidget {
  const MarketScreen({super.key});

  @override
  State<MarketScreen> createState() => _MarketScreenState();
}

class _MarketScreenState extends State<MarketScreen> {
  final _searchCtrl = TextEditingController();

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    final q = state.selectedQuote;

    return RefreshIndicator(
      onRefresh: state.refreshAll,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (state.error != null)
            Card(
              color: Colors.orange.shade50,
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Text('连接失败: ${state.error}\n请确认 Go 后端已启动 :8080'),
              ),
            ),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _searchCtrl,
                  decoration: const InputDecoration(
                    hintText: '代码/名称，如 600519',
                    prefixIcon: Icon(Icons.search),
                  ),
                  onSubmitted: (v) async {
                    final err = await state.searchAndSelect(v);
                    if (err != null && context.mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
                    }
                  },
                ),
              ),
              const SizedBox(width: 8),
              FilledButton(
                onPressed: () async {
                  final err = await state.searchAndSelect(_searchCtrl.text);
                  if (err != null && context.mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
                  }
                },
                child: const Text('查'),
              ),
            ],
          ),
          const SizedBox(height: 16),
          if (q != null) ...[
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text('${q.name} (${q.code})', style: Theme.of(context).textTheme.titleLarge),
                        Chip(label: Text(q.board), visualDensity: VisualDensity.compact),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        PriceText(value: q.price, fontSize: 32),
                        const SizedBox(width: 12),
                        PriceText(value: q.changePct, isPercent: true),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        SummaryTile(label: '昨收', value: fmtMoney(q.prevClose)),
                        SummaryTile(label: '涨停', value: fmtMoney(q.limitUp), valueColor: AppColors.up),
                        SummaryTile(label: '跌停', value: fmtMoney(q.limitDown), valueColor: AppColors.down),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        SummaryTile(label: '买一', value: fmtMoney(q.bid1)),
                        SummaryTile(label: '卖一', value: fmtMoney(q.ask1)),
                        SummaryTile(label: '来源', value: q.source),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ],
          const SizedBox(height: 8),
          Text('涨幅榜', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          ...state.marketList.map((s) => ListTile(
                tileColor: Colors.white,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                title: Text('${s.name} (${s.code})'),
                subtitle: Text(fmtMoney(s.price)),
                trailing: PriceText(value: s.changePct, isPercent: true),
                onTap: () => state.loadQuote(s.code),
              )),
        ],
      ),
    );
  }
}
