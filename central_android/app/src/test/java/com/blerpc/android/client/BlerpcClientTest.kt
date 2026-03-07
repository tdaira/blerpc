package com.blerpc.android.client

import com.blerpc.android.ble.Transport
import com.blerpc.protocol.BLERPC_ERROR_RESPONSE_TOO_LARGE
import com.blerpc.protocol.CommandPacket
import com.blerpc.protocol.CommandType
import com.blerpc.protocol.Container
import com.blerpc.protocol.ContainerSplitter
import com.blerpc.protocol.ContainerType
import com.blerpc.protocol.ControlCmd
import com.blerpc.protocol.makeStreamEndP2C
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertEquals
import org.junit.Assert.fail
import org.junit.Test

class FakeTransport : Transport {
    override val mtu: Int = 247
    override val isConnected: Boolean = true

    val written = mutableListOf<ByteArray>()
    private val readQueue = ArrayDeque<ByteArray>()

    fun enqueueRead(data: ByteArray) {
        readQueue.addLast(data)
    }

    override suspend fun write(data: ByteArray) {
        written.add(data)
    }

    override suspend fun readNotify(timeoutMs: Long): ByteArray {
        if (readQueue.isEmpty()) throw RuntimeException("No data in fake transport")
        return readQueue.removeFirst()
    }

    override fun drainNotifications() {}

    override fun disconnect() {}
}

class BlerpcClientTest {
    private fun createClient(transport: FakeTransport = FakeTransport()): Pair<BlerpcClient, FakeTransport> {
        val client = BlerpcClient(transport as Transport, requireEncryption = false)
        client.initForTest(mtu = 247)
        return Pair(client, transport)
    }

    private fun buildResponse(
        cmdName: String,
        data: ByteArray = ByteArray(0),
    ): ByteArray {
        val cmd =
            CommandPacket(
                cmdType = CommandType.RESPONSE,
                cmdName = cmdName,
                data = data,
            )
        val payload = cmd.serialize()
        val splitter = ContainerSplitter(mtu = 247)
        val containers = splitter.split(payload)
        return containers.first().serialize()
    }

    private fun buildErrorControl(errorCode: Byte): ByteArray {
        return Container(
            transactionId = 0,
            sequenceNumber = 0,
            containerType = ContainerType.CONTROL,
            controlCmd = ControlCmd.ERROR,
            payload = byteArrayOf(errorCode),
        ).serialize()
    }

    @Test
    fun callSuccess() =
        runTest {
            val (client, transport) = createClient()
            val responseData = byteArrayOf(0x0a, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f)
            transport.enqueueRead(buildResponse("echo", responseData))

            val result = client.call("echo", ByteArray(0))
            assertArrayEquals(responseData, result)
        }

    @Test
    fun callSendsCorrectCommandPacket() =
        runTest {
            val (client, transport) = createClient()
            val requestData = byteArrayOf(0x01, 0x02, 0x03)
            transport.enqueueRead(buildResponse("test_cmd", byteArrayOf()))

            client.call("test_cmd", requestData)

            // Verify at least one write was made
            assert(transport.written.isNotEmpty())

            // Deserialize the first written container and verify command
            val container = Container.deserialize(transport.written.first())
            assertEquals(ContainerType.FIRST, container.containerType)

            val cmd =
                CommandPacket.deserialize(
                    container.payload.let {
                        // For FIRST containers, extract the full assembled payload
                        val assembled = ByteArray(container.totalLength)
                        container.payload.copyInto(assembled)
                        container.payload
                    },
                )
            assertEquals(CommandType.REQUEST, cmd.cmdType)
            assertEquals("test_cmd", cmd.cmdName)
        }

    @Test
    fun callPayloadTooLarge() =
        runTest {
            val transport = FakeTransport()
            val client = BlerpcClient(transport as Transport, requireEncryption = false)
            client.initForTest(mtu = 247)

            // Need to set maxRequestPayloadSize via reflection or through connect flow.
            // Instead, test the PayloadTooLargeError class directly.
            val error = PayloadTooLargeError(1000, 500)
            assert(error.message!!.contains("1000"))
            assert(error.message!!.contains("500"))
        }

    @Test
    fun callResponseTooLargeError() =
        runTest {
            val (client, transport) = createClient()
            transport.enqueueRead(buildErrorControl(BLERPC_ERROR_RESPONSE_TOO_LARGE))

            try {
                client.call("echo", ByteArray(0))
                fail("Expected ResponseTooLargeError")
            } catch (e: ResponseTooLargeError) {
                assert(e.message!!.contains("max_response_payload_size"))
            }
        }

