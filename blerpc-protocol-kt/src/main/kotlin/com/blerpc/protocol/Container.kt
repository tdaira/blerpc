package com.blerpc.protocol

import java.nio.ByteBuffer
import java.nio.ByteOrder

data class Container(
    val transactionId: Int,
    val sequenceNumber: Int,
    val containerType: ContainerType,
    val controlCmd: ControlCmd = ControlCmd.NONE,
    val totalLength: Int = 0,
    val payload: ByteArray = ByteArray(0)
) {
    fun serialize(): ByteArray {
        val flags = ((containerType.value and 0x03) shl 6) or ((controlCmd.value and 0x0F) shl 2)

        return if (containerType == ContainerType.FIRST) {
            val buf = ByteBuffer.allocate(FIRST_HEADER_SIZE + payload.size)
                .order(ByteOrder.LITTLE_ENDIAN)
            buf.put(transactionId.toByte())
            buf.put(sequenceNumber.toByte())
            buf.put(flags.toByte())
            buf.putShort(totalLength.toShort())
            buf.put(payload.size.toByte())
            buf.put(payload)
            buf.array()
        } else {
            val buf = ByteBuffer.allocate(SUBSEQUENT_HEADER_SIZE + payload.size)
                .order(ByteOrder.LITTLE_ENDIAN)
            buf.put(transactionId.toByte())
            buf.put(sequenceNumber.toByte())
            buf.put(flags.toByte())
            buf.put(payload.size.toByte())
            buf.put(payload)
            buf.array()
        }
    }

    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is Container) return false
        return transactionId == other.transactionId &&
                sequenceNumber == other.sequenceNumber &&
                containerType == other.containerType &&
                controlCmd == other.controlCmd &&
                totalLength == other.totalLength &&
                payload.contentEquals(other.payload)
    }

    override fun hashCode(): Int {
        var result = transactionId
        result = 31 * result + sequenceNumber
        result = 31 * result + containerType.hashCode()
        result = 31 * result + controlCmd.hashCode()
        result = 31 * result + totalLength
        result = 31 * result + payload.contentHashCode()
        return result
    }

    companion object {
        fun deserialize(data: ByteArray): Container {
            if (data.size < 4) {
                throw IllegalArgumentException("Container too short: ${data.size} bytes")
            }

            val transactionId = data[0].toInt() and 0xFF
            val sequenceNumber = data[1].toInt() and 0xFF
            val flagsByte = data[2].toInt() and 0xFF
            val containerType = ContainerType.fromValue((flagsByte shr 6) and 0x03)
            val controlCmd = ControlCmd.fromValue((flagsByte shr 2) and 0x0F)

            return if (containerType == ContainerType.FIRST) {
                if (data.size < FIRST_HEADER_SIZE) {
                    throw IllegalArgumentException("FIRST container too short: ${data.size} bytes")
                }
                val totalLength = ByteBuffer.wrap(data, 3, 2)
                    .order(ByteOrder.LITTLE_ENDIAN).short.toInt() and 0xFFFF
                val payloadLen = data[5].toInt() and 0xFF
                val payload = data.copyOfRange(FIRST_HEADER_SIZE, FIRST_HEADER_SIZE + payloadLen)
                Container(transactionId, sequenceNumber, containerType, controlCmd, totalLength, payload)
            } else {
                val payloadLen = data[3].toInt() and 0xFF
                val payload = data.copyOfRange(SUBSEQUENT_HEADER_SIZE, SUBSEQUENT_HEADER_SIZE + payloadLen)
                Container(transactionId, sequenceNumber, containerType, controlCmd, 0, payload)
            }
        }
    }
}
