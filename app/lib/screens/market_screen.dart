import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../layout/app_breakpoints.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';
import '../widgets/components.dart';
import '../widgets/kline_chart.dart';

class MarketScreen extends StatefulWidget {
  const MarketScreen({super.key});

  @override
  State<MarketScreen> createState() => _MarketScreenState();
}

class _MarketScreenState extends State<MarketScreen> {
  final _searchCtrl = TextEditingController();
  final _scrollCtrl = ScrollController();
  final _listScrollCtrl = ScrollController();

  @override
  void initState() {
    super.initState();
    _scrollCtrl.addListener(_onScroll);
    _listScrollCtrl.addListener(_onListScroll);
  }

  void _onScroll() {
    if (_scrollCtrl.position.pixels >= _scrollCtrl.position.maxScrollExtent - 200) {
      context.read<AppState>().loadMoreStocks();
    }
  }

  void _onListScroll() {
    if (_listScrollCtrl.position.pixels >= _listScrollCtrl.position.maxScrollExtent - 120) {
      context.read<AppState>().loadMoreStocks();
    }
  }

  @override
  void dispose() {
    _scrollCtrl.dispose();
    _listScrollCtrl.dispose();
    _searchCtrl.dispose();
    super.dispose();
  }

  Future<void> _search(AppState state, String v) async {
    await state.searchStockList(v);
    if (v.trim().length >= 2) {
      final err = await state.searchAndSelect(v);
      if (err != null && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AdaptiveLayoutBuilder(
      builder: (context, mode) {
        if (mode == AppLayoutMode.desktop) {
          return _DesktopMarket(
            searchCtrl: _searchCtrl,
            listScrollCtrl: _listScrollCtrl,
            onSearch: _search,
          );
        }
        return _MobileMarket(
          searchCtrl: _searchCtrl,
          scrollCtrl: _scrollCtrl,
          onSearch: _search,
        );
      },
    );
  }
}

class _MobileMarket extends StatelessWidget {
  final TextEditingController searchCtrl;
  final ScrollController scrollCtrl;
  final Future<void> Function(AppState state, String v) onSearch;

  const _MobileMarket({
    required this.searchCtrl,
    required this.scrollCtrl,
    required this.onSearch,
  });

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    return RefreshIndicator(
      onRefresh: state.refreshAll,
      color: AppColors.primary,
      child: CustomScrollView(
        controller: scrollCtrl,
        slivers: [
          SliverToBoxAdapter(child: _MarketFilters(searchCtrl: searchCtrl, onSearch: onSearch, compact: false)),
          if (state.selectedQuote != null)
            SliverToBoxAdapter(child: _QuotePanel(padding: AppBreakpoints.pagePadding(context))),
          SliverPadding(
            padding: AppBreakpoints.pagePadding(context),
            sliver: _StockSliverList(compact: false),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 24)),
        ],
      ),
    );
  }
}

class _DesktopMarket extends StatelessWidget {
  final TextEditingController searchCtrl;
  final ScrollController listScrollCtrl;
  final Future<void> Function(AppState state, String v) onSearch;

  const _DesktopMarket({
    required this.searchCtrl,
    required this.listScrollCtrl,
    required this.onSearch,
  });

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    return Padding(
      padding: const EdgeInsets.fromLTRB(8, 0, 24, 16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            flex: 4,
            child: GlassCard(
              padding: const EdgeInsets.all(16),
              margin: EdgeInsets.zero,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  _MarketFilters(searchCtrl: searchCtrl, onSearch: onSearch, compact: true),
                  const SizedBox(height: 12),
                  Expanded(
                    child: RefreshIndicator(
                      onRefresh: state.refreshAll,
                      color: AppColors.primary,
                      child: _StockListView(scrollCtrl: listScrollCtrl, compact: true),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            flex: 6,
            child: _QuotePanel(padding: EdgeInsets.zero, sticky: true),
          ),
        ],
      ),
    );
  }
}

