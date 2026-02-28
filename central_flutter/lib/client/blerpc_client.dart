import 'dart:async';
import 'dart:developer' as dev;
import 'dart:typed_data';

import 'package:blerpc_central/proto/blerpc.pb.dart';
import 'package:blerpc_protocol/blerpc_protocol.dart';

import '../ble/ble_transport.dart';
import 'generated_client.dart';

class PayloadTooLargeError implements Exception {
  final int actual;
  final int limit;
  PayloadTooLargeError(this.actual, this.limit);
  @override
  String toString() =>
      'PayloadTooLargeError: Request payload ($actual bytes) exceeds peripheral limit ($limit bytes)';
}

class ResponseTooLargeError implements Exception {
  final String message;
  ResponseTooLargeError(this.message);
  @override
  String toString() => 'ResponseTooLargeError: $message';
}

class PeripheralErrorException implements Exception {
  final int errorCode;
  PeripheralErrorException(this.errorCode);
  @override
  String toString() =>
      'PeripheralErrorException: 0x${errorCode.toRadixString(16).padLeft(2, '0')}';
}

class ProtocolException implements Exception {
  final String message;
  ProtocolException(this.message);
  @override
  String toString() => 'ProtocolException: $message';
}

class BlerpcClient with GeneratedClientMixin {
  final BleTransport transport = BleTransport();
  final bool requireEncryption;

  ContainerSplitter? _splitter;
  final _assembler = ContainerAssembler();
  Duration _timeout = const Duration(milliseconds: 100);
  int? _maxRequestPayloadSize;

  // Encryption state
  BlerpcCryptoSession? _session;

  BlerpcClient({this.requireEncryption = true});

  int get mtu => transport.mtu;
  bool get isEncrypted => _session != null;

  Duration _readTimeout(bool firstRead) {
    if (!firstRead) return _timeout;
    return _timeout > const Duration(seconds: 2)
        ? _timeout
        : const Duration(seconds: 2);
  }

  void _handleControlError(Container container) {
    if (container.controlCmd == ControlCmd.error &&
        container.payload.isNotEmpty) {
      final errorCode = container.payload[0];
      if (errorCode == blerpcErrorResponseTooLarge) {
        throw ResponseTooLargeError(
            "Response exceeds peripheral's max_response_payload_size");
      }
      throw PeripheralErrorException(errorCode);
    }
  }

  Future<List<ScannedDevice>> scan({Duration? timeout}) async {
    return transport.scan(timeout: timeout ?? const Duration(seconds: 5));
  }

  Future<void> connect(ScannedDevice device) async {
    await transport.connect(device);
    _splitter = ContainerSplitter(mtu: transport.mtu);

    try {
      await _requestTimeout();
    } on TimeoutException {
      dev.log('Peripheral did not respond to timeout request, using default');
    }
    try {
      await _requestCapabilities();
    } on TimeoutException {
      dev.log('Peripheral did not respond to capabilities request');
    }

    if (requireEncryption && _session == null) {
      throw StateError(
        'Encryption required but key exchange was not completed. '
        'The peripheral may not support encryption or a MitM may '
        'have stripped the encryption capability flag.',
      );
    }
  }

  Future<void> _requestTimeout() async {
    final s = _splitter!;
    final tid = s.nextTransactionId();
    final req = makeTimeoutRequest(tid);
    await transport.write(req.serialize());
    final data =
        await transport.readNotify(timeout: const Duration(seconds: 1));
    final resp = Container.deserialize(data);
    if (resp.containerType == ContainerType.control &&
        resp.controlCmd == ControlCmd.timeout &&
        resp.payload.length == 2) {
      final bd = ByteData.sublistView(resp.payload);
      final ms = bd.getUint16(0, Endian.little);
      _timeout = Duration(milliseconds: ms);
      dev.log('Peripheral timeout: ${ms}ms');
    }
  }

