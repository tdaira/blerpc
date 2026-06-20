@file:Suppress("ktlint:standard:function-naming")

package com.blerpc.android.design

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

// ============================================================
// bleRPC — example Compose components on the theme.
// Mirror the web primitives; copy as starting points.
// Ported from the "bleRPC" Claude Design project (themes/compose/Components.kt),
// plus a BleSignalBars atom for RSSI (components/mobile/SignalBars).
// ============================================================

enum class BleButtonVariant { Primary, Secondary, Ghost }

@Composable
fun BleButton(
    text: String,
    variant: BleButtonVariant = BleButtonVariant.Primary,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    onClick: () -> Unit = {},
) {
    val shape = RoundedCornerShape(BleRadius.md)
    when (variant) {
        BleButtonVariant.Primary ->
            Button(
                onClick = onClick,
                modifier = modifier,
                shape = shape,
                enabled = enabled,
                colors =
                    ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.primary,
                        contentColor = BleOnAccent,
                        disabledContainerColor = MaterialTheme.colorScheme.surface,
                        disabledContentColor = MaterialTheme.colorScheme.onSurfaceVariant,
                    ),
            ) { Text(text) }

        BleButtonVariant.Secondary ->
            OutlinedButton(
                onClick = onClick,
                modifier = modifier,
                shape = shape,
                enabled = enabled,
                border = BorderStroke(1.dp, MaterialTheme.colorScheme.outline),
                colors =
                    ButtonDefaults.outlinedButtonColors(
                        containerColor = MaterialTheme.colorScheme.surface,
                        contentColor = MaterialTheme.colorScheme.onSurface,
                    ),
            ) { Text(text) }

        BleButtonVariant.Ghost ->
            TextButton(
                onClick = onClick,
                modifier = modifier,
                shape = shape,
                enabled = enabled,
                colors =
                    ButtonDefaults.textButtonColors(
                        contentColor = MaterialTheme.colorScheme.onSurfaceVariant,
                    ),
            ) { Text(text) }
    }
}

enum class BleBadgeTone { Blue, Green, Yellow, Red, Neutral }

@Composable
fun BleBadge(
    text: String,
    tone: BleBadgeTone = BleBadgeTone.Blue,
) {
    val color: Color =
        when (tone) {
            BleBadgeTone.Blue -> MaterialTheme.colorScheme.primary
            BleBadgeTone.Green -> MaterialTheme.colorScheme.tertiary
            BleBadgeTone.Yellow -> BleDarkWarning
            BleBadgeTone.Red -> MaterialTheme.colorScheme.error
            BleBadgeTone.Neutral -> MaterialTheme.colorScheme.onSurfaceVariant
        }
    Surface(
        color = color.copy(alpha = 0.15f),
        contentColor = color,
        shape = RoundedCornerShape(BleRadius.pill),
    ) {
        Text(
            text,
            style = MaterialTheme.typography.labelSmall,
            modifier = Modifier.padding(horizontal = 10.dp, vertical = 3.dp),
        )
    }
}

/** Flat card: 1px outline, 10px radius (shadow only on elevation/press at call site). */
@Composable
fun BleCard(
    modifier: Modifier = Modifier,
    content: @Composable ColumnScope.() -> Unit,
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(BleRadius.xl),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = BorderStroke(1.dp, MaterialTheme.colorScheme.outline),
        elevation = CardDefaults.cardElevation(defaultElevation = 0.dp),
    ) {
        androidx.compose.foundation.layout.Column(Modifier.padding(BleSpacing.s6), content = content)
    }
}

/** Map an RSSI (dBm) reading to a 0–4 signal-strength level. */
fun rssiToLevel(rssi: Int): Int =
    when {
        rssi >= -55 -> 4
        rssi >= -65 -> 3
        rssi >= -75 -> 2
        rssi >= -85 -> 1
        else -> 0
    }

/**
 * bleRPC SignalBars — a 4-bar BLE/RSSI strength indicator. `level` is 0–4;
 * inactive bars use the border color. Tone shifts with strength.
 * Mirrors components/mobile/SignalBars.jsx.
 */
@Composable
fun BleSignalBars(
    level: Int,
    modifier: Modifier = Modifier,
    size: Dp = 16.dp,
) {
    val lv = level.coerceIn(0, 4)
    val active =
        when {
            lv >= 3 -> MaterialTheme.colorScheme.tertiary // success (green)
            lv == 2 -> BleDarkWarning // warning (yellow)
            else -> MaterialTheme.colorScheme.error // error (red)
        }
    val inactive = MaterialTheme.colorScheme.outline
    val gap = size * 0.18f
    val barW = (size - gap * 3) / 4
    Row(
        modifier = modifier.height(size),
        verticalAlignment = Alignment.Bottom,
    ) {
        for (i in 0 until 4) {
            Box(
                Modifier
                    .width(barW)
                    .fillMaxHeight((i + 1) * 0.25f)
                    .clip(RoundedCornerShape(1.dp))
                    .background(if (i < lv) active else inactive),
            )
            if (i < 3) Spacer(Modifier.width(gap))
        }
    }
}
