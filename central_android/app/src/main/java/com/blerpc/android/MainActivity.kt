package com.blerpc.android

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.*
import androidx.core.content.ContextCompat
import com.blerpc.android.ble.ScannedDevice
import com.blerpc.android.client.BlerpcClient
import com.blerpc.android.test.TestRunner
import com.blerpc.android.ui.LogScreen
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {
    private lateinit var testRunner: TestRunner
    private val scope = CoroutineScope(Dispatchers.IO)
    private var autoRunPending = false
    private var autoRunIterations = 1

    private val permissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { /* permissions granted or denied */ }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        testRunner = TestRunner(applicationContext)
        requestPermissions()

        // Check if launched with --es action run_tests --ei iterations N
        if (intent?.getStringExtra("action") == "run_tests") {
            autoRunPending = true
            autoRunIterations = intent.getIntExtra("iterations", 1)
        }

        setContent {
            MaterialTheme {
                val logs by testRunner.logs.collectAsState()
                var isRunning by remember { mutableStateOf(false) }
                var isScanning by remember { mutableStateOf(false) }
                var scannedDevices by remember { mutableStateOf<List<ScannedDevice>>(emptyList()) }

                // Auto-run tests if requested via intent
                LaunchedEffect(autoRunPending) {
                    if (autoRunPending) {
                        val iters = autoRunIterations
                        autoRunPending = false
                        isRunning = true
                        scope.launch {
                            try {
                                testRunner.runAll(iters)
                            } finally {
                                isRunning = false
                            }
                        }
                    }
                }

                LogScreen(
                    logs = logs,
                    isRunning = isRunning,
                    isScanning = isScanning,
                    scannedDevices = scannedDevices,
                    onScan = {
                        isScanning = true
                        scannedDevices = emptyList()
                        scope.launch {
                            try {
                                val client = BlerpcClient(applicationContext)
                                scannedDevices = client.scan()
                            } catch (_: Exception) {
                                // Scan failed
                            } finally {
                                isScanning = false
                            }
                        }
                    },
                    onRunTests = {
                        isRunning = true
                        scannedDevices = emptyList()
                        scope.launch {
                            try {
                                testRunner.runAll()
                            } finally {
                                isRunning = false
                            }
                        }
                    },
                    onSelectDevice = { device ->
                        isRunning = true
                        scannedDevices = emptyList()
                        scope.launch {
                            try {
                                testRunner.runAll(device = device)
                            } finally {
                                isRunning = false
                            }
                        }
                    }
                )
            }
        }
    }

    override fun onNewIntent(intent: Intent?) {
        super.onNewIntent(intent)
        if (intent?.getStringExtra("action") == "run_tests") {
            autoRunPending = true
            autoRunIterations = intent.getIntExtra("iterations", 1)
        }
    }

    private fun requestPermissions() {
        val needed = arrayOf(
            Manifest.permission.BLUETOOTH_SCAN,
            Manifest.permission.BLUETOOTH_CONNECT,
            Manifest.permission.ACCESS_FINE_LOCATION
        ).filter {
            ContextCompat.checkSelfPermission(this, it) != PackageManager.PERMISSION_GRANTED
        }
        if (needed.isNotEmpty()) {
            permissionLauncher.launch(needed.toTypedArray())
        }
    }
}
