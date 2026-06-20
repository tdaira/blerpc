@file:Suppress("ktlint:standard:function-naming")

package com.blerpc.android.design

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable

// ============================================================
// bleRPC Design System — Compose theme entry point.
// Maps the Tokyo Night palette onto a Material 3 ColorScheme so
// Material components inherit the brand. bleRPC is dark-first.
// ============================================================

private val BleDarkColors =
    darkColorScheme(
        primary = BleDarkAccent,
        onPrimary = BleOnAccent,
        background = BleDarkBg,
        onBackground = BleDarkText,
        surface = BleDarkBgSecondary,
        onSurface = BleDarkText,
        surfaceVariant = BleDarkBgCode,
        onSurfaceVariant = BleDarkTextSecondary,
        outline = BleDarkBorder,
        error = BleDarkError,
        tertiary = BleDarkSuccess,
        secondary = BleDarkNavBg,
    )

private val BleLightColors =
    lightColorScheme(
        primary = BleLightAccent,
        onPrimary = BleOnAccent,
        background = BleLightBg,
        onBackground = BleLightText,
        surface = BleLightBgSecondary,
        onSurface = BleLightText,
        surfaceVariant = BleLightBgCode,
        onSurfaceVariant = BleLightTextSecondary,
        outline = BleLightBorder,
        error = BleLightError,
        tertiary = BleLightSuccess,
        secondary = BleLightNavBg,
    )

@Composable
fun BleRpcTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit,
) {
    MaterialTheme(
        colorScheme = if (darkTheme) BleDarkColors else BleLightColors,
        typography = BleTypography,
        content = content,
    )
}
