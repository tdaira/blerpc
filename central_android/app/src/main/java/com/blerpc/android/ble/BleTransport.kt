package com.blerpc.android.ble

import android.annotation.SuppressLint
import android.bluetooth.*
import android.bluetooth.le.*
import android.content.Context
import android.os.Build
import android.util.Log
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.delay
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withTimeout
import java.util.UUID
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

private const val TAG = "BleTransport"

val SERVICE_UUID: UUID = UUID.fromString("12340001-0000-1000-8000-00805f9b34fb")
private val CHAR_UUID = UUID.fromString("12340002-0000-1000-8000-00805f9b34fb")
private val CCCD_UUID = UUID.fromString("00002902-0000-1000-8000-00805f9b34fb")

data class ScannedDevice(
    val name: String?,
    val address: String,
    val rssi: Int,
    val manufacturerData: Map<Int, ByteArray>,
    val serviceData: Map<String, ByteArray>,
    val serviceUuids: List<String>,
    val rawBytes: ByteArray?,
    internal val device: BluetoothDevice
)

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
            Log.d(TAG, "onConnectionStateChange: status=$status, newState=$newState (0=disconnected, 2=connected)")
            if (newState == BluetoothProfile.STATE_CONNECTED) {
                if (status != BluetoothGatt.GATT_SUCCESS) {
                    Log.w(TAG, "Connected but status=$status (not GATT_SUCCESS), proceeding with discoverServices")
                }
                g.discoverServices()
            } else {
                Log.w(TAG, "Connection failed or disconnected: status=$status, newState=$newState")
                connectCont?.invoke(false)
                connectCont = null
            }
        }

        override fun onServicesDiscovered(g: BluetoothGatt, status: Int) {
            Log.d(TAG, "onServicesDiscovered: status=$status, services=${g.services.map { it.uuid }}")
            if (status != BluetoothGatt.GATT_SUCCESS) {
                Log.e(TAG, "Service discovery failed with status=$status")
            }
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

        @Suppress("DEPRECATION")
        override fun onCharacteristicChanged(g: BluetoothGatt, characteristic: BluetoothGattCharacteristic) {
            // Required for API < 33; the 3-arg overload is only called on API 33+
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

    suspend fun scan(timeout: Long = 5000, serviceUuid: UUID? = SERVICE_UUID): List<ScannedDevice> {
        val adapter = (context.getSystemService(Context.BLUETOOTH_SERVICE) as BluetoothManager).adapter
        val scanner = adapter.bluetoothLeScanner
            ?: throw IllegalStateException("BLE scanner not available")

        val results = mutableMapOf<String, ScannedDevice>()

        val callback = object : ScanCallback() {
            override fun onScanResult(callbackType: Int, result: ScanResult) {
                val record = result.scanRecord
                val mfgData = mutableMapOf<Int, ByteArray>()
                record?.manufacturerSpecificData?.let { sparse ->
                    for (i in 0 until sparse.size()) {
                        mfgData[sparse.keyAt(i)] = sparse.valueAt(i)
                    }
                }
                val svcData = mutableMapOf<String, ByteArray>()
                record?.serviceData?.forEach { (uuid, data) ->
                    svcData[uuid.toString()] = data
                }
                val svcUuids = record?.serviceUuids?.map { it.toString() } ?: emptyList()

                results[result.device.address] = ScannedDevice(
                    name = result.device.name,
                    address = result.device.address,
                    rssi = result.rssi,
                    manufacturerData = mfgData,
                    serviceData = svcData,
                    serviceUuids = svcUuids,
                    rawBytes = record?.bytes,
                    device = result.device
                )
            }

            override fun onScanFailed(errorCode: Int) {
                // Scan failed, results will be empty
            }
        }

        val settings = ScanSettings.Builder()
            .setScanMode(ScanSettings.SCAN_MODE_LOW_LATENCY)
            .build()

        // Use unfiltered scan and filter manually — Android's ScanFilter
        // does not reliably match 128-bit service UUIDs on all devices.
        Log.d(TAG, "scan: starting (unfiltered, manual filter=${serviceUuid}), timeout=${timeout}ms")
        scanner.startScan(null, settings, callback)
        delay(timeout)
        scanner.stopScan(callback)

        val filtered = if (serviceUuid != null) {
            val target = serviceUuid.toString()
            results.values.filter { device ->
                device.serviceUuids.any { it.equals(target, ignoreCase = true) }
            }
        } else {
            results.values.toList()
        }

        Log.d(TAG, "scan: found ${filtered.size}/${results.size} devices: ${filtered.map { "${it.address}(${it.name},rssi=${it.rssi})" }}")
        return filtered.sortedByDescending { it.rssi }
    }

    suspend fun connect(device: ScannedDevice) {
        Log.d(TAG, "connect: address=${device.address}, name=${device.name}, bondState=${device.device.bondState}")

        // Connect
        val success = suspendCancellableCoroutine { cont ->
            connectCont = { ok -> cont.resume(ok) }
            gatt = device.device.connectGatt(context, false, gattCallback, BluetoothDevice.TRANSPORT_LE)
        }
        if (!success) {
            Log.e(TAG, "connect: GATT connection/service discovery failed")
            gatt?.close()
            gatt = null
            throw RuntimeException("Failed to connect/discover services")
        }

        val g = gatt!!
        Log.d(TAG, "connect: GATT connected, requesting MTU")

        // Request MTU
        mtu = suspendCancellableCoroutine { cont ->
            mtuCont = { newMtu -> cont.resume(newMtu) }
            if (!g.requestMtu(247)) {
                mtuCont = null
                cont.resume(23)
            }
        }
        Log.d(TAG, "connect: MTU=$mtu")

        // Find characteristic
        Log.d(TAG, "connect: looking for service $SERVICE_UUID among ${g.services.size} services: ${g.services.map { it.uuid }}")
        val service = g.getService(SERVICE_UUID)
            ?: throw RuntimeException("Service not found (discovered ${g.services.size} services: ${g.services.map { it.uuid }})")
        writeChar = service.getCharacteristic(CHAR_UUID)
            ?: throw RuntimeException("Characteristic not found")

        // Enable notifications
        val wc = writeChar!!
        g.setCharacteristicNotification(wc, true)
        val cccd = wc.getDescriptor(CCCD_UUID)
            ?: throw RuntimeException("CCCD not found")
        suspendCancellableCoroutine { cont ->
            descriptorWriteCont = { cont.resume(Unit) }
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                g.writeDescriptor(cccd, BluetoothGattDescriptor.ENABLE_NOTIFICATION_VALUE)
            } else {
                @Suppress("DEPRECATION")
                cccd.value = BluetoothGattDescriptor.ENABLE_NOTIFICATION_VALUE
                @Suppress("DEPRECATION")
                g.writeDescriptor(cccd)
            }
        }
    }

    suspend fun write(data: ByteArray) {
        val g = gatt ?: throw IllegalStateException("Not connected")
        val c = writeChar ?: throw IllegalStateException("No characteristic")

        suspendCancellableCoroutine { cont ->
            writeComplete = { cont.resume(Unit) }
            val ok = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                g.writeCharacteristic(
                    c, data, BluetoothGattCharacteristic.WRITE_TYPE_NO_RESPONSE
                ) == BluetoothStatusCodes.SUCCESS
            } else {
                @Suppress("DEPRECATION")
                c.value = data
                c.writeType = BluetoothGattCharacteristic.WRITE_TYPE_NO_RESPONSE
                @Suppress("DEPRECATION")
                g.writeCharacteristic(c)
            }
            if (!ok) {
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
