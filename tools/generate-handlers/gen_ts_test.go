package main

import (
	"strings"
	"testing"
)

func TestGenerateTsClient_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateTsClient(cmds, nil, "blerpc")

	mustContain := []string{
		"export abstract class GeneratedClient",
		"async echo(",
		"message = ''",
		"blerpc.EchoRequest.create({ message })",
		"blerpc.EchoRequest.encode(req).finish()",
		"blerpc.EchoResponse.decode(respData)",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("TypeScript client missing %q", s)
		}
	}
}

func TestGenerateTsClient_CustomPkg(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateTsClient(cmds, nil, "myapp")

	mustContain := []string{
		"import { myapp } from '../proto/myapp'",
		"myapp.EchoRequest.create(",
		"myapp.EchoResponse.decode(",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("TS client custom pkg missing %q\nGot:\n%s", s, out)
		}
	}
	if strings.Contains(out, "blerpc") {
		t.Errorf("TS client custom pkg should not contain 'blerpc'")
	}
}

func TestGenerateTsClient_MessageField(t *testing.T) {
	cmds := []Command{messageFieldCommand()}
	out := generateTsClient(cmds, nil, "blerpc")

	mustContain := []string{
		"address?: Address",
		"address = {}",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("TS client message field missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateTsClient_Map(t *testing.T) {
	cmds := []Command{mapCommand()}
	out := generateTsClient(cmds, nil, "blerpc")

	mustContain := []string{
		"labels?: Record<string, string>",
		"counts?: Record<string, number>",
		"labels = {}",
		"counts = {}",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("TS client map missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateTsClient_Repeated(t *testing.T) {
	cmds := []Command{repeatedCommand()}
	out := generateTsClient(cmds, nil, "blerpc")

	mustContain := []string{
		"names?: string[]",
		"ids?: number[]",
		"names = []",
		"ids = []",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("TypeScript client repeated missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateTsClient_Enum(t *testing.T) {
	cmds := []Command{enumCommand()}
	out := generateTsClient(cmds, nil, "blerpc")

	if !strings.Contains(out, "async getStatus(") {
		t.Errorf("TS client enum missing getStatus method\nGot:\n%s", out)
	}
}

func TestGenerateTsClient_StreamP2C(t *testing.T) {
	cmds := []Command{streamP2CCommand()}
	streaming := map[string]string{"counter_stream": "p2c"}
	out := generateTsClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"async counterStream(",
		"blerpc.CounterStreamResponse[]",
		"this.streamReceive(",
		"blerpc.CounterStreamResponse.decode(data)",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("TS client p2c missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateTsClient_StreamC2P(t *testing.T) {
	cmds := []Command{streamC2PCommand()}
	streaming := map[string]string{"counter_upload": "c2p"}
	out := generateTsClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"async counterUpload(",
		"blerpc.ICounterUploadRequest[]",
		"this.streamSend(",
		"blerpc.CounterUploadResponse.decode(respData)",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("TS client c2p missing %q\nGot:\n%s", s, out)
		}
	}
}
