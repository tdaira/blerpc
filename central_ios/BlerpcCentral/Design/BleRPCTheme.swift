import SwiftUI

// ============================================================
// bleRPC Design System — SwiftUI theme
// Tokyo Night palette, dark-first. Values mirror the design system's tokens/*.css.
// Ported from the "bleRPC" Claude Design project (themes/swiftui/BleRPCTheme.swift).
// ============================================================

// MARK: - Colors

enum BleRPCColor {
    // Surfaces
    static let bg          = bleRPCDynamic(light: 0xF5F5F5, dark: 0x1A1B26)
    static let bgSecondary = bleRPCDynamic(light: 0xFFFFFF, dark: 0x24283B)
    static let bgCode      = bleRPCDynamic(light: 0xE8E8E8, dark: 0x1E2030)
    static let navBg       = bleRPCDynamic(light: 0xE8E8E8, dark: 0x16161E)

    // Text
    static let text          = bleRPCDynamic(light: 0x1A1B26, dark: 0xC0CAF5)
    static let textSecondary = bleRPCDynamic(light: 0x4E5173, dark: 0xA9B1D6)
    static let onAccent      = Color(hex: 0xFFFFFF)

    // Brand accent
    static let accent      = bleRPCDynamic(light: 0x0070D6, dark: 0x0082FC)
    static let accentHover = bleRPCDynamic(light: 0x005CB8, dark: 0x339DFF)

    // Lines
    static let border = bleRPCDynamic(light: 0xD0D0D0, dark: 0x3B4261)

    // Status
    static let success = bleRPCDynamic(light: 0x4A7A1E, dark: 0x9ECE6A)
    static let warning = bleRPCDynamic(light: 0xA07020, dark: 0xE0AF68)
    static let error   = bleRPCDynamic(light: 0xC44040, dark: 0xF7768E)
}

// MARK: - Spacing

enum BleRPCSpacing {
    static let s1: CGFloat = 4
    static let s2: CGFloat = 8
    static let s3: CGFloat = 12
    static let s4: CGFloat = 16
    static let s5: CGFloat = 20
    static let s6: CGFloat = 24
    static let s8: CGFloat = 32
    static let s12: CGFloat = 48
}

// MARK: - Radius

enum BleRPCRadius {
    static let sm: CGFloat = 4   // inline code
    static let md: CGFloat = 6   // buttons, inputs
    static let lg: CGFloat = 8   // code blocks
    static let xl: CGFloat = 10  // cards
    static let pill: CGFloat = 12
}

// MARK: - Typography
// Body uses the system UI font (San Francisco). Code uses a monospaced face;
// register Fira Code as "FiraCode-Regular" to match the web exactly, else the
// system monospaced font is used.

enum BleRPCFont {
    static let hero    = Font.system(size: 40, weight: .black)
    static let display = Font.system(size: 32, weight: .bold)
    static let h2      = Font.system(size: 24, weight: .semibold)
    static let h3      = Font.system(size: 19, weight: .semibold)
    static let body    = Font.system(size: 16, weight: .regular)
    static let caption = Font.system(size: 14, weight: .regular)
    static let badge   = Font.system(size: 12, weight: .semibold)

    /// Monospace — Fira Code if installed, otherwise the system monospaced font.
    static func mono(_ size: CGFloat = 14) -> Font {
        if UIFont(name: "FiraCode-Regular", size: size) != nil {
            return .custom("FiraCode-Regular", size: size)
        }
        return .system(size: size, design: .monospaced)
    }
}

// MARK: - Card modifier (resting → flat, 1px border, 10px radius)

struct BleRPCCard: ViewModifier {
    var padding: CGFloat = BleRPCSpacing.s6
    func body(content: Content) -> some View {
        content
            .padding(padding)
            .background(BleRPCColor.bgSecondary)
            .clipShape(RoundedRectangle(cornerRadius: BleRPCRadius.xl, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: BleRPCRadius.xl, style: .continuous)
                    .stroke(BleRPCColor.border, lineWidth: 1)
            )
    }
}

extension View {
    func bleRPCCard(padding: CGFloat = BleRPCSpacing.s6) -> some View {
        modifier(BleRPCCard(padding: padding))
    }
}
