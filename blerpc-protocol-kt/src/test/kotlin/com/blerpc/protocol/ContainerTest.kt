package com.blerpc.protocol

import org.junit.Assert.*
import org.junit.Test
import java.nio.ByteBuffer
import java.nio.ByteOrder

class ContainerTest {

    @Test
    fun firstContainerRoundtrip() {
        val c = Container(
            transactionId = 42,
            sequenceNumber = 0,
            containerType = ContainerType.FIRST,
            totalLength = 100,
            payload = byteArrayOf(0x01, 0x02, 0x03)
        )
        val data = c.serialize()
        val c2 = Container.deserialize(data)
        assertEquals(42, c2.transactionId)
        assertEquals(0, c2.sequenceNumber)
        assertEquals(ContainerType.FIRST, c2.containerType)
        assertEquals(100, c2.totalLength)
        assertArrayEquals(byteArrayOf(0x01, 0x02, 0x03), c2.payload)
    }

    @Test
    fun subsequentContainerRoundtrip() {
        val c = Container(
            transactionId = 7,
            sequenceNumber = 3,
            containerType = ContainerType.SUBSEQUENT,
            payload = byteArrayOf(0xaa.toByte(), 0xbb.toByte())
        )
        val data = c.serialize()
        val c2 = Container.deserialize(data)
        assertEquals(7, c2.transactionId)
        assertEquals(3, c2.sequenceNumber)
        assertEquals(ContainerType.SUBSEQUENT, c2.containerType)
        assertArrayEquals(byteArrayOf(0xaa.toByte(), 0xbb.toByte()), c2.payload)
    }

    @Test
    fun controlContainerRoundtrip() {
        val payload = ByteBuffer.allocate(2).order(ByteOrder.LITTLE_ENDIAN).putShort(500).array()
        val c = Container(
            transactionId = 1,
            sequenceNumber = 0,
            containerType = ContainerType.CONTROL,
            controlCmd = ControlCmd.TIMEOUT,
            payload = payload
        )
        val data = c.serialize()
        val c2 = Container.deserialize(data)
        assertEquals(ContainerType.CONTROL, c2.containerType)
        assertEquals(ControlCmd.TIMEOUT, c2.controlCmd)
        val ms = ByteBuffer.wrap(c2.payload).order(ByteOrder.LITTLE_ENDIAN).short.toInt() and 0xFFFF
        assertEquals(500, ms)
    }

    @Test
    fun flagsByteEncoding() {
        val c = Container(
            transactionId = 0,
            sequenceNumber = 0,
            containerType = ContainerType.CONTROL,
            controlCmd = ControlCmd.STREAM_END_C2P,
            payload = ByteArray(0)
        )
        val data = c.serialize()
        val flags = data[2].toInt() and 0xFF
        // type=0b11 in bits 7-6 => 0xC0
        // control_cmd=0x2 in bits 5-2 => 0x08
        assertEquals(0xC8, flags)
    }

    @Test(expected = IllegalArgumentException::class)
    fun deserializeTooShort() {
        Container.deserialize(byteArrayOf(0x00, 0x01))
    }

    @Test
    fun firstContainerHeaderSize() {
        val c = Container(
            transactionId = 0,
            sequenceNumber = 0,
            containerType = ContainerType.FIRST,
            totalLength = 0,
            payload = ByteArray(0)
        )
        assertEquals(FIRST_HEADER_SIZE, c.serialize().size)
    }

    @Test
    fun subsequentContainerHeaderSize() {
        val c = Container(
            transactionId = 0,
            sequenceNumber = 0,
            containerType = ContainerType.SUBSEQUENT,
            payload = ByteArray(0)
        )
        assertEquals(SUBSEQUENT_HEADER_SIZE, c.serialize().size)
    }
}
