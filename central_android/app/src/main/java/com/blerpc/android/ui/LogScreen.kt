package com.blerpc.android.ui

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.KeyboardArrowRight
import androidx.compose.material.icons.filled.Share
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Divider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.blerpc.android.ble.ScannedDevice
import com.blerpc.android.design.BleBadge
import com.blerpc.android.design.BleBadgeTone
import com.blerpc.android.design.BleMono
import com.blerpc.android.design.BleOnAccent
import com.blerpc.android.design.BleRadius
import com.blerpc.android.design.BleSignalBars
import com.blerpc.android.design.BleSpacing
import com.blerpc.android.design.rssiToLevel
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@Suppress("ktlint:standard:function-naming")
@Composable
fun LogScreen(
    logs: List<String>,
    isRunning: Boolean,
    isScanning: Boolean,
    scannedDevices: List<ScannedDevice>,
    onScan: () -> Unit,
    onSelectDevice: (ScannedDevice) -> Unit,
) {
    val colors = MaterialTheme.colorScheme
    val listState = rememberLazyListState()

    LaunchedEffect(logs.size) {
        if (logs.isNotEmpty()) {
            listState.animateScrollToItem(logs.size - 1)
        }
    }

    Column(
        modifier =
            Modifier
                .fillMaxSize()
                .background(colors.background)
                .padding(BleSpacing.s4),
    ) {
        // ── Brand wordmark ──────────────────────────────────────
        Text(
            text =
                buildAnnotatedString {
                    withStyle(SpanStyle(color = colors.primary, fontWeight = FontWeight.Black)) {
                        append("ble")
                    }
                    withStyle(SpanStyle(color = colors.onBackground, fontWeight = FontWeight.Black)) {
                        append("RPC")
                    }
                    withStyle(SpanStyle(color = colors.onBackground, fontWeight = FontWeight.Normal)) {
                        append(" Central")
                    }
                },
            fontSize = 24.sp,
            modifier = Modifier.padding(bottom = BleSpacing.s4),
        )

        // ── Scan action ─────────────────────────────────────────
        Button(
            onClick = onScan,
            enabled = !isScanning && !isRunning,
            shape = RoundedCornerShape(BleRadius.md),
            colors =
                ButtonDefaults.buttonColors(
                    containerColor = colors.primary,
                    contentColor = BleOnAccent,
                    disabledContainerColor = colors.surface,
                    disabledContentColor = colors.onSurfaceVariant,
                ),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(if (isScanning) "Scanning…" else "Scan")
        }

        // ── Device list ─────────────────────────────────────────
        if (scannedDevices.isNotEmpty()) {
            Spacer(modifier = Modifier.height(BleSpacing.s3))

            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.padding(bottom = BleSpacing.s2),
            ) {
                Text(
                    text = "DEVICES",
                    color = colors.onSurfaceVariant,
                    fontWeight = FontWeight.SemiBold,
                    fontSize = 12.sp,
                    letterSpacing = 0.8.sp,
                )
                Spacer(Modifier.width(BleSpacing.s2))
                BleBadge(text = scannedDevices.size.toString(), tone = BleBadgeTone.Blue)
            }

            LazyColumn(
                modifier =
                    Modifier
                        .fillMaxWidth()
                        .heightIn(max = 220.dp)
                        .clip(RoundedCornerShape(BleRadius.xl))
                        .background(colors.surface)
                        .border(1.dp, colors.outline, RoundedCornerShape(BleRadius.xl)),
            ) {
                items(scannedDevices, key = { it.address }) { device ->
                    Row(
                        modifier =
                            Modifier
                                .fillMaxWidth()
                                .clickable(enabled = !isRunning) { onSelectDevice(device) }
                                .padding(horizontal = BleSpacing.s3, vertical = 10.dp),
                        horizontalArrangement = Arrangement.spacedBy(BleSpacing.s3),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        BleSignalBars(level = rssiToLevel(device.rssi))
                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = device.name ?: "Unknown",
                                color = colors.onSurface,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Medium,
                            )
                            Text(
                                text = device.address,
                                color = colors.onSurfaceVariant,
                                fontSize = 11.sp,
                                fontFamily = BleMono,
                            )
                        }
                        Text(
                            text = "${device.rssi} dBm",
                            color = colors.onSurfaceVariant,
                            fontSize = 13.sp,
                            fontFamily = BleMono,
                        )
                        Icon(
                            imageVector = Icons.Default.KeyboardArrowRight,
                            contentDescription = null,
                            tint = colors.onSurfaceVariant,
                            modifier = Modifier.size(18.dp),
                        )
                    }
                    Divider(color = colors.outline)
                }
            }
        }

        Spacer(modifier = Modifier.height(BleSpacing.s3))

        // ── Copy logs ───────────────────────────────────────────
        val context = LocalContext.current
        val scope = rememberCoroutineScope()
        var showCopied by remember { mutableStateOf(false) }

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.End,
        ) {
            TextButton(
                onClick = {
                    val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                    clipboard.setPrimaryClip(ClipData.newPlainText("logs", logs.joinToString("\n")))
                    showCopied = true
                    scope.launch {
                        delay(1500)
                        showCopied = false
                    }
                },
                enabled = logs.isNotEmpty(),
                shape = RoundedCornerShape(BleRadius.md),
            ) {
                Icon(
                    imageVector = if (showCopied) Icons.Default.Check else Icons.Default.Share,
                    contentDescription = null,
                    tint = if (logs.isEmpty()) colors.onSurfaceVariant else colors.primary,
                    modifier = Modifier.size(16.dp),
                )
                Spacer(modifier = Modifier.width(BleSpacing.s1))
                Text(
                    text = if (showCopied) "Copied!" else "Copy Logs",
                    color = if (logs.isEmpty()) colors.onSurfaceVariant else colors.primary,
                    fontSize = 13.sp,
                )
            }
        }

        // ── Log console (code surface) ──────────────────────────
        LazyColumn(
            state = listState,
            modifier =
                Modifier
                    .fillMaxSize()
                    .clip(RoundedCornerShape(BleRadius.lg))
                    .background(colors.surfaceVariant)
                    .border(1.dp, colors.outline, RoundedCornerShape(BleRadius.lg))
                    .padding(BleSpacing.s3),
        ) {
            if (logs.isEmpty()) {
                item {
                    Text(
                        text = "Scan for devices, then tap one to run tests.",
                        color = colors.onSurfaceVariant,
                        fontSize = 13.sp,
                    )
                }
            }
            items(logs) { line ->
                val color =
                    when {
                        line.startsWith("[PASS]") -> colors.tertiary
                        line.startsWith("[FAIL]") || line.startsWith("[ERROR]") -> colors.error
                        line.startsWith("[BENCH]") -> colors.primary
                        else -> colors.onSurface
                    }
                Text(
                    text = line,
                    color = color,
                    fontFamily = BleMono,
                    fontSize = 13.sp,
                    modifier = Modifier.padding(vertical = 1.dp),
                )
            }
        }
    }
}
