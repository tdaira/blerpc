import 'package:flutter/material.dart';

import 'ble_colors.dart';
import 'ble_dimens.dart';

/// bleRPC Design System — Flutter theme (Tokyo Night, dark-first).
/// Maps the palette onto a Material 3 [ThemeData]/[ColorScheme].
class BleTheme {
  BleTheme._();

  static ThemeData dark() {
    return ThemeData(
      brightness: Brightness.dark,
      scaffoldBackgroundColor: BleColors.bg,
      colorScheme: const ColorScheme.dark(
        primary: BleColors.accent,
        onPrimary: BleColors.onAccent,
        surface: BleColors.bgSecondary,
        onSurface: BleColors.text,
        secondary: BleColors.accent,
        error: BleColors.error,
        outline: BleColors.border,
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: BleColors.navBg,
        foregroundColor: BleColors.text,
        elevation: 0,
      ),
      dividerTheme: const DividerThemeData(
        color: BleColors.border,
        space: 1,
        thickness: 1,
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: BleColors.accent,
          foregroundColor: BleColors.onAccent,
          disabledBackgroundColor: BleColors.bgSecondary,
          disabledForegroundColor: BleColors.textSecondary,
          textStyle: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
          padding: const EdgeInsets.symmetric(vertical: 12),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(BleRadius.md),
          ),
        ),
      ),
      useMaterial3: true,
    );
  }
}
