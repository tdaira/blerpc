import SwiftUI

// blerpc.net dark theme colors
private let bgPrimary = Color(red: 0x1A/255.0, green: 0x1B/255.0, blue: 0x26/255.0)
private let bgSecondary = Color(red: 0x24/255.0, green: 0x28/255.0, blue: 0x3B/255.0)
private let bgCode = Color(red: 0x1E/255.0, green: 0x20/255.0, blue: 0x30/255.0)
private let textPrimary = Color(red: 0xC0/255.0, green: 0xCA/255.0, blue: 0xF5/255.0)
private let textSecondary = Color(red: 0xA9/255.0, green: 0xB1/255.0, blue: 0xD6/255.0)
private let accent = Color(red: 0x00/255.0, green: 0x82/255.0, blue: 0xFC/255.0)
private let borderColor = Color(red: 0x3B/255.0, green: 0x42/255.0, blue: 0x61/255.0)
private let success = Color(red: 0x9E/255.0, green: 0xCE/255.0, blue: 0x6A/255.0)
private let error = Color(red: 0xF7/255.0, green: 0x76/255.0, blue: 0x8E/255.0)

struct ContentView: View {
    @ObservedObject var testRunner: TestRunner
    @State private var isRunning = false
    @State private var isScanning = false
    @State private var scannedDevices: [ScannedDevice] = []
    @State private var showCopied = false

    var body: some View {
        VStack(spacing: 16) {
            HStack(spacing: 0) {
                Text("ble")
                    .font(.system(size: 28, weight: .black))
                    .foregroundColor(accent)
                Text("RPC")
                    .font(.system(size: 28, weight: .black))
                    .foregroundColor(textPrimary)
                Text(" Central")
                    .font(.system(size: 28, weight: .regular))
                    .foregroundColor(textPrimary)
            }
            .padding(.top, 16)

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
                    .foregroundColor(.white)
            }
            .buttonStyle(.borderedProminent)
            .tint(accent)
            .disabled(isScanning || isRunning)
            .padding(.horizontal, 16)

            if !scannedDevices.isEmpty {
                VStack(alignment: .leading, spacing: 4) {
                    Text("Devices (\(scannedDevices.count))")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(textPrimary)
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
                                                .foregroundColor(textPrimary)
                                            Text(device.id.uuidString)
                                                .font(.system(size: 11, design: .monospaced))
                                                .foregroundColor(textSecondary)
                                        }
                                        Spacer()
                                        Text("\(device.rssi) dBm")
                                            .font(.system(size: 13, design: .monospaced))
                                            .foregroundColor(textSecondary)
                                    }
                                    .padding(.horizontal, 16)
                                    .padding(.vertical, 8)
                                }
                                .disabled(isRunning)
                                Divider()
                                    .background(borderColor)
                                    .padding(.leading, 16)
                            }
                        }
                    }
                    .frame(maxHeight: 200)
                    .background(bgSecondary)
                    .clipShape(RoundedRectangle(cornerRadius: 8))
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(borderColor, lineWidth: 1)
                    )
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
                        .foregroundColor(accent)
                }
                .disabled(testRunner.logs.isEmpty)
            }
            .padding(.horizontal, 16)

            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 2) {
                        ForEach(Array(testRunner.logs.enumerated()), id: \.offset) { index, line in
                            Text(line)
                                .foregroundColor(colorForLine(line))
                                .font(.system(size: 13, design: .monospaced))
                                .id(index)
                        }
                    }
                    .padding(12)
                }
                .background(bgCode)
                .clipShape(RoundedRectangle(cornerRadius: 8))
                .overlay(
                    RoundedRectangle(cornerRadius: 8)
                        .stroke(borderColor, lineWidth: 1)
                )
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
        .background(bgPrimary.ignoresSafeArea())
    }

    private func colorForLine(_ line: String) -> Color {
        if line.hasPrefix("[PASS]") {
            return success
        } else if line.hasPrefix("[FAIL]") || line.hasPrefix("[ERROR]") {
            return error
        } else if line.hasPrefix("[BENCH]") {
            return accent
        } else {
            return textPrimary
        }
    }
}
