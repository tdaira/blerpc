package com.blerpc.protocol

import org.junit.Assert.*
import org.junit.Test

class ContainerSplitterTest {

    @Test
    fun smallPayloadSingleContainer() {
        val splitter = ContainerSplitter(mtu = 247)
        val payload = "hello".toByteArray()
        val containers = splitter.split(payload, transactionId = 0)
        assertEquals(1, containers.size)
        assertEquals(ContainerType.FIRST, containers[0].containerType)
        assertEquals(5, containers[0].totalLength)
        assertArrayEquals("hello".toByteArray(), containers[0].payload)
    }

    @Test
    fun largePayloadMultipleContainers() {
        val mtu = 27
        val splitter = ContainerSplitter(mtu = mtu)
        val effective = mtu - ATT_OVERHEAD // 24
        val firstPayloadMax = effective - FIRST_HEADER_SIZE // 18
        val subsequentPayloadMax = effective - SUBSEQUENT_HEADER_SIZE // 20

        val payload = ByteArray(512) { (it % 256).toByte() }
        val containers = splitter.split(payload, transactionId = 5)

        assertEquals(ContainerType.FIRST, containers[0].containerType)
        assertEquals(512, containers[0].totalLength)
        assertEquals(firstPayloadMax, containers[0].payload.size)

        for (c in containers.drop(1)) {
            assertEquals(ContainerType.SUBSEQUENT, c.containerType)
            assertTrue(c.payload.size <= subsequentPayloadMax)
        }

        // Verify all data is accounted for
        val reassembled = containers.flatMap { it.payload.toList() }.toByteArray()
        assertArrayEquals(payload, reassembled)
    }

    @Test
    fun boundaryPayloadExactlyFirstMax() {
        val mtu = 30
        val splitter = ContainerSplitter(mtu = mtu)
        val effective = mtu - ATT_OVERHEAD // 27
        val firstMax = effective - FIRST_HEADER_SIZE // 21

        val payload = ByteArray(firstMax) { 'A'.code.toByte() }
        val containers = splitter.split(payload, transactionId = 0)
        assertEquals(1, containers.size)
        assertArrayEquals(payload, containers[0].payload)
    }

    @Test
    fun boundaryPayloadOneByteOverFirstMax() {
        val mtu = 30
        val splitter = ContainerSplitter(mtu = mtu)
        val effective = mtu - ATT_OVERHEAD
        val firstMax = effective - FIRST_HEADER_SIZE

        val payload = ByteArray(firstMax + 1) { 'A'.code.toByte() }
        val containers = splitter.split(payload, transactionId = 0)
        assertEquals(2, containers.size)
        assertEquals(firstMax, containers[0].payload.size)
        assertEquals(1, containers[1].payload.size)
    }

    @Test
    fun emptyPayload() {
        val splitter = ContainerSplitter(mtu = 247)
        val containers = splitter.split(ByteArray(0), transactionId = 0)
        assertEquals(1, containers.size)
        assertEquals(0, containers[0].totalLength)
        assertEquals(0, containers[0].payload.size)
    }

    @Test
    fun transactionIdAutoIncrement() {
        val splitter = ContainerSplitter(mtu = 247)
        val c1 = splitter.split("a".toByteArray())
        val c2 = splitter.split("b".toByteArray())
        assertEquals(0, c1[0].transactionId)
        assertEquals(1, c2[0].transactionId)
    }

    @Test
    fun transactionIdWrapsAt256() {
        val splitter = ContainerSplitter(mtu = 247)
        // Consume IDs 0..254
        repeat(255) { splitter.nextTransactionId() }
        val c1 = splitter.split("a".toByteArray())
        val c2 = splitter.split("b".toByteArray())
        assertEquals(255, c1[0].transactionId)
        assertEquals(0, c2[0].transactionId)
    }
}