class _MarketFilters extends StatelessWidget {
  final TextEditingController searchCtrl;
  final Future<void> Function(AppState state, String v) onSearch;
  final bool compact;

  const _MarketFilters({required this.searchCtrl, required this.onSearch, required this.compact});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    return Padding(
      padding: compact ? EdgeInsets.zero : AppBreakpoints.pagePadding(context).copyWith(bottom: 0, top: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (state.error != null)
            GlassCard(
              child: Text('后端未连接: ${state.error}', style: const TextStyle(color: Color(0xFFEA580C))),
            ),
          if (state.marketWarning != null && state.error == null)
            GlassCard(
              child: Text('部分数据加载失败: ${state.marketWarning}', style: const TextStyle(color: Color(0xFFD97706))),
            ),
          Row(
            children: [
              Expanded(
                child: GlassTextField(
                  controller: searchCtrl,
                  hintText: '搜股票：名称或代码',
                  prefixIcon: Icons.search_rounded,
                  onSubmitted: (v) => onSearch(state, v),
                ),
              ),
              const SizedBox(width: 8),
              GlassButton(
                label: '搜索',
                icon: Icons.search,
                minWidth: 88,
                onPressed: () => onSearch(state, searchCtrl.text),
              ),
            ],
          ),
          const SizedBox(height: 12),
          const SectionHeader(title: '市场板块'),
          GlassChipRow(
            chips: state.boardNames.entries
                .map((e) => GlassChip(
                      label: '${e.value}${state.boardCounts[e.key] != null ? ' (${state.boardCounts[e.key]})' : ''}',
                      selected: state.stockBoard == e.key,
                      onPressed: state.loading ? null : () => state.setStockBoard(e.key),
                    ))
                .toList(),
          ),
          if (state.hotSectors.isNotEmpty) ...[
            const SizedBox(height: 10),
            SectionHeader(
              title: state.hotSectorsSource == 'leaderboard_fallback' ? '今日领涨' : '热门行业',
            ),
            GlassChipRow(
              chips: state.hotSectors
                  .map((s) => GlassChip(
                        label: '${s.name} ${s.changePct >= 0 ? '+' : ''}${s.changePct.toStringAsFixed(2)}%',
                        onPressed: () {
                          if (state.loading) return;
                          if (state.hotSectorsSource == 'leaderboard_fallback') {
                            state.loadQuote(s.code);
                            state.loadKline();
                          } else {
                            state.openSector(s);
                          }
                        },
                      ))
                  .toList(),
            ),
          ],
          if (state.activeSectorName != null)
            Row(
              children: [
                Expanded(child: Text('行业：${state.activeSectorName}（${state.stockTotal} 只）')),
                TextButton(onPressed: state.clearSectorView, child: const Text('返回全市场')),
              ],
            )
          else if (state.stockSearchQ != null)
            Row(
              children: [
                Expanded(child: Text('搜索：${state.stockSearchQ}')),
                TextButton(
                  onPressed: () {
                    searchCtrl.clear();
                    state.searchStockList('');
                  },
                  child: const Text('清除'),
                ),
              ],
            ),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                state.activeSectorName != null
                    ? '成分股 ${state.stockList.length}/${state.stockTotal}'
                    : '全市场 ${state.stockTotal} 只',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600),
              ),
              DropdownButton<String>(
                value: state.stockSort,
                underline: const SizedBox.shrink(),
                borderRadius: BorderRadius.circular(12),
                items: const [
                  DropdownMenuItem(value: 'change_pct', child: Text('涨幅↓')),
                  DropdownMenuItem(value: 'change_pct_asc', child: Text('跌幅↓')),
                  DropdownMenuItem(value: 'volume', child: Text('成交量')),
                  DropdownMenuItem(value: 'code', child: Text('代码')),
                ],
                onChanged: state.loading ? null : (v) { if (v != null) state.setStockSort(v); },
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _QuotePanel extends StatelessWidget {
  final EdgeInsets padding;
  final bool sticky;

  const _QuotePanel({required this.padding, this.sticky = false});

  Widget _klineHeader(BuildContext context, AppState state) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text('K 线', style: Theme.of(context).textTheme.titleMedium),
        Wrap(
          spacing: 6,
          children: [
            for (final p in ['day', 'week', 'month', 'm60'])
              GlassChip(
                label: {'day': '日', 'week': '周', 'month': '月', 'm60': '60分'}[p]!,
                selected: state.klinePeriod == p,
                onPressed: state.loading ? null : () => state.setKlinePeriod(p),
              ),
          ],
        ),
      ],
    );
  }

  Widget _quoteCard(BuildContext context, dynamic q) {
    return GlassCard(
      margin: EdgeInsets.zero,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Flexible(child: Text('${q.name} (${q.code})', style: Theme.of(context).textTheme.titleLarge)),
              StatusBadge(label: q.board, color: AppColors.secondary),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              PriceText(value: q.price, fontSize: AppBreakpoints.isDesktop(context) ? 40 : 36),
              const SizedBox(width: 12),
              PriceText(value: q.changePct, isPercent: true, fontSize: 18),
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
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    final q = state.selectedQuote;
    final desktop = sticky && AppBreakpoints.isDesktop(context);

    if (q == null) {
      return Padding(
        padding: padding,
        child: const GlassCard(
          margin: EdgeInsets.zero,
          child: Center(
            child: Padding(padding: EdgeInsets.all(48), child: Text('选择一只股票查看详情')),
          ),
        ),
      );
    }

    if (desktop) {
      return Padding(
        padding: padding,
        child: Column(
          children: [
            _quoteCard(context, q),
            const SizedBox(height: 12),
            Expanded(
              child: GlassCard(
                margin: EdgeInsets.zero,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    _klineHeader(context, state),
                    const SizedBox(height: 8),
                    Expanded(
                      child: LayoutBuilder(
                        builder: (context, c) => KlineChart(
                          bars: state.klineData?.bars ?? [],
                          height: c.maxHeight.clamp(200, 800),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      );
    }

    return Padding(
      padding: padding,
      child: Column(
        children: [
          _quoteCard(context, q),
          const SizedBox(height: 12),
          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _klineHeader(context, state),
                const SizedBox(height: 8),
                KlineChart(bars: state.klineData?.bars ?? [], height: 320),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _StockSliverList extends StatelessWidget {
  final bool compact;
  const _StockSliverList({required this.compact});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    return SliverList(
      delegate: SliverChildBuilderDelegate(
        (context, i) {
          if (i >= state.stockList.length) {
            return state.loadingMore
                ? const Padding(padding: EdgeInsets.all(16), child: Center(child: CircularProgressIndicator()))
                : const SizedBox.shrink();
          }
          final s = state.stockList[i];
          return StockListTile(
            index: i,
            compact: compact,
            code: s.code,
            name: s.name,
            subtitle: s.boardName.isNotEmpty ? s.boardName : s.board,
            price: s.price,
            changePct: s.changePct,
            onTap: () async {
              await state.loadQuote(s.code);
              await state.loadKline();
            },
          );
        },
        childCount: state.stockList.length + (state.loadingMore ? 1 : 0),
      ),
    );
  }
}

class _StockListView extends StatelessWidget {
  final ScrollController scrollCtrl;
  final bool compact;
  const _StockListView({required this.scrollCtrl, required this.compact});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    return ListView.builder(
      controller: scrollCtrl,
      itemCount: state.stockList.length + (state.loadingMore ? 1 : 0),
      itemBuilder: (context, i) {
        if (i >= state.stockList.length) {
          return const Padding(padding: EdgeInsets.all(16), child: Center(child: CircularProgressIndicator()));
        }
        final s = state.stockList[i];
        return StockListTile(
          index: i,
          compact: compact,
          code: s.code,
          name: s.name,
          subtitle: s.boardName.isNotEmpty ? s.boardName : s.board,
          price: s.price,
          changePct: s.changePct,
          onTap: () async {
            await state.loadQuote(s.code);
            await state.loadKline();
          },
        );
      },
    );
  }
}
