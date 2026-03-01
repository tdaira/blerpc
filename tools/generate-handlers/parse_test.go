package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const echoProto = `syntax = "proto3";
package test;

message EchoRequest {
  string message = 1;
}

message EchoResponse {
  string message = 1;
}
`

const multiFieldProto = `syntax = "proto3";
package test;

message FlashReadRequest {
  uint32 address = 1;
  uint32 length = 2;
}

message FlashReadResponse {
  uint32 address = 1;
  bytes data = 2;
}
`

const noMatchProto = `syntax = "proto3";
package test;

message Foo {
  string name = 1;
}

message Bar {
  string name = 1;
}
`

const enumProto = `syntax = "proto3";
package test;

enum Status {
  STATUS_UNKNOWN = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}

message GetStatusRequest {
  string name = 1;
}

message GetStatusResponse {
  Status status = 1;
}
`

const nestedEnumProto = `syntax = "proto3";
package test;

message SetModeRequest {
  enum Mode {
    MODE_UNKNOWN = 0;
    MODE_FAST = 1;
    MODE_SLOW = 2;
  }
  Mode mode = 1;
}

message SetModeResponse {
  bool ok = 1;
}
`

const repeatedProto = `syntax = "proto3";
package test;

message BatchRequest {
  repeated string names = 1;
  repeated uint32 ids = 2;
}

message BatchResponse {
  repeated string results = 1;
}
`

const repeatedEnumProto = `syntax = "proto3";
package test;

enum Color {
  COLOR_UNKNOWN = 0;
  COLOR_RED = 1;
  COLOR_BLUE = 2;
}

message PaintRequest {
  repeated Color colors = 1;
}

message PaintResponse {
  bool ok = 1;
}
`

const messageFieldProto = `syntax = "proto3";
package test;

message Address {
  string street = 1;
  string city = 2;
}

message UpdateAddressRequest {
  string user_id = 1;
  Address address = 2;
}

message UpdateAddressResponse {
  bool ok = 1;
}
`

const nestedMessageProto = `syntax = "proto3";
package test;

message Outer {
  message Inner {
    string value = 1;
  }
}

message GetInnerRequest {
  string id = 1;
}

message GetInnerResponse {
  Outer.Inner result = 1;
}
`

const oneofProto = `syntax = "proto3";
package test;

message SearchRequest {
  oneof query {
    string text = 1;
    uint32 id = 2;
  }
}

message SearchResponse {
  string result = 1;
}
`

const mapProto = `syntax = "proto3";
package test;

message SetLabelsRequest {
  map<string, string> labels = 1;
  map<string, uint32> counts = 2;
}

message SetLabelsResponse {
  bool ok = 1;
}
`

