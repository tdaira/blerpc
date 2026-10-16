package main

// kotlinTypes maps proto field types to Kotlin types.
var kotlinTypes = map[string]string{
	"string":   "String",
	"bytes":    "com.google.protobuf.ByteString",
	"uint32":   "Int",
	"int32":    "Int",
	"uint64":   "Long",
	"int64":    "Long",
	"float":    "Float",
	"double":   "Double",
	"bool":     "Boolean",
	"sint32":   "Int",
	"sint64":   "Long",
	"fixed32":  "Int",
	"fixed64":  "Long",
	"sfixed32": "Int",
	"sfixed64": "Long",
}

// kotlinDefaults maps proto field types to Kotlin default values.
var kotlinDefaults = map[string]string{
	"string":   "\"\"",
	"bytes":    "com.google.protobuf.ByteString.EMPTY",
	"uint32":   "0",
	"int32":    "0",
	"uint64":   "0L",
	"int64":    "0L",
	"float":    "0.0f",
	"double":   "0.0",
	"bool":     "false",
	"sint32":   "0",
	"sint64":   "0L",
	"fixed32":  "0",
	"fixed64":  "0L",
	"sfixed32": "0",
	"sfixed64": "0L",
}

// swiftTypes maps proto field types to Swift types.
var swiftTypes = map[string]string{
	"string":   "String",
	"bytes":    "Data",
	"uint32":   "UInt32",
	"int32":    "Int32",
	"uint64":   "UInt64",
	"int64":    "Int64",
	"float":    "Float",
	"double":   "Double",
	"bool":     "Bool",
	"sint32":   "Int32",
	"sint64":   "Int64",
	"fixed32":  "UInt32",
	"fixed64":  "UInt64",
	"sfixed32": "Int32",
	"sfixed64": "Int64",
}

// swiftDefaults maps proto field types to Swift default values.
var swiftDefaults = map[string]string{
	"string":   "\"\"",
	"bytes":    "Data()",
	"uint32":   "0",
	"int32":    "0",
	"uint64":   "0",
	"int64":    "0",
	"float":    "0.0",
	"double":   "0.0",
	"bool":     "false",
	"sint32":   "0",
	"sint64":   "0",
	"fixed32":  "0",
	"fixed64":  "0",
	"sfixed32": "0",
	"sfixed64": "0",
}

// dartTypes maps proto field types to Dart types.
var dartTypes = map[string]string{
	"string":   "String",
	"bytes":    "List<int>",
	"uint32":   "int",
	"int32":    "int",
	"uint64":   "int",
	"int64":    "int",
	"float":    "double",
	"double":   "double",
	"bool":     "bool",
	"sint32":   "int",
	"sint64":   "int",
	"fixed32":  "int",
	"fixed64":  "int",
	"sfixed32": "int",
	"sfixed64": "int",
}

// dartDefaults maps proto field types to Dart default values.
var dartDefaults = map[string]string{
	"string":   "''",
	"bytes":    "const <int>[]",
	"uint32":   "0",
	"int32":    "0",
	"uint64":   "0",
	"int64":    "0",
	"float":    "0.0",
	"double":   "0.0",
	"bool":     "false",
	"sint32":   "0",
	"sint64":   "0",
	"fixed32":  "0",
	"fixed64":  "0",
	"sfixed32": "0",
	"sfixed64": "0",
}

// tsTypes maps proto field types to TypeScript types.
var tsTypes = map[string]string{
	"string":   "string",
	"bytes":    "Uint8Array",
	"uint32":   "number",
	"int32":    "number",
	"uint64":   "number",
	"int64":    "number",
	"float":    "number",
	"double":   "number",
	"bool":     "boolean",
	"sint32":   "number",
	"sint64":   "number",
	"fixed32":  "number",
	"fixed64":  "number",
	"sfixed32": "number",
	"sfixed64": "number",
}

// tsDefaults maps proto field types to TypeScript default values.
var tsDefaults = map[string]string{
	"string":   "''",
	"bytes":    "new Uint8Array(0)",
	"uint32":   "0",
	"int32":    "0",
	"uint64":   "0",
	"int64":    "0",
	"float":    "0",
	"double":   "0",
	"bool":     "false",
	"sint32":   "0",
	"sint64":   "0",
	"fixed32":  "0",
	"fixed64":  "0",
	"sfixed32": "0",
	"sfixed64": "0",
}

// cTypes maps proto field types to C types (for function parameters).
var cTypes = map[string]string{
	"string":   "const char *",
	"bytes":    "const uint8_t *",
	"uint32":   "uint32_t",
	"int32":    "int32_t",
	"uint64":   "uint64_t",
	"int64":    "int64_t",
	"float":    "float",
	"double":   "double",
	"bool":     "bool",
	"sint32":   "int32_t",
	"sint64":   "int64_t",
	"fixed32":  "uint32_t",
	"fixed64":  "uint64_t",
	"sfixed32": "int32_t",
	"sfixed64": "int64_t",
}

// pythonDefaults maps proto field types to Python default values.
var pythonDefaults = map[string]string{
	"string":   `""`,
	"bytes":    `b""`,
	"uint32":   "0",
	"int32":    "0",
	"uint64":   "0",
	"int64":    "0",
	"float":    "0.0",
	"double":   "0.0",
	"bool":     "False",
	"sint32":   "0",
	"sint64":   "0",
	"fixed32":  "0",
	"fixed64":  "0",
	"sfixed32": "0",
	"sfixed64": "0",
}

