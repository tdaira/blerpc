import BlerpcProtocol
import CoreBluetooth
import Foundation
import os.log

enum BlerpcClientError: Error {
    case notConnected
    case payloadTooLarge(actual: Int, limit: Int)
    case responseTooLarge
    case peripheralError(code: UInt8)
    case unexpectedResponseType(UInt8)
    case commandNameMismatch(expected: String, got: String)
    case keyExchangeFailed(String)
    case encryptionError(String)
    case replayDetected(counter: UInt32)
}

private let logger = Logger(subsystem: "com.blerpc", category: "BlerpcClient")

final class BlerpcClient: GeneratedClientProtocol {
    let transport = BleTransport()
    private var splitter: ContainerSplitter?
    private let assembler = ContainerAssembler()
    private var timeoutMs: Int = 100
    private var maxRequestPayloadSize: Int?
    private var maxResponsePayloadSize: Int?

    private var session: BlerpcCryptoSession?

    var mtu: Int { transport.mtu }
    var isEncrypted: Bool { session != nil }

    func scan(
        timeout: TimeInterval = 5,
        serviceUUID filterUUID: CBUUID? = serviceUUID
    ) async throws -> [ScannedDevice] {
        return try await transport.scan(timeout: timeout, serviceUUID: filterUUID)
    }

    func connect(device: ScannedDevice) async throws {
        try await transport.connect(device: device)
        let mtuVal = transport.mtu
        splitter = ContainerSplitter(mtu: mtuVal)

        do {
            try await requestTimeout()
        } catch is BleTransportError {
            logger.debug("Peripheral did not respond to timeout request, using default")
        }
        do {
            try await requestCapabilities()
        } catch is BleTransportError {
            logger.debug("Peripheral did not respond to capabilities request")
        }
    }

    private func requestTimeout() async throws {
        guard let s = splitter else { throw BlerpcClientError.notConnected }
        let tid = s.nextTransactionId()
        let req = makeTimeoutRequest(transactionId: tid)
        try transport.write(req.serialize())
        let data = try await transport.readNotify(timeoutMs: 1000)
        let resp = try Container.deserialize(data)
        if resp.containerType == .control,
           resp.controlCmd == .timeout,
           resp.payload.count == 2 {
            let ms = Int(resp.payload[0]) | (Int(resp.payload[1]) << 8)
            timeoutMs = ms
        } else {
            logger.warning("Unexpected timeout response: type=\(resp.containerType.rawValue), cmd=\(resp.controlCmd.rawValue), payload_len=\(resp.payload.count)")
        }
    }

    private func requestCapabilities() async throws {
        guard let s = splitter else { throw BlerpcClientError.notConnected }
        let tid = s.nextTransactionId()
        let req = makeCapabilitiesRequest(transactionId: tid)
        try transport.write(req.serialize())
        let data = try await transport.readNotify(timeoutMs: 1000)
        let resp = try Container.deserialize(data)
        if resp.containerType == .control,
           resp.controlCmd == .capabilities,
           resp.payload.count >= 6 {
            let maxReq = Int(resp.payload[0]) | (Int(resp.payload[1]) << 8)
            let maxResp = Int(resp.payload[2]) | (Int(resp.payload[3]) << 8)
            let flags = Int(resp.payload[4]) | (Int(resp.payload[5]) << 8)
            if maxReq == 0 || maxResp == 0 {
                logger.warning("Peripheral reported zero capability: max_request=\(maxReq), max_response=\(maxResp)")
            }
            maxRequestPayloadSize = maxReq
            maxResponsePayloadSize = maxResp
            logger.info("Peripheral capabilities: max_request=\(maxReq), max_response=\(maxResp), flags=0x\(String(flags, radix: 16, uppercase: false))")

            if flags & Int(capabilityFlagEncryptionSupported) != 0 {
                try await performKeyExchange()
            }
        } else {
            logger.warning("Unexpected capabilities response: type=\(resp.containerType.rawValue), cmd=\(resp.controlCmd.rawValue), payload_len=\(resp.payload.count)")
        }
    }

    private func performKeyExchange() async throws {
        guard let s = splitter else { throw BlerpcClientError.notConnected }

        session = try await BlerpcCrypto.centralPerformKeyExchange(
            send: { payload in
                let tid = s.nextTransactionId()
                let req = makeKeyExchange(transactionId: tid, payload: payload)
                try self.transport.write(req.serialize())
            },
            receive: {
                let data = try await self.transport.readNotify(timeoutMs: 2000)
                let resp = try Container.deserialize(data)
                guard resp.containerType == .control,
                      resp.controlCmd == .keyExchange else {
                    throw BlerpcClientError.keyExchangeFailed("Expected KEY_EXCHANGE response")
                }
                return resp.payload
            }
        )
        logger.info("E2E encryption established")
    }

    private func encryptPayload(_ payload: Data) throws -> Data {
        guard let s = session else { return payload }
        return try s.encrypt(payload)
    }

    private func decryptPayload(_ payload: Data) throws -> Data {
        guard let s = session else { return payload }
        return try s.decrypt(payload)
    }

