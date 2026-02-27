import 'dart:async';
import 'dart:collection';
import 'dart:io' show Platform;

import 'package:flutter/foundation.dart';
import 'package:flutter_blue_plus/flutter_blue_plus.dart';

const String serviceUuid = '12340001-0000-1000-8000-00805f9b34fb';
const String charUuid = '12340002-0000-1000-8000-00805f9b34fb';

class ScannedDevice {
  final BluetoothDevice device;
  final String? name;
  final String address;

  ScannedDevice({required this.device, this.name, required this.address});
}

class BleTransport {
  BluetoothDevice? _device;
  BluetoothCharacteristic? _char;
  StreamSubscription<List<int>>? _notifySub;
  final _notifyQueue = Queue<Uint8List>();
  Completer<Uint8List>? _notifyWaiter;
  int _mtu = 23;

  int get mtu => _mtu;

  Future<List<ScannedDevice>> scan({
    Duration timeout = const Duration(seconds: 5),
    String? filterServiceUuid,
  }) async {
    final allDevices = <String, ScanResult>{};

    // Use unfiltered scan and filter manually — Android's ScanFilter
    // does not reliably match 128-bit service UUIDs on all devices.
    final sub = FlutterBluePlus.onScanResults.listen((results) {
      for (final r in results) {
        allDevices[r.device.remoteId.str] = r;
      }
    });

    await FlutterBluePlus.startScan(timeout: timeout);
    await FlutterBluePlus.isScanning.firstWhere((s) => !s);
    sub.cancel();

    debugPrint('[BLE] Scan complete: ${allDevices.length} device(s) found');
    for (final r in allDevices.values) {
      final name = r.device.platformName.isNotEmpty
          ? r.device.platformName
          : r.advertisementData.advName;
      debugPrint(
          '[BLE]   ${r.device.remoteId.str} "$name" services=${r.advertisementData.serviceUuids.map((u) => u.str).toList()}');
    }

    final targetGuid = Guid(filterServiceUuid ?? serviceUuid);
    final devices = <ScannedDevice>[];
    for (final r in allDevices.values) {
      final hasService =
          r.advertisementData.serviceUuids.any((u) => u == targetGuid);
      if (!hasService) continue;
      devices.add(ScannedDevice(
        device: r.device,
        name: r.device.platformName.isNotEmpty
            ? r.device.platformName
            : r.advertisementData.advName.isNotEmpty
                ? r.advertisementData.advName
                : null,
        address: r.device.remoteId.str,
      ));
    }
    return devices;
  }

  Future<void> connect(ScannedDevice scannedDevice) async {
    _device = scannedDevice.device;
    await _device!.connect(autoConnect: false, mtu: null);

    // Request MTU (Android only — iOS negotiates MTU automatically)
    if (Platform.isAndroid) {
      _mtu = await _device!.requestMtu(247);
    } else {
      _mtu = _device!.mtuNow;
    }

    // Discover services
    final services = await _device!.discoverServices();
    BluetoothCharacteristic? targetChar;
    for (final s in services) {
      if (s.uuid == Guid(serviceUuid)) {
        for (final c in s.characteristics) {
          if (c.uuid == Guid(charUuid)) {
            targetChar = c;
            break;
          }
        }
      }
    }
    if (targetChar == null) {
      throw StateError('blerpc characteristic not found');
    }
    _char = targetChar;

    // Enable notifications — use a queue so that events arriving between
    // consecutive readNotify() calls are buffered instead of lost.
    _notifyQueue.clear();
    _notifyWaiter = null;
    await _char!.setNotifyValue(true);
    _notifySub = _char!.onValueReceived.listen((value) {
      final data = Uint8List.fromList(value);
      if (_notifyWaiter != null && !_notifyWaiter!.isCompleted) {
        _notifyWaiter!.complete(data);
        _notifyWaiter = null;
      } else {
        _notifyQueue.add(data);
      }
    });
  }

  Future<void> write(Uint8List data) async {
    if (_char == null) throw StateError('Not connected');
    await _char!.write(data, withoutResponse: true);
  }

  Future<Uint8List> readNotify({Duration? timeout}) async {
    if (_char == null) throw StateError('Not connected');
    if (_notifyQueue.isNotEmpty) {
      return _notifyQueue.removeFirst();
    }
    final waiter = Completer<Uint8List>();
    _notifyWaiter = waiter;
    timeout ??= const Duration(seconds: 2);
    try {
      return await waiter.future.timeout(timeout);
    } on TimeoutException {
      _notifyWaiter = null;
      rethrow;
    }
  }

  /// Drain any pending notifications.
  Future<void> drainNotifications() async {
    _notifyQueue.clear();
    // Also wait briefly for any in-flight notifications
    try {
      while (true) {
        await readNotify(timeout: const Duration(milliseconds: 100));
      }
    } on TimeoutException {
      // Done draining
    }
  }

  void disconnect() {
    _notifySub?.cancel();
    _notifySub = null;
    _notifyQueue.clear();
    _notifyWaiter = null;
    _device?.disconnect();
    _device = null;
    _char = null;
  }
}
