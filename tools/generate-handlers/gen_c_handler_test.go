package main

import (
	"strings"
	"testing"
)

func echoCommand() Command {
	return Command{
		Camel:      "Echo",
		Snake:      "echo",
		RequestMsg: "EchoRequest",
		ResponseMsg: "EchoResponse",
		RequestFields: []Field{
			{Type: "string", Name: "message", Number: 1},
		},
		ResponseFields: []Field{
			{Type: "string", Name: "message", Number: 1},
		},
	}
}

func messageFieldCommand() Command {
	return Command{
		Camel:      "UpdateAddress",
		Snake:      "update_address",
		RequestMsg: "UpdateAddressRequest",
		ResponseMsg: "UpdateAddressResponse",
		RequestFields: []Field{
			{Type: "string", Name: "user_id", Number: 1},
			{Type: "Address", Name: "address", Number: 2, IsMessage: true},
		},
		ResponseFields: []Field{
			{Type: "bool", Name: "ok", Number: 1},
		},
	}
}

func mapCommand() Command {
	return Command{
		Camel:      "SetLabels",
		Snake:      "set_labels",
		RequestMsg: "SetLabelsRequest",
		ResponseMsg: "SetLabelsResponse",
		RequestFields: []Field{
			{Name: "labels", Number: 1, IsMap: true, KeyType: "string", ValueType: "string"},
			{Name: "counts", Number: 2, IsMap: true, KeyType: "string", ValueType: "uint32"},
		},
		ResponseFields: []Field{
			{Type: "bool", Name: "ok", Number: 1},
		},
	}
}

func repeatedCommand() Command {
	return Command{
		Camel:      "Batch",
		Snake:      "batch",
		RequestMsg: "BatchRequest",
		ResponseMsg: "BatchResponse",
		RequestFields: []Field{
			{Type: "string", Name: "names", Number: 1, IsRepeated: true},
			{Type: "uint32", Name: "ids", Number: 2, IsRepeated: true},
		},
		ResponseFields: []Field{
			{Type: "string", Name: "results", Number: 1, IsRepeated: true},
		},
	}
}

func enumCommand() Command {
	return Command{
		Camel:      "GetStatus",
		Snake:      "get_status",
		RequestMsg: "GetStatusRequest",
		ResponseMsg: "GetStatusResponse",
		RequestFields: []Field{
			{Type: "string", Name: "name", Number: 1},
		},
		ResponseFields: []Field{
			{Type: "Status", Name: "status", Number: 1, IsEnum: true},
		},
	}
}

func streamP2CCommand() Command {
	return Command{
		Camel:      "CounterStream",
		Snake:      "counter_stream",
		RequestMsg: "CounterStreamRequest",
		ResponseMsg: "CounterStreamResponse",
		RequestFields: []Field{
			{Type: "uint32", Name: "start", Number: 1},
		},
		ResponseFields: []Field{
			{Type: "uint32", Name: "count", Number: 1},
		},
	}
}

func streamC2PCommand() Command {
	return Command{
		Camel:      "CounterUpload",
		Snake:      "counter_upload",
		RequestMsg: "CounterUploadRequest",
		ResponseMsg: "CounterUploadResponse",
		RequestFields: []Field{
			{Type: "uint32", Name: "value", Number: 1},
		},
		ResponseFields: []Field{
			{Type: "uint32", Name: "total", Number: 1},
		},
	}
}

func callbackCommand() Command {
	return Command{
		Camel:      "DataWrite",
		Snake:      "data_write",
		RequestMsg: "DataWriteRequest",
		ResponseMsg: "DataWriteResponse",
		RequestFields: []Field{
			{Type: "uint32", Name: "address", Number: 1},
			{Type: "bytes", Name: "data", Number: 2},
		},
		ResponseFields: []Field{
			{Type: "bool", Name: "ok", Number: 1},
		},
	}
}

func TestGenerateCHeader_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateCHeader(cmds, "blerpc")

	mustContain := []string{
		"#ifndef BLERPC_GENERATED_HANDLERS_H",
		"int handle_echo(const uint8_t *req_data, size_t req_len,",
		"pb_ostream_t *ostream);",
		"handlers_lookup",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C header missing %q", s)
		}
	}
}

func TestGenerateCHeader_CustomPkg(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateCHeader(cmds, "myapp")

	mustContain := []string{
		"#ifndef MYAPP_GENERATED_HANDLERS_H",
		"#define MYAPP_GENERATED_HANDLERS_H",
		"MYAPP_GENERATED_HANDLERS_H",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C header custom pkg missing %q\nGot:\n%s", s, out)
		}
	}
	mustNotContain := []string{
		"BLERPC_",
		"blerpc.pb.h",
	}
	for _, s := range mustNotContain {
		if strings.Contains(out, s) {
			t.Errorf("C header custom pkg should not contain %q", s)
		}
	}
}

func TestGenerateCHeader_MultipleCommands(t *testing.T) {
	cmds := []Command{echoCommand(), enumCommand()}
	out := generateCHeader(cmds, "blerpc")

	mustContain := []string{
		"int handle_echo(",
		"int handle_get_status(",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C header multiple commands missing %q", s)
		}
	}
}

func TestGenerateCSource_Echo(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateCSource(cmds, nil, "blerpc")

	mustContain := []string{
		"__attribute__((weak))",
		"int handle_echo(",
		"blerpc_EchoRequest req = blerpc_EchoRequest_init_zero;",
		"blerpc_EchoResponse resp = blerpc_EchoResponse_init_zero;",
		`{"echo", 4, handle_echo}`,
		"handlers_lookup",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C source missing %q", s)
		}
	}
}

func TestGenerateCSource_Callback(t *testing.T) {
	cmds := []Command{callbackCommand()}
	callbacks := map[string]bool{
		"DataWriteRequest.data": true,
	}
	out := generateCSource(cmds, callbacks, "blerpc")

	mustContain := []string{
		"req.data.funcs.decode = discard_bytes_cb;",
		"handle_data_write",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C source callback missing %q\nGot:\n%s", s, out)
		}
	}
}

func TestGenerateCSource_CustomPkg(t *testing.T) {
	cmds := []Command{echoCommand()}
	out := generateCSource(cmds, nil, "myapp")

	mustContain := []string{
		"myapp.pb.h",
		"myapp_EchoRequest req = myapp_EchoRequest_init_zero;",
		"myapp_EchoResponse resp = myapp_EchoResponse_init_zero;",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("C source custom pkg missing %q\nGot:\n%s", s, out)
		}
	}
	if strings.Contains(out, "blerpc_") {
		t.Error("C source custom pkg should not contain 'blerpc_'")
	}
}
