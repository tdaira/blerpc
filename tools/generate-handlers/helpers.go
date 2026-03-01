package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reSub1 = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	reSub2 = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func camelToSnake(name string) string {
	s := reSub1.ReplaceAllString(name, "${1}_${2}")
	s = reSub2.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
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

// cParamStr formats a C type and parameter name, handling pointer types.
func cParamStr(cType, name string) string {
	if strings.HasSuffix(cType, "*") {
		return cType + name
	}
	return cType + " " + name
}

// cClientParams builds the parameter list for a C client function.
func cClientParams(cmd Command, streaming map[string]string, callbacks map[string]bool, pkg string) []string {
	dir, isStreaming := streaming[cmd.Snake]
	reqMsg := pkg + "_" + cmd.RequestMsg
	respMsg := pkg + "_" + cmd.ResponseMsg

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
			cType := resolveCType(f)
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
