import 'dart:typed_data';

import 'package:blerpc_central/proto/blerpc.pb.dart';

/// Typed RPC method wrappers for blerpc commands.
mixin GeneratedClientMixin {
  Future<Uint8List> call(String cmdName, Uint8List requestData);
  Future<List<Uint8List>> streamReceive(String cmdName, Uint8List requestData);
  Future<Uint8List> streamSend(
      String cmdName, List<Uint8List> messages, String finalCmdName);

  Future<EchoResponse> echo({required String message}) async {
    final req = EchoRequest()..message = message;
    final data = await call('echo', Uint8List.fromList(req.writeToBuffer()));
    return EchoResponse.fromBuffer(data);
  }

  Future<FlashReadResponse> flashRead({
    required int address,
    required int length,
  }) async {
    final req = FlashReadRequest()
      ..address = address
      ..length = length;
    final data =
        await call('flash_read', Uint8List.fromList(req.writeToBuffer()));
    return FlashReadResponse.fromBuffer(data);
  }

  Future<DataWriteResponse> dataWrite({required List<int> data}) async {
    final req = DataWriteRequest()..data = data;
    final respData =
        await call('data_write', Uint8List.fromList(req.writeToBuffer()));
    return DataWriteResponse.fromBuffer(respData);
  }

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
}
