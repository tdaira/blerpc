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
		"address: Address = Address()",
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

	// Enum fields don't affect request params much (name is string),
	// but method should be generated
	if !strings.Contains(out, "func getStatus(") {
		t.Errorf("Swift client enum missing getStatus method\nGot:\n%s", out)
	}
}

func TestGenerateSwiftClient_StreamP2C(t *testing.T) {
	cmds := []Command{streamP2CCommand()}
	streaming := map[string]string{"counter_stream": "p2c"}
	out := generateSwiftClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"func counterStream(",
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
