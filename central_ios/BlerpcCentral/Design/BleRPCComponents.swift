import SwiftUI

// ============================================================
// bleRPC — example SwiftUI components built on the theme.
// Copy these as starting points; they mirror the web primitives.
// Ported from the "bleRPC" Claude Design project (themes/swiftui/BleRPCComponents.swift),
// plus a BleRPCSignalBars atom for RSSI (components/mobile/SignalBars).
// ============================================================

/// Primary / secondary / ghost button.
struct BleRPCButton: View {
    enum Variant { case primary, secondary, ghost }
    let title: String
    var variant: Variant = .primary
    var action: () -> Void = {}

    var body: some View {
        Button(action: action) {
            Text(title)
                .font(.system(size: 15, weight: .semibold))
                .padding(.horizontal, BleRPCSpacing.s4)
                .padding(.vertical, 10)
                .frame(maxWidth: .infinity)
                .foregroundColor(fg)
                .background(bg)
                .clipShape(RoundedRectangle(cornerRadius: BleRPCRadius.md, style: .continuous))
                .overlay(
                    RoundedRectangle(cornerRadius: BleRPCRadius.md, style: .continuous)
                        .stroke(variant == .secondary ? BleRPCColor.border : .clear, lineWidth: 1)
                )
        }
        .buttonStyle(.plain)
    }

    private var bg: Color {
        switch variant {
        case .primary: return BleRPCColor.accent
        case .secondary: return BleRPCColor.bgSecondary
        case .ghost: return .clear
        }
    }
    private var fg: Color {
        switch variant {
        case .primary: return BleRPCColor.onAccent
        case .secondary: return BleRPCColor.text
        case .ghost: return BleRPCColor.textSecondary
        }
    }
}

/// Status / category pill.
struct BleRPCBadge: View {
    enum Tone { case blue, green, yellow, red, neutral }
    let text: String
    var tone: Tone = .blue

    var body: some View {
        Text(text)
            .font(BleRPCFont.badge)
            .padding(.horizontal, 10)
            .padding(.vertical, 3)
            .foregroundColor(fg)
            .background(fg.opacity(0.15))
            .clipShape(Capsule())
    }

    private var fg: Color {
        switch tone {
        case .blue: return BleRPCColor.accent
        case .green: return BleRPCColor.success
        case .yellow: return BleRPCColor.warning
        case .red: return BleRPCColor.error
        case .neutral: return BleRPCColor.textSecondary
        }
    }
}

/// Map an RSSI (dBm) reading to a 0–4 signal-strength level.
func bleRPCRssiToLevel(_ rssi: Int) -> Int {
    switch rssi {
    case let r where r >= -55: return 4
    case let r where r >= -65: return 3
    case let r where r >= -75: return 2
    case let r where r >= -85: return 1
    default: return 0
    }
}

/// bleRPC SignalBars — a 4-bar BLE/RSSI strength indicator. `level` is 0–4;
/// inactive bars use the border color. Mirrors components/mobile/SignalBars.jsx.
struct BleRPCSignalBars: View {
    var level: Int
    var size: CGFloat = 16

    var body: some View {
        let lv = max(0, min(4, level))
        let gap = size * 0.18
        let barW = (size - gap * 3) / 4
        HStack(alignment: .bottom, spacing: gap) {
            ForEach(0..<4, id: \.self) { i in
                RoundedRectangle(cornerRadius: 1)
                    .fill(i < lv ? color(lv) : BleRPCColor.border)
                    .frame(width: barW, height: size * CGFloat(i + 1) / 4)
            }
        }
        .frame(height: size, alignment: .bottom)
    }

    private func color(_ lv: Int) -> Color {
        if lv >= 3 { return BleRPCColor.success }
        if lv == 2 { return BleRPCColor.warning }
        return BleRPCColor.error
    }
}

// MARK: - Preview

struct BleRPC_Previews: PreviewProvider {
    static var previews: some View {
        VStack(alignment: .leading, spacing: BleRPCSpacing.s4) {
            Text("bleRPC").font(BleRPCFont.display).foregroundColor(BleRPCColor.text)
            HStack { BleRPCBadge(text: "Stable", tone: .green); BleRPCBadge(text: "Swift", tone: .blue) }
            HStack(spacing: BleRPCSpacing.s4) {
                BleRPCSignalBars(level: 4)
                BleRPCSignalBars(level: 2)
                BleRPCSignalBars(level: 1)
            }
            VStack(spacing: BleRPCSpacing.s3) {
                Text("nRF54L15 DK").font(BleRPCFont.body).foregroundColor(BleRPCColor.text)
                BleRPCButton(title: "Connect device", variant: .primary)
                BleRPCButton(title: "Disconnect", variant: .secondary)
            }.bleRPCCard()
        }
        .padding()
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(BleRPCColor.bg)
        .preferredColorScheme(.dark)
    }
}