  Future<void> _requestCapabilities() async {
    final s = _splitter!;
    final tid = s.nextTransactionId();
    final req = makeCapabilitiesRequest(tid);
    await transport.write(req.serialize());
    final data =
        await transport.readNotify(timeout: const Duration(seconds: 1));
    final resp = Container.deserialize(data);
    if (resp.containerType == ContainerType.control &&
        resp.controlCmd == ControlCmd.capabilities &&
        resp.payload.length >= 6) {
      final bd = ByteData.sublistView(resp.payload);
      final maxReq = bd.getUint16(0, Endian.little);
      final maxResp = bd.getUint16(2, Endian.little);
      final flags = bd.getUint16(4, Endian.little);
      _maxRequestPayloadSize = maxReq;
      dev.log(
          'Capabilities: max_req=$maxReq, max_resp=$maxResp, flags=0x${flags.toRadixString(16).padLeft(4, '0')}');

      if (flags & capabilityFlagEncryptionSupported != 0) {
        await _performKeyExchange();
      }
    }
  }

  Future<void> _performKeyExchange() async {
    final s = _splitter!;

    try {
      _session = await centralPerformKeyExchange(
        send: (payload) async {
          final tid = s.nextTransactionId();
          final req = makeKeyExchange(tid, payload);
          await transport.write(req.serialize());
        },
        receive: () async {
          final data =
              await transport.readNotify(timeout: const Duration(seconds: 2));
          final resp = Container.deserialize(data);
          if (resp.containerType != ContainerType.control ||
              resp.controlCmd != ControlCmd.keyExchange) {
            throw StateError(
                'Expected KEY_EXCHANGE response, got something else');
          }
          return resp.payload;
        },
      );
      dev.log('E2E encryption established');
    } catch (e) {
      dev.log('Key exchange failed: $e');
      if (requireEncryption) rethrow;
    }
  }

  Future<Uint8List> _encryptPayload(Uint8List payload) async {
    if (_session == null) {
      if (requireEncryption) {
        throw StateError('Encryption required but no session established');
      }
      return payload;
    }
    return _session!.encrypt(payload);
  }

  Future<Uint8List> _decryptPayload(Uint8List payload) async {
    if (_session == null) {
      if (requireEncryption) {
        throw StateError('Encryption required but no session established');
      }
      return payload;
    }
    return _session!.decrypt(payload);
  }

  @override
  Future<Uint8List> call(String cmdName, Uint8List requestData) async {
    final s = _splitter ?? (throw StateError('Not connected'));

    final cmd = CommandPacket(
      cmdType: CommandType.request,
      cmdName: cmdName,
      data: requestData,
    );
    final payload = cmd.serialize();

    if (_maxRequestPayloadSize != null &&
        payload.length > _maxRequestPayloadSize!) {
      throw PayloadTooLargeError(payload.length, _maxRequestPayloadSize!);
    }

    final sendPayload = await _encryptPayload(payload);
    final containers = s.split(sendPayload);
    for (final c in containers) {
      await transport.write(c.serialize());
    }

    _assembler.reset();
    var firstRead = true;
    while (true) {
      final notifyData =
          await transport.readNotify(timeout: _readTimeout(firstRead));
      firstRead = false;
      final container = Container.deserialize(notifyData);

      if (container.containerType == ContainerType.control) {
        _handleControlError(container);
        continue;
      }

      final result = _assembler.feed(container);
      if (result != null) {
        final decrypted = await _decryptPayload(result);
        final resp = CommandPacket.deserialize(decrypted);
        if (resp.cmdType != CommandType.response) {
          throw ProtocolException(
              'Expected response, got type=${resp.cmdType}');
        }
        if (resp.cmdName != cmdName) {
          throw ProtocolException(
              "Command name mismatch: expected '$cmdName', got '${resp.cmdName}'");
        }
        return resp.data;
      }
    }
  }

