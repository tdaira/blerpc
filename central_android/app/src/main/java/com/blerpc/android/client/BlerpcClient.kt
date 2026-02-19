package com.blerpc.android.client

import android.content.Context
import android.util.Log
import com.blerpc.android.ble.BleTransport
import com.blerpc.android.ble.SERVICE_UUID
import com.blerpc.android.ble.ScannedDevice
import com.blerpc.protocol.*
import com.blerpc.protocol.BlerpcCryptoSession
import com.blerpc.protocol.CentralKeyExchange
import com.blerpc.protocol.CAPABILITY_FLAG_ENCRYPTION_SUPPORTED
import com.blerpc.protocol.makeKeyExchange
import kotlinx.coroutines.TimeoutCancellationException
import java.nio.ByteBuffer
import java.nio.ByteOrder
import java.util.UUID

class PayloadTooLargeError(actual: Int, limit: Int) :
    Exception("Request payload ($actual bytes) exceeds peripheral limit ($limit bytes)")

class ResponseTooLargeError(message: String) : Exception(message)

class PeripheralErrorException(val errorCode: Byte) :
    Exception("Peripheral error: 0x${errorCode.toInt().and(0xFF).toString(16).padStart(2, '0')}")

class ProtocolException(message: String) : Exception(message)

class BlerpcClient(context: Context) : GeneratedClient() {
    val transport = BleTransport(context)
    private var splitter: ContainerSplitter? = null
    private val assembler = ContainerAssembler()
    private var timeoutMs: Long = 100
    private var maxRequestPayloadSize: Int? = null
    private var maxResponsePayloadSize: Int? = null

    // Encryption state
    private var session: BlerpcCryptoSession? = null

    val mtu: Int get() = transport.mtu

    suspend fun scan(timeout: Long = 5000, serviceUuid: UUID? = SERVICE_UUID): List<ScannedDevice> {
        return transport.scan(timeout, serviceUuid)
    }

    suspend fun connect(device: ScannedDevice) {
        transport.connect(device)
        splitter = ContainerSplitter(mtu = transport.mtu)

        try {
            requestTimeout()
        } catch (_: TimeoutCancellationException) {
            Log.d(TAG, "Peripheral did not respond to timeout request, using default")
        }
        try {
            requestCapabilities()
        } catch (_: TimeoutCancellationException) {
            Log.d(TAG, "Peripheral did not respond to capabilities request")
        }
    }

    private suspend fun requestTimeout() {
        val s = splitter ?: throw IllegalStateException("Not connected")
        val tid = s.nextTransactionId()
        val req = makeTimeoutRequest(transactionId = tid)
        transport.write(req.serialize())
        val data = transport.readNotify(1000)
        val resp = Container.deserialize(data)
        if (resp.containerType == ContainerType.CONTROL &&
            resp.controlCmd == ControlCmd.TIMEOUT &&
            resp.payload.size == 2
        ) {
            val ms = ByteBuffer.wrap(resp.payload).order(ByteOrder.LITTLE_ENDIAN).short.toInt() and 0xFFFF
            timeoutMs = ms.toLong()
        } else {
            Log.w(TAG, "Unexpected timeout response: type=${resp.containerType}, cmd=${resp.controlCmd}, payload_len=${resp.payload.size}")
        }
    }

    private suspend fun requestCapabilities() {
        val s = splitter ?: throw IllegalStateException("Not connected")
        val tid = s.nextTransactionId()
        val req = makeCapabilitiesRequest(transactionId = tid)
        transport.write(req.serialize())
        val data = transport.readNotify(1000)
        val resp = Container.deserialize(data)
        if (resp.containerType == ContainerType.CONTROL &&
            resp.controlCmd == ControlCmd.CAPABILITIES &&
            resp.payload.size >= 6
        ) {
            val buf = ByteBuffer.wrap(resp.payload).order(ByteOrder.LITTLE_ENDIAN)
            val maxReq = buf.short.toInt() and 0xFFFF
            val maxResp = buf.short.toInt() and 0xFFFF
            val flags = buf.short.toInt() and 0xFFFF
            if (maxReq == 0 || maxResp == 0) {
                Log.w(TAG, "Peripheral reported zero capability: max_request=$maxReq, max_response=$maxResp")
            }
            maxRequestPayloadSize = maxReq
            maxResponsePayloadSize = maxResp
            Log.d(TAG, "Peripheral capabilities: max_request=$maxReq, max_response=$maxResp, flags=0x${flags.toString(16).padStart(4, '0')}")

            // Initiate key exchange if peripheral supports encryption
            if (flags and CAPABILITY_FLAG_ENCRYPTION_SUPPORTED != 0) {
                performKeyExchange()
            }
        } else {
            Log.w(TAG, "Unexpected capabilities response: type=${resp.containerType}, cmd=${resp.controlCmd}, payload_len=${resp.payload.size}")
        }
    }

