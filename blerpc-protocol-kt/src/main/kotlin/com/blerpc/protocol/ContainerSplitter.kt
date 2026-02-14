package com.blerpc.protocol

class ContainerSplitter(private val mtu: Int = 247) {
    private var transactionCounter = 0

    val effectiveMtu: Int get() = mtu - ATT_OVERHEAD

    fun nextTransactionId(): Int {
        val tid = transactionCounter
        transactionCounter = (transactionCounter + 1) and 0xFF
        return tid
    }

    fun split(payload: ByteArray, transactionId: Int? = null): List<Container> {
        val tid = transactionId ?: nextTransactionId()
        val totalLength = payload.size
        if (totalLength > 65535) {
            throw IllegalArgumentException("Payload too large: $totalLength > 65535")
        }

        val containers = mutableListOf<Container>()

        // First container
        val firstMaxPayload = effectiveMtu - FIRST_HEADER_SIZE
        val firstPayloadSize = minOf(payload.size, firstMaxPayload)
        val firstPayload = payload.copyOfRange(0, firstPayloadSize)
        containers.add(
            Container(
                transactionId = tid,
                sequenceNumber = 0,
                containerType = ContainerType.FIRST,
                totalLength = totalLength,
                payload = firstPayload
            )
        )

        var offset = firstPayloadSize
        var seq = 1
        val subsequentMaxPayload = effectiveMtu - SUBSEQUENT_HEADER_SIZE

        while (offset < totalLength) {
            if (seq > 255) {
                throw IllegalArgumentException(
                    "Payload requires more than 256 containers (seq=$seq), " +
                            "exceeding 8-bit sequence_number limit"
                )
            }
            val chunkSize = minOf(payload.size - offset, subsequentMaxPayload)
            val chunk = payload.copyOfRange(offset, offset + chunkSize)
            containers.add(
                Container(
                    transactionId = tid,
                    sequenceNumber = seq,
                    containerType = ContainerType.SUBSEQUENT,
                    payload = chunk
                )
            )
            offset += chunkSize
            seq++
        }

        return containers
    }
}
