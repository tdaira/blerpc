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

func TestKotlinSetterName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"message", "setMessage"},
		{"received_count", "setReceivedCount"},
		{"address", "setAddress"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := kotlinSetterName(tt.input)
			if got != tt.want {
				t.Errorf("kotlinSetterName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSwiftPropertyName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"message", "message"},
		{"received_count", "receivedCount"},
		{"address", "address"},
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