// Type resolution helpers.
// These handle scalar, enum, repeated, and map types for each target language.

// Helper to resolve a scalar type name from a proto type for a given language map.
func lookupScalar(typeMaps map[string]string, protoType, fallback string) string {
	if t, ok := typeMaps[protoType]; ok {
		return t
	}
	return fallback
}

func scalarKotlinType(f Field) string {
	if f.IsEnum {
		return "Int"
	}
	if f.IsMessage {
		return f.Type
	}
	if t, ok := kotlinTypes[f.Type]; ok {
		return t
	}
	return "Any"
}

func resolveKotlinType(f Field) string {
	if f.IsMap {
		k := lookupScalar(kotlinTypes, f.KeyType, "Any")
		v := lookupScalar(kotlinTypes, f.ValueType, f.ValueType)
		return "Map<" + k + ", " + v + ">"
	}
	base := scalarKotlinType(f)
	if f.IsRepeated {
		return "List<" + base + ">"
	}
	return base
}

func resolveKotlinDefault(f Field) string {
	if f.IsMap {
		return "emptyMap()"
	}
	if f.IsRepeated {
		return "emptyList()"
	}
	if f.IsEnum {
		return "0"
	}
	if f.IsMessage {
		return f.Type + ".getDefaultInstance()"
	}
	if d, ok := kotlinDefaults[f.Type]; ok {
		return d
	}
	return "0"
}

// typePath is the package-relative path of an enum / message field's type.
func typePath(f Field) string {
	if f.TypePath != "" {
		return f.TypePath
	}
	return f.Type
}

// scalarSwiftType is the Swift type of one value of field f; enum and message
// types are those protoc-gen-swift generates (typePrefix from swiftTypePrefix).
func scalarSwiftType(f Field, typePrefix string) string {
	if f.IsEnum || f.IsMessage {
		return typePrefix + typePath(f)
	}
	if t, ok := swiftTypes[f.Type]; ok {
		return t
	}
	return "Any"
}

func resolveSwiftType(f Field, typePrefix string) string {
	if f.IsMap {
		k := lookupScalar(swiftTypes, f.KeyType, "Any")
		v := lookupScalar(swiftTypes, f.ValueType, f.ValueType)
		return "[" + k + ": " + v + "]"
	}
	base := scalarSwiftType(f, typePrefix)
	if f.IsRepeated {
		return "[" + base + "]"
	}
	return base
}

func resolveSwiftDefault(f Field) string {
	if f.IsMap {
		return "[:]"
	}
	if f.IsRepeated {
		return "[]"
	}
	if f.IsEnum || f.IsMessage {
		return ".init()" // the enum's zero value / an empty message
	}
	if d, ok := swiftDefaults[f.Type]; ok {
		return d
	}
	return "nil"
}

func scalarDartType(f Field) string {
	if f.IsEnum {
		return "int"
	}
	if f.IsMessage {
		return f.Type
	}
	if t, ok := dartTypes[f.Type]; ok {
		return t
	}
	return "dynamic"
}

func resolveDartType(f Field) string {
	if f.IsMap {
		k := lookupScalar(dartTypes, f.KeyType, "dynamic")
		v := lookupScalar(dartTypes, f.ValueType, f.ValueType)
		return "Map<" + k + ", " + v + ">"
	}
	base := scalarDartType(f)
	if f.IsRepeated {
		return "List<" + base + ">"
	}
	return base
}

func resolveDartDefault(f Field) string {
	if f.IsMap {
		return "const {}"
	}
	if f.IsRepeated {
		return "const []"
	}
	if f.IsEnum {
		return "0"
	}
	if f.IsMessage {
		return f.Type + "()"
	}
	if d, ok := dartDefaults[f.Type]; ok {
		return d
	}
	return "null"
}

func scalarTsType(f Field) string {
	if f.IsEnum {
		return "number"
	}
	if f.IsMessage {
		return f.Type
	}
	if t, ok := tsTypes[f.Type]; ok {
		return t
	}
	return "unknown"
}

func resolveTsType(f Field) string {
	if f.IsMap {
		k := lookupScalar(tsTypes, f.KeyType, "string")
		v := lookupScalar(tsTypes, f.ValueType, f.ValueType)
		return "Record<" + k + ", " + v + ">"
	}
	base := scalarTsType(f)
	if f.IsRepeated {
		return base + "[]"
	}
	return base
}

func resolveTsDefault(f Field) string {
	if f.IsMap {
		return "{}"
	}
	if f.IsRepeated {
		return "[]"
	}
	if f.IsEnum {
		return "0"
	}
	if f.IsMessage {
		return "{}"
	}
	if d, ok := tsDefaults[f.Type]; ok {
		return d
	}
	return "undefined"
}

func resolvePythonDefault(f Field) string {
	if f.IsMap {
		return "None"
	}
	if f.IsRepeated {
		return "None"
	}
	if f.IsEnum {
		return "0"
	}
	if f.IsMessage {
		return "None"
	}
	if d, ok := pythonDefaults[f.Type]; ok {
		return d
	}
	return "None"
}

func resolveCType(f Field) string {
	if f.IsEnum {
		return "int32_t"
	}
	if f.IsMessage {
		return f.Type
	}
	if t, ok := cTypes[f.Type]; ok {
		return t
	}
	return "uint32_t"
}
