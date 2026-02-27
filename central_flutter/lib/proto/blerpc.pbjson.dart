// This is a generated file - do not edit.
//
// Generated from blerpc.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use echoRequestDescriptor instead')
const EchoRequest$json = {
  '1': 'EchoRequest',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `EchoRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List echoRequestDescriptor = $convert
    .base64Decode('CgtFY2hvUmVxdWVzdBIYCgdtZXNzYWdlGAEgASgJUgdtZXNzYWdl');

@$core.Deprecated('Use echoResponseDescriptor instead')
const EchoResponse$json = {
  '1': 'EchoResponse',
  '2': [
    {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `EchoResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List echoResponseDescriptor = $convert
    .base64Decode('CgxFY2hvUmVzcG9uc2USGAoHbWVzc2FnZRgBIAEoCVIHbWVzc2FnZQ==');

@$core.Deprecated('Use flashReadRequestDescriptor instead')
const FlashReadRequest$json = {
  '1': 'FlashReadRequest',
  '2': [
    {'1': 'address', '3': 1, '4': 1, '5': 13, '10': 'address'},
    {'1': 'length', '3': 2, '4': 1, '5': 13, '10': 'length'},
  ],
};

/// Descriptor for `FlashReadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List flashReadRequestDescriptor = $convert.base64Decode(
    'ChBGbGFzaFJlYWRSZXF1ZXN0EhgKB2FkZHJlc3MYASABKA1SB2FkZHJlc3MSFgoGbGVuZ3RoGA'
    'IgASgNUgZsZW5ndGg=');

@$core.Deprecated('Use flashReadResponseDescriptor instead')
const FlashReadResponse$json = {
  '1': 'FlashReadResponse',
  '2': [
    {'1': 'address', '3': 1, '4': 1, '5': 13, '10': 'address'},
    {'1': 'data', '3': 2, '4': 1, '5': 12, '10': 'data'},
  ],
};

/// Descriptor for `FlashReadResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List flashReadResponseDescriptor = $convert.base64Decode(
    'ChFGbGFzaFJlYWRSZXNwb25zZRIYCgdhZGRyZXNzGAEgASgNUgdhZGRyZXNzEhIKBGRhdGEYAi'
    'ABKAxSBGRhdGE=');

@$core.Deprecated('Use dataWriteRequestDescriptor instead')
const DataWriteRequest$json = {
  '1': 'DataWriteRequest',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 12, '10': 'data'},
  ],
};

/// Descriptor for `DataWriteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List dataWriteRequestDescriptor = $convert
    .base64Decode('ChBEYXRhV3JpdGVSZXF1ZXN0EhIKBGRhdGEYASABKAxSBGRhdGE=');

@$core.Deprecated('Use dataWriteResponseDescriptor instead')
const DataWriteResponse$json = {
  '1': 'DataWriteResponse',
  '2': [
    {'1': 'length', '3': 1, '4': 1, '5': 13, '10': 'length'},
  ],
};

/// Descriptor for `DataWriteResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List dataWriteResponseDescriptor = $convert.base64Decode(
    'ChFEYXRhV3JpdGVSZXNwb25zZRIWCgZsZW5ndGgYASABKA1SBmxlbmd0aA==');

@$core.Deprecated('Use counterStreamRequestDescriptor instead')
const CounterStreamRequest$json = {
  '1': 'CounterStreamRequest',
  '2': [
    {'1': 'count', '3': 1, '4': 1, '5': 13, '10': 'count'},
  ],
};

/// Descriptor for `CounterStreamRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List counterStreamRequestDescriptor =
    $convert.base64Decode(
        'ChRDb3VudGVyU3RyZWFtUmVxdWVzdBIUCgVjb3VudBgBIAEoDVIFY291bnQ=');

@$core.Deprecated('Use counterStreamResponseDescriptor instead')
const CounterStreamResponse$json = {
  '1': 'CounterStreamResponse',
  '2': [
    {'1': 'seq', '3': 1, '4': 1, '5': 13, '10': 'seq'},
    {'1': 'value', '3': 2, '4': 1, '5': 5, '10': 'value'},
  ],
};

/// Descriptor for `CounterStreamResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List counterStreamResponseDescriptor = $convert.base64Decode(
    'ChVDb3VudGVyU3RyZWFtUmVzcG9uc2USEAoDc2VxGAEgASgNUgNzZXESFAoFdmFsdWUYAiABKA'
    'VSBXZhbHVl');

@$core.Deprecated('Use counterUploadRequestDescriptor instead')
const CounterUploadRequest$json = {
  '1': 'CounterUploadRequest',
  '2': [
    {'1': 'seq', '3': 1, '4': 1, '5': 13, '10': 'seq'},
    {'1': 'value', '3': 2, '4': 1, '5': 5, '10': 'value'},
  ],
};

/// Descriptor for `CounterUploadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List counterUploadRequestDescriptor = $convert.base64Decode(
    'ChRDb3VudGVyVXBsb2FkUmVxdWVzdBIQCgNzZXEYASABKA1SA3NlcRIUCgV2YWx1ZRgCIAEoBV'
    'IFdmFsdWU=');

@$core.Deprecated('Use counterUploadResponseDescriptor instead')
const CounterUploadResponse$json = {
  '1': 'CounterUploadResponse',
  '2': [
    {'1': 'received_count', '3': 1, '4': 1, '5': 13, '10': 'receivedCount'},
  ],
};

/// Descriptor for `CounterUploadResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List counterUploadResponseDescriptor = $convert.base64Decode(
    'ChVDb3VudGVyVXBsb2FkUmVzcG9uc2USJQoOcmVjZWl2ZWRfY291bnQYASABKA1SDXJlY2Vpdm'
    'VkQ291bnQ=');
