package main

import (
	"strings"
	"testing"
)

func TestGenerateSwiftClient_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateSwiftClient(cmds, nil, "blerpc")

	mustContain := []string{
		"protocol GeneratedClientProtocol",
		"extension GeneratedClientProtocol",
		`func echo(message: String = "")`,
		"Blerpc_EchoRequest()",
		"req.message = message",
		`call(cmdName: "echo"`,
		"Blerpc_EchoResponse(serializedBytes:",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Swift client missing %q", s)
		}
	}
}

func TestGenerateSwiftClient_CustomPkg(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateSwiftClient(cmds, nil, "myapp")

	mustContain := []string{
		"Myapp_EchoRequest()",
		"Myapp_EchoResponse(serializedBytes:",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Swift client custom pkg missing %q\nGot:\n%s", s, out)
		}
	}
	if strings.Contains(out, "Blerpc_") {
		t.Errorf("Swift client custom pkg should not contain 'Blerpc_'")
	}
}

func TestGenerateSwiftClient_Repeated(t *testing.T) {
	cmds := []Command{repeatedCommand()}
	out := generateSwiftClient(cmds, nil, "blerpc")

	mustContain := []string{
		"names: [String] = []",
		"ids: [UInt32] = []",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Swift client repeated missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateSwiftClient_MessageField(t *testing.T) {
	cmds := []Command{messageFieldCommand()}
	out := generateSwiftClient(cmds, nil, "blerpc")

	mustContain := []string{
		"address: Blerpc_Address = .init()",
		"req.address = address",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Swift client message field missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateSwiftClient_Map(t *testing.T) {
	cmds := []Command{mapCommand()}
	out := generateSwiftClient(cmds, nil, "blerpc")

	mustContain := []string{
		"labels: [String: String] = [:]",
		"counts: [String: UInt32] = [:]",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Swift client map missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateSwiftClient_Enum(t *testing.T) {
	cmds := []Command{enumCommand()}
	out := generateSwiftClient(cmds, nil, "blerpc")

	if !strings.Contains(out, "func getStatus(") {
		t.Errorf("Swift client enum missing getStatus method\nGot:\n%s", out)
	}
}

// Enum, message and nested request fields are typed as protoc-gen-swift
// generates them, and the parameters / properties follow its naming.
func TestGenerateSwiftClient_TypedFields(t *testing.T) {
	cmd := Command{
		Camel:       "SetMode",
		Snake:       "set_mode",
		RequestMsg:  "SetModeRequest",
		ResponseMsg: "SetModeResponse",
		RequestFields: []Field{
			{Type: "Mode", Name: "mode", Number: 1, IsEnum: true, TypePath: "Mode"},
			{Type: "Level", Name: "level", Number: 2, IsEnum: true, TypePath: "SetModeRequest.Level"},
			{Type: "Mode", Name: "modes", Number: 3, IsEnum: true, IsRepeated: true, TypePath: "Mode"},
			{Type: "uint32", Name: "device_id", Number: 4},
			{Type: "sint64", Name: "offset_us", Number: 5},
			{Type: "fixed32", Name: "crc", Number: 6},
		},
	}
	out := generateSwiftClient([]Command{cmd}, nil, "my_app.v1")

	mustContain := []string{
		"func setMode(mode: MyApp_V1_Mode = .init(), level: MyApp_V1_SetModeRequest.Level = .init(), " +
			"modes: [MyApp_V1_Mode] = [], deviceID: UInt32 = 0, offsetUs: Int64 = 0, crc: UInt32 = 0)",
		"var req = MyApp_V1_SetModeRequest()",
		"req.deviceID = deviceID",
		"req.offsetUs = offsetUs",
		"MyApp_V1_SetModeResponse(serializedBytes:",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Swift client missing %q\nGot:\n%s", s, out)
		}
	}
	if strings.Contains(out, "Any") {
		t.Errorf("Swift client has an untyped parameter\nGot:\n%s", out)
	}
}

func TestGenerateSwiftClient_EmptyRequestIsLet(t *testing.T) {
	cmd := Command{Camel: "GetTime", Snake: "get_time", RequestMsg: "GetTimeRequest", ResponseMsg: "GetTimeResponse"}
	out := generateSwiftClient([]Command{cmd}, nil, "blerpc")
	if !strings.Contains(out, "let req = Blerpc_GetTimeRequest()") {
		t.Errorf("empty request should be a let\nGot:\n%s", out)
	}
}

func TestGenerateSwiftClient_StreamP2C(t *testing.T) {
	cmds := []Command{streamP2CCommand()}
	streaming := map[string]string{"counter_stream": "p2c"}
	out := generateSwiftClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"func counterStream(start: UInt32 = 0)",
		"[Blerpc_CounterStreamResponse]",
		"streamReceive(",
		"Blerpc_CounterStreamResponse(serializedBytes:",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Swift client p2c missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateSwiftClient_StreamC2P(t *testing.T) {
	cmds := []Command{streamC2PCommand()}
	streaming := map[string]string{"counter_upload": "c2p"}
	out := generateSwiftClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"func counterUpload(",
		"messages: [Blerpc_CounterUploadRequest]",
		"streamSend(",
		"Blerpc_CounterUploadResponse(serializedBytes:",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Swift client c2p missing %q\nGot:\n%s", s, out)
		}
	}
}
