package dynamic

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type SchemaError struct {
	Message error
	Field   string
	Offset  int64
}

func FormatError(input []byte, err error) error {
	var te *json.UnmarshalTypeError
	if errors.As(err, &te) {
		return formatUnmarshalTypeError(input, te)
	}
	var se *json.SyntaxError
	if errors.As(err, &se) {
		return formatSyntaxError(input, se)
	}
	var e *SchemaError
	if errors.As(err, &e) {
		return formatSchemaError(input, e)
	}
	return err
}

func formatSyntaxError(input []byte, err *json.SyntaxError) error {
	msg := err.Error()
	if msg == "unexpected end of JSON input" {
		return err
	}
	return fmt.Errorf("%s%s", err.Error(), location(input, int(err.Offset)))
}

func formatUnmarshalTypeError(input []byte, err *json.UnmarshalTypeError) error {
	if err.Field == "" {
		return fmt.Errorf("expected %s but received %s%s", ToTypeName(err.Type), err.Value, location(input, int(err.Offset)))
	}
	return fmt.Errorf("field '%s' expected type %s, but got %s%s", err.Field, ToTypeName(err.Type), err.Value, location(input, int(err.Offset)))
}

func formatSchemaError(input []byte, err *SchemaError) error {
	return fmt.Errorf("%s%s", err.Error(), location(input, int(err.Offset)))
}

func (e *SchemaError) Error() string {
	path, msg := e.flatten()
	if path == "" {
		return msg
	}
	return fmt.Sprintf("schema error at field '%s': %s", path, msg)
}

func (e *SchemaError) flatten() (path string, msg string) {
	var se *SchemaError
	if errors.As(e.Message, &se) {
		subPath, subMsg := se.flatten()
		return joinPath(e.Field, subPath), subMsg
	}
	return e.Field, e.Message.Error()
}

func joinPath(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	case strings.HasPrefix(b, "["):
		return a + b // "type" + "[0]" -> "type[0]"
	default:
		return a + "." + b // "properties" + "age" -> "properties.age"
	}
}

func (e *SchemaError) Unwrap() error {
	return e.Message
}

func location(input []byte, offset int) string {
	if offset > len(input) || offset < 0 {
		return ""
	}

	line := 1
	column := 0
	for i, b := range input {
		if i == offset {
			break
		}
		if b == '\n' {
			line++
			column = 0
		} else {
			column++
		}
	}
	return fmt.Sprintf(" at line %d, column %d", line, column)
}

func ToTypeName(v reflect.Type) string {
	switch v.Kind() {
	case reflect.Slice:
		return "array"
	case reflect.Struct, reflect.Map:
		return "object"
	case reflect.Float64:
		return "number"
	case reflect.Int64, reflect.Int:
		return "integer"
	case reflect.Bool:
		return "boolean"
	default:
		return v.Kind().String()
	}
}
