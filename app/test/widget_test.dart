import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:biga_app/main.dart';

void main() {
  testWidgets('app smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(const BigAApp());
    expect(find.byType(MaterialApp), findsOneWidget);
  });
}
