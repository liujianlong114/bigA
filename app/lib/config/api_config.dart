/// 后端 API 地址
/// - Web / macOS 桌面: localhost
/// - Android 模拟器: 10.0.2.2
/// - 真机: 改成电脑局域网 IP，如 http://192.168.1.100:8080
class ApiConfig {
  static const String baseUrl = String.fromEnvironment(
    'API_BASE',
    defaultValue: 'http://localhost:8080',
  );
}
