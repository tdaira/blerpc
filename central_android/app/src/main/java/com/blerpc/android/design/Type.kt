package com.blerpc.android.design

import androidx.compose.material3.Typography
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

// ============================================================
// bleRPC Design System — typography (mirrors tokens/typography.css)
// Body uses the platform default (Roboto). Code uses monospace —
// swap FontFamily.Monospace for a bundled Fira Code family to match
// the web exactly.
// ============================================================

val BleMono = FontFamily.Monospace

val BleTypography =
    Typography(
        displayLarge = TextStyle(fontFamily = FontFamily.Default, fontWeight = FontWeight.Black, fontSize = 40.sp),
        headlineLarge = TextStyle(fontFamily = FontFamily.Default, fontWeight = FontWeight.Bold, fontSize = 32.sp),
        headlineMedium = TextStyle(fontFamily = FontFamily.Default, fontWeight = FontWeight.SemiBold, fontSize = 24.sp),
        titleLarge = TextStyle(fontFamily = FontFamily.Default, fontWeight = FontWeight.SemiBold, fontSize = 19.sp),
        bodyLarge = TextStyle(fontFamily = FontFamily.Default, fontWeight = FontWeight.Normal, fontSize = 16.sp, lineHeight = 27.sp),
        bodyMedium = TextStyle(fontFamily = FontFamily.Default, fontWeight = FontWeight.Normal, fontSize = 14.sp),
        labelSmall = TextStyle(fontFamily = FontFamily.Default, fontWeight = FontWeight.SemiBold, fontSize = 12.sp),
    )
