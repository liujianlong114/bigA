import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../layout/responsive_page.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';
import '../theme/glass_theme.dart';
import '../theme/glass_tokens.dart';
import '../widgets/components.dart';

class AnalysisScreen extends StatefulWidget {
  const AnalysisScreen({super.key});

  @override
  State<AnalysisScreen> createState() => _AnalysisScreenState();
}

class _AnalysisScreenState extends State<AnalysisScreen> with SingleTickerProviderStateMixin {
  late TabController _tabs;
  bool loading = false;
  String? error;
  Map<String, dynamic>? market;
  Map<String, dynamic>? portfolio;
  Map<String, dynamic>? anomalies;
  Map<String, dynamic>? report;
  Map<String, dynamic>? predict;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 4, vsync: this);
    _load();
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      loading = true;
      error = null;
    });
    final api = context.read<AppState>().api;
    try {
      final code = context.read<AppState>().selectedCode;
      final results = await Future.wait([
        api.marketAnalysis(),
        api.portfolioAnalysis(),
        api.anomalies(),
        api.dailyReport(),
        api.predict(code),
      ]);
      if (!mounted) return;
      setState(() {
        market = results[0];
        portfolio = results[1];
        anomalies = results[2];
        report = results[3];
        predict = results[4];
        loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        error = e.toString();
        loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final g = GlassTheme.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final primary = isDark ? AppColors.primaryDark : AppColors.primary;
    return Column(
      children: [
        GlassSurface.flat(
          radius: 0,
          padding: EdgeInsets.zero,
          child: TabBar(
            controller: _tabs,
            isScrollable: true,
            indicatorColor: primary,
            labelColor: primary,
            unselectedLabelColor: g.labelSecondary,
            tabs: const [
              Tab(text: '大盘'),
              Tab(text: '持仓'),
              Tab(text: '异动'),
              Tab(text: '复盘'),
            ],
          ),
        ),
        if (loading) LinearProgressIndicator(color: AppColors.primary, backgroundColor: AppColors.primary.withValues(alpha: 0.12)),
        if (error != null)
          Padding(
            padding: const EdgeInsets.all(12),
            child: Text(error!, style: const TextStyle(color: AppColors.up)),
          ),
        Expanded(
          child: TabBarView(
            controller: _tabs,
            children: [
              _marketTab(),
              _portfolioTab(),
              _anomaliesTab(),
              _reportTab(),
            ],
          ),
        ),
      ],
    );
  }

  Widget _marketTab() {
    if (market == null) return const Center(child: Text('加载中...'));
    final breadth = market!['breadth'] as Map<String, dynamic>? ?? {};
    final sentiment = market!['sentiment'] as Map<String, dynamic>? ?? {};
    final indices = market!['indices'] as List? ?? [];
    return RefreshIndicator(
      onRefresh: _load,
      color: AppColors.primary,
      child: ResponsivePage(
        scrollable: true,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            ResponsiveGrid(
              desktopColumns: 2,
              children: [
                GlassCard(
                  margin: EdgeInsets.zero,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('市场宽度', style: Theme.of(context).textTheme.titleMedium),
                      const SizedBox(height: 8),
                      Text('上涨 ${breadth['up']} / 下跌 ${breadth['down']} / 平盘 ${breadth['flat']}'),
                      Text('涨跌比 ${breadth['up_down_ratio']}'),
                    ],
                  ),
                ),
                GlassCard(
                  margin: EdgeInsets.zero,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('情绪', style: Theme.of(context).textTheme.titleMedium),
                      const SizedBox(height: 8),
                      Text('涨停 ${sentiment['limit_up_count']} / 跌停 ${sentiment['limit_down_count']}'),
                      Text('均涨幅 ${sentiment['avg_change_pct']}%'),
                    ],
                  ),
                ),
              ],
            ),
            SectionHeader(title: '主要指数'),
            ResponsiveGrid(
              desktopColumns: 2,
              children: indices.map((e) {
                final m = e as Map<String, dynamic>;
                return GlassCard(
                  margin: EdgeInsets.zero,
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  child: ListTile(
                    title: Text('${m['name']} (${m['code']})'),
                    trailing: Text('${m['change_pct']}%', style: TextStyle(color: priceColor((m['change_pct'] as num?)?.toDouble() ?? 0))),
                    subtitle: Text('${m['price']}'),
                  ),
                );
              }).toList(),
            ),
            if (predict != null) ...[
              SectionHeader(title: '${context.read<AppState>().selectedCode} 涨跌推测'),
              GlassCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('评分 ${predict!['score']} | 信号 ${predict!['signal']}'),
                    Text('1日上涨概率 ${((predict!['prob_up_1d'] as num) * 100).toStringAsFixed(1)}%'),
                  ],
                ),
              ),
            ],
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _portfolioTab() {
    if (portfolio == null) return const Center(child: Text('加载中...'));
    final positions = portfolio!['positions'] as List? ?? [];
    final warnings = portfolio!['risk_warnings'] as List? ?? [];
    return RefreshIndicator(
      onRefresh: _load,
      color: AppColors.primary,
      child: ResponsivePage(
        scrollable: true,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            GlassCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('总市值 ${portfolio!['total_value']} | 盈亏 ${portfolio!['total_pl']} (${portfolio!['total_pl_pct']}%)'),
                  const SizedBox(height: 4),
                  Text('Beta ${portfolio!['beta']} | 最大回撤估算 ${portfolio!['max_drawdown_90d_est']}%'),
                ],
              ),
            ),
            if (warnings.isNotEmpty) ...[
              SectionHeader(title: '风险提示'),
              ResponsiveGrid(
                desktopColumns: 2,
                children: warnings
                    .map((w) => GlassCard(margin: EdgeInsets.zero, child: Text(w.toString())))
                    .toList(),
              ),
            ],
            SectionHeader(title: '持仓明细'),
            ResponsiveGrid(
              desktopColumns: 2,
              children: positions.map((p) {
                final m = p as Map<String, dynamic>;
                return GlassCard(
                  margin: EdgeInsets.zero,
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  child: ListTile(
                    title: Text('${m['name']} (${m['code']})'),
                    subtitle: Text('质量 ${m['quality_grade']} | MA20 ${m['ma20']}'),
                    trailing: Text('${m['pl_pct']}%', style: TextStyle(color: priceColor((m['pl_pct'] as num?)?.toDouble() ?? 0))),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _anomaliesTab() {
    if (anomalies == null) return const Center(child: Text('加载中...'));
    final items = anomalies!['anomalies'] as List? ?? [];
    return RefreshIndicator(
      onRefresh: _load,
      color: AppColors.primary,
      child: ResponsivePage(
        scrollable: true,
        child: ResponsiveGrid(
          desktopColumns: 2,
          children: items.map((e) {
            final m = e as Map<String, dynamic>;
            return GlassCard(
              margin: EdgeInsets.zero,
              child: ListTile(
                title: Text('${m['name']} (${m['code']}) ${m['change_pct']}%'),
                subtitle: Text('${(m['anomaly_types'] as List?)?.join(' · ') ?? ''}\n${(m['possible_causes'] as List?)?.first ?? ''}'),
                isThreeLine: true,
              ),
            );
          }).toList(),
        ),
      ),
    );
  }

  Widget _reportTab() {
    if (report == null) return const Center(child: Text('加载中...'));
    final summary = report!['market_summary'] as Map<String, dynamic>? ?? {};
    final top = report!['top_movers'] as List? ?? [];
    final watch = report!['tomorrow_watchlist'] as List? ?? [];
    return RefreshIndicator(
      onRefresh: _load,
      color: AppColors.primary,
      child: ResponsivePage(
        scrollable: true,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text('${report!['date']} 复盘', style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 8),
            GlassCard(child: Text('${summary['status']} — ${summary['description']}')),
            SectionHeader(title: '涨幅榜'),
            ResponsiveGrid(
              desktopColumns: 2,
              children: top.take(5).map((e) {
                final m = e as Map<String, dynamic>;
                return GlassCard(
                  margin: EdgeInsets.zero,
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  child: ListTile(title: Text('${m['name']}'), trailing: Text('${m['change_pct']}%')),
                );
              }).toList(),
            ),
            SectionHeader(title: '明日关注'),
            ResponsiveGrid(
              desktopColumns: 2,
              children: watch.take(10).map((e) {
                final m = e as Map<String, dynamic>;
                return GlassCard(
                  margin: EdgeInsets.zero,
                  child: ListTile(
                    title: Text('${m['name']} (${m['code']})'),
                    subtitle: Text('${m['reason']}'),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
}
