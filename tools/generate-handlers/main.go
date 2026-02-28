// Generate handler stubs and client code from blerpc.proto.
//
// Parses proto file with go-protoparser (proper AST) and generates:
//   - peripheral_fw/src/generated_handlers.h  — C declarations + handler_entry + lookup
//   - peripheral_fw/src/generated_handlers.c  — weak handler stubs + handler table
//   - peripheral_py/generated_handlers.py  — Python handler stubs + HANDLERS dict
//   - central_py/blerpc/generated/generated_client.py — Python client mixin class
//   - central_android/.../generated/GeneratedClient.kt — Kotlin/Android client
//   - central_ios/.../GeneratedClient.swift — Swift/iOS client
//   - central_flutter/lib/client/generated_client.dart — Dart/Flutter client mixin
//   - central_rn/src/client/GeneratedClient.ts — TypeScript/React Native abstract client
//   - central_fw/src/generated_client.h — C central client header (extern transport + typed wrappers)
//   - central_fw/src/generated_client.c — C central client implementation
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// Field represents a protobuf message field.
type Field struct {
	Type   string
	Name   string
	Number int
}

// Message represents a protobuf message.
type Message struct {
	Name   string
	Fields []Field
}

// Command represents a matched Request/Response pair.
type Command struct {
	Camel          string
	Snake          string
	RequestMsg     string
	ResponseMsg    string
	RequestFields  []Field
	ResponseFields []Field
}

// kotlinTypes maps proto field types to Kotlin types.
var kotlinTypes = map[string]string{
	"string": "String",
	"bytes":  "com.google.protobuf.ByteString",
	"uint32": "Int",
	"int32":  "Int",
	"uint64": "Long",
	"int64":  "Long",
	"float":  "Float",
	"double": "Double",
	"bool":   "Boolean",
}

// kotlinDefaults maps proto field types to Kotlin default values.
var kotlinDefaults = map[string]string{
	"string": "\"\"",
	"bytes":  "com.google.protobuf.ByteString.EMPTY",
	"uint32": "0",
	"int32":  "0",
	"uint64": "0L",
	"int64":  "0L",
	"float":  "0.0f",
	"double": "0.0",
	"bool":   "false",
}

// swiftTypes maps proto field types to Swift types.
var swiftTypes = map[string]string{
	"string": "String",
	"bytes":  "Data",
	"uint32": "UInt32",
	"int32":  "Int32",
	"uint64": "UInt64",
	"int64":  "Int64",
	"float":  "Float",
	"double": "Double",
	"bool":   "Bool",
}

// swiftDefaults maps proto field types to Swift default values.
var swiftDefaults = map[string]string{
	"string": "\"\"",
	"bytes":  "Data()",
	"uint32": "0",
	"int32":  "0",
	"uint64": "0",
	"int64":  "0",
	"float":  "0.0",
	"double": "0.0",
	"bool":   "false",
}

// dartTypes maps proto field types to Dart types.
var dartTypes = map[string]string{
	"string": "String",
	"bytes":  "List<int>",
	"uint32": "int",
	"int32":  "int",
	"uint64": "int",
	"int64":  "int",
	"float":  "double",
	"double": "double",
	"bool":   "bool",
}

// dartDefaults maps proto field types to Dart default values.
var dartDefaults = map[string]string{
	"string": "''",
	"bytes":  "const <int>[]",
	"uint32": "0",
	"int32":  "0",
	"uint64": "0",
	"int64":  "0",
	"float":  "0.0",
	"double": "0.0",
	"bool":   "false",
}

// tsTypes maps proto field types to TypeScript types.
var tsTypes = map[string]string{
	"string": "string",
	"bytes":  "Uint8Array",
	"uint32": "number",
	"int32":  "number",
	"uint64": "number",
	"int64":  "number",
	"float":  "number",
	"double": "number",
	"bool":   "boolean",
}

// tsDefaults maps proto field types to TypeScript default values.
var tsDefaults = map[string]string{
	"string": "''",
	"bytes":  "new Uint8Array(0)",
	"uint32": "0",
	"int32":  "0",
	"uint64": "0",
	"int64":  "0",
	"float":  "0",
	"double": "0",
	"bool":   "false",
}

// cTypes maps proto field types to C types (for function parameters).
var cTypes = map[string]string{
	"string": "const char *",
	"bytes":  "const uint8_t *",
	"uint32": "uint32_t",
	"int32":  "int32_t",
	"uint64": "uint64_t",
	"int64":  "int64_t",
	"float":  "float",
	"double": "double",
	"bool":   "bool",
}

// pythonDefaults maps proto field types to Python default values.
var pythonDefaults = map[string]string{
	"string": `""`,
	"bytes":  `b""`,
	"uint32": "0",
	"int32":  "0",
	"uint64": "0",
	"int64":  "0",
	"float":  "0.0",
	"double": "0.0",
	"bool":   "False",
}

var (
	reSub1 = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	reSub2 = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func camelToSnake(name string) string {
	s := reSub1.ReplaceAllString(name, "${1}_${2}")
	s = reSub2.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}

func parseProto(path string) ([]Message, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open proto: %w", err)
	}
	defer reader.Close()

	proto, err := protoparser.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("parse proto: %w", err)
	}

	var messages []Message
	for _, item := range proto.ProtoBody {
		msg, ok := item.(*parser.Message)
		if !ok {
			continue
		}
		m := Message{Name: msg.MessageName}
		for _, body := range msg.MessageBody {
			f, ok := body.(*parser.Field)
			if !ok {
				continue
			}
			num := 0
			_, _ = fmt.Sscanf(f.FieldNumber, "%d", &num)
			m.Fields = append(m.Fields, Field{
				Type:   f.Type,
				Name:   f.FieldName,
				Number: num,
			})
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func parseStreamingCommands(path string) (map[string]string, error) {
	streaming := make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return streaming, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid streaming line (expected 'name direction'): %q", line)
		}
		dir := parts[1]
		if dir != "p2c" && dir != "c2p" {
			return nil, fmt.Errorf("invalid direction %q (must be p2c or c2p)", dir)
		}
		streaming[parts[0]] = dir
	}
	return streaming, scanner.Err()
}

func parseOptions(path string) (map[string]bool, error) {
	callbacks := make(map[string]bool)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return callbacks, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "FT_CALLBACK") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				qualified := strings.TrimPrefix(parts[0], "blerpc.")
				callbacks[qualified] = true
			}
		}
	}
	return callbacks, scanner.Err()
}

func discoverCommands(messages []Message) []Command {
	msgByName := make(map[string]Message)
	for _, m := range messages {
		msgByName[m.Name] = m
	}

	var commands []Command
	for _, msg := range messages {
		if !strings.HasSuffix(msg.Name, "Request") {
			continue
		}
		camel := msg.Name[:len(msg.Name)-len("Request")]
		respName := camel + "Response"
		resp, ok := msgByName[respName]
		if !ok {
			continue
		}
		commands = append(commands, Command{
			Camel:          camel,
			Snake:          camelToSnake(camel),
			RequestMsg:     msg.Name,
			ResponseMsg:    respName,
			RequestFields:  msg.Fields,
			ResponseFields: resp.Fields,
		})
	}
	return commands
}

func generateCHeader(commands []Command) string {
	var b strings.Builder
	lines := []string{
		"/* Auto-generated by generate-handlers — DO NOT EDIT */",
		"#ifndef BLERPC_GENERATED_HANDLERS_H",
		"#define BLERPC_GENERATED_HANDLERS_H",
		"",
		"#include <stdint.h>",
		"#include <stddef.h>",
		"#include <pb_encode.h>",
		"",
		"#ifdef __cplusplus",
		`extern "C" {`,
		"#endif",
		"",
		"typedef int (*command_handler_fn)(const uint8_t *req_data, size_t req_len,",
		"                                  pb_ostream_t *ostream);",
		"",
		"struct handler_entry {",
		"    const char *name;",
		"    uint8_t name_len;",
		"    command_handler_fn handler;",
		"};",
		"",
		"command_handler_fn handlers_lookup(const char *name, uint8_t name_len);",
		"",
	}
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}

	for _, cmd := range commands {
		pad := strings.Repeat(" ", len(cmd.Snake))
		b.WriteString(fmt.Sprintf("int handle_%s(const uint8_t *req_data, size_t req_len,\n", cmd.Snake))
		b.WriteString(fmt.Sprintf("                %spb_ostream_t *ostream);\n", pad))
		b.WriteByte('\n')
	}

	tail := []string{
		"#ifdef __cplusplus",
		"}",
		"#endif",
		"",
		"#endif /* BLERPC_GENERATED_HANDLERS_H */",
	}
	for _, l := range tail {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	return b.String()
}

