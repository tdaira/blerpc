import SwiftUI

struct ContentView: View {
    @ObservedObject var testRunner: TestRunner
    @State private var isRunning = false
    @State private var isScanning = false
    @State private var scannedDevices: [ScannedDevice] = []
    @State private var showCopied = false

    // Device list sizing: fixed-height rows so the container hugs its content
    // (no empty space for one device) and scrolls only past the cap — matching
    // Android's LazyColumn + heightIn(max:).
    private let deviceRowHeight: CGFloat = 56
    private let deviceListMaxHeight: CGFloat = 220

    /// Exact list height for the current device count, capped at the max.
    private var deviceListHeight: CGFloat {
        let n = scannedDevices.count
        let dividers = CGFloat(max(0, n - 1)) // 1pt Divider between rows
        let content = CGFloat(n) * deviceRowHeight + dividers
        return min(content, deviceListMaxHeight)
    }

    var body: some View {
        VStack(spacing: BleRPCSpacing.s4) {
            // ── Brand wordmark ──────────────────────────────────
            HStack(spacing: 0) {
                Text("ble")
                    .font(.system(size: 28, weight: .black))
                    .foregroundColor(BleRPCColor.accent)
                Text("RPC")
                    .font(.system(size: 28, weight: .black))
                    .foregroundColor(BleRPCColor.text)
                Text(" Central")
                    .font(.system(size: 28, weight: .regular))
                    .foregroundColor(BleRPCColor.text)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, BleRPCSpacing.s4)
            .padding(.top, BleRPCSpacing.s4)

            // ── Scan action ─────────────────────────────────────
            scanButton
                .padding(.horizontal, BleRPCSpacing.s4)

            // ── Device list ─────────────────────────────────────
            if !scannedDevices.isEmpty {
                deviceList
                    .padding(.horizontal, BleRPCSpacing.s4)
            }

            // ── Copy logs ───────────────────────────────────────
            HStack {
                Spacer()
                Button(action: copyLogs) {
                    Label(showCopied ? "Copied!" : "Copy Logs", systemImage: showCopied ? "checkmark" : "doc.on.doc")
                        .font(.system(size: 13))
                        .foregroundColor(testRunner.logs.isEmpty ? BleRPCColor.textSecondary : BleRPCColor.accent)
                }
                .disabled(testRunner.logs.isEmpty)
            }
            .padding(.horizontal, BleRPCSpacing.s4)

            // ── Log console (code surface) ──────────────────────
            logConsole
                .padding(.horizontal, BleRPCSpacing.s4)
        }
        .padding(.bottom, BleRPCSpacing.s4)
        .background(BleRPCColor.bg.ignoresSafeArea())
        .preferredColorScheme(.dark)
    }

    // MARK: - Scan button

    private var scanButton: some View {
        Button(action: startScan) {
            Group {
                if isScanning {
                    HStack(spacing: BleRPCSpacing.s2) {
                        ProgressView()
                            .progressViewStyle(.circular)
                            .tint(BleRPCColor.textSecondary)
                            .controlSize(.small)
                        Text("Scanning…")
                    }
                } else {
                    Text("Scan")
                }
            }
            .font(.system(size: 15, weight: .semibold))
            .frame(maxWidth: .infinity)
            .padding(.vertical, 10)
            .foregroundColor((isScanning || isRunning) ? BleRPCColor.textSecondary : BleRPCColor.onAccent)
            .background((isScanning || isRunning) ? BleRPCColor.bgSecondary : BleRPCColor.accent)
            .clipShape(RoundedRectangle(cornerRadius: BleRPCRadius.md, style: .continuous))
        }
        .buttonStyle(.plain)
        .disabled(isScanning || isRunning)
    }

    // MARK: - Device list

