package com.blerpc.protocol

import org.junit.Assert.*
import org.junit.Test
import java.nio.ByteBuffer
import java.nio.ByteOrder

class ControlContainersTest {

    @Test
    fun timeoutRequest() {
        val c = makeTimeoutRequest(transactionId = 5)
        assertEquals(ContainerType.CONTROL, c.containerType)
        assertEquals(ControlCmd.TIMEOUT, c.controlCmd)
        assertEquals(0, c.payload.size)
    }

    @Test
    fun timeoutResponse() {
        val c = makeTimeoutResponse(transactionId = 5, timeoutMs = 200)
        assertEquals(ContainerType.CONTROL, c.containerType)
        assertEquals(ControlCmd.TIMEOUT, c.controlCmd)
        val ms = ByteBuffer.wrap(c.payload).order(ByteOrder.LITTLE_ENDIAN).short.toInt() and 0xFFFF
        assertEquals(200, ms)
    }

    @Test
    fun streamEndC2P() {
        val c = makeStreamEndC2P(transactionId = 3)
        assertEquals(ControlCmd.STREAM_END_C2P, c.controlCmd)
        val data = c.serialize()
        val c2 = Container.deserialize(data)
        assertEquals(ControlCmd.STREAM_END_C2P, c2.controlCmd)
    }

    @Test
    fun streamEndP2C() {
        val c = makeStreamEndP2C(transactionId = 3)
        assertEquals(ControlCmd.STREAM_END_P2C, c.controlCmd)
    }

    @Test
    fun capabilitiesRequest() {
        val c = makeCapabilitiesRequest(transactionId = 7)
        assertEquals(ContainerType.CONTROL, c.containerType)
        assertEquals(ControlCmd.CAPABILITIES, c.controlCmd)
        assertEquals(0, c.payload.size)
    }

    @Test
    fun capabilitiesResponse() {
        val c = makeCapabilitiesResponse(
            transactionId = 7,
            maxRequestPayloadSize = 256,
            maxResponsePayloadSize = 65535
        )
        assertEquals(ContainerType.CONTROL, c.containerType)
        assertEquals(ControlCmd.CAPABILITIES, c.controlCmd)
        assertEquals(4, c.payload.size)
        val buf = ByteBuffer.wrap(c.payload).order(ByteOrder.LITTLE_ENDIAN)
        assertEquals(256, buf.short.toInt() and 0xFFFF)
        assertEquals(65535, buf.short.toInt() and 0xFFFF)
    }

    @Test
    fun errorResponse() {
        val c = makeErrorResponse(
            transactionId = 10,
            errorCode = BLERPC_ERROR_RESPONSE_TOO_LARGE.toInt()
        )
        assertEquals(ContainerType.CONTROL, c.containerType)
        assertEquals(ControlCmd.ERROR, c.controlCmd)
        assertArrayEquals(byteArrayOf(0x01), c.payload)

        val data = c.serialize()
        val c2 = Container.deserialize(data)
        assertEquals(ContainerType.CONTROL, c2.containerType)
        assertEquals(ControlCmd.ERROR, c2.controlCmd)
        assertArrayEquals(byteArrayOf(BLERPC_ERROR_RESPONSE_TOO_LARGE), c2.payload)
    }
}
