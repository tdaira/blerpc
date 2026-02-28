import 'dart:async';

import '../ble/ble_transport.dart';
import '../client/blerpc_client.dart';
import '../proto/blerpc.pb.dart';

typedef LogCallback = void Function(String message);

class TestRunner {
  final LogCallback _log;
  bool _running = false;
  int _passCount = 0;
  int _failCount = 0;

  TestRunner({required LogCallback log}) : _log = log;

  bool get isRunning => _running;

  Future<void> runAll({
    int iterations = 1,
    ScannedDevice? device,
  }) async {
    if (_running) return;
    _running = true;
    _passCount = 0;
    _failCount = 0;

    final client = BlerpcClient();
    try {
      final ScannedDevice target;
      if (device != null) {
        target = device;
      } else {
        _log('Scanning for bleRPC peripherals...');
        final devices = await client.scan();
        if (devices.isEmpty) {
          _log('[ERROR] No bleRPC devices found');
          _running = false;
          return;
        }
        target = devices.first;
      }
      _log('Connecting to ${target.name ?? target.address}...');
      await client.connect(target);
      _log('Connected. MTU=${client.mtu}, encrypted=${client.isEncrypted}');

      for (var iter = 1; iter <= iterations; iter++) {
        if (iterations > 1) _log('--- Iteration $iter/$iterations ---');

        await _runTest(client, 'echo_basic', () async {
          final resp = await client.echo(message: 'hello');
          _check(resp.message == 'hello',
              "Expected 'hello', got '${resp.message}'");
        });

        await _runTest(client, 'echo_empty', () async {
          final resp = await client.echo(message: '');
          _check(resp.message == '', "Expected empty, got '${resp.message}'");
        });

        await _runTest(client, 'flash_read_basic', () async {
          final resp = await client.flashRead(address: 0, length: 64);
          _check(resp.data.length == 64,
              'Expected 64 bytes, got ${resp.data.length}');
        });

        await _runTest(client, 'flash_read_8kb', () async {
          final resp = await client.flashRead(address: 0, length: 8192);
          _check(resp.data.length == 8192,
              'Expected 8192 bytes, got ${resp.data.length}');
        });

        await _runTest(client, 'data_write', () async {
          final testData = List.generate(64, (i) => i);
          final resp = await client.dataWrite(data: testData);
          _check(resp.length == 64, 'Expected length 64, got ${resp.length}');
        });

        await _runTest(client, 'counter_stream', () async {
          final results = await client.counterStream(count: 5);
          _check(
              results.length == 5, 'Expected 5 results, got ${results.length}');
          for (var i = 0; i < 5; i++) {
            _check(
                results[i].seq == i, 'Expected seq=$i, got ${results[i].seq}');
            _check(results[i].value == i * 10,
                'Expected value=${i * 10}, got ${results[i].value}');
          }
        });

        await _runTest(client, 'counter_upload', () async {
          final messages = List.generate(5, (i) {
            return CounterUploadRequest()
              ..seq = i
              ..value = i * 10;
          });
          final resp = await client.counterUpload(messages);
          _check(resp.receivedCount == 5,
              'Expected received_count=5, got ${resp.receivedCount}');
        });
      }

      _log(
          '=== Functional: $_passCount passed, $_failCount failed ($iterations iterations) ===');

      // Throughput benchmarks
      _log('');
      _log('=== Throughput Benchmarks ===');
      await _benchFlashReadThroughput(client);
      await _benchFlashReadOverhead(client);
      await _benchEchoRoundtrip(client);
      await _benchDataWriteThroughput(client);
      await _benchStreamThroughput(client);
    } catch (e) {
      _log('[ERROR] $e');
    } finally {
      client.disconnect();
      _running = false;
    }
  }

