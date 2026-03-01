package main

import (
	"strings"
	"testing"
)

func TestGeneratePyHandlers_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generatePyHandlers(cmds, "blerpc")

	mustContain := []string{
		"def handle_echo(req_data):",
		"blerpc_pb2.EchoRequest()",
		"blerpc_pb2.EchoResponse().SerializeToString()",
		`"echo": handle_echo`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Python handlers missing %q", s)
		}
	}
}

func TestGeneratePyHandlers_MultipleCommands(t *testing.T) {
	cmds := []Command{echoCommand(), enumCommand()}
	out := generatePyHandlers(cmds, "blerpc")

	mustContain := []string{
		"def handle_echo(req_data):",
		"def handle_get_status(req_data):",
		`"echo": handle_echo`,
		`"get_status": handle_get_status`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Python handlers multiple missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGeneratePyHandlers_CustomPkg(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generatePyHandlers(cmds, "myapp")

	mustContain := []string{
		"myapp_pb2.EchoRequest()",
		"myapp_pb2.EchoResponse()",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Python handlers custom pkg missing %q\nGot:\n%s", s, out)
		}
	}
	if strings.Contains(out, "blerpc") {
		t.Error("Python handlers custom pkg should not contain 'blerpc'")
	}
}

func TestGeneratePyClient_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generatePyClient(cmds, nil, "blerpc")

	mustContain := []string{
		"class GeneratedClientMixin:",
		`async def echo(self, *, message=""):`,
		"blerpc_pb2.EchoRequest(message=message)",
		`await self._call("echo"`,
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Python client missing %q", s)
		}
	}
}

func TestGeneratePyClient_CustomPkg(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generatePyClient(cmds, nil, "myapp")

	mustContain := []string{
		"from . import myapp_pb2",
		"myapp_pb2.EchoRequest(",
		"myapp_pb2.EchoResponse()",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Python client custom pkg missing %q\nGot:\n%s", s, out)
		}
	}
	if strings.Contains(out, "blerpc") {
		t.Errorf("Python client custom pkg should not contain 'blerpc'")
	}
}

func TestGeneratePyClient_Repeated(t *testing.T) {
	cmds := []Command{repeatedCommand()}
	out := generatePyClient(cmds, nil, "blerpc")

	mustContain := []string{
		"names=None",
		"ids=None",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Python client repeated missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGeneratePyClient_Enum(t *testing.T) {
	cmds := []Command{enumCommand()}
	out := generatePyClient(cmds, nil, "blerpc")

	if !strings.Contains(out, "async def get_status(") {
		t.Errorf("Python client enum missing get_status method\nGot:\n%s", out)
	}
}

func TestGeneratePyClient_StreamP2C(t *testing.T) {
	cmds := []Command{streamP2CCommand()}
	streaming := map[string]string{"counter_stream": "p2c"}
	out := generatePyClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"async def counter_stream(self",
		"P2C stream:",
		"async for data in self.stream_receive(",
		"ParseFromString(data)",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Python client p2c missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGeneratePyClient_StreamC2P(t *testing.T) {
	cmds := []Command{streamC2PCommand()}
	streaming := map[string]string{"counter_upload": "c2p"}
	out := generatePyClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"async def counter_upload(self, messages):",
		"C2P stream:",
		"self.stream_send(",
		"SerializeToString()",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Python client c2p missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGeneratePyClient_Map(t *testing.T) {
	cmds := []Command{mapCommand()}
	out := generatePyClient(cmds, nil, "blerpc")

	mustContain := []string{
		"labels=None",
		"counts=None",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Python client map missing %q\nGot:\n%s", s, out)
		}
	}
}
