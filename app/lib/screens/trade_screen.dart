import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../layout/app_breakpoints.dart';
import '../layout/responsive_page.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';
import '../theme/glass_theme.dart';
import '../widgets/components.dart';

class TradeScreen extends StatefulWidget {
  const TradeScreen({super.key});

  @override
  State<TradeScreen> createState() => _TradeScreenState();
}

class _TradeScreenState extends State<TradeScreen> {
  final _codeCtrl = TextEditingController(text: '600519');
  final _qtyCtrl = TextEditingController(text: '100');
  final _priceCtrl = TextEditingController();
  final _triggerCtrl = TextEditingController();
  String _side = 'buy';
  String _orderType = 'market';
  String _condType = 'price_gte';
  bool _showConditional = false;

  @override
  void dispose() {
    _codeCtrl.dispose();
    _qtyCtrl.dispose();
    _priceCtrl.dispose();
    _triggerCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit(AppState state) async {
    final code = _codeCtrl.text.trim();
    if (code.length == 6) await state.loadQuote(code);
    final qty = int.tryParse(_qtyCtrl.text) ?? 0;
    final price = double.tryParse(_priceCtrl.text);
    final err = await state.submitOrder(
      side: _side,
      orderType: _orderType,
      quantity: qty,
      price: _orderType == 'limit' ? price : null,
    );
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err ?? '委托成功')));
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    final q = state.selectedQuote;
    final expandBtn = AppBreakpoints.shouldExpandButtons(context);

    final orderForm = GlassCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text('下单', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 12),
          GlassTextField(controller: _codeCtrl, labelText: '股票代码'),
          const SizedBox(height: 12),
          AdaptiveSegmentRow(
            segments: [
              GlassChip(label: '买入', selected: _side == 'buy', selectedColor: AppColors.up, onPressed: () => setState(() => _side = 'buy')),
              GlassChip(label: '卖出', selected: _side == 'sell', selectedColor: AppColors.down, onPressed: () => setState(() => _side = 'sell')),
            ],
          ),
          const SizedBox(height: 12),
          AdaptiveSegmentRow(
            segments: [
              GlassChip(label: '市价', selected: _orderType == 'market', onPressed: () => setState(() => _orderType = 'market')),
              GlassChip(label: '限价', selected: _orderType == 'limit', onPressed: () => setState(() => _orderType = 'limit')),
            ],
          ),
          const SizedBox(height: 12),
          GlassTextField(controller: _qtyCtrl, labelText: '数量（100 整数倍）'),
          if (_orderType == 'limit') ...[
            const SizedBox(height: 12),
            GlassTextField(controller: _priceCtrl, labelText: '限价'),
          ],
          const SizedBox(height: 16),
          AdaptiveButtonRow(
            buttons: [
              GlassButton(
                label: _side == 'buy' ? '确认买入' : '确认卖出',
                variant: _side == 'buy' ? GlassButtonVariant.danger : GlassButtonVariant.primary,
                expanded: expandBtn,
                minWidth: expandBtn ? null : 160,
                onPressed: state.loading ? null : () => _submit(state),
              ),
            ],
          ),
        ],
      ),
    );

    final sidePanel = Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (q != null)
          GlassCard(
            child: ListTile(
              title: Text('当前 ${q.name}', style: const TextStyle(fontWeight: FontWeight.w600)),
              subtitle: Text('现价 ${fmtMoney(q.price)} · 卖一 ${fmtMoney(q.ask1)} · 买一 ${fmtMoney(q.bid1)}'),
            ),
          ),
        GlassCard(
          child: Text(
            'A 股规则：100 股一手 · T+1 当日买入不可卖 · 佣金/印花税/过户费按模拟盘计算',
            style: TextStyle(fontSize: 13, color: GlassTheme.of(context).labelSecondary, height: 1.4),
          ),
        ),
        _conditionalCard(state, expandBtn),
        AdaptiveButtonRow(
          buttons: [
            GlassButton(
              label: '模拟下一交易日 (T+1 交割)',
              variant: GlassButtonVariant.secondary,
              icon: Icons.calendar_today_rounded,
              expanded: expandBtn,
              onPressed: state.loading
                  ? null
                  : () async {
                      final err = await state.doSettle();
                      if (!context.mounted) return;
                      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err ?? 'T+1 交割完成，持仓可卖')));
                    },
            ),
          ],
        ),
      ],
    );

    return ResponsivePage(
      scrollable: true,
      child: DesktopSplitView(
        primaryFlex: 5,
        secondaryFlex: 6,
        scrollSecondary: false,
        primary: orderForm,
        secondary: sidePanel,
      ),
    );
  }

  Widget _conditionalCard(AppState state, bool expandBtn) {
    return GlassCard(
      padding: EdgeInsets.zero,
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          initiallyExpanded: _showConditional,
          onExpansionChanged: (v) => setState(() => _showConditional = v),
          title: const Text('条件单', style: TextStyle(fontWeight: FontWeight.w600)),
          subtitle: Text('待触发 ${state.conditionalOrders.where((c) => c.status == 'pending').length} 笔'),
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  GlassChipRow(
                    chips: [
                      GlassChip(label: '价格 ≥', selected: _condType == 'price_gte', onPressed: () => setState(() => _condType = 'price_gte')),
                      GlassChip(label: '价格 ≤', selected: _condType == 'price_lte', onPressed: () => setState(() => _condType = 'price_lte')),
                      GlassChip(label: '涨幅 ≥', selected: _condType == 'change_pct_gte', onPressed: () => setState(() => _condType = 'change_pct_gte')),
                      GlassChip(label: '跌幅 ≤', selected: _condType == 'change_pct_lte', onPressed: () => setState(() => _condType = 'change_pct_lte')),
                    ],
                  ),
                  const SizedBox(height: 8),
                  GlassTextField(controller: _triggerCtrl, labelText: '触发值'),
                  const SizedBox(height: 12),
                  AdaptiveButtonRow(
                    buttons: [
                      GlassButton(
                        label: '创建条件单',
                        expanded: expandBtn,
                        minWidth: expandBtn ? null : 140,
                        onPressed: state.loading ? null : () => _createConditional(state),
                      ),
                    ],
                  ),
                  ...state.conditionalOrders.take(10).map((co) => ListTile(
                        dense: true,
                        title: Text('${co.name} (${co.code})'),
                        subtitle: Text('${co.conditionLabel} · ${co.status}'),
                        trailing: co.status == 'pending'
                            ? IconButton(
                                icon: const Icon(Icons.close_rounded, size: 20),
                                onPressed: () async {
                                  final err = await state.cancelConditionalOrder(co.id);
                                  if (context.mounted) {
                                    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err ?? '已取消')));
                                  }
                                },
                              )
                            : null,
                      )),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _createConditional(AppState state) async {
    final code = _codeCtrl.text.trim();
    final qty = int.tryParse(_qtyCtrl.text) ?? 0;
    final trigger = double.tryParse(_triggerCtrl.text);
    if (code.length != 6 || qty <= 0 || trigger == null) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请填写 6 位代码、数量和触发值')));
      return;
    }
    final err = await state.createConditionalOrder(
      code: code,
      conditionType: _condType,
      triggerValue: trigger,
      side: _side,
      orderType: _orderType,
      quantity: qty,
      price: _orderType == 'limit' ? double.tryParse(_priceCtrl.text) : null,
    );
    if (!context.mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err ?? '条件单已创建')));
  }
}
