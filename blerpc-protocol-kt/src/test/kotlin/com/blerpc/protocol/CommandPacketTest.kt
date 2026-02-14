package com.blerpc.protocol

import org.junit.Assert.*
import org.junit.Test

class CommandPacketTest {

    @Test
    fun requestRoundtrip() {
        val cmd = CommandPacket(
            cmdType = CommandType.REQUEST,
            cmdName = "echo",
            data = byteArrayOf(0x0a, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f)
        )
        val serialized = cmd.serialize()
        val cmd2 = CommandPacket.deserialize(serialized)
        assertEquals(CommandType.REQUEST, cmd2.cmdType)
        assertEquals("echo", cmd2.cmdName)
        assertArrayEquals(cmd.data, cmd2.data)
    }

    @Test
    fun responseRoundtrip() {
        val cmd = CommandPacket(
            cmdType = CommandType.RESPONSE,
            cmdName = "flash_read",
            data = byteArrayOf(0x01, 0x02, 0x03)
        )
        val serialized = cmd.serialize()
        val cmd2 = CommandPacket.deserialize(serialized)
        assertEquals(CommandType.RESPONSE, cmd2.cmdType)
        assertEquals("flash_read", cmd2.cmdName)
        assertArrayEquals(cmd.data, cmd2.data)
    }

    @Test
    fun requestTypeBit() {
        val req = CommandPacket(CommandType.REQUEST, "test", ByteArray(0))
        val data = req.serialize()
        // Bit 7 of byte 0 should be 0 for REQUEST
        assertEquals(0, data[0].toInt() and 0x80)

        val resp = CommandPacket(CommandType.RESPONSE, "test", ByteArray(0))
        val data2 = resp.serialize()
        // Bit 7 of byte 0 should be 1 for RESPONSE
        assertEquals(0x80, data2[0].toInt() and 0x80)
    }

    @Test
    fun emptyData() {
        val cmd = CommandPacket(CommandType.REQUEST, "ping", ByteArray(0))
        val serialized = cmd.serialize()
        val cmd2 = CommandPacket.deserialize(serialized)
        assertEquals("ping", cmd2.cmdName)
        assertEquals(0, cmd2.data.size)
    }

    @Test(expected = IllegalArgumentException::class)
    fun deserializeTooShort() {
        CommandPacket.deserialize(byteArrayOf(0x00))
    }

    @Test
    fun dataLenLittleEndian() {
        val cmd = CommandPacket(CommandType.REQUEST, "x", ByteArray(300))
        val serialized = cmd.serialize()
        // Byte layout: [type(1)][name_len(1)][name(1)][data_len_lo][data_len_hi][data(300)]
        val lo = serialized[3].toInt() and 0xFF
        val hi = serialized[4].toInt() and 0xFF
        assertEquals(300, lo + hi * 256)
    }
}
