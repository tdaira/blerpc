package main

// EnumValue represents a single value in an enum.
type EnumValue struct {
	Name   string
	Number int
}

// Enum represents a protobuf enum.
type Enum struct {
	Name   string
	Values []EnumValue
}

// OneofGroup represents a protobuf oneof.
type OneofGroup struct {
	Name   string
	Fields []Field
}

// Field represents a protobuf message field.
type Field struct {
	Type       string
	Name       string
	Number     int
	IsEnum     bool
	IsRepeated bool
	IsMessage  bool
	IsMap      bool
	KeyType    string
	ValueType  string
}

// Message represents a protobuf message.
type Message struct {
	Name   string
	Fields []Field
	Oneofs []OneofGroup
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

// ServiceRPC represents a single RPC method within a service.
type ServiceRPC struct {
	Name         string
	RequestType  string
	ResponseType string
	ClientStream bool // stream on request
	ServerStream bool // stream on response
}

// Service represents a protobuf service definition.
type Service struct {
	Name string
	RPCs []ServiceRPC
}
