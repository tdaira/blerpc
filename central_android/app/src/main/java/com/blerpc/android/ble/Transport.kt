package com.blerpc.android.ble

interface Transport {
    val mtu: Int
    val isConnected: Boolean

    suspend fun write(data: ByteArray)

    suspend fun readNotify(timeoutMs: Long): ByteArray

    fun drainNotifications()

    fun disconnect()
}
