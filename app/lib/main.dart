import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/app_state.dart';
import 'screens/home_shell.dart';
import 'screens/login_screen.dart';
import 'services/api_service.dart';
import 'theme/app_theme.dart';
import 'widgets/components.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const BigAApp());
}

class BigAApp extends StatelessWidget {
  const BigAApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => AppState(ApiService())..init(),
      child: MaterialApp(
        title: 'bigA 模拟盘',
        debugShowCheckedModeBanner: false,
        theme: buildAppTheme(),
        home: const AuthGate(),
      ),
    );
  }
}

class AuthGate extends StatelessWidget {
  const AuthGate({super.key});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    if (!state.authReady) {
      return AppBackground(
        child: Scaffold(
          backgroundColor: Colors.transparent,
          body: Center(
            child: GlassCard(
              margin: EdgeInsets.zero,
              child: const SizedBox(
                width: 32,
                height: 32,
                child: CircularProgressIndicator(strokeWidth: 2.5, color: AppColors.primary),
              ),
            ),
          ),
        ),
      );
    }
    if (state.isLoggedIn || state.isGuest) {
      return const HomeShell();
    }
    return const LoginScreen();
  }
}
