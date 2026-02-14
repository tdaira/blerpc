package com.blerpc.android.client

import android.content.Context
import com.blerpc.android.ble.BleTransport
import com.blerpc.protocol.*
import java.nio.ByteBuffer
import java.nio.ByteOrder

class PayloadTooLargeError(actual: Int, limit: Int) :
    Exception("Request payload ($actual bytes) exceeds peripheral limit ($limit bytes)")

class ResponseTooLargeError(message: String) : Exception(message)

class BlerpcClient(context: Context) : GeneratedClient() {
    val transport = BleTransport(context)
    private var splitter: ContainerSplitter? = null
    private val assembler = ContainerAssembler()
    private var timeoutMs: Long = 100
    private var maxRequestPayloadSize: Int? = null
    private var maxResponsePayloadSize: Int? = null

    val mtu: Int get() = transport.mtu

    suspend fun connect(deviceName: String = "blerpc", timeout: Long = 10000) {
        transport.connect(deviceName, timeout)
        splitter = ContainerSplitter(mtu = transport.mtu)

        try {
            requestTimeout()
        } catch (_: Exception) {
            // Peripheral may not support timeout request
        }
        try {
            requestCapabilities()
        } catch (_: Exception) {
            // Peripheral may not support capabilities request
        }
    }

    private suspend fun requestTimeout() {
        val s = splitter!!
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
        }
    }

    private suspend fun requestCapabilities() {
        val s = splitter!!
        val tid = s.nextTransactionId()
        val req = makeCapabilitiesRequest(transactionId = tid)
        transport.write(req.serialize())
        val data = transport.readNotify(1000)
        val resp = Container.deserialize(data)
        if (resp.containerType == ContainerType.CONTROL &&
            resp.controlCmd == ControlCmd.CAPABILITIES &&
            resp.payload.size == 4
        ) {
            val buf = ByteBuffer.wrap(resp.payload).order(ByteOrder.LITTLE_ENDIAN)
            maxRequestPayloadSize = buf.short.toInt() and 0xFFFF
            maxResponsePayloadSize = buf.short.toInt() and 0xFFFF
        }
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

        val containers = s.split(payload)
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
                    throw RuntimeException("Peripheral error: 0x${errorCode.toString(16).padStart(2, '0')}")
                }
                continue
            }

            val result = assembler.feed(container)
            if (result != null) {
                val resp = CommandPacket.deserialize(result)
                if (resp.cmdType != CommandType.RESPONSE) {
                    throw RuntimeException("Expected response, got type=${resp.cmdType}")
                }
                if (resp.cmdName != cmdName) {
                    throw RuntimeException("Command name mismatch: expected '$cmdName', got '${resp.cmdName}'")
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

        val containers = s.split(payload)
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
                    throw RuntimeException("Peripheral error: 0x${errorCode.toString(16).padStart(2, '0')}")
                }
                continue
            }

            val result = assembler.feed(container)
            if (result != null) {
                val resp = CommandPacket.deserialize(result)
                if (resp.cmdType != CommandType.RESPONSE) {
                    throw RuntimeException("Expected response, got type=${resp.cmdType}")
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

        for (msgData in messages) {
            val cmd = CommandPacket(
                cmdType = CommandType.REQUEST,
                cmdName = cmdName,
                data = msgData
            )
            val payload = cmd.serialize()
            val containers = s.split(payload)
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
                    throw RuntimeException("Peripheral error: 0x${errorCode.toString(16).padStart(2, '0')}")
                }
                continue
            }

            val result = assembler.feed(container)
            if (result != null) {
                val resp = CommandPacket.deserialize(result)
                if (resp.cmdType != CommandType.RESPONSE) {
                    throw RuntimeException("Expected response, got type=${resp.cmdType}")
                }
                if (resp.cmdName != finalCmdName) {
                    throw RuntimeException("Command name mismatch: expected '$finalCmdName', got '${resp.cmdName}'")
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
}
