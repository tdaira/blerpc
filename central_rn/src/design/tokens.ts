// bleRPC Design System — tokens (Tokyo Night, dark-first).
// Values mirror the "bleRPC" Claude Design project tokens/*.css.

export const BleColors = {
  // Surfaces
  bg: '#1A1B26',
  bgSecondary: '#24283B',
  bgCode: '#1E2030',
  navBg: '#16161E',
  // Text
  text: '#C0CAF5',
  textSecondary: '#A9B1D6',
  onAccent: '#FFFFFF',
  // Brand accent
  accent: '#0082FC',
  accentHover: '#339DFF',
  // Lines
  border: '#3B4261',
  // Status
  success: '#9ECE6A',
  warning: '#E0AF68',
  error: '#F7768E',
} as const;

export const BleSpacing = {
  s1: 4,
  s2: 8,
  s3: 12,
  s4: 16,
  s5: 20,
  s6: 24,
  s8: 32,
  s12: 48,
} as const;

export const BleRadius = {
  sm: 4, // inline code
  md: 6, // buttons, inputs
  lg: 8, // code blocks
  xl: 10, // cards
  pill: 12, // badges
} as const;

/** Map an RSSI (dBm) reading to a 0–4 signal-strength level. */
export function rssiToLevel(rssi: number): number {
  if (rssi >= -55) return 4;
  if (rssi >= -65) return 3;
  if (rssi >= -75) return 2;
  if (rssi >= -85) return 1;
  return 0;
}
