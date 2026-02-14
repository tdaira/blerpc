package com.blerpc.android.test

import android.content.Context
import android.util.Log
import com.blerpc.android.ble.ScannedDevice
import com.blerpc.android.client.BlerpcClient
import com.google.protobuf.ByteString
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class TestRunner(private val context: Context) {
    private val _logs = MutableStateFlow<List<String>>(emptyList())
    val logs: StateFlow<List<String>> = _logs.asStateFlow()

    private var running = false
    private var passCount = 0
    private var failCount = 0

    private fun log(msg: String) {
        Log.i("BlerpcTest", msg)
        _logs.value = _logs.value + msg
    }

    suspend fun runAll(iterations: Int = 1, device: ScannedDevice? = null) {
        if (running) return
        running = true
        _logs.value = emptyList()
        passCount = 0
        failCount = 0

        val client = BlerpcClient(context)
        try {
            val target: ScannedDevice
            if (device != null) {
                target = device
            } else {
                log("Scanning for blerpc peripherals...")
                val devices = client.scan()
                if (devices.isEmpty()) {
                    log("[ERROR] No blerpc devices found")
                    running = false
                    return
                }
                target = devices.first()
            }
            log("Connecting to ${target.name ?: target.address}...")
            client.connect(target)
            log("Connected. MTU=${client.mtu}")

            for (iter in 1..iterations) {
                if (iterations > 1) {
                    log("--- Iteration $iter/$iterations ---")
                }

                runTest(client, "echo_basic") {
                    val resp = client.echo(message = "hello")
                    check(resp.message == "hello") { "Expected 'hello', got '${resp.message}'" }
                }

                runTest(client, "echo_empty") {
                    val resp = client.echo(message = "")
                    check(resp.message == "") { "Expected empty, got '${resp.message}'" }
                }

                runTest(client, "flash_read_basic") {
                    val resp = client.flashRead(address = 0, length = 64)
                    check(resp.data.size() == 64) { "Expected 64 bytes, got ${resp.data.size()}" }
                }

                runTest(client, "flash_read_8kb") {
                    val resp = client.flashRead(address = 0, length = 8192)
                    check(resp.data.size() == 8192) { "Expected 8192 bytes, got ${resp.data.size()}" }
                }

                runTest(client, "data_write") {
                    val testData = ByteArray(64) { it.toByte() }
                    val resp = client.dataWrite(data = ByteString.copyFrom(testData))
                    check(resp.length == 64) { "Expected length 64, got ${resp.length}" }
                }

                runTest(client, "counter_stream") {
                    val results = client.counterStreamAll(5)
                    check(results.size == 5) { "Expected 5 results, got ${results.size}" }
                    for (i in 0 until 5) {
                        check(results[i].first == i) { "Expected seq=$i, got ${results[i].first}" }
                        check(results[i].second == i * 10) { "Expected value=${i * 10}, got ${results[i].second}" }
                    }
                }

                runTest(client, "counter_upload") {
                    val resp = client.counterUploadAll(5)
                    check(resp.receivedCount == 5) { "Expected received_count=5, got ${resp.receivedCount}" }
                }
            }

            log("=== Functional: $passCount passed, $failCount failed ($iterations iterations) ===")

            // Throughput benchmarks
            log("")
            log("=== Throughput Benchmarks ===")
            benchmarkFlashReadThroughput(client)
            benchmarkFlashReadOverhead(client)
            benchmarkEchoRoundtrip(client)
            benchmarkDataWriteThroughput(client)
            benchmarkStreamThroughput(client)

        } catch (e: Exception) {
            log("[ERROR] ${e.message}")
        } finally {
            client.disconnect()
            running = false
        }
    }

    private suspend fun benchmarkFlashReadThroughput(client: BlerpcClient) {
        val readSize = 8192
        val count = 10
        val totalBytes = readSize * count

        // Warmup
        client.flashRead(address = 0, length = readSize)

        val startMs = System.currentTimeMillis()
        for (i in 0 until count) {
            val resp = client.flashRead(address = 0, length = readSize)
            check(resp.data.size() == readSize)
        }
        val elapsedMs = System.currentTimeMillis() - startMs
        val kbPerSec = totalBytes.toDouble() / 1024.0 / (elapsedMs.toDouble() / 1000.0)
        val msPerCall = elapsedMs.toDouble() / count

        log("[BENCH] flash_read_throughput: %.1f KB/s (%d bytes in %d ms, %.1f ms/call)".format(
            kbPerSec, totalBytes, elapsedMs, msPerCall))
    }

    private suspend fun benchmarkFlashReadOverhead(client: BlerpcClient) {
        val count = 20

        // Warmup
        client.flashRead(address = 0, length = 1)

        val startMs = System.currentTimeMillis()
        for (i in 0 until count) {
            client.flashRead(address = 0, length = 1)
        }
        val elapsedMs = System.currentTimeMillis() - startMs
        val msPerCall = elapsedMs.toDouble() / count

        log("[BENCH] flash_read_overhead: %.1f ms/call (1 byte × %d calls in %d ms)".format(
            msPerCall, count, elapsedMs))
    }

    private suspend fun benchmarkEchoRoundtrip(client: BlerpcClient) {
        val count = 50

        // Warmup
        client.echo(message = "x")

        val startMs = System.currentTimeMillis()
        for (i in 0 until count) {
            client.echo(message = "hello")
        }
        val elapsedMs = System.currentTimeMillis() - startMs
        val msPerCall = elapsedMs.toDouble() / count

        log("[BENCH] echo_roundtrip: %.1f ms/call (%d calls in %d ms)".format(
            msPerCall, count, elapsedMs))
    }

    private suspend fun benchmarkDataWriteThroughput(client: BlerpcClient) {
        val writeSize = 200
        val count = 20
        val totalBytes = writeSize * count
        val testData = ByteString.copyFrom(ByteArray(writeSize) { (it % 256).toByte() })

        // Warmup
        client.dataWrite(data = testData)

        val startMs = System.currentTimeMillis()
        for (i in 0 until count) {
            client.dataWrite(data = testData)
        }
        val elapsedMs = System.currentTimeMillis() - startMs
        val kbPerSec = totalBytes.toDouble() / 1024.0 / (elapsedMs.toDouble() / 1000.0)
        val msPerCall = elapsedMs.toDouble() / count

        log("[BENCH] data_write_throughput: %.1f KB/s (%d bytes in %d ms, %.1f ms/call)".format(
            kbPerSec, totalBytes, elapsedMs, msPerCall))
    }

    private suspend fun benchmarkStreamThroughput(client: BlerpcClient) {
        val count = 20

        // counter_stream (P→C): peripheral sends 'count' responses
        val startMs1 = System.currentTimeMillis()
        val results = client.counterStreamAll(count)
        val elapsedMs1 = System.currentTimeMillis() - startMs1
        check(results.size == count)
        log("[BENCH] counter_stream (P→C): %d items in %d ms (%.1f ms/item)".format(
            count, elapsedMs1, elapsedMs1.toDouble() / count))

        // counter_upload (C→P): central sends 'count' requests
        val startMs2 = System.currentTimeMillis()
        val resp = client.counterUploadAll(count)
        val elapsedMs2 = System.currentTimeMillis() - startMs2
        check(resp.receivedCount == count)
        log("[BENCH] counter_upload (C→P): %d items in %d ms (%.1f ms/item)".format(
            count, elapsedMs2, elapsedMs2.toDouble() / count))
    }

    private suspend inline fun runTest(client: BlerpcClient, name: String, block: () -> Unit) {
        try {
            block()
            passCount++
            log("[PASS] $name")
        } catch (e: Exception) {
            failCount++
            log("[FAIL] $name: ${e.message}")
            delay(500)
            client.transport.drainNotifications()
        }
    }
}
