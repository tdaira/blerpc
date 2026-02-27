import 'package:flutter_test/flutter_test.dart';

import 'package:blerpc_central/main.dart';

void main() {
  testWidgets('App renders without crashing', (WidgetTester tester) async {
    await tester.pumpWidget(const BlerpcCentralApp());
    expect(find.text('blerpc Central'), findsOneWidget);
    expect(find.text('Scan'), findsOneWidget);
  });
}
