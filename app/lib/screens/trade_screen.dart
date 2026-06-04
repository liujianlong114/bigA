import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';

class TradeScreen extends StatefulWidget {
  const TradeScreen({super.key});

  @override
  State<TradeScreen> createState() => _TradeScreenState();
}

class _TradeScreenState extends State<TradeScreen> {
  final _codeCtrl = TextEditingController(text: '600519');
  final _qtyCtrl = TextEditingController(text: '100');
  final _priceCtrl = TextEditingController();
  String _side = 'buy';
  String _orderType = 'market';

  @override
  void dispose() {
    _codeCtrl.dispose();
    _qtyCtrl.dispose();
    _priceCtrl.dispose();
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
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(err ?? '委托成功')),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    final q = state.selectedQuote;

    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text('下单', style: Theme.of(context).textTheme.titleLarge),
                const SizedBox(height: 12),
                TextField(
                  controller: _codeCtrl,
                  decoration: const InputDecoration(labelText: '股票代码'),
                  keyboardType: TextInputType.number,
                ),
                const SizedBox(height: 12),
                SegmentedButton<String>(
                  segments: const [
                    ButtonSegment(value: 'buy', label: Text('买入'), icon: Icon(Icons.arrow_upward)),
                    ButtonSegment(value: 'sell', label: Text('卖出'), icon: Icon(Icons.arrow_downward)),
                  ],
                  selected: {_side},
                  onSelectionChanged: (s) => setState(() => _side = s.first),
                ),
                const SizedBox(height: 12),
                SegmentedButton<String>(
                  segments: const [
                    ButtonSegment(value: 'market', label: Text('市价')),
                    ButtonSegment(value: 'limit', label: Text('限价')),
                  ],
                  selected: {_orderType},
                  onSelectionChanged: (s) => setState(() => _orderType = s.first),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: _qtyCtrl,
                  decoration: const InputDecoration(labelText: '数量（股，100 整数倍）'),
                  keyboardType: TextInputType.number,
                ),
                if (_orderType == 'limit') ...[
                  const SizedBox(height: 12),
                  TextField(
                    controller: _priceCtrl,
                    decoration: const InputDecoration(labelText: '限价'),
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                ],
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: state.loading ? null : () => _submit(state),
                  style: FilledButton.styleFrom(
                    backgroundColor: _side == 'buy' ? AppColors.up : AppColors.down,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                  child: Text(_side == 'buy' ? '确认买入' : '确认卖出'),
                ),
              ],
            ),
          ),
        ),
        if (q != null)
          Card(
            child: ListTile(
              title: Text('当前 ${q.name}'),
              subtitle: Text('现价 ${fmtMoney(q.price)}  卖一 ${fmtMoney(q.ask1)}  买一 ${fmtMoney(q.bid1)}'),
            ),
          ),
        Card(
          color: Colors.blue.shade50,
          child: const Padding(
            padding: EdgeInsets.all(12),
            child: Text(
              'A 股规则：100 股一手 · T+1 当日买入不可卖 · 佣金/印花税/过户费按模拟盘计算',
              style: TextStyle(fontSize: 13),
            ),
          ),
        ),
        OutlinedButton.icon(
          onPressed: state.loading
              ? null
              : () async {
                  final err = await state.doSettle();
                  if (!context.mounted) return;
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text(err ?? 'T+1 交割完成，持仓可卖')),
                  );
                },
          icon: const Icon(Icons.calendar_today),
          label: const Text('模拟下一交易日 (T+1 交割)'),
        ),
      ],
    );
  }
}
