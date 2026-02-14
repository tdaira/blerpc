import Foundation
import SwiftProtobuf

@MainActor
final class TestRunner: ObservableObject {
    @Published var logs: [String] = []
    private var running = false
    private var passCount = 0
    private var failCount = 0

    private func log(_ msg: String) {
        print("[BlerpcTest] \(msg)")
        logs.append(msg)
    }

    func runAll(iterations: Int = 1) async {
        guard !running else { return }
        running = true
        logs = []
        passCount = 0
        failCount = 0

        let client = BlerpcClient()
        do {
            log("Connecting to blerpc peripheral...")
            try await client.connect()
            let mtu = client.mtu
            log("Connected. MTU=\(mtu)")

            for iter in 1...iterations {
                if iterations > 1 {
                    log("--- Iteration \(iter)/\(iterations) ---")
                }

                await runTest(client: client, name: "echo_basic") {
                    let resp = try await client.echo(message: "hello")
                    guard resp.message == "hello" else {
                        throw TestError.assertion("Expected 'hello', got '\(resp.message)'")
                    }
                }

                await runTest(client: client, name: "echo_empty") {
                    let resp = try await client.echo(message: "")
                    guard resp.message == "" else {
                        throw TestError.assertion("Expected empty, got '\(resp.message)'")
                    }
                }

                await runTest(client: client, name: "flash_read_basic") {
                    let resp = try await client.flashRead(address: 0, length: 64)
                    guard resp.data.count == 64 else {
                        throw TestError.assertion("Expected 64 bytes, got \(resp.data.count)")
                    }
                }

                await runTest(client: client, name: "flash_read_8kb") {
                    let resp = try await client.flashRead(address: 0, length: 8192)
                    guard resp.data.count == 8192 else {
                        throw TestError.assertion("Expected 8192 bytes, got \(resp.data.count)")
                    }
                }

                await runTest(client: client, name: "data_write") {
                    let testData = Data(0..<64)
                    let resp = try await client.dataWrite(data: testData)
                    guard resp.length == 64 else {
                        throw TestError.assertion("Expected length 64, got \(resp.length)")
                    }
                }

                await runTest(client: client, name: "counter_stream") {
                    let results = try await client.counterStreamAll(count: 5)
                    guard results.count == 5 else {
                        throw TestError.assertion("Expected 5 results, got \(results.count)")
                    }
                    for i in 0..<5 {
                        guard results[i].seq == UInt32(i) else {
                            throw TestError.assertion("Expected seq=\(i), got \(results[i].seq)")
                        }
                        guard results[i].value == Int32(i * 10) else {
                            throw TestError.assertion(
                                "Expected value=\(i * 10), got \(results[i].value)"
                            )
                        }
                    }
                }

                await runTest(client: client, name: "counter_upload") {
                    let resp = try await client.counterUploadAll(count: 5)
                    guard resp.receivedCount == 5 else {
                        throw TestError.assertion(
                            "Expected received_count=5, got \(resp.receivedCount)"
                        )
                    }
                }
            }

            log("=== Functional: \(passCount) passed, \(failCount) failed (\(iterations) iterations) ===")

            // Throughput benchmarks
            log("")
            log("=== Throughput Benchmarks ===")
            await benchmarkFlashReadThroughput(client: client)
            await benchmarkFlashReadOverhead(client: client)
            await benchmarkEchoRoundtrip(client: client)
            await benchmarkDataWriteThroughput(client: client)
            await benchmarkStreamThroughput(client: client)

        } catch {
            log("[ERROR] \(error)")
        }
        client.disconnect()
        running = false
    }

    private func benchmarkFlashReadThroughput(client: BlerpcClient) async {
        let readSize: UInt32 = 8192
        let count = 10
        let totalBytes = Int(readSize) * count

        // Warmup
        _ = try? await client.flashRead(address: 0, length: readSize)

        let start = ContinuousClock.now
        for _ in 0..<count {
            let resp = try? await client.flashRead(address: 0, length: readSize)
            guard resp?.data.count == Int(readSize) else { continue }
        }
        let elapsed = start.duration(to: .now)
        let elapsedMs = Double(elapsed.components.attoseconds) / 1e15 + Double(elapsed.components.seconds) * 1000
        let kbPerSec = Double(totalBytes) / 1024.0 / (elapsedMs / 1000.0)
        let msPerCall = elapsedMs / Double(count)

        log(String(format: "[BENCH] flash_read_throughput: %.1f KB/s (%d bytes in %.0f ms, %.1f ms/call)",
                   kbPerSec, totalBytes, elapsedMs, msPerCall))
    }