    private var deviceList: some View {
        VStack(alignment: .leading, spacing: BleRPCSpacing.s2) {
            HStack(spacing: BleRPCSpacing.s2) {
                Text("DEVICES")
                    .font(.system(size: 12, weight: .semibold))
                    .tracking(0.8)
                    .foregroundColor(BleRPCColor.textSecondary)
                BleRPCBadge(text: "\(scannedDevices.count)", tone: .blue)
            }

            ScrollView {
                VStack(spacing: 0) {
                    ForEach(Array(scannedDevices.enumerated()), id: \.element.id) { index, device in
                        Button(action: { select(device) }) {
                            HStack(spacing: BleRPCSpacing.s3) {
                                BleRPCSignalBars(level: bleRPCRssiToLevel(device.rssi))
                                VStack(alignment: .leading, spacing: 2) {
                                    Text(device.name ?? "Unknown")
                                        .font(.system(size: 15, weight: .medium))
                                        .foregroundColor(BleRPCColor.text)
                                    Text(device.id.uuidString)
                                        .font(BleRPCFont.mono(11))
                                        .foregroundColor(BleRPCColor.textSecondary)
                                        .lineLimit(1)
                                }
                                Spacer()
                                Text("\(device.rssi) dBm")
                                    .font(BleRPCFont.mono(13))
                                    .foregroundColor(BleRPCColor.textSecondary)
                                Image(systemName: "chevron.right")
                                    .font(.system(size: 13, weight: .semibold))
                                    .foregroundColor(BleRPCColor.textSecondary)
                            }
                            .padding(.horizontal, BleRPCSpacing.s3)
                            .frame(height: deviceRowHeight)
                            .contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)
                        .disabled(isRunning)
                        if index < scannedDevices.count - 1 {
                            Divider().overlay(BleRPCColor.border)
                        }
                    }
                }
            }
            // Hug the rows (no empty space for one device); scroll past the cap.
            .frame(height: deviceListHeight)
            .background(BleRPCColor.bgSecondary)
            .clipShape(RoundedRectangle(cornerRadius: BleRPCRadius.xl, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: BleRPCRadius.xl, style: .continuous)
                    .stroke(BleRPCColor.border, lineWidth: 1)
            )
        }
    }

    // MARK: - Log console

    private var logConsole: some View {
        ScrollViewReader { proxy in
            ScrollView {
                if testRunner.logs.isEmpty {
                    Text("Scan for devices, then tap one to run tests.")
                        .font(.system(size: 13))
                        .foregroundColor(BleRPCColor.textSecondary)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(BleRPCSpacing.s3)
                } else {
                    LazyVStack(alignment: .leading, spacing: 2) {
                        ForEach(Array(testRunner.logs.enumerated()), id: \.offset) { index, line in
                            Text(line)
                                .foregroundColor(colorForLine(line))
                                .font(BleRPCFont.mono(13))
                                .id(index)
                        }
                    }
                    .padding(BleRPCSpacing.s3)
                }
            }
            .background(BleRPCColor.bgCode)
            .clipShape(RoundedRectangle(cornerRadius: BleRPCRadius.lg, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: BleRPCRadius.lg, style: .continuous)
                    .stroke(BleRPCColor.border, lineWidth: 1)
            )
            .onChange(of: testRunner.logs.count) { newCount in
                if newCount > 0 {
                    withAnimation {
                        proxy.scrollTo(newCount - 1, anchor: .bottom)
                    }
                }
            }
        }
    }

    // MARK: - Actions

    private func startScan() {
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
    }

    private func select(_ device: ScannedDevice) {
        scannedDevices = []
        isRunning = true
        Task {
            await testRunner.runAll(device: device)
            isRunning = false
        }
    }

    private func copyLogs() {
        UIPasteboard.general.string = testRunner.logs.joined(separator: "\n")
        showCopied = true
        Task {
            try? await Task.sleep(nanoseconds: 1_500_000_000)
            showCopied = false
        }
    }

    private func colorForLine(_ line: String) -> Color {
        if line.hasPrefix("[PASS]") {
            return BleRPCColor.success
        } else if line.hasPrefix("[FAIL]") || line.hasPrefix("[ERROR]") {
            return BleRPCColor.error
        } else if line.hasPrefix("[BENCH]") {
            return BleRPCColor.accent
        } else {
            return BleRPCColor.text
        }
    }
}