    private suspend fun performKeyExchange() {
        val s = splitter ?: throw IllegalStateException("Not connected")
        val kx = CentralKeyExchange()

        // Step 1: Send central's ephemeral public key
        val step1 = kx.start()
        val tid1 = s.nextTransactionId()
        val req1 = makeKeyExchange(transactionId = tid1, payload = step1)
        transport.write(req1.serialize())

        // Step 2: Receive peripheral's response
        val data2 = transport.readNotify(2000)
        val resp2 = Container.deserialize(data2)
        if (resp2.containerType != ContainerType.CONTROL ||
            resp2.controlCmd != ControlCmd.KEY_EXCHANGE
        ) {
            Log.e(TAG, "Expected KEY_EXCHANGE step 2, got something else")
            return
        }

        val step3 = try {
            kx.processStep2(resp2.payload)
        } catch (e: IllegalArgumentException) {
            Log.e(TAG, "Key exchange step 2 failed: ${e.message}")
            return
        }

        // Step 3: Send encrypted confirmation
        val tid3 = s.nextTransactionId()
        val req3 = makeKeyExchange(transactionId = tid3, payload = step3)
        transport.write(req3.serialize())

        // Step 4: Receive peripheral's confirmation
        val data4 = transport.readNotify(2000)
        val resp4 = Container.deserialize(data4)
        if (resp4.containerType != ContainerType.CONTROL ||
            resp4.controlCmd != ControlCmd.KEY_EXCHANGE
        ) {
            Log.e(TAG, "Expected KEY_EXCHANGE step 4, got something else")
            return
        }

        session = try {
            kx.finish(resp4.payload)
        } catch (e: IllegalArgumentException) {
            Log.e(TAG, "Key exchange step 4 failed: ${e.message}")
            return
        }
        Log.i(TAG, "E2E encryption established")
    }

    private fun encryptPayload(payload: ByteArray): ByteArray {
        return session?.encrypt(payload) ?: payload
    }

    private fun decryptPayload(payload: ByteArray): ByteArray {
        return session?.decrypt(payload) ?: payload
    }

    override suspend fun call(cmdName: String, requestData: ByteArray): ByteArray {
        val s = splitter ?: throw IllegalStateException("Not connected")

        val cmd = CommandPacket(
            cmdType = CommandType.REQUEST,
            cmdName = cmdName,
            data = requestData
        )
        val payload = cmd.serialize()

        maxRequestPayloadSize?.let { limit ->
            if (payload.size > limit) throw PayloadTooLargeError(payload.size, limit)
        }

        // Encrypt if active, then split into containers and send
        val sendPayload = encryptPayload(payload)
        val containers = s.split(sendPayload)
        for (c in containers) {
            transport.write(c.serialize())
        }

        assembler.reset()
        var firstRead = true
        while (true) {
            // First read uses longer timeout (peripheral processing + BLE latency)
            val t = if (firstRead) maxOf(timeoutMs, 2000) else timeoutMs
            firstRead = false
            val notifyData = transport.readNotify(t)
            val container = Container.deserialize(notifyData)

            if (container.containerType == ContainerType.CONTROL) {
                if (container.controlCmd == ControlCmd.ERROR && container.payload.isNotEmpty()) {
                    val errorCode = container.payload[0]
                    if (errorCode == BLERPC_ERROR_RESPONSE_TOO_LARGE) {
                        throw ResponseTooLargeError("Response exceeds peripheral's max_response_payload_size")
                    }
                    throw PeripheralErrorException(errorCode)
                }
                continue
            }

            val result = assembler.feed(container)
            if (result != null) {
                // Decrypt if active
                val decrypted = decryptPayload(result)
                val resp = CommandPacket.deserialize(decrypted)
                if (resp.cmdType != CommandType.RESPONSE) {
                    throw ProtocolException("Expected response, got type=${resp.cmdType}")
                }
                if (resp.cmdName != cmdName) {
                    throw ProtocolException("Command name mismatch: expected '$cmdName', got '${resp.cmdName}'")
                }
                return resp.data
            }
        }
    }