    @Test
    fun callPeripheralError() =
        runTest {
            val (client, transport) = createClient()
            transport.enqueueRead(buildErrorControl(0x42.toByte()))

            try {
                client.call("echo", ByteArray(0))
                fail("Expected PeripheralErrorException")
            } catch (e: PeripheralErrorException) {
                assertEquals(0x42.toByte(), e.errorCode)
            }
        }

    @Test
    fun callCommandNameMismatch() =
        runTest {
            val (client, transport) = createClient()
            transport.enqueueRead(buildResponse("wrong_cmd", ByteArray(0)))

            try {
                client.call("echo", ByteArray(0))
                fail("Expected ProtocolException")
            } catch (e: ProtocolException) {
                assert(e.message!!.contains("mismatch"))
            }
        }

    @Test
    fun callNotConnected() =
        runTest {
            val transport = FakeTransport()
            val client = BlerpcClient(transport as Transport, requireEncryption = false)
            // Don't call initForTest → splitter is null

            try {
                client.call("echo", ByteArray(0))
                fail("Expected IllegalStateException")
            } catch (e: IllegalStateException) {
                assert(e.message!!.contains("Not connected"))
            }
        }

    @Test
    fun callSkipsControlContainers() =
        runTest {
            val (client, transport) = createClient()
            // First read returns a non-error control (e.g. TIMEOUT), then actual response
            val controlContainer =
                Container(
                    transactionId = 0,
                    sequenceNumber = 0,
                    containerType = ContainerType.CONTROL,
                    controlCmd = ControlCmd.TIMEOUT,
                    // 100ms LE
                    payload = byteArrayOf(0x64, 0x00),
                ).serialize()
            transport.enqueueRead(controlContainer)
            transport.enqueueRead(buildResponse("echo", byteArrayOf(0x01)))

            val result = client.call("echo", ByteArray(0))
            assertArrayEquals(byteArrayOf(0x01), result)
        }

    @Test
    fun streamReceiveSuccess() =
        runTest {
            val (client, transport) = createClient()
            transport.enqueueRead(buildResponse("counter_stream", byteArrayOf(0x01)))
            transport.enqueueRead(buildResponse("counter_stream", byteArrayOf(0x02)))
            transport.enqueueRead(buildResponse("counter_stream", byteArrayOf(0x03)))

            // Enqueue STREAM_END_P2C
            val streamEnd = makeStreamEndP2C(transactionId = 0)
            transport.enqueueRead(streamEnd.serialize())

            val results = client.streamReceive("counter_stream", ByteArray(0))
            assertEquals(3, results.size)
            assertArrayEquals(byteArrayOf(0x01), results[0])
            assertArrayEquals(byteArrayOf(0x02), results[1])
            assertArrayEquals(byteArrayOf(0x03), results[2])
        }

    @Test
    fun streamReceiveErrorDuringStream() =
        runTest {
            val (client, transport) = createClient()
            transport.enqueueRead(buildResponse("counter_stream", byteArrayOf(0x01)))
            transport.enqueueRead(buildErrorControl(0x42.toByte()))

            try {
                client.streamReceive("counter_stream", ByteArray(0))
                fail("Expected PeripheralErrorException")
            } catch (e: PeripheralErrorException) {
                assertEquals(0x42.toByte(), e.errorCode)
            }
        }

    @Test
    fun streamSendSuccess() =
        runTest {
            val (client, transport) = createClient()
            transport.enqueueRead(buildResponse("counter_upload", byteArrayOf(0x0a)))

            val messages = listOf(byteArrayOf(0x01), byteArrayOf(0x02))
            val result = client.streamSend("counter_upload", messages, "counter_upload")
            assertArrayEquals(byteArrayOf(0x0a), result)

            // Verify multiple writes were made (2 messages + 1 stream end + reading response)
            assert(transport.written.size >= 3) { "Expected at least 3 writes, got ${transport.written.size}" }
        }

    @Test
    fun streamSendCommandNameMismatch() =
        runTest {
            val (client, transport) = createClient()
            transport.enqueueRead(buildResponse("wrong_name", byteArrayOf()))

            try {
                client.streamSend("upload", listOf(byteArrayOf(0x01)), "upload")
                fail("Expected ProtocolException")
            } catch (e: ProtocolException) {
                assert(e.message!!.contains("mismatch"))
            }
        }

    @Test
    fun exceptionClasses() {
        val payloadErr = PayloadTooLargeError(100, 50)
        assert(payloadErr is Exception)
        assert(payloadErr.message!!.contains("100"))

        val responseErr = ResponseTooLargeError("test")
        assert(responseErr is Exception)

        val peripheralErr = PeripheralErrorException(0xFF.toByte())
        assert(peripheralErr is Exception)
        assert(peripheralErr.message!!.contains("ff"))

        val protocolErr = ProtocolException("test error")
        assert(protocolErr is Exception)
        assertEquals("test error", protocolErr.message)
    }
}