func generateCSource(commands []Command, callbacks map[string]bool) string {
	var b strings.Builder

	header := []string{
		"/* Auto-generated by generate-handlers — DO NOT EDIT */",
		`#include "generated_handlers.h"`,
		`#include "blerpc.pb.h"`,
		"#include <pb_encode.h>",
		"#include <pb_decode.h>",
		"#include <string.h>",
		"",
		"/* Discard callback for FT_CALLBACK fields during decode */",
		"static bool discard_bytes_cb(pb_istream_t *stream, const pb_field_t *field,",
		"                             void **arg)",
		"{",
		"    (void)field;",
		"    (void)arg;",
		"    uint8_t buf[64];",
		"    size_t left = stream->bytes_left;",
		"    while (left > 0) {",
		"        size_t n = left < sizeof(buf) ? left : sizeof(buf);",
		"        if (!pb_read(stream, buf, n)) return false;",
		"        left -= n;",
		"    }",
		"    return true;",
		"}",
		"",
	}
	for _, l := range header {
		b.WriteString(l)
		b.WriteByte('\n')
	}

	// Weak handler stubs
	for _, cmd := range commands {
		reqMsg := "blerpc_" + cmd.RequestMsg
		respMsg := "blerpc_" + cmd.ResponseMsg
		pad := strings.Repeat(" ", len(cmd.Snake))

		b.WriteString("__attribute__((weak))\n")
		b.WriteString(fmt.Sprintf("int handle_%s(const uint8_t *req_data, size_t req_len,\n", cmd.Snake))
		b.WriteString(fmt.Sprintf("                %spb_ostream_t *ostream)\n", pad))
		b.WriteString("{\n")

		// Decode request
		b.WriteString(fmt.Sprintf("    %s req = %s_init_zero;\n", reqMsg, reqMsg))

		// Discard callbacks for FT_CALLBACK request fields
		for _, field := range cmd.RequestFields {
			key := cmd.RequestMsg + "." + field.Name
			if callbacks[key] {
				b.WriteString(fmt.Sprintf("    req.%s.funcs.decode = discard_bytes_cb;\n", field.Name))
			}
		}

		b.WriteString("    pb_istream_t stream = pb_istream_from_buffer(req_data, req_len);\n")
		b.WriteString(fmt.Sprintf("    if (!pb_decode(&stream, %s_fields, &req)) return -1;\n", reqMsg))
		b.WriteByte('\n')

		// Encode response
		b.WriteString(fmt.Sprintf("    %s resp = %s_init_zero;\n", respMsg, respMsg))
		b.WriteString(fmt.Sprintf("    if (!pb_encode(ostream, %s_fields, &resp)) return -1;\n", respMsg))
		b.WriteString("    return 0;\n")
		b.WriteString("}\n")
		b.WriteByte('\n')
	}

	// Handler table
	b.WriteString("static const struct handler_entry handler_table[] = {\n")
	for _, cmd := range commands {
		b.WriteString(fmt.Sprintf("    {\"%s\", %d, handle_%s},\n", cmd.Snake, len(cmd.Snake), cmd.Snake))
	}
	b.WriteString("};\n")
	b.WriteByte('\n')

	// Lookup function
	b.WriteString("command_handler_fn handlers_lookup(const char *name, uint8_t name_len)\n")
	b.WriteString("{\n")
	b.WriteString("    size_t i;\n")
	b.WriteString("    for (i = 0; i < sizeof(handler_table) / sizeof(handler_table[0]); i++) {\n")
	b.WriteString("        if (handler_table[i].name_len == name_len &&\n")
	b.WriteString("            memcmp(handler_table[i].name, name, name_len) == 0) {\n")
	b.WriteString("            return handler_table[i].handler;\n")
	b.WriteString("        }\n")
	b.WriteString("    }\n")
	b.WriteString("    return NULL;\n")
	b.WriteString("}\n")

	return b.String()
}

func generatePyHandlers(commands []Command) string {
	var b strings.Builder

	b.WriteString("\"\"\"Auto-generated by generate-handlers — DO NOT EDIT.\"\"\"\n")
	b.WriteByte('\n')
	b.WriteString("import os\n")
	b.WriteString("import sys\n")
	b.WriteByte('\n')
	b.WriteString("sys.path.insert(0, os.path.join(os.path.dirname(__file__), \"..\", \"central_py\"))\n")
	b.WriteString("from blerpc.generated import blerpc_pb2\n")
	b.WriteByte('\n')
	b.WriteByte('\n')

	for _, cmd := range commands {
		reqCls := "blerpc_pb2." + cmd.RequestMsg
		respCls := "blerpc_pb2." + cmd.ResponseMsg
		b.WriteString(fmt.Sprintf("def handle_%s(req_data):\n", cmd.Snake))
		b.WriteString(fmt.Sprintf("    req = %s()\n", reqCls))
		b.WriteString("    req.ParseFromString(req_data)\n")
		b.WriteString(fmt.Sprintf("    return %s().SerializeToString()\n", respCls))
		b.WriteByte('\n')
		b.WriteByte('\n')
	}

	// HANDLERS dict
	b.WriteString("HANDLERS = {\n")
	for _, cmd := range commands {
		b.WriteString(fmt.Sprintf("    \"%s\": handle_%s,\n", cmd.Snake, cmd.Snake))
	}
	b.WriteString("}\n")

	return b.String()
}