    override suspend fun streamReceive(cmdName: String, requestData: ByteArray): List<ByteArray> {
        val s = splitter ?: throw IllegalStateException("Not connected")

        val cmd = CommandPacket(
            cmdType = CommandType.REQUEST,
            cmdName = cmdName,
            data = requestData
        )
        val payload = cmd.serialize()

        maxRequestPayloadSize?.let { limit ->
            if (payload.size > limit) throw PayloadTooLargeError(payload.size, limit)
        }

        // Encrypt if active, then split into containers and send
        val sendPayload = encryptPayload(payload)
        val containers = s.split(sendPayload)
        for (c in containers) {
            transport.write(c.serialize())
        }

        val results = mutableListOf<ByteArray>()
        assembler.reset()
        var firstRead = true
        while (true) {
            // First read uses longer timeout (peripheral processing time)
            val t = if (firstRead) maxOf(timeoutMs, 2000) else timeoutMs
            firstRead = false
            val notifyData = transport.readNotify(t)
            val container = Container.deserialize(notifyData)

            if (container.containerType == ContainerType.CONTROL) {
                if (container.controlCmd == ControlCmd.STREAM_END_P2C) {
                    break
                }
                if (container.controlCmd == ControlCmd.ERROR && container.payload.isNotEmpty()) {
                    val errorCode = container.payload[0]
                    if (errorCode == BLERPC_ERROR_RESPONSE_TOO_LARGE) {
                        throw ResponseTooLargeError("Response exceeds peripheral's max_response_payload_size")
                    }
                    throw PeripheralErrorException(errorCode)
                }
                continue
            }

            val result = assembler.feed(container)
            if (result != null) {
                // Decrypt each received response
                val decrypted = decryptPayload(result)
                val resp = CommandPacket.deserialize(decrypted)
                if (resp.cmdType != CommandType.RESPONSE) {
                    throw ProtocolException("Expected response, got type=${resp.cmdType}")
                }
                results.add(resp.data)
            }
        }
        return results
    }

    override suspend fun streamSend(
        cmdName: String,
        messages: List<ByteArray>,
        finalCmdName: String
    ): ByteArray {
        val s = splitter ?: throw IllegalStateException("Not connected")

        // Encrypt each message
        for (msgData in messages) {
            val cmd = CommandPacket(
                cmdType = CommandType.REQUEST,
                cmdName = cmdName,
                data = msgData
            )
            val payload = cmd.serialize()
            val sendPayload = encryptPayload(payload)
            val containers = s.split(sendPayload)
            for (c in containers) {
                transport.write(c.serialize())
            }
        }

        // Send STREAM_END_C2P
        val tid = s.nextTransactionId()
        val streamEnd = makeStreamEndC2P(transactionId = tid)
        transport.write(streamEnd.serialize())

        // Wait for final response
        assembler.reset()
        var firstRead = true
        while (true) {
            val t = if (firstRead) maxOf(timeoutMs, 2000) else timeoutMs
            firstRead = false
            val notifyData = transport.readNotify(t)
            val container = Container.deserialize(notifyData)

            if (container.containerType == ContainerType.CONTROL) {
                if (container.controlCmd == ControlCmd.ERROR && container.payload.isNotEmpty()) {
                    val errorCode = container.payload[0]
                    if (errorCode == BLERPC_ERROR_RESPONSE_TOO_LARGE) {
                        throw ResponseTooLargeError("Response exceeds peripheral's max_response_payload_size")
                    }
                    throw PeripheralErrorException(errorCode)
                }
                continue
            }

            val result = assembler.feed(container)
            if (result != null) {
                // Decrypt final response
                val decrypted = decryptPayload(result)
                val resp = CommandPacket.deserialize(decrypted)
                if (resp.cmdType != CommandType.RESPONSE) {
                    throw ProtocolException("Expected response, got type=${resp.cmdType}")
                }
                if (resp.cmdName != finalCmdName) {
                    throw ProtocolException("Command name mismatch: expected '$finalCmdName', got '${resp.cmdName}'")
                }
                return resp.data
            }
        }
    }

    suspend fun counterStreamAll(count: Int): List<Pair<Int, Int>> {
        val req = blerpc.Blerpc.CounterStreamRequest.newBuilder().setCount(count).build()
        val responses = streamReceive("counter_stream", req.toByteArray())
        return responses.map { data ->
            val resp = blerpc.Blerpc.CounterStreamResponse.parseFrom(data)
            Pair(resp.seq, resp.value)
        }
    }

    suspend fun counterUploadAll(count: Int): blerpc.Blerpc.CounterUploadResponse {
        val messages = (0 until count).map { i ->
            blerpc.Blerpc.CounterUploadRequest.newBuilder()
                .setSeq(i)
                .setValue(i * 10)
                .build()
                .toByteArray()
        }
        val respData = streamSend("counter_upload", messages, "counter_upload")
        return blerpc.Blerpc.CounterUploadResponse.parseFrom(respData)
    }

    fun disconnect() {
        transport.disconnect()
    }

    companion object {
        private const val TAG = "BlerpcClient"
    }
}
