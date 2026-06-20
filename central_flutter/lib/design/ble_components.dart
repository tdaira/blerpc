import 'package:flutter/material.dart';

import 'ble_colors.dart';
import 'ble_dimens.dart';

/// Map an RSSI (dBm) reading to a 0–4 signal-strength level.
int rssiToLevel(int rssi) {
  if (rssi >= -55) return 4;
  if (rssi >= -65) return 3;
  if (rssi >= -75) return 2;
  if (rssi >= -85) return 1;
  return 0;
}

/// The bleRPC wordmark: `ble` in accent blue, `RPC` in text color (both 900),
/// then an optional trailing suffix in regular weight.
class BleWordmark extends StatelessWidget {
  const BleWordmark({super.key, this.suffix = ' Central', this.fontSize = 24});

  final String suffix;
  final double fontSize;

  @override
  Widget build(BuildContext context) {
    return Text.rich(
      TextSpan(
        children: [
          TextSpan(
            text: 'ble',
            style: TextStyle(
              color: BleColors.accent,
              fontWeight: FontWeight.w900,
              fontSize: fontSize,
            ),
          ),
          TextSpan(
            text: 'RPC',
            style: TextStyle(
              color: BleColors.text,
              fontWeight: FontWeight.w900,
              fontSize: fontSize,
            ),
          ),
          TextSpan(
            text: suffix,
            style: TextStyle(
              color: BleColors.text,
              fontWeight: FontWeight.w400,
              fontSize: fontSize,
            ),
          ),
        ],
      ),
    );
  }
}

enum BleBadgeTone { blue, green, yellow, red, neutral }

/// A status/category pill — tone color at 15% alpha background.
class BleBadge extends StatelessWidget {
  const BleBadge(this.text, {super.key, this.tone = BleBadgeTone.blue});

  final String text;
  final BleBadgeTone tone;

  Color get _color {
    switch (tone) {
      case BleBadgeTone.blue:
        return BleColors.accent;
      case BleBadgeTone.green:
        return BleColors.success;
      case BleBadgeTone.yellow:
        return BleColors.warning;
      case BleBadgeTone.red:
        return BleColors.error;
      case BleBadgeTone.neutral:
        return BleColors.textSecondary;
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = _color;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 3),
      decoration: BoxDecoration(
        color: c.withAlpha(38), // ~15% alpha
        borderRadius: BorderRadius.circular(BleRadius.pill),
      ),
      child: Text(
        text,
        style: TextStyle(color: c, fontSize: 12, fontWeight: FontWeight.w600),
      ),
    );
  }
}

/// bleRPC SignalBars — a 4-bar BLE/RSSI strength indicator. `level` is 0–4;
/// inactive bars use the border color. Mirrors components/mobile/SignalBars.jsx.
class BleSignalBars extends StatelessWidget {
  const BleSignalBars({super.key, required this.level, this.size = 16});

  final int level;
  final double size;

  Color _activeColor(int lv) {
    if (lv >= 3) return BleColors.success;
    if (lv == 2) return BleColors.warning;
    return BleColors.error;
  }

  @override
  Widget build(BuildContext context) {
    final lv = level.clamp(0, 4);
    final gap = size * 0.18;
    final barW = (size - gap * 3) / 4;
    final active = _activeColor(lv);
    return SizedBox(
      height: size,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          for (var i = 0; i < 4; i++) ...[
            Container(
              width: barW,
              height: size * (i + 1) / 4,
              decoration: BoxDecoration(
                color: i < lv ? active : BleColors.border,
                borderRadius: BorderRadius.circular(1),
              ),
            ),
            if (i < 3) SizedBox(width: gap),
          ],
        ],
      ),
    );
  }
}
