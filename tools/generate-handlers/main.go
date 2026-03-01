// Generate handler stubs and client code from a proto file.
//
// Parses proto file with go-protoparser (proper AST) and generates code for
// multiple target platforms. All output paths are configurable via CLI flags.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// flagOrDefault returns the flag value if non-empty, otherwise the default.
func flagOrDefault(flagVal, defaultVal string) string {
	if flagVal != "" {
		return flagVal
	}
	return defaultVal
}

func main() {
	root := flag.String("root", ".", "project root directory")

	// Input flags
	protoFlag := flag.String("proto", "", "path to .proto file (default: <root>/proto/blerpc.proto)")
	optionsFlag := flag.String("options", "", "path to .options file (default: <root>/proto/blerpc.options)")
	streamingFlag := flag.String("streaming", "", "path to streaming.txt (default: <root>/proto/streaming.txt)")

	// Import path flags
	protoPathDirs := flag.String("proto-path", "", "comma-separated proto import search paths")

	// Output flags
	outCHeaderFlag := flag.String("out-c-header", "", "C handler header output path")
	outCSourceFlag := flag.String("out-c-source", "", "C handler source output path")
	outPyHandlersFlag := flag.String("out-py-handlers", "", "Python handlers output path")
	outPyClientFlag := flag.String("out-py-client", "", "Python client output path")
	outKtClientFlag := flag.String("out-kt-client", "", "Kotlin client output path")
	outSwiftClientFlag := flag.String("out-swift-client", "", "Swift client output path")
	outDartClientFlag := flag.String("out-dart-client", "", "Dart client output path")
	outTsClientFlag := flag.String("out-ts-client", "", "TypeScript client output path")
	outCClientHeaderFlag := flag.String("out-c-client-header", "", "C client header output path")
	outCClientSourceFlag := flag.String("out-c-client-source", "", "C client source output path")

	flag.Parse()

	protoPath := flagOrDefault(*protoFlag, filepath.Join(*root, "proto", "blerpc.proto"))
	optionsFile := flagOrDefault(*optionsFlag, filepath.Join(*root, "proto", "blerpc.options"))
	streamingFile := flagOrDefault(*streamingFlag, filepath.Join(*root, "proto", "streaming.txt"))

	outCHeader := flagOrDefault(*outCHeaderFlag, filepath.Join(*root, "peripheral_fw", "src", "generated_handlers.h"))
	outCSource := flagOrDefault(*outCSourceFlag, filepath.Join(*root, "peripheral_fw", "src", "generated_handlers.c"))
	outPyHandlers := flagOrDefault(*outPyHandlersFlag, filepath.Join(*root, "peripheral_py", "generated_handlers.py"))
	outPyClient := flagOrDefault(*outPyClientFlag, filepath.Join(*root, "central_py", "blerpc", "generated", "generated_client.py"))
	outKtClient := flagOrDefault(*outKtClientFlag, filepath.Join(*root, "central_android", "app", "src", "main", "java", "com", "blerpc", "android", "client", "GeneratedClient.kt"))
	outSwiftClient := flagOrDefault(*outSwiftClientFlag, filepath.Join(*root, "central_ios", "BlerpcCentral", "Client", "GeneratedClient.swift"))
	outDartClient := flagOrDefault(*outDartClientFlag, filepath.Join(*root, "central_flutter", "lib", "client", "generated_client.dart"))
	outTsClient := flagOrDefault(*outTsClientFlag, filepath.Join(*root, "central_rn", "src", "client", "GeneratedClient.ts"))
	outCClientHeader := flagOrDefault(*outCClientHeaderFlag, filepath.Join(*root, "central_fw", "src", "generated_client.h"))
	outCClientSource := flagOrDefault(*outCClientSourceFlag, filepath.Join(*root, "central_fw", "src", "generated_client.c"))

	var importPaths []string
	if *protoPathDirs != "" {
		importPaths = strings.Split(*protoPathDirs, ",")
	}

	protoFile, err := parseProtoWithImports(protoPath, importPaths)
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

	pkg := protoFile.Package
	if pkg == "" {
		pkg = "blerpc"
	}

	// Discover commands: prefer service definitions, fall back to naming convention
	var commands []Command
	if len(protoFile.Services) > 0 {
		msgByName := make(map[string]Message)
		for _, m := range protoFile.Messages {
			msgByName[m.Name] = m
		}
		commands = discoverCommandsFromServices(protoFile.Services, msgByName)
		// Merge streaming info from service definitions into the streaming map
		svcStreaming := streamingFromServices(protoFile.Services)
		for k, v := range svcStreaming {
			if _, exists := streaming[k]; !exists {
				streaming[k] = v
			}
		}
	} else {
		commands = discoverCommands(protoFile.Messages)
	}
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
		{outCHeader, generateCHeader(commands, pkg)},
		{outCSource, generateCSource(commands, callbacks, pkg)},
		{outPyHandlers, generatePyHandlers(commands, pkg)},
		{outPyClient, generatePyClient(commands, streaming, pkg)},
		{outKtClient, generateKotlinClient(commands, streaming, pkg)},
		{outSwiftClient, generateSwiftClient(commands, streaming, pkg)},
		{outDartClient, generateDartClient(commands, streaming, pkg)},
		{outTsClient, generateTsClient(commands, streaming, pkg)},
		{outCClientHeader, generateCClientHeader(commands, streaming, callbacks, pkg)},
		{outCClientSource, generateCClientSource(commands, streaming, callbacks, pkg)},
	}

	for _, out := range outputs {
		if err := writeFile(out.path, out.content); err != nil {
			log.Fatalf("Failed to write %s: %v", out.path, err)
		}
		rel, _ := filepath.Rel(*root, out.path)
		fmt.Printf("  Generated %s\n", rel)
	}
}
