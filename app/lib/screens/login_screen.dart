import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/app_state.dart';
import '../theme/app_theme.dart';
import '../widgets/components.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> with SingleTickerProviderStateMixin {
  late final TabController _tabs;
  final _loginUser = TextEditingController();
  final _loginPass = TextEditingController();
  final _regUser = TextEditingController();
  final _regNick = TextEditingController();
  final _regPass = TextEditingController();
  bool _busy = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabs.dispose();
    _loginUser.dispose();
    _loginPass.dispose();
    _regUser.dispose();
    _regNick.dispose();
    _regPass.dispose();
    super.dispose();
  }

  Future<void> _run(Future<String?> Function() action) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    final err = await action();
    if (!mounted) return;
    setState(() {
      _busy = false;
      _error = err;
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();

    return AppBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        body: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 440),
                child: GlassCard(
                  margin: EdgeInsets.zero,
                  padding: const EdgeInsets.all(28),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        'bigA 模拟盘',
                        style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                              fontWeight: FontWeight.w700,
                              letterSpacing: -0.5,
                            ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        '注册独立账户，100 万模拟资金起步。不知道股票代码？首页可按板块浏览全市场。',
                        style: TextStyle(color: Colors.black.withValues(alpha: 0.5), height: 1.4),
                      ),
                      const SizedBox(height: 24),
                      GlassSurface(
                        radius: 14,
                        padding: const EdgeInsets.all(4),
                        child: TabBar(
                          controller: _tabs,
                          indicator: BoxDecoration(
                            borderRadius: BorderRadius.circular(10),
                            color: AppColors.primary.withValues(alpha: 0.12),
                          ),
                          labelColor: AppColors.primary,
                          unselectedLabelColor: const Color(0xFF64748B),
                          dividerColor: Colors.transparent,
                          tabs: const [
                            Tab(text: '登录'),
                            Tab(text: '注册'),
                          ],
                        ),
                      ),
                      const SizedBox(height: 20),
                      if (_error != null)
                        Padding(
                          padding: const EdgeInsets.only(bottom: 12),
                          child: Text(_error!, style: const TextStyle(color: AppColors.up)),
                        ),
                      SizedBox(
                        height: 260,
                        child: TabBarView(
                          controller: _tabs,
                          children: [
                            _loginForm(state),
                            _registerForm(state),
                          ],
                        ),
                      ),
                      GlassButton(
                        label: '游客进入（本地开发）',
                        variant: GlassButtonVariant.ghost,
                        expanded: true,
                        onPressed: _busy ? null : () => state.enterGuest(),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _loginForm(AppState state) {
    return ListView(
      children: [
        GlassTextField(controller: _loginUser, labelText: '用户名', textInputAction: TextInputAction.next),
        const SizedBox(height: 12),
        GlassTextField(
          controller: _loginPass,
          labelText: '密码',
          obscureText: true,
          onSubmitted: (_) => _submitLogin(state),
        ),
        const SizedBox(height: 20),
        GlassButton(
          label: '登录',
          expanded: true,
          loading: _busy,
          onPressed: _busy ? null : () => _submitLogin(state),
        ),
      ],
    );
  }

  Widget _registerForm(AppState state) {
    return ListView(
      children: [
        GlassTextField(controller: _regUser, labelText: '用户名（至少 3 字）'),
        const SizedBox(height: 12),
        GlassTextField(controller: _regNick, labelText: '昵称（可选）'),
        const SizedBox(height: 12),
        GlassTextField(
          controller: _regPass,
          labelText: '密码（至少 6 位）',
          obscureText: true,
          onSubmitted: (_) => _submitRegister(state),
        ),
        const SizedBox(height: 20),
        GlassButton(
          label: '注册并进入',
          expanded: true,
          loading: _busy,
          onPressed: _busy ? null : () => _submitRegister(state),
        ),
      ],
    );
  }

  void _submitLogin(AppState state) {
    _run(() => state.login(_loginUser.text.trim(), _loginPass.text));
  }

  void _submitRegister(AppState state) {
    _run(() => state.register(
          _regUser.text.trim(),
          _regPass.text,
          nickname: _regNick.text.trim(),
        ));
  }
}
