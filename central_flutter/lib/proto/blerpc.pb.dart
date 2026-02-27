// This is a generated file - do not edit.
//
// Generated from blerpc.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// Echo — loopback test. Returns the same message string.
class EchoRequest extends $pb.GeneratedMessage {
  factory EchoRequest({
    $core.String? message,
  }) {
    final result = create();
    if (message != null) result.message = message;
    return result;
  }

  EchoRequest._();

  factory EchoRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory EchoRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EchoRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EchoRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EchoRequest copyWith(void Function(EchoRequest) updates) =>
      super.copyWith((message) => updates(message as EchoRequest))
          as EchoRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EchoRequest create() => EchoRequest._();
  @$core.override
  EchoRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static EchoRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EchoRequest>(create);
  static EchoRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);
}

class EchoResponse extends $pb.GeneratedMessage {
  factory EchoResponse({
    $core.String? message,
  }) {
    final result = create();
    if (message != null) result.message = message;
    return result;
  }

  EchoResponse._();

  factory EchoResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory EchoResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EchoResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EchoResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EchoResponse copyWith(void Function(EchoResponse) updates) =>
      super.copyWith((message) => updates(message as EchoResponse))
          as EchoResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EchoResponse create() => EchoResponse._();
  @$core.override
  EchoResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static EchoResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EchoResponse>(create);
  static EchoResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get message => $_getSZ(0);
  @$pb.TagNumber(1)
  set message($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMessage() => $_has(0);
  @$pb.TagNumber(1)
  void clearMessage() => $_clearField(1);
}

/// FlashRead — read raw bytes from peripheral flash.
/// The peripheral returns data starting at the given address.
class FlashReadRequest extends $pb.GeneratedMessage {
  factory FlashReadRequest({
    $core.int? address,
    $core.int? length,
  }) {
    final result = create();
    if (address != null) result.address = address;
    if (length != null) result.length = length;
    return result;
  }

  FlashReadRequest._();

