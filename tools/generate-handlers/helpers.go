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

// kotlinBuilderCall is the protobuf-java builder call that stores parameter
// `value` into field f: setX for singular fields, setXValue for enums (the
// parameter is the enum's number), addAllX / addAllXValue for repeated fields
// and putAllX for maps. Snake-case names camel-case: "received_count" ->
// "setReceivedCount".
func kotlinBuilderCall(f Field, value string) string {
	var name strings.Builder
	for _, p := range strings.Split(f.Name, "_") {
		if p != "" {
			name.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	verb, suffix := "set", ""
	switch {
	case f.IsMap:
		verb = "putAll"
	case f.IsRepeated:
		verb = "addAll"
	}
	if f.IsEnum {
		suffix = "Value"
	}
	return fmt.Sprintf(".%s%s%s(%s)", verb, name.String(), suffix, value)
}

// lowerCamelFromSnake converts a snake_case field name to lowerCamelCase
// ("device_id" -> "deviceId"), the property naming of protobuf-dart and
// protobuf.js.
func lowerCamelFromSnake(fieldName string) string {
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

// swiftAbbreviations are upper-cased whole by protoc-gen-swift when they are
// not the first word ("device_id" -> "deviceID", "image_url" -> "imageURL").
var swiftAbbreviations = map[string]bool{"url": true, "http": true, "https": true, "id": true}

// swiftCamelCase reproduces protoc-gen-swift's CamelCaser
// (SwiftProtobufPluginLibrary/NamingUtils.swift) for protobuf identifiers:
// the name splits into words at underscores, case changes and digit runs;
// each word after the first is capitalized (abbreviations upper-cased) and
// the first one too when initialUpper is set. A repeated underscore is kept
// (minus one), as is a trailing one.
func swiftCamelCase(s string, initialUpper bool) string {
	const (
		clsNone = iota
		clsDigit
		clsLower
		clsUpper
		clsUnderscore
		clsOther
	)
	classOf := func(r rune) int {
		switch {
		case r >= '0' && r <= '9':
			return clsDigit
		case r >= 'a' && r <= 'z':
			return clsLower
		case r >= 'A' && r <= 'Z':
			return clsUpper
		case r == '_':
			return clsUnderscore
		}
		return clsOther
	}

	var result, current strings.Builder
	addCurrent := func() {
		word := current.String()
		if word == "" {
			return
		}
		switch {
		case result.Len() == 0 && !initialUpper:
		case swiftAbbreviations[word]:
			word = strings.ToUpper(word)
		default:
			word = strings.ToUpper(word[:1]) + word[1:]
		}
		result.WriteString(word)
		current.Reset()
	}

	last := clsNone
	for _, r := range s {
		cls := classOf(r)
		switch cls {
		case clsDigit:
			if last != clsDigit {
				addCurrent()
			}
			if result.Len() == 0 {
				result.WriteByte('_') // an identifier cannot start with a digit
			}
			current.WriteRune(r)
		case clsUpper:
			if last != clsUpper {
				addCurrent()
			}
			current.WriteRune(r + ('a' - 'A'))
		case clsLower:
			if last != clsLower && last != clsUpper {
				addCurrent()
			}
			current.WriteRune(r)
		case clsUnderscore:
			addCurrent()
			if last == clsUnderscore {
				result.WriteByte('_')
			}
		default:
			addCurrent()
			current.WriteRune(r)
		}
		last = cls
	}
	addCurrent()
	if last == clsUnderscore {
		result.WriteByte('_')
	}
	return result.String()
}

// swiftPropertyName is the property protoc-gen-swift generates for a field.
func swiftPropertyName(fieldName string) string {
	return swiftCamelCase(fieldName, false)
}

// swiftTypePrefix is the prefix protoc-gen-swift puts on the types of a proto
// package: each dotted component upper-camel-cased, joined and followed by
// "_" ("blerpc" -> "Blerpc_", "my_app.v1" -> "MyApp_V1_").
func swiftTypePrefix(pkg string) string {
	if pkg == "" {
		return ""
	}
	parts := strings.Split(pkg, ".")
	for i, p := range parts {
		parts[i] = swiftCamelCase(p, true)
	}
	return strings.Join(parts, "_") + "_"
}

// dartPropertyName is the property protoc-gen-dart generates for a field.
func dartPropertyName(fieldName string) string {
	return lowerCamelFromSnake(fieldName)
}

// tsPropertyName is the property protobuf.js generates for a field.
func tsPropertyName(fieldName string) string {
	return lowerCamelFromSnake(fieldName)
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
