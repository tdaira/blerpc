import React from 'react';
import { View, Text, Platform, StyleSheet } from 'react-native';
import { BleColors, BleRadius } from './tokens';

/** Monospace family — Menlo on iOS, the platform monospace elsewhere. */
export const monoFont = Platform.OS === 'ios' ? 'Menlo' : 'monospace';

/** The bleRPC wordmark: `ble` accent-blue, `RPC` text-color (both 900), suffix regular. */
export function Wordmark({
  suffix = ' Central',
  fontSize = 20,
}: {
  suffix?: string;
  fontSize?: number;
}) {
  return (
    <Text style={{ fontSize }}>
      <Text style={styles.wordmarkBle}>ble</Text>
      <Text style={styles.wordmarkRpc}>RPC</Text>
      <Text style={styles.wordmarkSuffix}>{suffix}</Text>
    </Text>
  );
}

export type BadgeTone = 'blue' | 'green' | 'yellow' | 'red' | 'neutral';

const toneColor: Record<BadgeTone, string> = {
  blue: BleColors.accent,
  green: BleColors.success,
  yellow: BleColors.warning,
  red: BleColors.error,
  neutral: BleColors.textSecondary,
};

/** Status/category pill — tone color at ~15% alpha background. */
export function Badge({ text, tone = 'blue' }: { text: string; tone?: BadgeTone }) {
  const c = toneColor[tone];
  return (
    <View
      style={{
        backgroundColor: c + '26', // ~15% alpha
        borderRadius: BleRadius.pill,
        paddingHorizontal: 10,
        paddingVertical: 3,
        alignSelf: 'flex-start',
      }}
    >
      <Text style={{ color: c, fontSize: 12, fontWeight: '600' }}>{text}</Text>
    </View>
  );
}

/**
 * bleRPC SignalBars — a 4-bar BLE/RSSI strength indicator. `level` is 0–4;
 * inactive bars use the border color. Mirrors components/mobile/SignalBars.jsx.
 */
export function SignalBars({ level, size = 16 }: { level: number; size?: number }) {
  const lv = Math.max(0, Math.min(4, level));
  const active = lv >= 3 ? BleColors.success : lv === 2 ? BleColors.warning : BleColors.error;
  const gap = size * 0.18;
  const barW = (size - gap * 3) / 4;
  return (
    <View style={{ flexDirection: 'row', alignItems: 'flex-end', height: size }}>
      {[0, 1, 2, 3].map((i) => (
        <View
          key={i}
          style={{
            width: barW,
            height: (size * (i + 1)) / 4,
            borderRadius: 1,
            backgroundColor: i < lv ? active : BleColors.border,
            marginLeft: i === 0 ? 0 : gap,
          }}
        />
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  wordmarkBle: { color: BleColors.accent, fontWeight: '900' },
  wordmarkRpc: { color: BleColors.text, fontWeight: '900' },
  wordmarkSuffix: { color: BleColors.text, fontWeight: '400' },
});
