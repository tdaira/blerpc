package com.blerpc.android.ble

import android.annotation.SuppressLint
import android.bluetooth.*
import android.bluetooth.le.*
import android.content.Context
import android.os.ParcelUuid
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withTimeout
import java.util.UUID
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

private val SERVICE_UUID = UUID.fromString("12340001-0000-1000-8000-00805f9b34fb")
private val CHAR_UUID = UUID.fromString("12340002-0000-1000-8000-00805f9b34fb")
private val CCCD_UUID = UUID.fromString("00002902-0000-1000-8000-00805f9b34fb")

@SuppressLint("MissingPermission")
class BleTransport(private val context: Context) {
    private var gatt: BluetoothGatt? = null
    private var writeChar: BluetoothGattCharacteristic? = null
    private val notifyChannel = Channel<ByteArray>(Channel.UNLIMITED)
    var mtu: Int = 23
        private set
    val isConnected: Boolean get() = gatt != null

    private var writeComplete: (() -> Unit)? = null
    private var connectCont: ((Boolean) -> Unit)? = null
    private var mtuCont: ((Int) -> Unit)? = null
    private var descriptorWriteCont: (() -> Unit)? = null

    private val gattCallback = object : BluetoothGattCallback() {
        override fun onConnectionStateChange(g: BluetoothGatt, status: Int, newState: Int) {
            if (newState == BluetoothProfile.STATE_CONNECTED) {
                g.discoverServices()
            } else {
                connectCont?.invoke(false)
                connectCont = null
            }
        }

        override fun onServicesDiscovered(g: BluetoothGatt, status: Int) {
            connectCont?.invoke(status == BluetoothGatt.GATT_SUCCESS)
            connectCont = null
        }

        override fun onMtuChanged(g: BluetoothGatt, newMtu: Int, status: Int) {
            if (status == BluetoothGatt.GATT_SUCCESS) {
                mtu = newMtu
            }
            mtuCont?.invoke(mtu)
            mtuCont = null
        }

        override fun onDescriptorWrite(g: BluetoothGatt, descriptor: BluetoothGattDescriptor, status: Int) {
            descriptorWriteCont?.invoke()
            descriptorWriteCont = null
        }

        override fun onCharacteristicWrite(g: BluetoothGatt, characteristic: BluetoothGattCharacteristic, status: Int) {
            writeComplete?.invoke()
            writeComplete = null
        }

        @Deprecated("Deprecated in API 33+", ReplaceWith("onCharacteristicChanged(g, characteristic, value)"))
        @Suppress("DEPRECATION")
        override fun onCharacteristicChanged(g: BluetoothGatt, characteristic: BluetoothGattCharacteristic) {
            // API 31-32 callback
            notifyChannel.trySend(characteristic.value)
        }

        override fun onCharacteristicChanged(
            g: BluetoothGatt,
            characteristic: BluetoothGattCharacteristic,
            value: ByteArray
        ) {
            // API 33+ callback
            notifyChannel.trySend(value)
        }
    }

    suspend fun connect(deviceName: String = "blerpc", timeout: Long = 10000) {
        val adapter = (context.getSystemService(Context.BLUETOOTH_SERVICE) as BluetoothManager).adapter
        val scanner = adapter.bluetoothLeScanner
            ?: throw IllegalStateException("BLE scanner not available")

        // Scan for device
        val device = withTimeout(timeout) {
            suspendCancellableCoroutine { cont ->
                var resumed = false
                val callback = object : ScanCallback() {
                    override fun onScanResult(callbackType: Int, result: ScanResult) {
                        if (!resumed && result.device.name == deviceName) {
                            resumed = true
                            scanner.stopScan(this)
                            cont.resume(result.device)
                        }
                    }

                    override fun onScanFailed(errorCode: Int) {
                        if (!resumed) {
                            resumed = true
                            cont.resumeWithException(RuntimeException("Scan failed: $errorCode"))
                        }
                    }
                }
                val settings = ScanSettings.Builder()
                    .setScanMode(ScanSettings.SCAN_MODE_LOW_LATENCY)
                    .build()
                scanner.startScan(null, settings, callback)
                cont.invokeOnCancellation { scanner.stopScan(callback) }
            }
        }

        // Connect
        val success = suspendCancellableCoroutine { cont ->
            connectCont = { ok -> cont.resume(ok) }
            gatt = device.connectGatt(context, false, gattCallback, BluetoothDevice.TRANSPORT_LE)
        }
        if (!success) throw RuntimeException("Failed to connect/discover services")

        val g = gatt!!

        // Request MTU
        mtu = suspendCancellableCoroutine { cont ->
            mtuCont = { newMtu -> cont.resume(newMtu) }
            if (!g.requestMtu(247)) {
                mtuCont = null
                cont.resume(23)
            }
        }

        // Find characteristic
        val service = g.getService(SERVICE_UUID)
            ?: throw RuntimeException("Service not found")
        writeChar = service.getCharacteristic(CHAR_UUID)
            ?: throw RuntimeException("Characteristic not found")

        // Enable notifications
        g.setCharacteristicNotification(writeChar, true)
        val cccd = writeChar!!.getDescriptor(CCCD_UUID)
            ?: throw RuntimeException("CCCD not found")
        suspendCancellableCoroutine { cont ->
            descriptorWriteCont = { cont.resume(Unit) }
            @Suppress("DEPRECATION")
            cccd.value = BluetoothGattDescriptor.ENABLE_NOTIFICATION_VALUE
            @Suppress("DEPRECATION")
            g.writeDescriptor(cccd)
        }
    }

    suspend fun write(data: ByteArray) {
        val g = gatt ?: throw IllegalStateException("Not connected")
        val c = writeChar ?: throw IllegalStateException("No characteristic")

        suspendCancellableCoroutine { cont ->
            writeComplete = { cont.resume(Unit) }
            @Suppress("DEPRECATION")
            c.value = data
            c.writeType = BluetoothGattCharacteristic.WRITE_TYPE_NO_RESPONSE
            @Suppress("DEPRECATION")
            if (!g.writeCharacteristic(c)) {
                writeComplete = null
                cont.resumeWithException(RuntimeException("Write failed"))
            }
        }
    }

    suspend fun readNotify(timeoutMs: Long): ByteArray {
        return withTimeout(timeoutMs) {
            notifyChannel.receive()
        }
    }

    fun drainNotifications() {
        while (notifyChannel.tryReceive().isSuccess) { /* discard */ }
    }

    fun disconnect() {
        gatt?.let {
            it.close()
            gatt = null
        }
        writeChar = null
    }
}
