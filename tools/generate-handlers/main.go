// Generate handler stubs and client code from blerpc.proto.
//
// Parses proto file with go-protoparser (proper AST) and generates:
//   - peripheral_fw/src/generated_handlers.h  — C declarations + handler_entry + lookup
//   - peripheral_fw/src/generated_handlers.c  — weak handler stubs + handler table
//   - peripheral_py/generated_handlers.py  — Python handler stubs + HANDLERS dict
//   - central_py/blerpc/generated/generated_client.py — Python client mixin class
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

func parseStreamingCommands(path string) (map[string]bool, error) {
	streaming := make(map[string]bool)
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
		streaming[line] = true
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

func generatePyClient(commands []Command, streaming map[string]bool) string {
	var b strings.Builder

	b.WriteString("\"\"\"Auto-generated by generate-handlers — DO NOT EDIT.\"\"\"\n")
	b.WriteByte('\n')
	b.WriteString("from __future__ import annotations\n")
	b.WriteByte('\n')
	b.WriteString("from . import blerpc_pb2\n")
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString("class GeneratedClientMixin:\n")
	b.WriteString("    \"\"\"Auto-generated unary RPC methods.\n")
	b.WriteByte('\n')
	b.WriteString("    Streaming RPCs are implemented manually in BlerpcClient.\n")
	b.WriteString("    \"\"\"\n")
	b.WriteByte('\n')

	first := true
	for _, cmd := range commands {
		if streaming[cmd.Snake] {
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

func generateKotlinClient(commands []Command, streaming map[string]bool) string {
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
		if streaming[cmd.Snake] {
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
				def = "TODO()"
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

func generateSwiftClient(commands []Command, streaming map[string]bool) string {
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
		if streaming[cmd.Snake] {
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

	b.WriteString("}\n")

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
	}

	for _, out := range outputs {
		if err := writeFile(out.path, out.content); err != nil {
			log.Fatalf("Failed to write %s: %v", out.path, err)
		}
		rel, _ := filepath.Rel(*root, out.path)
		fmt.Printf("  Generated %s\n", rel)
	}
}