  factory FlashReadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory FlashReadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'FlashReadRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'address', fieldType: $pb.PbFieldType.OU3)
    ..aI(2, _omitFieldNames ? '' : 'length', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FlashReadRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FlashReadRequest copyWith(void Function(FlashReadRequest) updates) =>
      super.copyWith((message) => updates(message as FlashReadRequest))
          as FlashReadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static FlashReadRequest create() => FlashReadRequest._();
  @$core.override
  FlashReadRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static FlashReadRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<FlashReadRequest>(create);
  static FlashReadRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get address => $_getIZ(0);
  @$pb.TagNumber(1)
  set address($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get length => $_getIZ(1);
  @$pb.TagNumber(2)
  set length($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLength() => $_has(1);
  @$pb.TagNumber(2)
  void clearLength() => $_clearField(2);
}

class FlashReadResponse extends $pb.GeneratedMessage {
  factory FlashReadResponse({
    $core.int? address,
    $core.List<$core.int>? data,
  }) {
    final result = create();
    if (address != null) result.address = address;
    if (data != null) result.data = data;
    return result;
  }

  FlashReadResponse._();

  factory FlashReadResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory FlashReadResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'FlashReadResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'address', fieldType: $pb.PbFieldType.OU3)
    ..a<$core.List<$core.int>>(
        2, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FlashReadResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FlashReadResponse copyWith(void Function(FlashReadResponse) updates) =>
      super.copyWith((message) => updates(message as FlashReadResponse))
          as FlashReadResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static FlashReadResponse create() => FlashReadResponse._();
  @$core.override
  FlashReadResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static FlashReadResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<FlashReadResponse>(create);
  static FlashReadResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get address => $_getIZ(0);
  @$pb.TagNumber(1)
  set address($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get data => $_getN(1);
  @$pb.TagNumber(2)
  set data($core.List<$core.int> value) => $_setBytes(1, value);
  @$pb.TagNumber(2)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(2)
  void clearData() => $_clearField(2);
}

/// DataWrite — write raw bytes to peripheral (sink test).
/// The peripheral acknowledges with the number of bytes received.
class DataWriteRequest extends $pb.GeneratedMessage {
  factory DataWriteRequest({
    $core.List<$core.int>? data,
  }) {
    final result = create();
    if (data != null) result.data = data;
    return result;
  }

  DataWriteRequest._();

  factory DataWriteRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DataWriteRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DataWriteRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..a<$core.List<$core.int>>(
        1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DataWriteRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DataWriteRequest copyWith(void Function(DataWriteRequest) updates) =>
      super.copyWith((message) => updates(message as DataWriteRequest))
          as DataWriteRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DataWriteRequest create() => DataWriteRequest._();
  @$core.override
  DataWriteRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static DataWriteRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DataWriteRequest>(create);
  static DataWriteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get data => $_getN(0);
  @$pb.TagNumber(1)
  set data($core.List<$core.int> value) => $_setBytes(0, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
}

class DataWriteResponse extends $pb.GeneratedMessage {
  factory DataWriteResponse({
    $core.int? length,
  }) {
    final result = create();
    if (length != null) result.length = length;
    return result;
  }

  DataWriteResponse._();

  factory DataWriteResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DataWriteResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DataWriteResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'length', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DataWriteResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DataWriteResponse copyWith(void Function(DataWriteResponse) updates) =>
      super.copyWith((message) => updates(message as DataWriteResponse))
          as DataWriteResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DataWriteResponse create() => DataWriteResponse._();
  @$core.override
  DataWriteResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static DataWriteResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DataWriteResponse>(create);
  static DataWriteResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get length => $_getIZ(0);
  @$pb.TagNumber(1)
  set length($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLength() => $_has(0);
  @$pb.TagNumber(1)
  void clearLength() => $_clearField(1);
}

/// CounterStream (P→C stream) — peripheral sends `count` responses,
/// each with an incrementing seq and value = seq * 10.
class CounterStreamRequest extends $pb.GeneratedMessage {
  factory CounterStreamRequest({
    $core.int? count,
  }) {
    final result = create();
    if (count != null) result.count = count;
    return result;
  }

  CounterStreamRequest._();

  factory CounterStreamRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CounterStreamRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CounterStreamRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'count', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CounterStreamRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CounterStreamRequest copyWith(void Function(CounterStreamRequest) updates) =>
      super.copyWith((message) => updates(message as CounterStreamRequest))
          as CounterStreamRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CounterStreamRequest create() => CounterStreamRequest._();
  @$core.override
  CounterStreamRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CounterStreamRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CounterStreamRequest>(create);
  static CounterStreamRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get count => $_getIZ(0);
  @$pb.TagNumber(1)
  set count($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCount() => $_has(0);
  @$pb.TagNumber(1)
  void clearCount() => $_clearField(1);
}

class CounterStreamResponse extends $pb.GeneratedMessage {
  factory CounterStreamResponse({
    $core.int? seq,
    $core.int? value,
  }) {
    final result = create();
    if (seq != null) result.seq = seq;
    if (value != null) result.value = value;
    return result;
  }

  CounterStreamResponse._();

  factory CounterStreamResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CounterStreamResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CounterStreamResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'seq', fieldType: $pb.PbFieldType.OU3)
    ..aI(2, _omitFieldNames ? '' : 'value')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CounterStreamResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CounterStreamResponse copyWith(
          void Function(CounterStreamResponse) updates) =>
      super.copyWith((message) => updates(message as CounterStreamResponse))
          as CounterStreamResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CounterStreamResponse create() => CounterStreamResponse._();
  @$core.override
  CounterStreamResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CounterStreamResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CounterStreamResponse>(create);
  static CounterStreamResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get seq => $_getIZ(0);
  @$pb.TagNumber(1)
  set seq($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSeq() => $_has(0);
  @$pb.TagNumber(1)
  void clearSeq() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get value => $_getIZ(1);
  @$pb.TagNumber(2)
  set value($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => $_clearField(2);
}

/// CounterUpload (C→P stream) — central sends `count` requests,
/// peripheral responds with the total received count.
class CounterUploadRequest extends $pb.GeneratedMessage {
  factory CounterUploadRequest({
    $core.int? seq,
    $core.int? value,
  }) {
    final result = create();
    if (seq != null) result.seq = seq;
    if (value != null) result.value = value;
    return result;
  }

  CounterUploadRequest._();

  factory CounterUploadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CounterUploadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CounterUploadRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'seq', fieldType: $pb.PbFieldType.OU3)
    ..aI(2, _omitFieldNames ? '' : 'value')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CounterUploadRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CounterUploadRequest copyWith(void Function(CounterUploadRequest) updates) =>
      super.copyWith((message) => updates(message as CounterUploadRequest))
          as CounterUploadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CounterUploadRequest create() => CounterUploadRequest._();
  @$core.override
  CounterUploadRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CounterUploadRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CounterUploadRequest>(create);
  static CounterUploadRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get seq => $_getIZ(0);
  @$pb.TagNumber(1)
  set seq($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSeq() => $_has(0);
  @$pb.TagNumber(1)
  void clearSeq() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get value => $_getIZ(1);
  @$pb.TagNumber(2)
  set value($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => $_clearField(2);
}

class CounterUploadResponse extends $pb.GeneratedMessage {
  factory CounterUploadResponse({
    $core.int? receivedCount,
  }) {
    final result = create();
    if (receivedCount != null) result.receivedCount = receivedCount;
    return result;
  }

  CounterUploadResponse._();

  factory CounterUploadResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CounterUploadResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CounterUploadResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'blerpc'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'receivedCount',
        fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CounterUploadResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CounterUploadResponse copyWith(
          void Function(CounterUploadResponse) updates) =>
      super.copyWith((message) => updates(message as CounterUploadResponse))
          as CounterUploadResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CounterUploadResponse create() => CounterUploadResponse._();
  @$core.override
  CounterUploadResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CounterUploadResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CounterUploadResponse>(create);
  static CounterUploadResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get receivedCount => $_getIZ(0);
  @$pb.TagNumber(1)
  set receivedCount($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasReceivedCount() => $_has(0);
  @$pb.TagNumber(1)
  void clearReceivedCount() => $_clearField(1);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