  Future<void> _benchFlashReadThroughput(BlerpcClient client) async {
    const readSize = 8192;
    const count = 10;
    const totalBytes = readSize * count;

    // Warmup
    await client.flashRead(address: 0, length: readSize);

    final sw = Stopwatch()..start();
    for (var i = 0; i < count; i++) {
      final resp = await client.flashRead(address: 0, length: readSize);
      _check(resp.data.length == readSize, 'flash_read size mismatch');
    }
    final elapsedMs = sw.elapsedMilliseconds;
    final kbPerSec = totalBytes / 1024.0 / (elapsedMs / 1000.0);
    final msPerCall = elapsedMs / count;
    _log(
        '[BENCH] flash_read_throughput: ${kbPerSec.toStringAsFixed(1)} KB/s ($totalBytes bytes in ${elapsedMs}ms, ${msPerCall.toStringAsFixed(1)} ms/call)');
  }

  Future<void> _benchFlashReadOverhead(BlerpcClient client) async {
    const count = 20;

    await client.flashRead(address: 0, length: 1);

    final sw = Stopwatch()..start();
    for (var i = 0; i < count; i++) {
      await client.flashRead(address: 0, length: 1);
    }
    final elapsedMs = sw.elapsedMilliseconds;
    final msPerCall = elapsedMs / count;
    _log(
        '[BENCH] flash_read_overhead: ${msPerCall.toStringAsFixed(1)} ms/call (1 byte x $count calls in ${elapsedMs}ms)');
  }

  Future<void> _benchEchoRoundtrip(BlerpcClient client) async {
    const count = 50;

    await client.echo(message: 'x');

    final sw = Stopwatch()..start();
    for (var i = 0; i < count; i++) {
      await client.echo(message: 'hello');
    }
    final elapsedMs = sw.elapsedMilliseconds;
    final msPerCall = elapsedMs / count;
    _log(
        '[BENCH] echo_roundtrip: ${msPerCall.toStringAsFixed(1)} ms/call ($count calls in ${elapsedMs}ms)');
  }

  Future<void> _benchDataWriteThroughput(BlerpcClient client) async {
    const writeSize = 200;
    const count = 20;
    const totalBytes = writeSize * count;
    final testData = List.generate(writeSize, (i) => i % 256);

    await client.dataWrite(data: testData);

    final sw = Stopwatch()..start();
    for (var i = 0; i < count; i++) {
      await client.dataWrite(data: testData);
    }
    final elapsedMs = sw.elapsedMilliseconds;
    final kbPerSec = totalBytes / 1024.0 / (elapsedMs / 1000.0);
    final msPerCall = elapsedMs / count;
    _log(
        '[BENCH] data_write_throughput: ${kbPerSec.toStringAsFixed(1)} KB/s ($totalBytes bytes in ${elapsedMs}ms, ${msPerCall.toStringAsFixed(1)} ms/call)');
  }

  Future<void> _benchStreamThroughput(BlerpcClient client) async {
    const count = 20;

    final sw1 = Stopwatch()..start();
    final results = await client.counterStream(count: count);
    final elapsed1 = sw1.elapsedMilliseconds;
    _check(results.length == count, 'stream count mismatch');
    _log(
        '[BENCH] counter_stream (P->C): $count items in ${elapsed1}ms (${(elapsed1 / count).toStringAsFixed(1)} ms/item)');

    final messages = List.generate(count, (i) {
      return CounterUploadRequest()
        ..seq = i
        ..value = i * 10;
    });
    final sw2 = Stopwatch()..start();
    final resp = await client.counterUpload(messages);
    final elapsed2 = sw2.elapsedMilliseconds;
    _check(resp.receivedCount == count, 'upload count mismatch');
    _log(
        '[BENCH] counter_upload (C->P): $count items in ${elapsed2}ms (${(elapsed2 / count).toStringAsFixed(1)} ms/item)');
  }

  Future<void> _runTest(
      BlerpcClient client, String name, Future<void> Function() block) async {
    try {
      await block();
      _passCount++;
      _log('[PASS] $name');
    } catch (e) {
      _failCount++;
      _log('[FAIL] $name: $e');
      await Future<void>.delayed(const Duration(milliseconds: 500));
      await client.transport.drainNotifications();
    }
  }

  void _check(bool condition, String message) {
    if (!condition) throw StateError(message);
  }
}
