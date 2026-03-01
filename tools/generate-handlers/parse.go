package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// ProtoFile holds the parsed result of a proto file.
type ProtoFile struct {
	Package  string
	Messages []Message
	Enums    []Enum
	Services []Service
	Imports  []string // import paths (for recursive resolution)
}

// collectEnums extracts enum definitions from parser enum body items.
func collectEnums(e *parser.Enum) Enum {
	en := Enum{Name: e.EnumName}
	for _, body := range e.EnumBody {
		ef, ok := body.(*parser.EnumField)
		if !ok {
			continue
		}
		num := 0
		_, _ = fmt.Sscanf(ef.Number, "%d", &num)
		en.Values = append(en.Values, EnumValue{
			Name:   ef.Ident,
			Number: num,
		})
	}
	return en
}

func parseProtoReader(r io.Reader) (*ProtoFile, error) {
	proto, err := protoparser.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse proto: %w", err)
	}

	// Extract package name and imports
	var pkgName string
	var imports []string
	for _, item := range proto.ProtoBody {
		if pkg, ok := item.(*parser.Package); ok {
			pkgName = pkg.Name
		}
		if imp, ok := item.(*parser.Import); ok {
			loc := strings.Trim(imp.Location, "\"")
			imports = append(imports, loc)
		}
	}

	// Collect all enums (top-level + nested inside messages)
	enumSet := make(map[string]bool)
	msgSet := make(map[string]bool)

	var enums []Enum
	for _, item := range proto.ProtoBody {
		if e, ok := item.(*parser.Enum); ok {
			en := collectEnums(e)
			enums = append(enums, en)
			enumSet[en.Name] = true
		}
	}

	// Collect message names and nested enums/messages
	for _, item := range proto.ProtoBody {
		msg, ok := item.(*parser.Message)
		if !ok {
			continue
		}
		msgSet[msg.MessageName] = true
		for _, body := range msg.MessageBody {
			if e, ok := body.(*parser.Enum); ok {
				en := collectEnums(e)
				enums = append(enums, en)
				enumSet[en.Name] = true
			}
			if nested, ok := body.(*parser.Message); ok {
				msgSet[nested.MessageName] = true
			}
		}
	}

	var messages []Message
	for _, item := range proto.ProtoBody {
		msg, ok := item.(*parser.Message)
		if !ok {
			continue
		}
		m := Message{Name: msg.MessageName}
		for _, body := range msg.MessageBody {
			switch f := body.(type) {
			case *parser.Field:
				num := 0
				_, _ = fmt.Sscanf(f.FieldNumber, "%d", &num)
				m.Fields = append(m.Fields, Field{
					Type:       f.Type,
					Name:       f.FieldName,
					Number:     num,
					IsEnum:     enumSet[f.Type],
					IsRepeated: f.IsRepeated,
					IsMessage:  msgSet[f.Type],
				})
			case *parser.MapField:
				num := 0
				_, _ = fmt.Sscanf(f.FieldNumber, "%d", &num)
				m.Fields = append(m.Fields, Field{
					Name:      f.MapName,
					Number:    num,
					IsMap:     true,
					KeyType:   f.KeyType,
					ValueType: f.Type,
				})
			case *parser.Oneof:
				og := OneofGroup{Name: f.OneofName}
				for _, of := range f.OneofFields {
					num := 0
					_, _ = fmt.Sscanf(of.FieldNumber, "%d", &num)
					field := Field{
						Type:      of.Type,
						Name:      of.FieldName,
						Number:    num,
						IsEnum:    enumSet[of.Type],
						IsMessage: msgSet[of.Type],
					}
					og.Fields = append(og.Fields, field)
					// Also add oneof fields to the message's flat field list
					m.Fields = append(m.Fields, field)
				}
				m.Oneofs = append(m.Oneofs, og)
			}
		}
		messages = append(messages, m)
	}
	// Collect service definitions
	var services []Service
	for _, item := range proto.ProtoBody {
		svc, ok := item.(*parser.Service)
		if !ok {
			continue
		}
		s := Service{Name: svc.ServiceName}
		for _, body := range svc.ServiceBody {
			rpc, ok := body.(*parser.RPC)
			if !ok {
				continue
			}
			sr := ServiceRPC{
				Name:         rpc.RPCName,
				RequestType:  rpc.RPCRequest.MessageType,
				ResponseType: rpc.RPCResponse.MessageType,
				ClientStream: rpc.RPCRequest.IsStream,
				ServerStream: rpc.RPCResponse.IsStream,
			}
			s.RPCs = append(s.RPCs, sr)
		}
		services = append(services, s)
	}

	return &ProtoFile{Package: pkgName, Messages: messages, Enums: enums, Services: services, Imports: imports}, nil
}