func TestParseProtoReader_Echo(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(echoProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	msgs := pf.Messages
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Name != "EchoRequest" {
		t.Errorf("expected EchoRequest, got %s", msgs[0].Name)
	}
	if msgs[1].Name != "EchoResponse" {
		t.Errorf("expected EchoResponse, got %s", msgs[1].Name)
	}
	if len(msgs[0].Fields) != 1 {
		t.Fatalf("expected 1 field in EchoRequest, got %d", len(msgs[0].Fields))
	}
	f := msgs[0].Fields[0]
	if f.Type != "string" || f.Name != "message" || f.Number != 1 {
		t.Errorf("unexpected field: %+v", f)
	}
}

func TestParseProtoReader_MultiField(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(multiFieldProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	msgs := pf.Messages
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	req := msgs[0]
	if len(req.Fields) != 2 {
		t.Fatalf("expected 2 fields in FlashReadRequest, got %d", len(req.Fields))
	}
	if req.Fields[0].Type != "uint32" || req.Fields[0].Name != "address" {
		t.Errorf("unexpected field[0]: %+v", req.Fields[0])
	}
	if req.Fields[1].Type != "uint32" || req.Fields[1].Name != "length" {
		t.Errorf("unexpected field[1]: %+v", req.Fields[1])
	}
}

func TestParseProtoReader_Enum(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(enumProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	if len(pf.Enums) != 1 {
		t.Fatalf("expected 1 enum, got %d", len(pf.Enums))
	}
	e := pf.Enums[0]
	if e.Name != "Status" {
		t.Errorf("expected enum name Status, got %s", e.Name)
	}
	if len(e.Values) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(e.Values))
	}
	if e.Values[0].Name != "STATUS_UNKNOWN" || e.Values[0].Number != 0 {
		t.Errorf("unexpected enum value[0]: %+v", e.Values[0])
	}
	if e.Values[1].Name != "STATUS_ACTIVE" || e.Values[1].Number != 1 {
		t.Errorf("unexpected enum value[1]: %+v", e.Values[1])
	}

	// Check that the Status field is marked as enum
	msgs := pf.Messages
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	resp := msgs[1]
	if len(resp.Fields) != 1 {
		t.Fatalf("expected 1 field in GetStatusResponse, got %d", len(resp.Fields))
	}
	if !resp.Fields[0].IsEnum {
		t.Error("expected Status field to be marked as enum")
	}
	if resp.Fields[0].Type != "Status" {
		t.Errorf("expected field type Status, got %s", resp.Fields[0].Type)
	}
}

func TestParseProtoReader_NestedEnum(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(nestedEnumProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	if len(pf.Enums) != 1 {
		t.Fatalf("expected 1 enum, got %d", len(pf.Enums))
	}
	e := pf.Enums[0]
	if e.Name != "Mode" {
		t.Errorf("expected enum name Mode, got %s", e.Name)
	}
	if len(e.Values) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(e.Values))
	}

	// Check that the Mode field is marked as enum
	msgs := pf.Messages
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	req := msgs[0]
	if len(req.Fields) != 1 {
		t.Fatalf("expected 1 field in SetModeRequest, got %d", len(req.Fields))
	}
	if !req.Fields[0].IsEnum {
		t.Error("expected Mode field to be marked as enum")
	}
}