    func call(cmdName: String, requestData: Data) async throws -> Data {
        guard let s = splitter else { throw BlerpcClientError.notConnected }

        let cmd = CommandPacket(cmdType: .request, cmdName: cmdName, data: requestData)
        let payload = try cmd.serialize()

        if let limit = maxRequestPayloadSize, payload.count > limit {
            throw BlerpcClientError.payloadTooLarge(actual: payload.count, limit: limit)
        }

        let sendPayload = try encryptPayload(payload)
        let containers = try s.split(sendPayload)
        for c in containers {
            try transport.write(c.serialize())
        }

        assembler.reset()
        var firstRead = true
        while true {
            let t = firstRead ? max(timeoutMs, 2000) : timeoutMs
            firstRead = false
            let notifyData = try await transport.readNotify(timeoutMs: t)
            let container = try Container.deserialize(notifyData)

            if container.containerType == .control {
                if container.controlCmd == .error, !container.payload.isEmpty {
                    let errorCode = container.payload[0]
                    if errorCode == blerpcErrorResponseTooLarge {
                        throw BlerpcClientError.responseTooLarge
                    }
                    throw BlerpcClientError.peripheralError(code: errorCode)
                }
                continue
            }

            if let result = assembler.feed(container) {
                let decrypted = try decryptPayload(result)
                let resp = try CommandPacket.deserialize(decrypted)
                guard resp.cmdType == .response else {
                    throw BlerpcClientError.unexpectedResponseType(resp.cmdType.rawValue)
                }
                guard resp.cmdName == cmdName else {
                    throw BlerpcClientError.commandNameMismatch(
                        expected: cmdName, got: resp.cmdName
                    )
                }
                return resp.data
            }
        }
    }

    func streamReceive(cmdName: String, requestData: Data) async throws -> [Data] {
        guard let s = splitter else { throw BlerpcClientError.notConnected }

        let cmd = CommandPacket(cmdType: .request, cmdName: cmdName, data: requestData)
        let payload = try cmd.serialize()

        if let limit = maxRequestPayloadSize, payload.count > limit {
            throw BlerpcClientError.payloadTooLarge(actual: payload.count, limit: limit)
        }

        let sendPayload = try encryptPayload(payload)
        let containers = try s.split(sendPayload)
        for c in containers {
            try transport.write(c.serialize())
        }

        var results: [Data] = []
        assembler.reset()
        var firstRead = true
        while true {
            let t = firstRead ? max(timeoutMs, 2000) : timeoutMs
            firstRead = false
            let notifyData = try await transport.readNotify(timeoutMs: t)
            let container = try Container.deserialize(notifyData)

            if container.containerType == .control {
                if container.controlCmd == .streamEndP2C {
                    break
                }
                if container.controlCmd == .error, !container.payload.isEmpty {
                    let errorCode = container.payload[0]
                    if errorCode == blerpcErrorResponseTooLarge {
                        throw BlerpcClientError.responseTooLarge
                    }
                    throw BlerpcClientError.peripheralError(code: errorCode)
                }
                continue
            }

            if let result = assembler.feed(container) {
                let decrypted = try decryptPayload(result)
                let resp = try CommandPacket.deserialize(decrypted)
                guard resp.cmdType == .response else {
                    throw BlerpcClientError.unexpectedResponseType(resp.cmdType.rawValue)
                }
                results.append(resp.data)
            }
        }
        return results
    }

    func streamSend(
        cmdName: String,
        messages: [Data],
        finalCmdName: String
    ) async throws -> Data {
        guard let s = splitter else { throw BlerpcClientError.notConnected }

        for msgData in messages {
            let cmd = CommandPacket(cmdType: .request, cmdName: cmdName, data: msgData)
            let payload = try cmd.serialize()
            let sendPayload = try encryptPayload(payload)
            let containers = try s.split(sendPayload)
            for c in containers {
                try transport.write(c.serialize())
            }
        }

        // Send STREAM_END_C2P
        let tid = s.nextTransactionId()
        let streamEnd = makeStreamEndC2P(transactionId: tid)
        try transport.write(streamEnd.serialize())

        // Wait for final response
        assembler.reset()
        var firstRead = true
        while true {
            let t = firstRead ? max(timeoutMs, 2000) : timeoutMs
            firstRead = false
            let notifyData = try await transport.readNotify(timeoutMs: t)
            let container = try Container.deserialize(notifyData)

            if container.containerType == .control {
                if container.controlCmd == .error, !container.payload.isEmpty {
                    let errorCode = container.payload[0]
                    if errorCode == blerpcErrorResponseTooLarge {
                        throw BlerpcClientError.responseTooLarge
                    }
                    throw BlerpcClientError.peripheralError(code: errorCode)
                }
                continue
            }

            if let result = assembler.feed(container) {
                let decrypted = try decryptPayload(result)
                let resp = try CommandPacket.deserialize(decrypted)
                guard resp.cmdType == .response else {
                    throw BlerpcClientError.unexpectedResponseType(resp.cmdType.rawValue)
                }
                guard resp.cmdName == finalCmdName else {
                    throw BlerpcClientError.commandNameMismatch(
                        expected: finalCmdName, got: resp.cmdName
                    )
                }
                return resp.data
            }
        }
    }

    func counterStreamAll(count: UInt32) async throws -> [(seq: UInt32, value: Int32)] {
        var req = Blerpc_CounterStreamRequest()
        req.count = count
        let responses = try await streamReceive(
            cmdName: "counter_stream",
            requestData: try req.serializedData()
        )
        return try responses.map { data in
            let resp = try Blerpc_CounterStreamResponse(serializedBytes: data)
            return (seq: resp.seq, value: resp.value)
        }
    }

    func counterUploadAll(count: Int) async throws -> Blerpc_CounterUploadResponse {
        let messages = try (0..<count).map { i -> Data in
            var req = Blerpc_CounterUploadRequest()
            req.seq = UInt32(i)
            req.value = Int32(i * 10)
            return try req.serializedData()
        }
        let respData = try await streamSend(
            cmdName: "counter_upload",
            messages: messages,
            finalCmdName: "counter_upload"
        )
        return try Blerpc_CounterUploadResponse(serializedBytes: respData)
    }

    func disconnect() {
        transport.disconnect()
    }
}
