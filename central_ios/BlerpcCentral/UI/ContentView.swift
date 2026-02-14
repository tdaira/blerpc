import SwiftUI

struct ContentView: View {
    @ObservedObject var testRunner: TestRunner
    @State private var isRunning = false
    @State private var isScanning = false
    @State private var scannedDevices: [ScannedDevice] = []
    @State private var showCopied = false

    var body: some View {
        VStack(spacing: 16) {
            Text("blerpc iOS Central")
                .font(.title)
                .padding(.top, 16)

            HStack(spacing: 12) {
                Button(action: {
                    isScanning = true
                    scannedDevices = []
                    Task {
                        do {
                            let client = BlerpcClient()
                            scannedDevices = try await client.scan()
                        } catch {
                            testRunner.logs.append("[ERROR] Scan failed: \(error)")
                        }
                        isScanning = false
                    }
                }) {
                    Text(isScanning ? "Scanning..." : "Scan")
                        .frame(maxWidth: .infinity)
                }
                .buttonStyle(.borderedProminent)
                .disabled(isScanning || isRunning)

                Button(action: {
                    isRunning = true
                    Task {
                        await testRunner.runAll()
                        isRunning = false
                    }
                }) {
                    Text(isRunning ? "Running..." : "Run Tests")
                        .frame(maxWidth: .infinity)
                }
                .buttonStyle(.borderedProminent)
                .disabled(isRunning || isScanning)
            }
            .padding(.horizontal, 16)

            if !scannedDevices.isEmpty {
                VStack(alignment: .leading, spacing: 4) {
                    Text("Devices (\(scannedDevices.count))")
                        .font(.headline)
                        .padding(.horizontal, 16)

                    ScrollView {
                        LazyVStack(spacing: 0) {
                            ForEach(scannedDevices) { device in
                                Button(action: {
                                    scannedDevices = []
                                    isRunning = true
                                    Task {
                                        await testRunner.runAll(device: device)
                                        isRunning = false
                                    }
                                }) {
                                    HStack {
                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(device.name ?? "Unknown")
                                                .font(.system(size: 15, weight: .medium))
                                                .foregroundColor(.primary)
                                            Text(device.id.uuidString)
                                                .font(.system(size: 11, design: .monospaced))
                                                .foregroundColor(.secondary)
                                        }
                                        Spacer()
                                        Text("\(device.rssi) dBm")
                                            .font(.system(size: 13, design: .monospaced))
                                            .foregroundColor(.secondary)
                                    }
                                    .padding(.horizontal, 16)
                                    .padding(.vertical, 8)
                                }
                                .disabled(isRunning)
                                Divider().padding(.leading, 16)
                            }
                        }
                    }
                    .frame(maxHeight: 200)
                    .background(Color(UIColor.secondarySystemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 8))
                    .padding(.horizontal, 16)
                }
            }

            HStack {
                Spacer()
                Button(action: {
                    UIPasteboard.general.string = testRunner.logs.joined(separator: "\n")
                    showCopied = true
                    Task {
                        try? await Task.sleep(nanoseconds: 1_500_000_000)
                        showCopied = false
                    }
                }) {
                    Label(showCopied ? "Copied!" : "Copy Logs", systemImage: showCopied ? "checkmark" : "doc.on.doc")
                        .font(.system(size: 13))
                }
                .disabled(testRunner.logs.isEmpty)
            }
            .padding(.horizontal, 16)

            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 4) {
                        ForEach(Array(testRunner.logs.enumerated()), id: \.offset) { index, line in
                            Text(line)
                                .foregroundColor(colorForLine(line))
                                .font(.system(size: 13, design: .monospaced))
                                .id(index)
                        }
                    }
                    .padding(8)
                }
                .background(Color(white: 0.12))
                .clipShape(RoundedRectangle(cornerRadius: 8))
                .padding(.horizontal, 16)
                .onChange(of: testRunner.logs.count) { newCount in
                    if newCount > 0 {
                        withAnimation {
                            proxy.scrollTo(newCount - 1, anchor: .bottom)
                        }
                    }
                }
            }
        }
        .padding(.bottom, 16)
    }

    private func colorForLine(_ line: String) -> Color {
        if line.hasPrefix("[PASS]") {
            return Color(red: 0.31, green: 0.79, blue: 0.69)
        } else if line.hasPrefix("[FAIL]") || line.hasPrefix("[ERROR]") {
            return Color(red: 1.0, green: 0.42, blue: 0.42)
        } else if line.hasPrefix("[BENCH]") {
            return Color(red: 0.6, green: 0.8, blue: 1.0)
        } else {
            return Color(white: 0.83)
        }
    }
}