func TestDiscoverCommands_Echo(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(echoProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	cmds := discoverCommands(pf.Messages)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	if cmd.Camel != "Echo" {
		t.Errorf("expected Camel=Echo, got %s", cmd.Camel)
	}
	if cmd.Snake != "echo" {
		t.Errorf("expected Snake=echo, got %s", cmd.Snake)
	}
	if cmd.RequestMsg != "EchoRequest" {
		t.Errorf("expected RequestMsg=EchoRequest, got %s", cmd.RequestMsg)
	}
	if cmd.ResponseMsg != "EchoResponse" {
		t.Errorf("expected ResponseMsg=EchoResponse, got %s", cmd.ResponseMsg)
	}
}

func TestDiscoverCommands_NoMatch(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(noMatchProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	cmds := discoverCommands(pf.Messages)
	if len(cmds) != 0 {
		t.Fatalf("expected 0 commands, got %d", len(cmds))
	}
}

func TestDiscoverCommands_MultiField(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(multiFieldProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	cmds := discoverCommands(pf.Messages)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	if cmd.Camel != "FlashRead" {
		t.Errorf("expected Camel=FlashRead, got %s", cmd.Camel)
	}
	if cmd.Snake != "flash_read" {
		t.Errorf("expected Snake=flash_read, got %s", cmd.Snake)
	}
	if len(cmd.RequestFields) != 2 {
		t.Errorf("expected 2 request fields, got %d", len(cmd.RequestFields))
	}
	if len(cmd.ResponseFields) != 2 {
		t.Errorf("expected 2 response fields, got %d", len(cmd.ResponseFields))
	}
}

func TestDiscoverCommands_Enum(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(enumProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	cmds := discoverCommands(pf.Messages)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	if cmd.Camel != "GetStatus" {
		t.Errorf("expected Camel=GetStatus, got %s", cmd.Camel)
	}
	if len(cmd.ResponseFields) != 1 {
		t.Fatalf("expected 1 response field, got %d", len(cmd.ResponseFields))
	}
	if !cmd.ResponseFields[0].IsEnum {
		t.Error("expected response field to be marked as enum")
	}
}

func TestParseProtoReader_Repeated(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(repeatedProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	msgs := pf.Messages
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	req := msgs[0]
	if len(req.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(req.Fields))
	}
	if !req.Fields[0].IsRepeated {
		t.Error("expected names to be repeated")
	}
	if req.Fields[0].Type != "string" {
		t.Errorf("expected type string, got %s", req.Fields[0].Type)
	}
	if !req.Fields[1].IsRepeated {
		t.Error("expected ids to be repeated")
	}
}

func TestParseProtoReader_RepeatedEnum(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(repeatedEnumProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	msgs := pf.Messages
	req := msgs[0]
	if len(req.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(req.Fields))
	}
	f := req.Fields[0]
	if !f.IsRepeated {
		t.Error("expected colors to be repeated")
	}
	if !f.IsEnum {
		t.Error("expected colors to be an enum")
	}
}

func TestParseProtoReader_MessageField(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(messageFieldProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	msgs := pf.Messages
	// Should find Address, UpdateAddressRequest, UpdateAddressResponse
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	// UpdateAddressRequest is the second message
	req := msgs[1]
	if req.Name != "UpdateAddressRequest" {
		t.Fatalf("expected UpdateAddressRequest, got %s", req.Name)
	}
	if len(req.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(req.Fields))
	}
	// First field: user_id (string, not a message)
	if req.Fields[0].IsMessage {
		t.Error("expected user_id not to be a message type")
	}
	// Second field: address (Address, a message type)
	if !req.Fields[1].IsMessage {
		t.Error("expected address to be a message type")
	}
	if req.Fields[1].Type != "Address" {
		t.Errorf("expected type Address, got %s", req.Fields[1].Type)
	}
}

func TestDiscoverCommands_MessageField(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(messageFieldProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	cmds := discoverCommands(pf.Messages)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	if cmd.Camel != "UpdateAddress" {
		t.Errorf("expected Camel=UpdateAddress, got %s", cmd.Camel)
	}
	if len(cmd.RequestFields) != 2 {
		t.Fatalf("expected 2 request fields, got %d", len(cmd.RequestFields))
	}
	if !cmd.RequestFields[1].IsMessage {
		t.Error("expected second request field to be a message type")
	}
}

func TestParseProtoReader_Map(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(mapProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	msgs := pf.Messages
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	req := msgs[0]
	if len(req.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(req.Fields))
	}
	// First field: map<string, string> labels
	f0 := req.Fields[0]
	if !f0.IsMap {
		t.Error("expected labels to be a map")
	}
	if f0.KeyType != "string" {
		t.Errorf("expected KeyType=string, got %s", f0.KeyType)
	}
	if f0.ValueType != "string" {
		t.Errorf("expected ValueType=string, got %s", f0.ValueType)
	}
	if f0.Name != "labels" {
		t.Errorf("expected name=labels, got %s", f0.Name)
	}
	// Second field: map<string, uint32> counts
	f1 := req.Fields[1]
	if !f1.IsMap {
		t.Error("expected counts to be a map")
	}
	if f1.KeyType != "string" {
		t.Errorf("expected KeyType=string, got %s", f1.KeyType)
	}
	if f1.ValueType != "uint32" {
		t.Errorf("expected ValueType=uint32, got %s", f1.ValueType)
	}
}

func TestParseProtoReader_Oneof(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(oneofProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	msgs := pf.Messages
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	req := msgs[0]
	// Oneof fields should appear in the flat field list
	if len(req.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(req.Fields))
	}
	if req.Fields[0].Name != "text" || req.Fields[0].Type != "string" {
		t.Errorf("unexpected field[0]: %+v", req.Fields[0])
	}
	if req.Fields[1].Name != "id" || req.Fields[1].Type != "uint32" {
		t.Errorf("unexpected field[1]: %+v", req.Fields[1])
	}
	// Check oneof group
	if len(req.Oneofs) != 1 {
		t.Fatalf("expected 1 oneof group, got %d", len(req.Oneofs))
	}
	og := req.Oneofs[0]
	if og.Name != "query" {
		t.Errorf("expected oneof name query, got %s", og.Name)
	}
	if len(og.Fields) != 2 {
		t.Fatalf("expected 2 oneof fields, got %d", len(og.Fields))
	}
}

const serviceProto = `syntax = "proto3";
package test;

message EchoRequest {
  string message = 1;
}

message EchoResponse {
  string message = 1;
}

message CounterStreamRequest {
  uint32 start = 1;
}

message CounterStreamResponse {
  uint32 count = 1;
}

message CounterUploadRequest {
  uint32 value = 1;
}

message CounterUploadResponse {
  uint32 total = 1;
}

service TestService {
  rpc Echo(EchoRequest) returns (EchoResponse);
  rpc CounterStream(CounterStreamRequest) returns (stream CounterStreamResponse);
  rpc CounterUpload(stream CounterUploadRequest) returns (CounterUploadResponse);
}
`

func TestParseProtoReader_Service(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(serviceProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	if len(pf.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(pf.Services))
	}
	svc := pf.Services[0]
	if svc.Name != "TestService" {
		t.Errorf("expected service name TestService, got %s", svc.Name)
	}
	if len(svc.RPCs) != 3 {
		t.Fatalf("expected 3 RPCs, got %d", len(svc.RPCs))
	}

	// Echo: unary
	if svc.RPCs[0].Name != "Echo" {
		t.Errorf("expected RPC name Echo, got %s", svc.RPCs[0].Name)
	}
	if svc.RPCs[0].ClientStream || svc.RPCs[0].ServerStream {
		t.Error("Echo should not be streaming")
	}

	// CounterStream: server-side stream (p2c)
	if svc.RPCs[1].Name != "CounterStream" {
		t.Errorf("expected RPC name CounterStream, got %s", svc.RPCs[1].Name)
	}
	if svc.RPCs[1].ClientStream {
		t.Error("CounterStream should not have client stream")
	}
	if !svc.RPCs[1].ServerStream {
		t.Error("CounterStream should have server stream")
	}

	// CounterUpload: client-side stream (c2p)
	if svc.RPCs[2].Name != "CounterUpload" {
		t.Errorf("expected RPC name CounterUpload, got %s", svc.RPCs[2].Name)
	}
	if !svc.RPCs[2].ClientStream {
		t.Error("CounterUpload should have client stream")
	}
	if svc.RPCs[2].ServerStream {
		t.Error("CounterUpload should not have server stream")
	}
}

func TestStreamingFromServices(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(serviceProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	streaming := streamingFromServices(pf.Services)

	if _, ok := streaming["echo"]; ok {
		t.Error("echo should not be in streaming map")
	}
	if dir, ok := streaming["counter_stream"]; !ok || dir != "p2c" {
		t.Errorf("expected counter_stream=p2c, got %q", dir)
	}
	if dir, ok := streaming["counter_upload"]; !ok || dir != "c2p" {
		t.Errorf("expected counter_upload=c2p, got %q", dir)
	}
}

func TestDiscoverCommandsFromServices(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(serviceProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	msgByName := make(map[string]Message)
	for _, m := range pf.Messages {
		msgByName[m.Name] = m
	}
	cmds := discoverCommandsFromServices(pf.Services, msgByName)
	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(cmds))
	}
	if cmds[0].Camel != "Echo" || cmds[0].Snake != "echo" {
		t.Errorf("unexpected cmd[0]: %+v", cmds[0])
	}
	if cmds[1].Camel != "CounterStream" || cmds[1].Snake != "counter_stream" {
		t.Errorf("unexpected cmd[1]: %+v", cmds[1])
	}
	if cmds[2].Camel != "CounterUpload" || cmds[2].Snake != "counter_upload" {
		t.Errorf("unexpected cmd[2]: %+v", cmds[2])
	}
	// Verify request/response fields are populated
	if len(cmds[0].RequestFields) != 1 {
		t.Errorf("expected 1 request field for Echo, got %d", len(cmds[0].RequestFields))
	}
}

func TestParseProtoReader_Imports(t *testing.T) {
	proto := `syntax = "proto3";
package test;
import "common.proto";
import "other.proto";

message EchoRequest {
  string message = 1;
}

message EchoResponse {
  string message = 1;
}
`
	pf, err := parseProtoReader(strings.NewReader(proto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	if len(pf.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(pf.Imports))
	}
	if pf.Imports[0] != "common.proto" {
		t.Errorf("expected import[0]=common.proto, got %s", pf.Imports[0])
	}
	if pf.Imports[1] != "other.proto" {
		t.Errorf("expected import[1]=other.proto, got %s", pf.Imports[1])
	}
}

func TestParseProtoWithImports(t *testing.T) {
	// Create temporary directory with proto files
	tmpDir := t.TempDir()

	// Write shared types proto
	sharedProto := `syntax = "proto3";
package test;

enum Color {
  COLOR_UNKNOWN = 0;
  COLOR_RED = 1;
}

message Address {
  string street = 1;
  string city = 2;
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "shared.proto"), []byte(sharedProto), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write main proto that imports shared
	mainProto := `syntax = "proto3";
package test;
import "shared.proto";

message GetAddressRequest {
  string user_id = 1;
}

message GetAddressResponse {
  Address address = 1;
  Color color = 2;
}
`
	mainPath := filepath.Join(tmpDir, "main.proto")
	if err := os.WriteFile(mainPath, []byte(mainProto), 0o644); err != nil {
		t.Fatal(err)
	}

	pf, err := parseProtoWithImports(mainPath, nil)
	if err != nil {
		t.Fatalf("parseProtoWithImports: %v", err)
	}

	// Should have messages from both files
	if len(pf.Messages) < 3 {
		t.Fatalf("expected at least 3 messages (Address, GetAddressRequest, GetAddressResponse), got %d", len(pf.Messages))
	}

	// Should have enums from the imported file
	if len(pf.Enums) < 1 {
		t.Fatalf("expected at least 1 enum (Color), got %d", len(pf.Enums))
	}

	// The address field should be recognized as a message type
	// since Address is defined in the imported file
	cmds := discoverCommands(pf.Messages)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
}

func TestParseProtoWithImports_ProtoPath(t *testing.T) {
	// Create two directories: main dir and includes dir
	mainDir := t.TempDir()
	includesDir := t.TempDir()

	// Write shared proto in includes dir
	sharedProto := `syntax = "proto3";
package test;

message Shared {
  string value = 1;
}
`
	if err := os.WriteFile(filepath.Join(includesDir, "shared.proto"), []byte(sharedProto), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write main proto that imports shared
	mainProto := `syntax = "proto3";
package test;
import "shared.proto";

message UseSharedRequest {
  Shared data = 1;
}

message UseSharedResponse {
  bool ok = 1;
}
`
	mainPath := filepath.Join(mainDir, "main.proto")
	if err := os.WriteFile(mainPath, []byte(mainProto), 0o644); err != nil {
		t.Fatal(err)
	}

	// Without proto-path, import should be skipped (not found in main dir)
	pf, err := parseProtoWithImports(mainPath, nil)
	if err != nil {
		t.Fatalf("parseProtoWithImports: %v", err)
	}
	if len(pf.Messages) != 2 { // Only main file messages
		t.Fatalf("without proto-path: expected 2 messages, got %d", len(pf.Messages))
	}

	// With proto-path, import should resolve
	pf, err = parseProtoWithImports(mainPath, []string{includesDir})
	if err != nil {
		t.Fatalf("parseProtoWithImports: %v", err)
	}
	if len(pf.Messages) != 3 { // Main + imported
		t.Fatalf("with proto-path: expected 3 messages, got %d", len(pf.Messages))
	}
}

func TestParseProtoReader_Package(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(echoProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	if pf.Package != "test" {
		t.Errorf("expected package=test, got %s", pf.Package)
	}
}

func TestDiscoverCommands_Oneof(t *testing.T) {
	pf, err := parseProtoReader(strings.NewReader(oneofProto))
	if err != nil {
		t.Fatalf("parseProtoReader: %v", err)
	}
	cmds := discoverCommands(pf.Messages)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	if cmd.Camel != "Search" {
		t.Errorf("expected Camel=Search, got %s", cmd.Camel)
	}
	// Oneof fields show up as regular request fields
	if len(cmd.RequestFields) != 2 {
		t.Fatalf("expected 2 request fields, got %d", len(cmd.RequestFields))
	}
}
