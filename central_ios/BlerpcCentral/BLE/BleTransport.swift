import CoreBluetooth
import Foundation

private let serviceUUID = CBUUID(string: "12340001-0000-1000-8000-00805f9b34fb")
private let charUUID = CBUUID(string: "12340002-0000-1000-8000-00805f9b34fb")

enum BleTransportError: Error {
    case scanTimeout
    case connectionFailed
    case serviceNotFound
    case characteristicNotFound
    case notConnected
    case writeFailed
    case readTimeout
    case bluetoothNotAvailable
}

final class BleTransport: NSObject, CBCentralManagerDelegate, CBPeripheralDelegate {
    private var centralManager: CBCentralManager?
    private var peripheral: CBPeripheral?
    private var writeChar: CBCharacteristic?
    private(set) var mtu: Int = 23
    var isConnected: Bool { peripheral != nil }

    private var notifyContinuation: AsyncStream<Data>.Continuation?
    private var notifyIterator: AsyncStream<Data>.AsyncIterator?

    // Continuations for async bridging
    private var scanContinuation: CheckedContinuation<CBPeripheral, any Error>?
    private var connectContinuation: CheckedContinuation<Void, any Error>?
    private var stateContinuation: CheckedContinuation<Void, any Error>?

    override init() {
        super.init()
    }

    func connect(deviceName: String = "blerpc", timeout: TimeInterval = 10) async throws {
        let cm = CBCentralManager(delegate: self, queue: nil)
        self.centralManager = cm

        // Wait for powered-on state
        if cm.state != .poweredOn {
            try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, any Error>) in
                self.stateContinuation = cont
            }
        }

        // Scan for peripheral
        let foundPeripheral: CBPeripheral = try await withThrowingTaskGroup(of: CBPeripheral.self) { group in
            group.addTask { @Sendable in
                try await withCheckedThrowingContinuation { cont in
                    self.scanContinuation = cont
                    cm.scanForPeripherals(withServices: [serviceUUID])
                }
            }
            group.addTask { @Sendable in
                try await Task.sleep(nanoseconds: UInt64(timeout * 1_000_000_000))
                throw BleTransportError.scanTimeout
            }
            let result = try await group.next()!
            group.cancelAll()
            return result
        }

        cm.stopScan()
        self.peripheral = foundPeripheral
        foundPeripheral.delegate = self

        // Connect
        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, any Error>) in
            self.connectContinuation = cont
            cm.connect(foundPeripheral)
        }

        // Discover services
        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, any Error>) in
            self.connectContinuation = cont
            foundPeripheral.discoverServices([serviceUUID])
        }

        guard let service = foundPeripheral.services?.first(where: { $0.uuid == serviceUUID }) else {
            throw BleTransportError.serviceNotFound
        }

        // Discover characteristics
        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, any Error>) in
            self.connectContinuation = cont
            foundPeripheral.discoverCharacteristics([charUUID], for: service)
        }

        guard let char = service.characteristics?.first(where: { $0.uuid == charUUID }) else {
            throw BleTransportError.characteristicNotFound
        }
        self.writeChar = char

        // Get MTU
        let maxWrite = foundPeripheral.maximumWriteValueLength(for: .withoutResponse)
        // On iOS, maximumWriteValueLength already accounts for ATT overhead
        self.mtu = maxWrite + 3 // Add back ATT overhead for protocol calculations

        // Enable notifications
        foundPeripheral.setNotifyValue(true, for: char)

        // Set up notification stream
        setupNotifyStream()
    }

    private func setupNotifyStream() {
        let (stream, continuation) = AsyncStream<Data>.makeStream(bufferingPolicy: .unbounded)
        self.notifyContinuation = continuation
        self.notifyIterator = stream.makeAsyncIterator()
    }

    func write(_ data: Data) throws {
        guard let p = peripheral, let c = writeChar else {
            throw BleTransportError.notConnected
        }
        p.writeValue(data, for: c, type: .withoutResponse)
    }

    func readNotify(timeoutMs: Int) async throws -> Data {
        guard notifyIterator != nil else {
            throw BleTransportError.notConnected
        }

        return try await withThrowingTaskGroup(of: Data.self) { group in
            group.addTask { @Sendable in
                guard let data = await self.notifyIterator?.next() else {
                    throw BleTransportError.readTimeout
                }
                return data
            }
            group.addTask { @Sendable in
                try await Task.sleep(nanoseconds: UInt64(timeoutMs) * 1_000_000)
                throw BleTransportError.readTimeout
            }
            let result = try await group.next()!
            group.cancelAll()
            return result
        }
    }

    func drainNotifications() {
        // Re-create the stream to drain pending notifications
        setupNotifyStream()
    }

    func disconnect() {
        notifyContinuation?.finish()
        if let p = peripheral, let cm = centralManager {
            cm.cancelPeripheralConnection(p)
        }
        peripheral = nil
        writeChar = nil
        notifyContinuation = nil
        notifyIterator = nil
    }

    // MARK: - CBCentralManagerDelegate

    func centralManagerDidUpdateState(_ central: CBCentralManager) {
        if central.state == .poweredOn {
            stateContinuation?.resume()
            stateContinuation = nil
        }
    }

    func centralManager(
        _ central: CBCentralManager,
        didDiscover peripheral: CBPeripheral,
        advertisementData: [String: Any],
        rssi: NSNumber
    ) {
        scanContinuation?.resume(returning: peripheral)
        scanContinuation = nil
    }

    func centralManager(_ central: CBCentralManager, didConnect peripheral: CBPeripheral) {
        connectContinuation?.resume()
        connectContinuation = nil
    }

    func centralManager(
        _ central: CBCentralManager,
        didFailToConnect peripheral: CBPeripheral,
        error: (any Error)?
    ) {
        connectContinuation?.resume(throwing: BleTransportError.connectionFailed)
        connectContinuation = nil
    }

    // MARK: - CBPeripheralDelegate

    func peripheral(_ peripheral: CBPeripheral, didDiscoverServices error: (any Error)?) {
        connectContinuation?.resume()
        connectContinuation = nil
    }

    func peripheral(
        _ peripheral: CBPeripheral,
        didDiscoverCharacteristicsFor service: CBService,
        error: (any Error)?
    ) {
        connectContinuation?.resume()
        connectContinuation = nil
    }

    func peripheral(
        _ peripheral: CBPeripheral,
        didUpdateValueFor characteristic: CBCharacteristic,
        error: (any Error)?
    ) {
        guard let data = characteristic.value else { return }
        notifyContinuation?.yield(data)
    }
}
