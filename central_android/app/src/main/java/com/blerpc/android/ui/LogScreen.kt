package com.blerpc.android.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.blerpc.android.ble.ScannedDevice

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
            .padding(16.dp)
    ) {
        Text(
            text = "blerpc Android Central",
            style = MaterialTheme.typography.headlineMedium,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Button(
                onClick = onScan,
                enabled = !isScanning && !isRunning,
                modifier = Modifier.weight(1f)
            ) {
                Text(if (isScanning) "Scanning..." else "Scan")
            }

            Button(
                onClick = onRunTests,
                enabled = !isRunning && !isScanning,
                modifier = Modifier.weight(1f)
            ) {
                Text(if (isRunning) "Running..." else "Run Tests")
            }
        }

        if (scannedDevices.isNotEmpty()) {
            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = "Devices (${scannedDevices.size})",
                style = MaterialTheme.typography.titleSmall,
                modifier = Modifier.padding(bottom = 4.dp)
            )

            LazyColumn(
                modifier = Modifier
                    .fillMaxWidth()
                    .heightIn(max = 200.dp)
                    .background(Color(0xFF2D2D2D))
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
                                color = Color(0xFFD4D4D4),
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Medium
                            )
                            Text(
                                text = device.address,
                                color = Color(0xFF888888),
                                fontSize = 11.sp,
                                fontFamily = FontFamily.Monospace
                            )
                        }
                        Text(
                            text = "${device.rssi} dBm",
                            color = Color(0xFF888888),
                            fontSize = 13.sp,
                            fontFamily = FontFamily.Monospace
                        )
                    }
                    Divider(color = Color(0xFF444444))
                }
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        LazyColumn(
            state = listState,
            modifier = Modifier
                .fillMaxSize()
                .background(Color(0xFF1E1E1E))
                .padding(8.dp)
        ) {
            items(logs) { line ->
                val color = when {
                    line.startsWith("[PASS]") -> Color(0xFF4EC9B0)
                    line.startsWith("[FAIL]") -> Color(0xFFFF6B6B)
                    line.startsWith("[ERROR]") -> Color(0xFFFF6B6B)
                    else -> Color(0xFFD4D4D4)
                }
                Text(
                    text = line,
                    color = color,
                    fontFamily = FontFamily.Monospace,
                    fontSize = 13.sp,
                    modifier = Modifier.padding(vertical = 2.dp)
                )
            }
        }
    }
}
