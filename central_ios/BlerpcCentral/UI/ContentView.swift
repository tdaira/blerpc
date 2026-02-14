import SwiftUI

struct ContentView: View {
    @ObservedObject var testRunner: TestRunner
    @State private var isRunning = false

    var body: some View {
        VStack(spacing: 16) {
            Text("blerpc iOS Central")
                .font(.title)
                .padding(.top, 16)

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
            .disabled(isRunning)
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
