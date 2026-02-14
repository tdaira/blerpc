import SwiftUI

@main
struct BlerpcCentralApp: App {
    @StateObject private var testRunner = TestRunner()

    var body: some Scene {
        WindowGroup {
            ContentView(testRunner: testRunner)
                .onOpenURL { url in
                    guard url.scheme == "blerpc", url.host == "run_tests" else { return }
                    let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
                    let iterations = components?.queryItems?
                        .first(where: { $0.name == "iterations" })
                        .flatMap { Int($0.value ?? "1") } ?? 1
                    Task {
                        // Auto-run: scan and connect to first device
                        await testRunner.runAll(iterations: iterations)
                    }
                }
        }
    }
}
