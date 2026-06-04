import 'package:flutter_test/flutter_test.dart';
import 'package:biga_app/main.dart';

void main() {
  testWidgets('app smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(const BigAApp());
    await tester.pump();
    expect(find.text('bigA 模拟盘'), findsOneWidget);
  });
}
