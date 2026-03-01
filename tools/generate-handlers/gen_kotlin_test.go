package main

import (
	"strings"
	"testing"
)

func TestGenerateKotlinClient_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateKotlinClient(cmds, nil, "blerpc")

	mustContain := []string{
		"abstract class GeneratedClient",
		`open suspend fun echo(message: String = "")`,
		"blerpc.Blerpc.EchoRequest.newBuilder()",
		".setMessage(message)",
		`call("echo"`,
		"blerpc.Blerpc.EchoResponse.parseFrom",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Kotlin client missing %q", s)
		}
	}
}

func TestGenerateKotlinClient_CustomPkg(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateKotlinClient(cmds, nil, "myapp")

	mustContain := []string{
		"package com.myapp.android.client",
		"myapp.Myapp.EchoRequest.newBuilder()",
		"myapp.Myapp.EchoResponse.parseFrom",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Kotlin client custom pkg missing %q\nGot:\n%s", s, out)
		}
	}
	if strings.Contains(out, "blerpc") {
		t.Errorf("Kotlin client custom pkg should not contain 'blerpc'")
	}
}

func TestGenerateKotlinClient_MessageField(t *testing.T) {
	cmds := []Command{messageFieldCommand()}
	out := generateKotlinClient(cmds, nil, "blerpc")

	mustContain := []string{
		"address: Address = Address.getDefaultInstance()",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Kotlin client message field missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateKotlinClient_Map(t *testing.T) {
	cmds := []Command{mapCommand()}
	out := generateKotlinClient(cmds, nil, "blerpc")

	mustContain := []string{
		"labels: Map<String, String> = emptyMap()",
		"counts: Map<String, Int> = emptyMap()",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Kotlin client map missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateKotlinClient_Repeated(t *testing.T) {
	cmds := []Command{repeatedCommand()}
	out := generateKotlinClient(cmds, nil, "blerpc")

	mustContain := []string{
		"names: List<String> = emptyList()",
		"ids: List<Int> = emptyList()",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Kotlin client repeated missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateKotlinClient_Enum(t *testing.T) {
	cmds := []Command{enumCommand()}
	out := generateKotlinClient(cmds, nil, "blerpc")

	// enum → Int type, default 0
	if !strings.Contains(out, "name: String") {
		t.Errorf("Kotlin client enum missing name param\nGot:\n%s", out)
	}
}

func TestGenerateKotlinClient_StreamP2C(t *testing.T) {
	cmds := []Command{streamP2CCommand()}
	streaming := map[string]string{"counter_stream": "p2c"}
	out := generateKotlinClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"open suspend fun counterStream(",
		"List<blerpc.Blerpc.CounterStreamResponse>",
		"streamReceive(",
		".map {",
		"parseFrom(it)",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Kotlin client p2c missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateKotlinClient_StreamC2P(t *testing.T) {
	cmds := []Command{streamC2PCommand()}
	streaming := map[string]string{"counter_upload": "c2p"}
	out := generateKotlinClient(cmds, streaming, "blerpc")

	mustContain := []string{
		"open suspend fun counterUpload(",
		"messages: List<blerpc.Blerpc.CounterUploadRequest>",
		"streamSend(",
		"it.toByteArray()",
		"parseFrom(respData)",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Kotlin client c2p missing %q\nGot:\n%s", s, out)
		}
	}
}
