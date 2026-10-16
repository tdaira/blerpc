package main

import "testing"

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Echo", "echo"},
		{"FlashRead", "flash_read"},
		{"DataWrite", "data_write"},
		{"CounterStream", "counter_stream"},
		{"CounterUpload", "counter_upload"},
		{"HTMLParser", "html_parser"},
		{"getHTTPResponse", "get_http_response"},
		{"SimpleXML", "simple_xml"},
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"already_snake", "already_snake"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := camelToSnake(tt.input)
			if got != tt.want {
				t.Errorf("camelToSnake(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToLowerCamel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Echo", "echo"},
		{"FlashRead", "flashRead"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toLowerCamel(tt.input)
			if got != tt.want {
				t.Errorf("toLowerCamel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestKotlinBuilderCall(t *testing.T) {
	tests := []struct {
		field Field
		want  string
	}{
		{Field{Type: "string", Name: "message"}, ".setMessage(v)"},
		{Field{Type: "uint32", Name: "received_count"}, ".setReceivedCount(v)"},
		{Field{Type: "Address", Name: "address", IsMessage: true}, ".setAddress(v)"},
		{Field{Type: "Foot", Name: "foot", IsEnum: true}, ".setFootValue(v)"},
		{Field{Type: "uint32", Name: "ids", IsRepeated: true}, ".addAllIds(v)"},
		{Field{Type: "Foot", Name: "feet", IsEnum: true, IsRepeated: true}, ".addAllFeetValue(v)"},
		{Field{Name: "labels", IsMap: true, KeyType: "string", ValueType: "string"}, ".putAllLabels(v)"},
	}
	for _, tt := range tests {
		t.Run(tt.field.Name, func(t *testing.T) {
			if got := kotlinBuilderCall(tt.field, "v"); got != tt.want {
				t.Errorf("kotlinBuilderCall(%q) = %q, want %q", tt.field.Name, got, tt.want)
			}
		})
	}
}

// The expected names are what protoc-gen-swift 1.39.0 generates for these
// fields (and for package my_app.v1).
func TestSwiftPropertyName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"message", "message"},
		{"received_count", "receivedCount"},
		{"address", "address"},
		{"device_id", "deviceID"},
		{"id", "id"},
		{"image_url", "imageURL"},
		{"http_status_code", "httpStatusCode"},
		{"user_id_list", "userIDList"},
		{"field2", "field2"},
		{"v2_api", "v2Api"},
		{"foo__bar", "foo_Bar"},
		{"camelCase", "camelCase"},
		{"ABC_def", "abcDef"},
		{"trailing_", "trailing_"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := swiftPropertyName(tt.input)
			if got != tt.want {
				t.Errorf("swiftPropertyName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSwiftTypePrefix(t *testing.T) {
	tests := map[string]string{
		"blerpc":    "Blerpc_",
		"insole":    "Insole_",
		"my_app.v1": "MyApp_V1_",
		"":          "",
	}
	for pkg, want := range tests {
		if got := swiftTypePrefix(pkg); got != want {
			t.Errorf("swiftTypePrefix(%q) = %q, want %q", pkg, got, want)
		}
	}
}

func TestDartAndTsPropertyNamesKeepId(t *testing.T) {
	// protoc-gen-dart and protobuf.js do not upper-case abbreviations.
	if got := dartPropertyName("device_id"); got != "deviceId" {
		t.Errorf("dartPropertyName = %q", got)
	}
	if got := tsPropertyName("device_id"); got != "deviceId" {
		t.Errorf("tsPropertyName = %q", got)
	}
}

func TestCParamStr(t *testing.T) {
	tests := []struct {
		cType string
		name  string
		want  string
	}{
		{"uint32_t", "count", "uint32_t count"},
		{"const char *", "name", "const char *name"},
		{"const uint8_t *", "data", "const uint8_t *data"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cParamStr(tt.cType, tt.name)
			if got != tt.want {
				t.Errorf("cParamStr(%q, %q) = %q, want %q", tt.cType, tt.name, got, tt.want)
			}
		})
	}
}