  @override
  Future<List<Uint8List>> streamReceive(
      String cmdName, Uint8List requestData) async {
    final s = _splitter ?? (throw StateError('Not connected'));

    final cmd = CommandPacket(
      cmdType: CommandType.request,
      cmdName: cmdName,
      data: requestData,
    );
    final payload = cmd.serialize();

    if (_maxRequestPayloadSize != null &&
        payload.length > _maxRequestPayloadSize!) {
      throw PayloadTooLargeError(payload.length, _maxRequestPayloadSize!);
    }

    final sendPayload = await _encryptPayload(payload);
    final containers = s.split(sendPayload);
    for (final c in containers) {
      await transport.write(c.serialize());
    }

    final results = <Uint8List>[];
    _assembler.reset();
    var firstRead = true;
    while (true) {
      final notifyData =
          await transport.readNotify(timeout: _readTimeout(firstRead));
      firstRead = false;
      final container = Container.deserialize(notifyData);

      if (container.containerType == ContainerType.control) {
        if (container.controlCmd == ControlCmd.streamEndP2C) break;
        _handleControlError(container);
        continue;
      }

      final result = _assembler.feed(container);
      if (result != null) {
        final decrypted = await _decryptPayload(result);
        final resp = CommandPacket.deserialize(decrypted);
        if (resp.cmdType != CommandType.response) {
          throw ProtocolException(
              'Expected response, got type=${resp.cmdType}');
        }
        results.add(resp.data);
      }
    }
    return results;
  }

  @override
  Future<Uint8List> streamSend(
      String cmdName, List<Uint8List> messages, String finalCmdName) async {
    final s = _splitter ?? (throw StateError('Not connected'));

    for (final msgData in messages) {
      final cmd = CommandPacket(
        cmdType: CommandType.request,
        cmdName: cmdName,
        data: msgData,
      );
      final payload = cmd.serialize();
      final sendPayload = await _encryptPayload(payload);
      final containers = s.split(sendPayload);
      for (final c in containers) {
        await transport.write(c.serialize());
      }
    }

    // Send STREAM_END_C2P
    final tid = s.nextTransactionId();
    final streamEnd = makeStreamEndC2P(tid);
    await transport.write(streamEnd.serialize());

    // Wait for final response
    _assembler.reset();
    var firstRead = true;
    while (true) {
      final notifyData =
          await transport.readNotify(timeout: _readTimeout(firstRead));
      firstRead = false;
      final container = Container.deserialize(notifyData);

      if (container.containerType == ContainerType.control) {
        _handleControlError(container);
        continue;
      }

      final result = _assembler.feed(container);
      if (result != null) {
        final decrypted = await _decryptPayload(result);
        final resp = CommandPacket.deserialize(decrypted);
        if (resp.cmdType != CommandType.response) {
          throw ProtocolException(
              'Expected response, got type=${resp.cmdType}');
        }
        if (resp.cmdName != finalCmdName) {
          throw ProtocolException(
              "Command name mismatch: expected '$finalCmdName', got '${resp.cmdName}'");
        }
        return resp.data;
      }
    }
  }

  // Streaming convenience methods (manually implemented, not auto-generated)

  Future<List<(int, int)>> counterStreamAll(int count) async {
    final req = CounterStreamRequest()..count = count;
    final responses = await streamReceive(
        'counter_stream', Uint8List.fromList(req.writeToBuffer()));
    return responses.map((data) {
      final resp = CounterStreamResponse.fromBuffer(data);
      return (resp.seq, resp.value);
    }).toList();
  }

  Future<CounterUploadResponse> counterUploadAll(int count) async {
    final messages = List.generate(count, (i) {
      final req = CounterUploadRequest()
        ..seq = i
        ..value = i * 10;
      return Uint8List.fromList(req.writeToBuffer());
    });
    final respData =
        await streamSend('counter_upload', messages, 'counter_upload');
    return CounterUploadResponse.fromBuffer(respData);
  }

  void disconnect() {
    transport.disconnect();
    _session = null;
    _splitter = null;
  }
}
