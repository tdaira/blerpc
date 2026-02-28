package com.blerpc.android.ui

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Share
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.blerpc.android.ble.ScannedDevice
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

// blerpc.net dark theme colors
private val BgPrimary = Color(0xFF1A1B26)
private val BgSecondary = Color(0xFF24283B)
private val BgCode = Color(0xFF1E2030)
private val TextPrimary = Color(0xFFC0CAF5)
private val TextSecondary = Color(0xFFA9B1D6)
private val Accent = Color(0xFF0082FC)
private val Border = Color(0xFF3B4261)
private val Success = Color(0xFF9ECE6A)
private val Error = Color(0xFFF7768E)
private val NavBg = Color(0xFF16161E)

@Composable
fun LogScreen(
    logs: List<String>,
    isRunning: Boolean,
    isScanning: Boolean,
    scannedDevices: List<ScannedDevice>,
    onScan: () -> Unit,
    onRunTests: () -> Unit,
    onSelectDevice: (ScannedDevice) -> Unit
) {
    val listState = rememberLazyListState()

    LaunchedEffect(logs.size) {
        if (logs.isNotEmpty()) {
            listState.animateScrollToItem(logs.size - 1)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(BgPrimary)
            .padding(16.dp)
    ) {
        Text(
            text = buildAnnotatedString {
                withStyle(SpanStyle(color = Accent, fontWeight = FontWeight.Black)) {
                    append("ble")
                }
                withStyle(SpanStyle(color = TextPrimary, fontWeight = FontWeight.Black)) {
                    append("RPC")
                }
                withStyle(SpanStyle(color = TextPrimary, fontWeight = FontWeight.Normal)) {
                    append(" Central")
                }
            },
            fontSize = 24.sp,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Button(
                onClick = onScan,
                enabled = !isScanning && !isRunning,
                colors = ButtonDefaults.buttonColors(
                    containerColor = Accent,
                    contentColor = Color.White,
                    disabledContainerColor = BgSecondary,
                    disabledContentColor = TextSecondary
                ),
                modifier = Modifier.weight(1f)
            ) {
                Text(if (isScanning) "Scanning..." else "Scan")
            }

            Button(
                onClick = onRunTests,
                enabled = !isRunning && !isScanning,
                colors = ButtonDefaults.buttonColors(
                    containerColor = Accent,
                    contentColor = Color.White,
                    disabledContainerColor = BgSecondary,
                    disabledContentColor = TextSecondary
                ),
                modifier = Modifier.weight(1f)
            ) {
                Text(if (isRunning) "Running..." else "Run Tests")
            }
        }

        if (scannedDevices.isNotEmpty()) {
            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = "Devices (${scannedDevices.size})",
                color = TextPrimary,
                fontWeight = FontWeight.SemiBold,
                fontSize = 14.sp,
                modifier = Modifier.padding(bottom = 4.dp)
            )

            LazyColumn(
                modifier = Modifier
                    .fillMaxWidth()
                    .heightIn(max = 200.dp)
                    .clip(RoundedCornerShape(8.dp))
                    .background(BgSecondary)
                    .border(1.dp, Border, RoundedCornerShape(8.dp))
                    .padding(4.dp)
            ) {
                items(scannedDevices, key = { it.address }) { device ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable(enabled = !isRunning) { onSelectDevice(device) }
                            .padding(horizontal = 12.dp, vertical = 8.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = device.name ?: "Unknown",
                                color = TextPrimary,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Medium
                            )
                            Text(
                                text = device.address,
                                color = TextSecondary,
                                fontSize = 11.sp,
                                fontFamily = FontFamily.Monospace
                            )
                        }
                        Text(
                            text = "${device.rssi} dBm",
                            color = TextSecondary,
                            fontSize = 13.sp,
                            fontFamily = FontFamily.Monospace
                        )
                    }
                    Divider(color = Border)
                }
            }
        }

        Spacer(modifier = Modifier.height(12.dp))

        val context = LocalContext.current
        val scope = rememberCoroutineScope()
        var showCopied by remember { mutableStateOf(false) }

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.End
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
                enabled = logs.isNotEmpty()
            ) {
                Icon(
                    imageVector = if (showCopied) Icons.Default.Check else Icons.Default.Share,
                    contentDescription = null,
                    tint = if (logs.isEmpty()) TextSecondary else Accent,
                    modifier = Modifier.size(16.dp)
                )
                Spacer(modifier = Modifier.width(4.dp))
                Text(
                    text = if (showCopied) "Copied!" else "Copy Logs",
                    color = if (logs.isEmpty()) TextSecondary else Accent,
                    fontSize = 13.sp
                )
            }
        }

        LazyColumn(
            state = listState,
            modifier = Modifier
                .fillMaxSize()
                .clip(RoundedCornerShape(8.dp))
                .background(BgCode)
                .border(1.dp, Border, RoundedCornerShape(8.dp))
                .padding(12.dp)
        ) {
            items(logs) { line ->
                val color = when {
                    line.startsWith("[PASS]") -> Success
                    line.startsWith("[FAIL]") || line.startsWith("[ERROR]") -> Error
                    line.startsWith("[BENCH]") -> Accent
                    else -> TextPrimary
                }
                Text(
                    text = line,
                    color = color,
                    fontFamily = FontFamily.Monospace,
                    fontSize = 13.sp,
                    modifier = Modifier.padding(vertical = 1.dp)
                )
            }
        }
    }
}
