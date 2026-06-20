import SwiftUI
import UIKit

// MARK: - Hex initializers
// bleRPC Design System — ported from the "bleRPC" Claude Design project
// (themes/swiftui/Color+Hex.swift).

extension UIColor {
    /// Create a UIColor from a 0xRRGGBB hex literal.
    convenience init(hex: UInt, alpha: CGFloat = 1) {
        self.init(
            red: CGFloat((hex >> 16) & 0xFF) / 255,
            green: CGFloat((hex >> 8) & 0xFF) / 255,
            blue: CGFloat(hex & 0xFF) / 255,
            alpha: alpha
        )
    }
}

extension Color {
    /// Create a Color from a 0xRRGGBB hex literal.
    init(hex: UInt, opacity: Double = 1) {
        self.init(
            .sRGB,
            red: Double((hex >> 16) & 0xFF) / 255,
            green: Double((hex >> 8) & 0xFF) / 255,
            blue: Double(hex & 0xFF) / 255,
            opacity: opacity
        )
    }
}

/// Builds a Color that resolves to `light` or `dark` based on the active trait.
/// bleRPC ships dark-first, but both schemes are defined.
func bleRPCDynamic(light: UInt, dark: UInt) -> Color {
    Color(UIColor { traits in
        traits.userInterfaceStyle == .dark
            ? UIColor(hex: dark)
            : UIColor(hex: light)
    })
}
