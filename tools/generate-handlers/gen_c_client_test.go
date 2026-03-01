package main

import (
	"strings"
	"testing"
)

func TestGenerateCClientHeader_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateCClientHeader(cmds, nil, nil, "blerpc")

	mustContain := []string{
		"#ifndef BLERPC_GENERATED_CLIENT_H",
		"blerpc_rpc_call",
		"int blerpc_echo(",
		"const char *message",
		"blerpc_EchoResponse *resp",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C client header missing %q", s)
		}
	}
}

func TestGenerateCClientSource_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateCClientSource(cmds, nil, nil, "blerpc")

	mustContain := []string{
		`#include "generated_client.h"`,
		"int blerpc_echo(",
		"blerpc_EchoRequest req = blerpc_EchoRequest_init_zero",
		"strncpy(req.message, message",
		`blerpc_rpc_call("echo"`,
		"blerpc_EchoResponse_fields",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C client source missing %q", s)
		}
	}
}

func TestGenerateCClientHeader_StreamP2C(t *testing.T) {
	cmds := []Command{streamP2CCommand()}
	streaming := map[string]string{"counter_stream": "p2c"}
	out := generateCClientHeader(cmds, streaming, nil, "blerpc")

	mustContain := []string{
		"int blerpc_counter_stream(",
		"blerpc_CounterStreamResponse *results",
		"size_t max_results",
		"size_t *result_count",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C client header p2c missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateCClientSource_StreamP2C(t *testing.T) {
	cmds := []Command{streamP2CCommand()}
	streaming := map[string]string{"counter_stream": "p2c"}
	out := generateCClientSource(cmds, streaming, nil, "blerpc")

	mustContain := []string{
		"struct _blerpc_counter_stream_ctx",
		"_blerpc_counter_stream_on_resp",
		"blerpc_stream_receive(",
		"blerpc_CounterStreamRequest req = blerpc_CounterStreamRequest_init_zero",
		"req.start = start",
		"*result_count = ctx.count",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C client source p2c missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateCClientHeader_StreamC2P(t *testing.T) {
	cmds := []Command{streamC2PCommand()}
	streaming := map[string]string{"counter_upload": "c2p"}
	out := generateCClientHeader(cmds, streaming, nil, "blerpc")

	mustContain := []string{
		"int blerpc_counter_upload(",
		"const blerpc_CounterUploadRequest *messages",
		"size_t msg_count",
		"blerpc_CounterUploadResponse *resp",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C client header c2p missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateCClientSource_StreamC2P(t *testing.T) {
	cmds := []Command{streamC2PCommand()}
	streaming := map[string]string{"counter_upload": "c2p"}
	out := generateCClientSource(cmds, streaming, nil, "blerpc")

	mustContain := []string{
		"struct _blerpc_counter_upload_ctx",
		"_blerpc_counter_upload_next(",
		"blerpc_stream_send(",
		"blerpc_CounterUploadResponse_fields",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C client source c2p missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateCClientSource_Callback(t *testing.T) {
	cmds := []Command{callbackCommand()}
	callbacks := map[string]bool{
		"DataWriteRequest.data": true,
	}
	out := generateCClientSource(cmds, nil, callbacks, "blerpc")

	mustContain := []string{
		"_blerpc_encode_bytes_cb",
		"_blerpc_bytes_encode_ctx",
		"req.data.funcs.encode",
		"work_buf",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C client source callback missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateCClientHeader_CustomPkg(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateCClientHeader(cmds, nil, nil, "myapp")

	mustContain := []string{
		"#ifndef MYAPP_GENERATED_CLIENT_H",
		"myapp_rpc_call",
		"int myapp_echo(",
		"myapp_EchoResponse *resp",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C client header custom pkg missing %q\nGot:\n%s", s, out)
		}
	}
	if strings.Contains(out, "blerpc") {
		t.Error("C client header custom pkg should not contain 'blerpc'")
	}
}

func TestGenerateCClientSource_MultiField(t *testing.T) {
	cmds := []Command{messageFieldCommand()}
	out := generateCClientSource(cmds, nil, nil, "blerpc")

	mustContain := []string{
		"int blerpc_update_address(",
		"blerpc_UpdateAddressRequest req = blerpc_UpdateAddressRequest_init_zero",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C client source multi-field missing %q\nGot:\n%s", s, out)
		}
	}
}