// parseProtoWithImports parses a proto file and recursively resolves imports.
// protoPaths are additional directories to search for imported files.
func parseProtoWithImports(path string, protoPaths []string) (*ProtoFile, error) {
	visited := make(map[string]bool)
	return parseProtoRecursive(path, protoPaths, visited)
}

func parseProtoRecursive(path string, protoPaths []string, visited map[string]bool) (*ProtoFile, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("abs path: %w", err)
	}
	if visited[absPath] {
		return &ProtoFile{}, nil // already parsed, skip
	}
	visited[absPath] = true

	reader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open proto: %w", err)
	}
	defer reader.Close()

	pf, err := parseProtoReader(reader)
	if err != nil {
		return nil, err
	}

	// Resolve imports
	protoDir := filepath.Dir(path)
	searchPaths := append([]string{protoDir}, protoPaths...)

	for _, imp := range pf.Imports {
		impPath := resolveImportPath(imp, searchPaths)
		if impPath == "" {
			continue // skip unresolvable imports (e.g. google/protobuf/*)
		}
		imported, err := parseProtoRecursive(impPath, protoPaths, visited)
		if err != nil {
			return nil, fmt.Errorf("import %q: %w", imp, err)
		}
		// Merge imported types into this proto file
		pf.Messages = append(pf.Messages, imported.Messages...)
		pf.Enums = append(pf.Enums, imported.Enums...)
		pf.Services = append(pf.Services, imported.Services...)
	}

	return pf, nil
}

// resolveImportPath finds the file for an import path across search directories.
func resolveImportPath(importLoc string, searchPaths []string) string {
	for _, dir := range searchPaths {
		candidate := filepath.Join(dir, importLoc)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
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

// streamingFromServices derives streaming directions from service RPC definitions.
// server stream → p2c (peripheral-to-central), client stream → c2p (central-to-peripheral).
func streamingFromServices(services []Service) map[string]string {
	streaming := make(map[string]string)
	for _, svc := range services {
		for _, rpc := range svc.RPCs {
			snake := camelToSnake(rpc.Name)
			if rpc.ServerStream && !rpc.ClientStream {
				streaming[snake] = "p2c"
			} else if rpc.ClientStream && !rpc.ServerStream {
				streaming[snake] = "c2p"
			}
			// bidirectional streaming not supported yet; unary has no entry
		}
	}
	return streaming
}

// discoverCommandsFromServices builds commands from service RPC definitions.
func discoverCommandsFromServices(services []Service, msgByName map[string]Message) []Command {
	var commands []Command
	for _, svc := range services {
		for _, rpc := range svc.RPCs {
			reqMsg, reqOk := msgByName[rpc.RequestType]
			respMsg, respOk := msgByName[rpc.ResponseType]
			if !reqOk || !respOk {
				continue
			}
			commands = append(commands, Command{
				Camel:          rpc.Name,
				Snake:          camelToSnake(rpc.Name),
				RequestMsg:     rpc.RequestType,
				ResponseMsg:    rpc.ResponseType,
				RequestFields:  reqMsg.Fields,
				ResponseFields: respMsg.Fields,
			})
		}
	}
	return commands
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