func generatePyClient(commands []Command, streaming map[string]string) string {
	var b strings.Builder

	b.WriteString("\"\"\"Auto-generated by generate-handlers — DO NOT EDIT.\"\"\"\n")
	b.WriteByte('\n')
	b.WriteString("from __future__ import annotations\n")
	b.WriteByte('\n')
	b.WriteString("from . import blerpc_pb2\n")
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString("class GeneratedClientMixin:\n")
	b.WriteString("    \"\"\"Auto-generated RPC methods (unary and streaming).\n")
	b.WriteByte('\n')
	b.WriteString("    Requires _call, stream_receive, and stream_send from BlerpcClient.\n")
	b.WriteString("    \"\"\"\n")
	b.WriteByte('\n')

	first := true
	for _, cmd := range commands {
		if _, ok := streaming[cmd.Snake]; ok {
			continue
		}

		reqCls := "blerpc_pb2." + cmd.RequestMsg
		respCls := "blerpc_pb2." + cmd.ResponseMsg

		// Build keyword args
		var params []string
		for _, f := range cmd.RequestFields {
			def, ok := pythonDefaults[f.Type]
			if !ok {
				def = "None"
			}
			params = append(params, fmt.Sprintf("%s=%s", f.Name, def))
		}

		paramsStr := strings.Join(params, ", ")
		if paramsStr != "" {
			paramsStr = ", *, " + paramsStr
		}

		// Build request constructor kwargs
		var kwargs []string
		for _, f := range cmd.RequestFields {
			kwargs = append(kwargs, fmt.Sprintf("%s=%s", f.Name, f.Name))
		}
		kwargsStr := strings.Join(kwargs, ", ")

		if !first {
			b.WriteByte('\n')
		}
		first = false

		b.WriteString(fmt.Sprintf("    async def %s(self%s):\n", cmd.Snake, paramsStr))
		b.WriteString(fmt.Sprintf("        \"\"\"Call the %s command.\"\"\"\n", cmd.Snake))
		b.WriteString(fmt.Sprintf("        req = %s(%s)\n", reqCls, kwargsStr))
		b.WriteString(fmt.Sprintf("        resp_data = await self._call(\"%s\", req.SerializeToString())\n", cmd.Snake))
		b.WriteString(fmt.Sprintf("        resp = %s()\n", respCls))
		b.WriteString("        resp.ParseFromString(resp_data)\n")
		b.WriteString("        return resp\n")
	}

	// Streaming methods
	for _, cmd := range commands {
		dir, ok := streaming[cmd.Snake]
		if !ok {
			continue
		}

		reqCls := "blerpc_pb2." + cmd.RequestMsg
		respCls := "blerpc_pb2." + cmd.ResponseMsg

		b.WriteByte('\n')

		if dir == "p2c" {
			// Build keyword args (same as unary)
			var params []string
			for _, f := range cmd.RequestFields {
				def, ok := pythonDefaults[f.Type]
				if !ok {
					def = "None"
				}
				params = append(params, fmt.Sprintf("%s=%s", f.Name, def))
			}
			paramsStr := strings.Join(params, ", ")
			if paramsStr != "" {
				paramsStr = ", *, " + paramsStr
			}

			var kwargs []string
			for _, f := range cmd.RequestFields {
				kwargs = append(kwargs, fmt.Sprintf("%s=%s", f.Name, f.Name))
			}
			kwargsStr := strings.Join(kwargs, ", ")

			b.WriteString(fmt.Sprintf("    async def %s(self%s):\n", cmd.Snake, paramsStr))
			b.WriteString(fmt.Sprintf("        \"\"\"P2C stream: %s.\"\"\"\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("        req = %s(%s)\n", reqCls, kwargsStr))
			b.WriteString("        results = []\n")
			b.WriteString(fmt.Sprintf("        async for data in self.stream_receive(\n"))
			b.WriteString(fmt.Sprintf("            \"%s\", req.SerializeToString()\n", cmd.Snake))
			b.WriteString("        ):\n")
			b.WriteString(fmt.Sprintf("            resp = %s()\n", respCls))
			b.WriteString("            resp.ParseFromString(data)\n")
			b.WriteString("            results.append(resp)\n")
			b.WriteString("        return results\n")
		} else {
			// c2p: takes list of typed request messages
			b.WriteString(fmt.Sprintf("    async def %s(self, messages):\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("        \"\"\"C2P stream: %s.\"\"\"\n", cmd.Snake))
			b.WriteString("        raw = [m.SerializeToString() for m in messages]\n")
			b.WriteString(fmt.Sprintf("        resp_data = await self.stream_send(\"%s\", raw, \"%s\")\n", cmd.Snake, cmd.Snake))
			b.WriteString(fmt.Sprintf("        resp = %s()\n", respCls))
			b.WriteString("        resp.ParseFromString(resp_data)\n")
			b.WriteString("        return resp\n")
		}
	}

	return b.String()
}

func toLowerCamel(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// kotlinSetterName returns the protobuf-java setter for a field.
// For snake_case fields like "received_count", the setter is "setReceivedCount".
func kotlinSetterName(fieldName string) string {
	parts := strings.Split(fieldName, "_")
	var b strings.Builder
	b.WriteString("set")
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}

func generateKotlinClient(commands []Command, streaming map[string]string) string {
	var b strings.Builder

	b.WriteString("/* Auto-generated by generate-handlers — DO NOT EDIT */\n")
	b.WriteString("package com.blerpc.android.client\n")
	b.WriteByte('\n')
	b.WriteString("import com.google.protobuf.ByteString\n")
	b.WriteByte('\n')
	b.WriteString("/**\n")
	b.WriteString(" * Auto-generated RPC methods.\n")
	b.WriteString(" * Subclass and override for custom behavior.\n")
	b.WriteString(" */\n")
	b.WriteString("abstract class GeneratedClient {\n")
	b.WriteString("    protected abstract suspend fun call(cmdName: String, requestData: ByteArray): ByteArray\n")
	b.WriteString("    protected abstract suspend fun streamReceive(cmdName: String, requestData: ByteArray): List<ByteArray>\n")
	b.WriteString("    protected abstract suspend fun streamSend(cmdName: String, messages: List<ByteArray>, finalCmdName: String): ByteArray\n")
	b.WriteByte('\n')

	first := true
	for _, cmd := range commands {
		if _, ok := streaming[cmd.Snake]; ok {
			continue
		}

		reqCls := "blerpc.Blerpc." + cmd.RequestMsg
		respCls := "blerpc.Blerpc." + cmd.ResponseMsg
		methodName := toLowerCamel(cmd.Camel)

		// Build parameters
		var params []string
		for _, f := range cmd.RequestFields {
			ktType, ok := kotlinTypes[f.Type]
			if !ok {
				ktType = "Any"
			}
			def, ok := kotlinDefaults[f.Type]
			if !ok {
				log.Fatalf("unsupported Kotlin default for field type %q in command %q", f.Type, cmd.Snake)
			}
			params = append(params, fmt.Sprintf("%s: %s = %s", f.Name, ktType, def))
		}

		paramsStr := strings.Join(params, ", ")

		if !first {
			b.WriteByte('\n')
		}
		first = false

		b.WriteString(fmt.Sprintf("    open suspend fun %s(%s): %s {\n", methodName, paramsStr, respCls))
		b.WriteString(fmt.Sprintf("        val req = %s.newBuilder()\n", reqCls))
		for _, f := range cmd.RequestFields {
			setter := kotlinSetterName(f.Name)
			b.WriteString(fmt.Sprintf("            .%s(%s)\n", setter, f.Name))
		}
		b.WriteString("            .build()\n")
		b.WriteString(fmt.Sprintf("        val respData = call(\"%s\", req.toByteArray())\n", cmd.Snake))
		b.WriteString(fmt.Sprintf("        return %s.parseFrom(respData)\n", respCls))
		b.WriteString("    }\n")
	}

	// Streaming methods
	for _, cmd := range commands {
		dir, ok := streaming[cmd.Snake]
		if !ok {
			continue
		}

		reqCls := "blerpc.Blerpc." + cmd.RequestMsg
		respCls := "blerpc.Blerpc." + cmd.ResponseMsg
		methodName := toLowerCamel(cmd.Camel)

		b.WriteByte('\n')

		if dir == "p2c" {
			var params []string
			for _, f := range cmd.RequestFields {
				ktType, ok := kotlinTypes[f.Type]
				if !ok {
					ktType = "Any"
				}
				def, ok := kotlinDefaults[f.Type]
				if !ok {
					log.Fatalf("unsupported Kotlin default for field type %q in command %q", f.Type, cmd.Snake)
				}
				params = append(params, fmt.Sprintf("%s: %s = %s", f.Name, ktType, def))
			}
			paramsStr := strings.Join(params, ", ")

			b.WriteString(fmt.Sprintf("    open suspend fun %s(%s): List<%s> {\n", methodName, paramsStr, respCls))
			b.WriteString(fmt.Sprintf("        val req = %s.newBuilder()\n", reqCls))
			for _, f := range cmd.RequestFields {
				setter := kotlinSetterName(f.Name)
				b.WriteString(fmt.Sprintf("            .%s(%s)\n", setter, f.Name))
			}
			b.WriteString("            .build()\n")
			b.WriteString(fmt.Sprintf("        val responses = streamReceive(\"%s\", req.toByteArray())\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("        return responses.map { %s.parseFrom(it) }\n", respCls))
			b.WriteString("    }\n")
		} else {
			b.WriteString(fmt.Sprintf("    open suspend fun %s(messages: List<%s>): %s {\n", methodName, reqCls, respCls))
			b.WriteString("        val raw = messages.map { it.toByteArray() }\n")
			b.WriteString(fmt.Sprintf("        val respData = streamSend(\"%s\", raw, \"%s\")\n", cmd.Snake, cmd.Snake))
			b.WriteString(fmt.Sprintf("        return %s.parseFrom(respData)\n", respCls))
			b.WriteString("    }\n")
		}
	}

	b.WriteString("}\n")

	return b.String()
}

// swiftPropertyName converts a snake_case field name to lowerCamelCase.
func swiftPropertyName(fieldName string) string {
	parts := strings.Split(fieldName, "_")
	if len(parts) == 0 {
		return fieldName
	}
	var b strings.Builder
	b.WriteString(parts[0])
	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}

// dartPropertyName converts a snake_case field name to lowerCamelCase (same as Swift).
func dartPropertyName(fieldName string) string {
	return swiftPropertyName(fieldName)
}

// tsPropertyName converts a snake_case field name to lowerCamelCase (same as Swift).
func tsPropertyName(fieldName string) string {
	return swiftPropertyName(fieldName)
}

func generateSwiftClient(commands []Command, streaming map[string]string) string {
	var b strings.Builder

	b.WriteString("/* Auto-generated by generate-handlers — DO NOT EDIT */\n")
	b.WriteString("import Foundation\n")
	b.WriteString("import SwiftProtobuf\n")
	b.WriteByte('\n')
	b.WriteString("/// Auto-generated RPC method protocol.\n")
	b.WriteString("/// Conform to this protocol and implement call/streamReceive/streamSend.\n")
	b.WriteString("protocol GeneratedClientProtocol {\n")
	b.WriteString("    func call(cmdName: String, requestData: Data) async throws -> Data\n")
	b.WriteString("    func streamReceive(cmdName: String, requestData: Data) async throws -> [Data]\n")
	b.WriteString("    func streamSend(cmdName: String, messages: [Data], finalCmdName: String) async throws -> Data\n")
	b.WriteString("}\n")
	b.WriteByte('\n')
	b.WriteString("extension GeneratedClientProtocol {\n")

	first := true
	for _, cmd := range commands {
		if _, ok := streaming[cmd.Snake]; ok {
			continue
		}

		reqCls := "Blerpc_" + cmd.RequestMsg
		respCls := "Blerpc_" + cmd.ResponseMsg
		methodName := toLowerCamel(cmd.Camel)

		// Build parameters
		var params []string
		for _, f := range cmd.RequestFields {
			swType, ok := swiftTypes[f.Type]
			if !ok {
				swType = "Any"
			}
			def, ok := swiftDefaults[f.Type]
			if !ok {
				def = "nil"
			}
			propName := swiftPropertyName(f.Name)
			params = append(params, fmt.Sprintf("%s: %s = %s", propName, swType, def))
		}

		paramsStr := strings.Join(params, ", ")

		if !first {
			b.WriteByte('\n')
		}
		first = false

		b.WriteString(fmt.Sprintf("    func %s(%s) async throws -> %s {\n", methodName, paramsStr, respCls))
		b.WriteString(fmt.Sprintf("        var req = %s()\n", reqCls))
		for _, f := range cmd.RequestFields {
			propName := swiftPropertyName(f.Name)
			b.WriteString(fmt.Sprintf("        req.%s = %s\n", propName, propName))
		}
		b.WriteString(fmt.Sprintf("        let respData = try await call(cmdName: \"%s\", requestData: try req.serializedData())\n", cmd.Snake))
		b.WriteString(fmt.Sprintf("        return try %s(serializedBytes: respData)\n", respCls))
		b.WriteString("    }\n")
	}

	// Streaming methods
	for _, cmd := range commands {
		dir, ok := streaming[cmd.Snake]
		if !ok {
			continue
		}

		reqCls := "Blerpc_" + cmd.RequestMsg
		respCls := "Blerpc_" + cmd.ResponseMsg
		methodName := toLowerCamel(cmd.Camel)

		b.WriteByte('\n')

		if dir == "p2c" {
			var params []string
			for _, f := range cmd.RequestFields {
				swType, ok := swiftTypes[f.Type]
				if !ok {
					swType = "Any"
				}
				def, ok := swiftDefaults[f.Type]
				if !ok {
					def = "nil"
				}
				propName := swiftPropertyName(f.Name)
				params = append(params, fmt.Sprintf("%s: %s = %s", propName, swType, def))
			}
			paramsStr := strings.Join(params, ", ")

			b.WriteString(fmt.Sprintf("    func %s(%s) async throws -> [%s] {\n", methodName, paramsStr, respCls))
			b.WriteString(fmt.Sprintf("        var req = %s()\n", reqCls))
			for _, f := range cmd.RequestFields {
				propName := swiftPropertyName(f.Name)
				b.WriteString(fmt.Sprintf("        req.%s = %s\n", propName, propName))
			}
			b.WriteString(fmt.Sprintf("        let responses = try await streamReceive(cmdName: \"%s\", requestData: try req.serializedData())\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("        return try responses.map { try %s(serializedBytes: $0) }\n", respCls))
			b.WriteString("    }\n")
		} else {
			b.WriteString(fmt.Sprintf("    func %s(messages: [%s]) async throws -> %s {\n", methodName, reqCls, respCls))
			b.WriteString("        let raw = try messages.map { try $0.serializedData() }\n")
			b.WriteString(fmt.Sprintf("        let respData = try await streamSend(cmdName: \"%s\", messages: raw, finalCmdName: \"%s\")\n", cmd.Snake, cmd.Snake))
			b.WriteString(fmt.Sprintf("        return try %s(serializedBytes: respData)\n", respCls))
			b.WriteString("    }\n")
		}
	}

	b.WriteString("}\n")

	return b.String()
}

func generateDartClient(commands []Command, streaming map[string]string) string {
	var b strings.Builder

	b.WriteString("/* Auto-generated by generate-handlers — DO NOT EDIT */\n")
	b.WriteString("import 'dart:typed_data';\n")
	b.WriteByte('\n')
	b.WriteString("import 'package:blerpc_central/proto/blerpc.pb.dart';\n")
	b.WriteByte('\n')
	b.WriteString("/// Auto-generated RPC method wrappers for blerpc commands.\n")
	b.WriteString("mixin GeneratedClientMixin {\n")
	b.WriteString("  Future<Uint8List> call(String cmdName, Uint8List requestData);\n")
	b.WriteString("  Future<List<Uint8List>> streamReceive(String cmdName, Uint8List requestData);\n")
	b.WriteString("  Future<Uint8List> streamSend(\n")
	b.WriteString("      String cmdName, List<Uint8List> messages, String finalCmdName);\n")

	for _, cmd := range commands {
		if _, ok := streaming[cmd.Snake]; ok {
			continue
		}

		reqCls := cmd.RequestMsg
		respCls := cmd.ResponseMsg
		methodName := toLowerCamel(cmd.Camel)

		// Build parameters
		var params []string
		for _, f := range cmd.RequestFields {
			dtType, ok := dartTypes[f.Type]
			if !ok {
				dtType = "dynamic"
			}
			def, ok := dartDefaults[f.Type]
			if !ok {
				def = "null"
			}
			propName := dartPropertyName(f.Name)
			params = append(params, fmt.Sprintf("%s %s = %s", dtType, propName, def))
		}

		paramsStr := strings.Join(params, ", ")
		if paramsStr != "" {
			paramsStr = "{" + paramsStr + "}"
		}

		b.WriteByte('\n')
		b.WriteString(fmt.Sprintf("  Future<%s> %s(%s) async {\n", respCls, methodName, paramsStr))

		// Build cascade assignment — single field on one line, multiple fields multiline
		if len(cmd.RequestFields) <= 1 {
			if len(cmd.RequestFields) == 1 {
				propName := dartPropertyName(cmd.RequestFields[0].Name)
				b.WriteString(fmt.Sprintf("    final req = %s()..%s = %s;\n", reqCls, propName, propName))
			} else {
				b.WriteString(fmt.Sprintf("    final req = %s();\n", reqCls))
			}
		} else {
			b.WriteString(fmt.Sprintf("    final req = %s()\n", reqCls))
			for i, f := range cmd.RequestFields {
				propName := dartPropertyName(f.Name)
				if i < len(cmd.RequestFields)-1 {
					b.WriteString(fmt.Sprintf("      ..%s = %s\n", propName, propName))
				} else {
					b.WriteString(fmt.Sprintf("      ..%s = %s;\n", propName, propName))
				}
			}
		}

		b.WriteString("    final respData =\n")
		b.WriteString(fmt.Sprintf("        await call('%s', Uint8List.fromList(req.writeToBuffer()));\n", cmd.Snake))
		b.WriteString(fmt.Sprintf("    return %s.fromBuffer(respData);\n", respCls))
		b.WriteString("  }\n")
	}

	// Streaming methods
	for _, cmd := range commands {
		dir, ok := streaming[cmd.Snake]
		if !ok {
			continue
		}

		reqCls := cmd.RequestMsg
		respCls := cmd.ResponseMsg
		methodName := toLowerCamel(cmd.Camel)

		b.WriteByte('\n')

		if dir == "p2c" {
			var params []string
			for _, f := range cmd.RequestFields {
				dtType, ok := dartTypes[f.Type]
				if !ok {
					dtType = "dynamic"
				}
				def, ok := dartDefaults[f.Type]
				if !ok {
					def = "null"
				}
				propName := dartPropertyName(f.Name)
				params = append(params, fmt.Sprintf("%s %s = %s", dtType, propName, def))
			}
			paramsStr := strings.Join(params, ", ")
			if paramsStr != "" {
				paramsStr = "{" + paramsStr + "}"
			}

			b.WriteString(fmt.Sprintf("  Future<List<%s>> %s(%s) async {\n", respCls, methodName, paramsStr))

			if len(cmd.RequestFields) <= 1 {
				if len(cmd.RequestFields) == 1 {
					propName := dartPropertyName(cmd.RequestFields[0].Name)
					b.WriteString(fmt.Sprintf("    final req = %s()..%s = %s;\n", reqCls, propName, propName))
				} else {
					b.WriteString(fmt.Sprintf("    final req = %s();\n", reqCls))
				}
			} else {
				b.WriteString(fmt.Sprintf("    final req = %s()\n", reqCls))
				for i, f := range cmd.RequestFields {
					propName := dartPropertyName(f.Name)
					if i < len(cmd.RequestFields)-1 {
						b.WriteString(fmt.Sprintf("      ..%s = %s\n", propName, propName))
					} else {
						b.WriteString(fmt.Sprintf("      ..%s = %s;\n", propName, propName))
					}
				}
			}

			b.WriteString(fmt.Sprintf("    final responses = await streamReceive(\n"))
			b.WriteString(fmt.Sprintf("        '%s', Uint8List.fromList(req.writeToBuffer()));\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("    return responses.map((data) => %s.fromBuffer(data)).toList();\n", respCls))
			b.WriteString("  }\n")
		} else {
			b.WriteString(fmt.Sprintf("  Future<%s> %s(List<%s> messages) async {\n", respCls, methodName, reqCls))
			b.WriteString("    final raw =\n")
			b.WriteString("        messages.map((m) => Uint8List.fromList(m.writeToBuffer())).toList();\n")
			b.WriteString(fmt.Sprintf("    final respData = await streamSend('%s', raw, '%s');\n", cmd.Snake, cmd.Snake))
			b.WriteString(fmt.Sprintf("    return %s.fromBuffer(respData);\n", respCls))
			b.WriteString("  }\n")
		}
	}

	b.WriteString("}\n")

	return b.String()
}

func generateTsClient(commands []Command, streaming map[string]string) string {
	var b strings.Builder

	b.WriteString("/* Auto-generated by generate-handlers — DO NOT EDIT */\n")
	b.WriteString("import { blerpc } from '../proto/blerpc';\n")
	b.WriteByte('\n')
	b.WriteString("export abstract class GeneratedClient {\n")
	b.WriteString("  protected abstract call(cmdName: string, requestData: Uint8Array): Promise<Uint8Array>;\n")
	b.WriteString("  protected abstract streamReceive(cmdName: string, requestData: Uint8Array): Promise<Uint8Array[]>;\n")
	b.WriteString("  protected abstract streamSend(\n")
	b.WriteString("    cmdName: string,\n")
	b.WriteString("    messages: Uint8Array[],\n")
	b.WriteString("    finalCmdName: string,\n")
	b.WriteString("  ): Promise<Uint8Array>;\n")

	for _, cmd := range commands {
		if _, ok := streaming[cmd.Snake]; ok {
			continue
		}

		reqCls := "blerpc." + cmd.RequestMsg
		respCls := "blerpc." + cmd.ResponseMsg
		methodName := toLowerCamel(cmd.Camel)

		// Build parameters and type annotations
		var params []string
		var typeFields []string
		for _, f := range cmd.RequestFields {
			tsType, ok := tsTypes[f.Type]
			if !ok {
				tsType = "unknown"
			}
			def, ok := tsDefaults[f.Type]
			if !ok {
				def = "undefined"
			}
			propName := tsPropertyName(f.Name)
			params = append(params, fmt.Sprintf("%s = %s", propName, def))
			typeFields = append(typeFields, fmt.Sprintf("%s?: %s", propName, tsType))
		}

		b.WriteByte('\n')
		if len(cmd.RequestFields) > 0 {
			// Destructured parameter with defaults
			paramsStr := strings.Join(params, ", ")
			typeStr := strings.Join(typeFields, "; ")
			singleLine := fmt.Sprintf("  async %s({ %s }: { %s } = {}): Promise<%s> {",
				methodName, paramsStr, typeStr, respCls)
			if len(singleLine) <= 100 {
				b.WriteString(singleLine + "\n")
			} else {
				// Multi-line destructured parameters (Prettier-compatible)
				b.WriteString(fmt.Sprintf("  async %s({\n", methodName))
				for _, p := range params {
					b.WriteString(fmt.Sprintf("    %s,\n", p))
				}
				b.WriteString(fmt.Sprintf("  }: { %s } = {}): Promise<%s> {\n", typeStr, respCls))
			}
		} else {
			b.WriteString(fmt.Sprintf("  async %s(): Promise<%s> {\n", methodName, respCls))
		}

		// Create request
		if len(cmd.RequestFields) > 0 {
			var createFields []string
			for _, f := range cmd.RequestFields {
				propName := tsPropertyName(f.Name)
				createFields = append(createFields, propName)
			}
			b.WriteString(fmt.Sprintf("    const req = %s.create({ %s });\n", reqCls, strings.Join(createFields, ", ")))
		} else {
			b.WriteString(fmt.Sprintf("    const req = %s.create({});\n", reqCls))
		}

		b.WriteString(fmt.Sprintf("    const respData = await this.call('%s', %s.encode(req).finish());\n", cmd.Snake, reqCls))
		b.WriteString(fmt.Sprintf("    return %s.decode(respData);\n", respCls))
		b.WriteString("  }\n")
	}

	// Streaming methods
	for _, cmd := range commands {
		dir, ok := streaming[cmd.Snake]
		if !ok {
			continue
		}

		reqCls := "blerpc." + cmd.RequestMsg
		respCls := "blerpc." + cmd.ResponseMsg
		methodName := toLowerCamel(cmd.Camel)

		b.WriteByte('\n')

		if dir == "p2c" {
			var params []string
			var typeFields []string
			for _, f := range cmd.RequestFields {
				tsType, ok := tsTypes[f.Type]
				if !ok {
					tsType = "unknown"
				}
				def, ok := tsDefaults[f.Type]
				if !ok {
					def = "undefined"
				}
				propName := tsPropertyName(f.Name)
				params = append(params, fmt.Sprintf("%s = %s", propName, def))
				typeFields = append(typeFields, fmt.Sprintf("%s?: %s", propName, tsType))
			}

			if len(cmd.RequestFields) > 0 {
				paramsStr := strings.Join(params, ", ")
				typeStr := strings.Join(typeFields, "; ")
				singleLine := fmt.Sprintf("  async %s({ %s }: { %s } = {}): Promise<%s[]> {",
					methodName, paramsStr, typeStr, respCls)
				if len(singleLine) <= 100 {
					b.WriteString(singleLine + "\n")
				} else {
					b.WriteString(fmt.Sprintf("  async %s({\n", methodName))
					for _, p := range params {
						b.WriteString(fmt.Sprintf("    %s,\n", p))
					}
					b.WriteString(fmt.Sprintf("  }: { %s } = {}): Promise<%s[]> {\n", typeStr, respCls))
				}
			} else {
				b.WriteString(fmt.Sprintf("  async %s(): Promise<%s[]> {\n", methodName, respCls))
			}

			if len(cmd.RequestFields) > 0 {
				var createFields []string
				for _, f := range cmd.RequestFields {
					propName := tsPropertyName(f.Name)
					createFields = append(createFields, propName)
				}
				b.WriteString(fmt.Sprintf("    const req = %s.create({ %s });\n", reqCls, strings.Join(createFields, ", ")))
			} else {
				b.WriteString(fmt.Sprintf("    const req = %s.create({});\n", reqCls))
			}

			b.WriteString(fmt.Sprintf("    const responses = await this.streamReceive(\n"))
			b.WriteString(fmt.Sprintf("      '%s',\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("      %s.encode(req).finish(),\n", reqCls))
			b.WriteString("    );\n")
			b.WriteString(fmt.Sprintf("    return responses.map((data) => %s.decode(data));\n", respCls))
			b.WriteString("  }\n")
		} else {
			iReqCls := "blerpc.I" + cmd.RequestMsg
			b.WriteString(fmt.Sprintf("  async %s(messages: %s[]): Promise<%s> {\n", methodName, iReqCls, respCls))
			b.WriteString(fmt.Sprintf("    const raw = messages.map((m) =>\n"))
			b.WriteString(fmt.Sprintf("      %s.encode(%s.create(m)).finish(),\n", reqCls, reqCls))
			b.WriteString("    );\n")
			b.WriteString(fmt.Sprintf("    const respData = await this.streamSend('%s', raw, '%s');\n", cmd.Snake, cmd.Snake))
			b.WriteString(fmt.Sprintf("    return %s.decode(respData);\n", respCls))
			b.WriteString("  }\n")
		}
	}

	b.WriteString("}\n")

	return b.String()
}

// cParamStr formats a C type and parameter name, handling pointer types.
func cParamStr(cType, name string) string {
	if strings.HasSuffix(cType, "*") {
		return cType + name
	}
	return cType + " " + name
}

// cClientParams builds the parameter list for a C client function.
func cClientParams(cmd Command, streaming map[string]string, callbacks map[string]bool) []string {
	dir, isStreaming := streaming[cmd.Snake]
	reqMsg := "blerpc_" + cmd.RequestMsg
	respMsg := "blerpc_" + cmd.ResponseMsg

	if isStreaming && dir == "c2p" {
		return []string{
			fmt.Sprintf("const %s *messages", reqMsg),
			"size_t msg_count",
			fmt.Sprintf("%s *resp", respMsg),
		}
	}

	var params []string

	for _, f := range cmd.RequestFields {
		key := cmd.RequestMsg + "." + f.Name
		if callbacks[key] {
			params = append(params, fmt.Sprintf("const uint8_t *%s", f.Name))
			params = append(params, fmt.Sprintf("size_t %s_len", f.Name))
		} else {
			cType := cTypes[f.Type]
			if cType == "" {
				cType = "uint32_t"
			}
			params = append(params, cParamStr(cType, f.Name))
		}
	}

	hasCbReq := false
	for _, f := range cmd.RequestFields {
		if callbacks[cmd.RequestMsg+"."+f.Name] {
			hasCbReq = true
			break
		}
	}
	if hasCbReq {
		params = append(params, "uint8_t *work_buf", "size_t work_buf_size")
	}

	if isStreaming && dir == "p2c" {
		params = append(params,
			fmt.Sprintf("%s *results", respMsg),
			"size_t max_results",
			"size_t *result_count",
		)
	} else {
		params = append(params, fmt.Sprintf("%s *resp", respMsg))
		for _, f := range cmd.ResponseFields {
			key := cmd.ResponseMsg + "." + f.Name
			if callbacks[key] {
				params = append(params, fmt.Sprintf("uint8_t *%s_buf", f.Name))
				params = append(params, fmt.Sprintf("size_t %s_buf_size", f.Name))
				params = append(params, fmt.Sprintf("size_t *%s_len", f.Name))
			}
		}
	}

	return params
}

func generateCClientHeader(commands []Command, streaming map[string]string, callbacks map[string]bool) string {
	var b strings.Builder

	lines := []string{
		"/* Auto-generated by generate-handlers — DO NOT EDIT */",
		"#ifndef BLERPC_GENERATED_CLIENT_H",
		"#define BLERPC_GENERATED_CLIENT_H",
		"",
		`#include "blerpc.pb.h"`,
		"#include <pb_encode.h>",
		"#include <pb_decode.h>",
		"#include <stdint.h>",
		"#include <stddef.h>",
		"#include <stdbool.h>",
		"#include <string.h>",
		"",
		"#ifdef __cplusplus",
		`extern "C" {`,
		"#endif",
		"",
		"/* Callback for P2C streaming response payloads */",
		"typedef int (*blerpc_on_stream_resp_t)(const uint8_t *data, size_t len, void *ctx);",
		"",
		"/* Callback for C2P streaming message serialization */",
		"typedef int (*blerpc_next_msg_t)(size_t index, uint8_t *buf, size_t buf_size,",
		"                                 size_t *len, void *ctx);",
		"",
		"/* User-provided RPC transport functions */",
		"extern int blerpc_rpc_call(const char *cmd_name,",
		"                           const uint8_t *req_data, size_t req_len,",
		"                           uint8_t *resp_data, size_t resp_size, size_t *resp_len);",
		"",
		"extern int blerpc_stream_receive(const char *cmd_name,",
		"                                 const uint8_t *req_data, size_t req_len,",
		"                                 blerpc_on_stream_resp_t on_resp, void *ctx);",
		"",
		"extern int blerpc_stream_send(const char *cmd_name, size_t msg_count,",
		"                              blerpc_next_msg_t next_msg, void *msg_ctx,",
		"                              const char *final_cmd_name,",
		"                              uint8_t *resp_data, size_t resp_size, size_t *resp_len);",
		"",
		"/* Generated typed RPC functions */",
	}
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}

	for _, cmd := range commands {
		params := cClientParams(cmd, streaming, callbacks)
		b.WriteString(fmt.Sprintf("int blerpc_%s(%s);\n", cmd.Snake, strings.Join(params, ", ")))
	}

	tail := []string{
		"",
		"#ifdef __cplusplus",
		"}",
		"#endif",
		"",
		"#endif /* BLERPC_GENERATED_CLIENT_H */",
	}
	for _, l := range tail {
		b.WriteString(l)
		b.WriteByte('\n')
	}

	return b.String()
}

func generateCClientSource(commands []Command, streaming map[string]string, callbacks map[string]bool) string {
	var b strings.Builder

	b.WriteString("/* Auto-generated by generate-handlers — DO NOT EDIT */\n")
	b.WriteString("#include \"generated_client.h\"\n\n")

	// Check if we need helpers
	needDecode := false
	needEncode := false
	needRespBuf := false
	for _, cmd := range commands {
		if _, ok := streaming[cmd.Snake]; ok {
			continue
		}
		for _, f := range cmd.ResponseFields {
			if callbacks[cmd.ResponseMsg+"."+f.Name] {
				needDecode = true
				needRespBuf = true
			}
		}
		for _, f := range cmd.RequestFields {
			if callbacks[cmd.RequestMsg+"."+f.Name] {
				needEncode = true
			}
		}
	}

	// Internal response buffer for FT_CALLBACK response commands
	if needRespBuf {
		b.WriteString("#ifndef BLERPC_GENERATED_RESP_BUF_SIZE\n")
		b.WriteString("#define BLERPC_GENERATED_RESP_BUF_SIZE 4096\n")
		b.WriteString("#endif\n")
		b.WriteString("static uint8_t _blerpc_resp_buf[BLERPC_GENERATED_RESP_BUF_SIZE];\n\n")
	}

	if needDecode {
		b.WriteString("/* Decode context for FT_CALLBACK bytes fields */\n")
		b.WriteString("struct _blerpc_bytes_decode_ctx {\n")
		b.WriteString("    uint8_t *buf;\n")
		b.WriteString("    size_t buf_size;\n")
		b.WriteString("    size_t decoded_len;\n")
		b.WriteString("};\n\n")
		b.WriteString("static bool _blerpc_decode_bytes_cb(pb_istream_t *stream,\n")
		b.WriteString("                                     const pb_field_t *field, void **arg)\n")
		b.WriteString("{\n")
		b.WriteString("    (void)field;\n")
		b.WriteString("    struct _blerpc_bytes_decode_ctx *ctx =\n")
		b.WriteString("        (struct _blerpc_bytes_decode_ctx *)*arg;\n")
		b.WriteString("    size_t len = stream->bytes_left;\n")
		b.WriteString("    if (len > ctx->buf_size - ctx->decoded_len) return false;\n")
		b.WriteString("    if (!pb_read(stream, ctx->buf + ctx->decoded_len, len)) return false;\n")
		b.WriteString("    ctx->decoded_len += len;\n")
		b.WriteString("    return true;\n")
		b.WriteString("}\n\n")
	}

	if needEncode {
		b.WriteString("/* Encode context for FT_CALLBACK bytes fields */\n")
		b.WriteString("struct _blerpc_bytes_encode_ctx {\n")
		b.WriteString("    const uint8_t *data;\n")
		b.WriteString("    size_t data_len;\n")
		b.WriteString("};\n\n")
		b.WriteString("static bool _blerpc_encode_bytes_cb(pb_ostream_t *stream,\n")
		b.WriteString("                                     const pb_field_t *field,\n")
		b.WriteString("                                     void *const *arg)\n")
		b.WriteString("{\n")
		b.WriteString("    const struct _blerpc_bytes_encode_ctx *ctx =\n")
		b.WriteString("        *(const struct _blerpc_bytes_encode_ctx **)arg;\n")
		b.WriteString("    if (!pb_encode_tag_for_field(stream, field)) return false;\n")
		b.WriteString("    if (!pb_encode_varint(stream, ctx->data_len)) return false;\n")
		b.WriteString("    return pb_write(stream, ctx->data, ctx->data_len);\n")
		b.WriteString("}\n\n")
	}

	// Per-command generation
	for _, cmd := range commands {
		dir, isStreaming := streaming[cmd.Snake]
		reqMsg := "blerpc_" + cmd.RequestMsg
		respMsg := "blerpc_" + cmd.ResponseMsg
		params := cClientParams(cmd, streaming, callbacks)

		if isStreaming && dir == "p2c" {
			// P2C streaming: callback struct + on_resp function + main function
			b.WriteString(fmt.Sprintf("struct _blerpc_%s_ctx {\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("    %s *results;\n", respMsg))
			b.WriteString("    size_t max_results;\n")
			b.WriteString("    size_t count;\n")
			b.WriteString("};\n\n")

			b.WriteString(fmt.Sprintf("static int _blerpc_%s_on_resp(const uint8_t *data, size_t len,\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("                              %svoid *ctx)\n", strings.Repeat(" ", len(cmd.Snake))))
			b.WriteString("{\n")
			b.WriteString(fmt.Sprintf("    struct _blerpc_%s_ctx *c = (struct _blerpc_%s_ctx *)ctx;\n", cmd.Snake, cmd.Snake))
			b.WriteString("    if (c->count >= c->max_results) return -1;\n")
			b.WriteString(fmt.Sprintf("    c->results[c->count] = (%s)%s_init_zero;\n", respMsg, respMsg))
			b.WriteString("    pb_istream_t istream = pb_istream_from_buffer(data, len);\n")
			b.WriteString(fmt.Sprintf("    if (!pb_decode(&istream, %s_fields, &c->results[c->count])) return -1;\n", respMsg))
			b.WriteString("    c->count++;\n")
			b.WriteString("    return 0;\n")
			b.WriteString("}\n\n")

			b.WriteString(fmt.Sprintf("int blerpc_%s(%s)\n", cmd.Snake, strings.Join(params, ", ")))
			b.WriteString("{\n")
			b.WriteString(fmt.Sprintf("    %s req = %s_init_zero;\n", reqMsg, reqMsg))
			for _, f := range cmd.RequestFields {
				if f.Type == "string" {
					b.WriteString(fmt.Sprintf("    strncpy(req.%s, %s, sizeof(req.%s) - 1);\n", f.Name, f.Name, f.Name))
				} else {
					b.WriteString(fmt.Sprintf("    req.%s = %s;\n", f.Name, f.Name))
				}
			}
			b.WriteByte('\n')
			b.WriteString(fmt.Sprintf("    uint8_t req_buf[%s_size];\n", reqMsg))
			b.WriteString("    pb_ostream_t ostream = pb_ostream_from_buffer(req_buf, sizeof(req_buf));\n")
			b.WriteString(fmt.Sprintf("    if (!pb_encode(&ostream, %s_fields, &req)) return -1;\n", reqMsg))
			b.WriteByte('\n')
			b.WriteString(fmt.Sprintf("    struct _blerpc_%s_ctx ctx = {\n", cmd.Snake))
			b.WriteString("        .results = results, .max_results = max_results, .count = 0\n")
			b.WriteString("    };\n")
			b.WriteString(fmt.Sprintf("    if (blerpc_stream_receive(\"%s\", req_buf, ostream.bytes_written,\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("                              _blerpc_%s_on_resp, &ctx) != 0) return -1;\n", cmd.Snake))
			b.WriteByte('\n')
			b.WriteString("    *result_count = ctx.count;\n")
			b.WriteString("    return 0;\n")
			b.WriteString("}\n\n")

		} else if isStreaming && dir == "c2p" {
			// C2P streaming: next_msg struct + callback + main function
			b.WriteString(fmt.Sprintf("struct _blerpc_%s_ctx {\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("    const %s *messages;\n", reqMsg))
			b.WriteString("};\n\n")

			b.WriteString(fmt.Sprintf("static int _blerpc_%s_next(size_t index, uint8_t *buf,\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("                           %ssize_t buf_size, size_t *len, void *ctx)\n", strings.Repeat(" ", len(cmd.Snake))))
			b.WriteString("{\n")
			b.WriteString(fmt.Sprintf("    struct _blerpc_%s_ctx *c = (struct _blerpc_%s_ctx *)ctx;\n", cmd.Snake, cmd.Snake))
			b.WriteString("    pb_ostream_t ostream = pb_ostream_from_buffer(buf, buf_size);\n")
			b.WriteString(fmt.Sprintf("    if (!pb_encode(&ostream, %s_fields, &c->messages[index])) return -1;\n", reqMsg))
			b.WriteString("    *len = ostream.bytes_written;\n")
			b.WriteString("    return 0;\n")
			b.WriteString("}\n\n")

			b.WriteString(fmt.Sprintf("int blerpc_%s(%s)\n", cmd.Snake, strings.Join(params, ", ")))
			b.WriteString("{\n")
			b.WriteString(fmt.Sprintf("    struct _blerpc_%s_ctx ctx = { .messages = messages };\n", cmd.Snake))
			b.WriteByte('\n')
			b.WriteString(fmt.Sprintf("    uint8_t resp_buf[%s_size];\n", respMsg))
			b.WriteString("    size_t resp_len;\n")
			b.WriteString(fmt.Sprintf("    if (blerpc_stream_send(\"%s\", msg_count,\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("                           _blerpc_%s_next, &ctx,\n", cmd.Snake))
			b.WriteString(fmt.Sprintf("                           \"%s\", resp_buf, sizeof(resp_buf),\n", cmd.Snake))
			b.WriteString("                           &resp_len) != 0) return -1;\n")
			b.WriteByte('\n')
			b.WriteString(fmt.Sprintf("    *resp = (%s)%s_init_zero;\n", respMsg, respMsg))
			b.WriteString("    pb_istream_t istream = pb_istream_from_buffer(resp_buf, resp_len);\n")
			b.WriteString(fmt.Sprintf("    if (!pb_decode(&istream, %s_fields, resp)) return -1;\n", respMsg))
			b.WriteByte('\n')
			b.WriteString("    return 0;\n")
			b.WriteString("}\n\n")

		} else {
			// Unary command
			hasCbReq := false
			hasCbResp := false
			for _, f := range cmd.RequestFields {
				if callbacks[cmd.RequestMsg+"."+f.Name] {
					hasCbReq = true
				}
			}
			for _, f := range cmd.ResponseFields {
				if callbacks[cmd.ResponseMsg+"."+f.Name] {
					hasCbResp = true
				}
			}

			b.WriteString(fmt.Sprintf("int blerpc_%s(%s)\n", cmd.Snake, strings.Join(params, ", ")))
			b.WriteString("{\n")

			// Encode context setup for FT_CALLBACK request fields
			if hasCbReq {
				for _, f := range cmd.RequestFields {
					if callbacks[cmd.RequestMsg+"."+f.Name] {
						b.WriteString(fmt.Sprintf("    struct _blerpc_bytes_encode_ctx _%s_ctx = {\n", f.Name))
						b.WriteString(fmt.Sprintf("        .data = %s, .data_len = %s_len\n", f.Name, f.Name))
						b.WriteString("    };\n")
					}
				}
			}

			// Init request
			b.WriteString(fmt.Sprintf("    %s req = %s_init_zero;\n", reqMsg, reqMsg))

			// Set request fields
			for _, f := range cmd.RequestFields {
				key := cmd.RequestMsg + "." + f.Name
				if callbacks[key] {
					b.WriteString(fmt.Sprintf("    req.%s.funcs.encode = _blerpc_encode_bytes_cb;\n", f.Name))
					b.WriteString(fmt.Sprintf("    req.%s.arg = &_%s_ctx;\n", f.Name, f.Name))
				} else if f.Type == "string" {
					b.WriteString(fmt.Sprintf("    strncpy(req.%s, %s, sizeof(req.%s) - 1);\n", f.Name, f.Name, f.Name))
				} else {
					b.WriteString(fmt.Sprintf("    req.%s = %s;\n", f.Name, f.Name))
				}
			}
			b.WriteByte('\n')

			// Encode request
			if hasCbReq {
				b.WriteString(fmt.Sprintf("    pb_ostream_t sizing = PB_OSTREAM_SIZING;\n"))
				b.WriteString(fmt.Sprintf("    if (!pb_encode(&sizing, %s_fields, &req)) return -1;\n", reqMsg))
				b.WriteString("    if (sizing.bytes_written > work_buf_size) return -1;\n")
				b.WriteByte('\n')
				b.WriteString("    pb_ostream_t ostream = pb_ostream_from_buffer(work_buf, work_buf_size);\n")
				b.WriteString(fmt.Sprintf("    if (!pb_encode(&ostream, %s_fields, &req)) return -1;\n", reqMsg))
			} else {
				b.WriteString(fmt.Sprintf("    uint8_t req_buf[%s_size];\n", reqMsg))
				b.WriteString("    pb_ostream_t ostream = pb_ostream_from_buffer(req_buf, sizeof(req_buf));\n")
				b.WriteString(fmt.Sprintf("    if (!pb_encode(&ostream, %s_fields, &req)) return -1;\n", reqMsg))
			}
			b.WriteByte('\n')

			// RPC call
			reqBufName := "req_buf"
			if hasCbReq {
				reqBufName = "work_buf"
			}
			if hasCbResp {
				b.WriteString("    size_t resp_len;\n")
				b.WriteString(fmt.Sprintf("    if (blerpc_rpc_call(\"%s\", %s, ostream.bytes_written,\n", cmd.Snake, reqBufName))
				b.WriteString("                        _blerpc_resp_buf, sizeof(_blerpc_resp_buf),\n")
				b.WriteString("                        &resp_len) != 0) return -1;\n")
			} else {
				b.WriteString(fmt.Sprintf("    uint8_t resp_buf[%s_size];\n", respMsg))
				b.WriteString("    size_t resp_len;\n")
				b.WriteString(fmt.Sprintf("    if (blerpc_rpc_call(\"%s\", %s, ostream.bytes_written,\n", cmd.Snake, reqBufName))
				b.WriteString("                        resp_buf, sizeof(resp_buf), &resp_len) != 0) return -1;\n")
			}
			b.WriteByte('\n')

			// Decode context for FT_CALLBACK response fields
			if hasCbResp {
				for _, f := range cmd.ResponseFields {
					if callbacks[cmd.ResponseMsg+"."+f.Name] {
						b.WriteString(fmt.Sprintf("    struct _blerpc_bytes_decode_ctx _%s_ctx = {\n", f.Name))
						b.WriteString(fmt.Sprintf("        .buf = %s_buf, .buf_size = %s_buf_size, .decoded_len = 0\n", f.Name, f.Name))
						b.WriteString("    };\n")
					}
				}
			}

			// Decode response
			b.WriteString(fmt.Sprintf("    *resp = (%s)%s_init_zero;\n", respMsg, respMsg))
			if hasCbResp {
				for _, f := range cmd.ResponseFields {
					if callbacks[cmd.ResponseMsg+"."+f.Name] {
						b.WriteString(fmt.Sprintf("    resp->%s.funcs.decode = _blerpc_decode_bytes_cb;\n", f.Name))
						b.WriteString(fmt.Sprintf("    resp->%s.arg = &_%s_ctx;\n", f.Name, f.Name))
					}
				}
			}

			isBuf := "_blerpc_resp_buf"
			if !hasCbResp {
				isBuf = "resp_buf"
			}
			b.WriteString(fmt.Sprintf("    pb_istream_t istream = pb_istream_from_buffer(%s, resp_len);\n", isBuf))
			b.WriteString(fmt.Sprintf("    if (!pb_decode(&istream, %s_fields, resp)) return -1;\n", respMsg))

			// Set output lengths for FT_CALLBACK response fields
			if hasCbResp {
				b.WriteByte('\n')
				for _, f := range cmd.ResponseFields {
					if callbacks[cmd.ResponseMsg+"."+f.Name] {
						b.WriteString(fmt.Sprintf("    *%s_len = _%s_ctx.decoded_len;\n", f.Name, f.Name))
					}
				}
			}

			b.WriteByte('\n')
			b.WriteString("    return 0;\n")
			b.WriteString("}\n\n")
		}
	}

	return b.String()
}

func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func main() {
	root := flag.String("root", ".", "project root directory")
	flag.Parse()

	protoFile := filepath.Join(*root, "proto", "blerpc.proto")
	optionsFile := filepath.Join(*root, "proto", "blerpc.options")
	streamingFile := filepath.Join(*root, "proto", "streaming.txt")

	outCHeader := filepath.Join(*root, "peripheral_fw", "src", "generated_handlers.h")
	outCSource := filepath.Join(*root, "peripheral_fw", "src", "generated_handlers.c")
	outPyHandlers := filepath.Join(*root, "peripheral_py", "generated_handlers.py")
	outPyClient := filepath.Join(*root, "central_py", "blerpc", "generated", "generated_client.py")
	outKtClient := filepath.Join(*root, "central_android", "app", "src", "main", "java", "com", "blerpc", "android", "client", "GeneratedClient.kt")
	outSwiftClient := filepath.Join(*root, "central_ios", "BlerpcCentral", "Client", "GeneratedClient.swift")
	outDartClient := filepath.Join(*root, "central_flutter", "lib", "client", "generated_client.dart")
	outTsClient := filepath.Join(*root, "central_rn", "src", "client", "GeneratedClient.ts")
	outCClientHeader := filepath.Join(*root, "central_fw", "src", "generated_client.h")
	outCClientSource := filepath.Join(*root, "central_fw", "src", "generated_client.c")

	messages, err := parseProto(protoFile)
	if err != nil {
		log.Fatalf("Failed to parse proto: %v", err)
	}

	callbacks, err := parseOptions(optionsFile)
	if err != nil {
		log.Fatalf("Failed to parse options: %v", err)
	}

	streaming, err := parseStreamingCommands(streamingFile)
	if err != nil {
		log.Fatalf("Failed to parse streaming commands: %v", err)
	}

	commands := discoverCommands(messages)
	if len(commands) == 0 {
		fmt.Fprintln(os.Stderr, "No Request/Response pairs found in proto file.")
		os.Exit(1)
	}

	names := make([]string, len(commands))
	for i, c := range commands {
		names[i] = c.Snake
	}
	fmt.Printf("Found %d commands: %s\n", len(commands), strings.Join(names, ", "))

	outputs := []struct {
		path    string
		content string
	}{
		{outCHeader, generateCHeader(commands)},
		{outCSource, generateCSource(commands, callbacks)},
		{outPyHandlers, generatePyHandlers(commands)},
		{outPyClient, generatePyClient(commands, streaming)},
		{outKtClient, generateKotlinClient(commands, streaming)},
		{outSwiftClient, generateSwiftClient(commands, streaming)},
		{outDartClient, generateDartClient(commands, streaming)},
		{outTsClient, generateTsClient(commands, streaming)},
		{outCClientHeader, generateCClientHeader(commands, streaming, callbacks)},
		{outCClientSource, generateCClientSource(commands, streaming, callbacks)},
	}

	for _, out := range outputs {
		if err := writeFile(out.path, out.content); err != nil {
			log.Fatalf("Failed to write %s: %v", out.path, err)
		}
		rel, _ := filepath.Rel(*root, out.path)
		fmt.Printf("  Generated %s\n", rel)
	}
}
