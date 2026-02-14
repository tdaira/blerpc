package com.blerpc.protocol

import java.nio.ByteBuffer
import java.nio.ByteOrder

enum class CommandType(val value: Int) {
    REQUEST(0),
    RESPONSE(1);

    companion object {
        fun fromValue(v: Int): CommandType = entries.first { it.value == v }
    }
}

data class CommandPacket(
    val cmdType: CommandType,
    val cmdName: String,
    val data: ByteArray = ByteArray(0)
) {
    fun serialize(): ByteArray {
        val nameBytes = cmdName.toByteArray(Charsets.US_ASCII)
        require(nameBytes.size <= 255) { "cmd_name too long: ${nameBytes.size} > 255" }
        require(data.size <= 65535) { "data too long: ${data.size} > 65535" }

        val byte0 = (cmdType.value and 0x01) shl 7
        val buf = ByteBuffer.allocate(1 + 1 + nameBytes.size + 2 + data.size)
            .order(ByteOrder.LITTLE_ENDIAN)
        buf.put(byte0.toByte())
        buf.put(nameBytes.size.toByte())
        buf.put(nameBytes)
        buf.putShort(data.size.toShort())
        buf.put(data)
        return buf.array()
    }

    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is CommandPacket) return false
        return cmdType == other.cmdType && cmdName == other.cmdName && data.contentEquals(other.data)
    }

    override fun hashCode(): Int {
        var result = cmdType.hashCode()
        result = 31 * result + cmdName.hashCode()
        result = 31 * result + data.contentHashCode()
        return result
    }

    companion object {
        fun deserialize(data: ByteArray): CommandPacket {
            require(data.size >= 2) { "Command packet too short: ${data.size} bytes" }

            val cmdType = CommandType.fromValue((data[0].toInt() shr 7) and 0x01)
            val cmdNameLen = data[1].toInt() and 0xFF

            var offset = 2
            require(data.size >= offset + cmdNameLen + 2) { "Command packet truncated" }

            val cmdName = String(data, offset, cmdNameLen, Charsets.US_ASCII)
            offset += cmdNameLen

            val dataLen = ByteBuffer.wrap(data, offset, 2)
                .order(ByteOrder.LITTLE_ENDIAN).short.toInt() and 0xFFFF
            offset += 2

            val payload = data.copyOfRange(offset, offset + dataLen)
            return CommandPacket(cmdType, cmdName, payload)
        }
    }
}