    private func benchmarkFlashReadOverhead(client: BlerpcClient) async {
        let count = 20

        // Warmup
        _ = try? await client.flashRead(address: 0, length: 1)

        let start = ContinuousClock.now
        for _ in 0..<count {
            _ = try? await client.flashRead(address: 0, length: 1)
        }
        let elapsed = start.duration(to: .now)
        let elapsedMs = Double(elapsed.components.attoseconds) / 1e15 + Double(elapsed.components.seconds) * 1000
        let msPerCall = elapsedMs / Double(count)

        log(String(format: "[BENCH] flash_read_overhead: %.1f ms/call (1 byte x %d calls in %.0f ms)",
                   msPerCall, count, elapsedMs))
    }

    private func benchmarkEchoRoundtrip(client: BlerpcClient) async {
        let count = 50

        // Warmup
        _ = try? await client.echo(message: "x")

        let start = ContinuousClock.now
        for _ in 0..<count {
            _ = try? await client.echo(message: "hello")
        }
        let elapsed = start.duration(to: .now)
        let elapsedMs = Double(elapsed.components.attoseconds) / 1e15 + Double(elapsed.components.seconds) * 1000
        let msPerCall = elapsedMs / Double(count)

        log(String(format: "[BENCH] echo_roundtrip: %.1f ms/call (%d calls in %.0f ms)",
                   msPerCall, count, elapsedMs))
    }

    private func benchmarkDataWriteThroughput(client: BlerpcClient) async {
        let writeSize = 200
        let count = 20
        let totalBytes = writeSize * count
        let testData = Data((0..<writeSize).map { UInt8($0 % 256) })

        // Warmup
        _ = try? await client.dataWrite(data: testData)

        let start = ContinuousClock.now
        for _ in 0..<count {
            _ = try? await client.dataWrite(data: testData)
        }
        let elapsed = start.duration(to: .now)
        let elapsedMs = Double(elapsed.components.attoseconds) / 1e15 + Double(elapsed.components.seconds) * 1000
        let kbPerSec = Double(totalBytes) / 1024.0 / (elapsedMs / 1000.0)
        let msPerCall = elapsedMs / Double(count)

        log(String(format: "[BENCH] data_write_throughput: %.1f KB/s (%d bytes in %.0f ms, %.1f ms/call)",
                   kbPerSec, totalBytes, elapsedMs, msPerCall))
    }

    private func benchmarkStreamThroughput(client: BlerpcClient) async {
        let count: UInt32 = 20

        // counter_stream (P->C)
        let start1 = ContinuousClock.now
        let results = try? await client.counterStreamAll(count: count)
        let elapsed1 = start1.duration(to: .now)
        let elapsedMs1 = Double(elapsed1.components.attoseconds) / 1e15 + Double(elapsed1.components.seconds) * 1000
        if let results = results {
            log(String(format: "[BENCH] counter_stream (P->C): %d items in %.0f ms (%.1f ms/item)",
                       results.count, elapsedMs1, elapsedMs1 / Double(results.count)))
        }

        // counter_upload (C->P)
        let start2 = ContinuousClock.now
        let resp = try? await client.counterUploadAll(count: Int(count))
        let elapsed2 = start2.duration(to: .now)
        let elapsedMs2 = Double(elapsed2.components.attoseconds) / 1e15 + Double(elapsed2.components.seconds) * 1000
        if let resp = resp {
            log(String(format: "[BENCH] counter_upload (C->P): %d items in %.0f ms (%.1f ms/item)",
                       resp.receivedCount, elapsedMs2, elapsedMs2 / Double(count)))
        }
    }

    private func runTest(
        client: BlerpcClient,
        name: String,
        block: @escaping () async throws -> Void
    ) async {
        do {
            try await block()
            passCount += 1
            log("[PASS] \(name)")
        } catch {
            failCount += 1
            log("[FAIL] \(name): \(error)")
            try? await Task.sleep(nanoseconds: 500_000_000)
            client.transport.drainNotifications()
        }
    }
}

private enum TestError: Error {
    case assertion(String)
}
