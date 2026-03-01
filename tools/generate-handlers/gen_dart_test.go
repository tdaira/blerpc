package main

import (
	"strings"
	"testing"
)

func TestGenerateDartClient_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateDartClient(cmds, nil, "blerpc")

	mustContain := []string{
		"mixin GeneratedClientMixin",
		"Future<EchoResponse> echo(",
		"String message = ''",
		"final req = EchoRequest()..message = message;",
		"await call('echo'",
		"EchoResponse.fromBuffer(respData)",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Dart client missing %q", s)
		}
	}
}

func TestGenerateDartClient_CustomPkg(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateDartClient(cmds, nil, "myapp")

	mustContain := []string{
		"import 'package:myapp_central/proto/myapp.pb.dart'",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Dart client custom pkg missing %q\nGot:\n%s", s, out)
		}
	}
	if strings.Contains(out, "blerpc") {
		t.Errorf("Dart client custom pkg should not contain 'blerpc'")
	}
}

func TestGenerateDartClient_Repeated(t *testing.T) {
	cmds := []Command{repeatedCommand()}
	out := generateDartClient(cmds, nil, "blerpc")

	mustContain := []string{
		"List<String> names = const []",
		"List<int> ids = const []",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Dart client repeated missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateDartClient_MessageField(t *testing.T) {
	cmds := []Command{messageFieldCommand()}
	out := generateDartClient(cmds, nil, "blerpc")

	mustContain := []string{
		"Address address = Address()",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Dart client message field missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateDartClient_Map(t *testing.T) {
	cmds := []Command{mapCommand()}
	out := generateDartClient(cmds, nil, "blerpc")

	mustContain := []string{
		"Map<String, String> labels = const {}",
		"Map<String, int> counts = const {}",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Dart client map missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateDartClient_Enum(t *testing.T) {
	cmds := []Command{enumCommand()}
	out := generateDartClient(cmds, nil, "blerpc")

	if !strings.Contains(out, "Future<GetStatusResponse> getStatus(") {
		t.Errorf("Dart client enum missing getStatus method\nGot:\n%s", out)
	}
}

func TestGenerateDartClient_StreamP2C(t *testing.T) {
	cmds := []Command{streamP2CCommand()}
	streaming := map[string]string{"counter_stream": "p2c"}
	out := generateDartClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"Future<List<CounterStreamResponse>> counterStream(",
		"await streamReceive(",
		"CounterStreamResponse.fromBuffer(data)",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Dart client p2c missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateDartClient_StreamC2P(t *testing.T) {
	cmds := []Command{streamC2PCommand()}
	streaming := map[string]string{"counter_upload": "c2p"}
	out := generateDartClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"Future<CounterUploadResponse> counterUpload(",
		"List<CounterUploadRequest> messages",
		"await streamSend(",
		"CounterUploadResponse.fromBuffer(respData)",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Dart client c2p missing %q\nGot:\n%s", s, out)
		}
	}
}
