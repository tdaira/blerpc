package com.blerpc.protocol

import org.junit.Assert.*
import org.junit.Test

class ContainerAssemblerTest {

    @Test
    fun singleContainerAssembly() {
        val assembler = ContainerAssembler()
        val c = Container(
            transactionId = 0,
            sequenceNumber = 0,
            containerType = ContainerType.FIRST,
            totalLength = 5,
            payload = "hello".toByteArray()
        )
        val result = assembler.feed(c)
        assertArrayEquals("hello".toByteArray(), result)
    }

    @Test
    fun multiContainerAssembly() {
        val assembler = ContainerAssembler()
        val c1 = Container(
            transactionId = 1,
            sequenceNumber = 0,
            containerType = ContainerType.FIRST,
            totalLength = 8,
            payload = "hell".toByteArray()
        )
        val c2 = Container(
            transactionId = 1,
            sequenceNumber = 1,
            containerType = ContainerType.SUBSEQUENT,
            payload = "o wo".toByteArray()
        )
        assertNull(assembler.feed(c1))
        val result = assembler.feed(c2)
        assertArrayEquals("hello wo".toByteArray(), result)
    }

    @Test
    fun sequenceGapDiscardsTransaction() {
        val assembler = ContainerAssembler()
        val c1 = Container(
            transactionId = 2,
            sequenceNumber = 0,
            containerType = ContainerType.FIRST,
            totalLength = 10,
            payload = "abc".toByteArray()
        )
        val cBad = Container(
            transactionId = 2,
            sequenceNumber = 2, // Gap: expected 1
            containerType = ContainerType.SUBSEQUENT,
            payload = "def".toByteArray()
        )
        assertNull(assembler.feed(c1))
        val result = assembler.feed(cBad)
        assertNull(result)
    }

    @Test
    fun controlContainerIgnored() {
        val assembler = ContainerAssembler()
        val c = makeTimeoutRequest(transactionId = 0)
        assertNull(assembler.feed(c))
    }

    @Test
    fun subsequentWithoutFirstIgnored() {
        val assembler = ContainerAssembler()
        val c = Container(
            transactionId = 99,
            sequenceNumber = 1,
            containerType = ContainerType.SUBSEQUENT,
            payload = "orphan".toByteArray()
        )
        assertNull(assembler.feed(c))
    }

    @Test
    fun splitAssembleRoundtripSmall() {
        val splitter = ContainerSplitter(mtu = 247)
        val assembler = ContainerAssembler()
        val payload = "hello world".toByteArray()

        val containers = splitter.split(payload, transactionId = 0)
        var result: ByteArray? = null
        for (c in containers) {
            val serialized = c.serialize()
            val deserialized = Container.deserialize(serialized)
            result = assembler.feed(deserialized)
        }
        assertArrayEquals(payload, result)
    }

    @Test
    fun splitAssembleRoundtripLarge() {
        val splitter = ContainerSplitter(mtu = 27)
        val assembler = ContainerAssembler()
        val payload = ByteArray(1024) { (it % 256).toByte() }

        val containers = splitter.split(payload, transactionId = 10)
        var result: ByteArray? = null
        for (c in containers) {
            val serialized = c.serialize()
            val deserialized = Container.deserialize(serialized)
            result = assembler.feed(deserialized)
        }
        assertArrayEquals(payload, result)
    }

    @Test
    fun splitAssembleRoundtripLargePayload() {
        val splitter = ContainerSplitter(mtu = 247)
        val assembler = ContainerAssembler()
        val payload = ByteArray(60000) { 0xab.toByte() }

        val containers = splitter.split(payload, transactionId = 0)
        assertTrue(containers.size > 200)
        var result: ByteArray? = null
        for (c in containers) {
            val serialized = c.serialize()
            val deserialized = Container.deserialize(serialized)
            result = assembler.feed(deserialized)
        }
        assertArrayEquals(payload, result)
    }

    @Test(expected = IllegalArgumentException::class)
    fun payloadTooLargeRaises() {
        val splitter = ContainerSplitter(mtu = 27)
        val payload = ByteArray(10000) { 0 }
        splitter.split(payload, transactionId = 0)
    }

    @Test
    fun splitAssembleRoundtripEmpty() {
        val splitter = ContainerSplitter(mtu = 247)
        val assembler = ContainerAssembler()
        val payload = ByteArray(0)

        val containers = splitter.split(payload, transactionId = 0)
        var result: ByteArray? = null
        for (c in containers) {
            val serialized = c.serialize()
            val deserialized = Container.deserialize(serialized)
            result = assembler.feed(deserialized)
        }
        assertArrayEquals(payload, result)
    }
}
