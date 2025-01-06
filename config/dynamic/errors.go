package dynamic

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type StructuralError struct {
	Fields  []string
	Value   string
	Message string
	Offset  int64
	d       *json.Decoder
}

type SemanticError struct {
	Fields  []string
	Message string
}

func NewStructuralError(err error, offset int64, d *json.Decoder) error {
	var errType *json.UnmarshalTypeError
	if errors.As(err, &errType) {
		if len(errType.Field) == 0 {
			return err
		}
		return &StructuralError{Value: errType.Value, Fields: []string{errType.Field}, Offset: offset, d: d}
	}
	var semantic *StructuralError
	if errors.As(err, &semantic) {
		return semantic
	}

	return &StructuralError{Message: err.Error(), Offset: offset, d: d}
}

func NewStructuralErrorWithField(err error, offset int64, d *json.Decoder, field string) error {
	var errType *json.UnmarshalTypeError
	if errors.As(err, &errType) {
		return &StructuralError{Value: errType.Value, Fields: []string{field, errType.Field}, d: d}
	}
	var semantic *StructuralError
	if errors.As(err, &semantic) {
		return semantic.wrap(field, offset, d)
	}

	return &StructuralError{Fields: []string{field}, Message: err.Error(), Offset: d.InputOffset(), d: d}
}

func (s *StructuralError) Error() string {
	if len(s.Value) > 0 {
		return fmt.Sprintf("structural error at %s: %s", strings.Join(s.Fields, "."), s.Value)
	} else if len(s.Message) > 0 {
		return fmt.Sprintf("structural error at %s: %s", strings.Join(s.Fields, "."), s.Message)
	}
	return fmt.Sprintf("structural error at %s", strings.Join(s.Fields, "."))
}

func (s *StructuralError) wrap(field string, offset int64, d *json.Decoder) *StructuralError {
	s.Fields = append([]string{field}, s.Fields...)
	if s.d != d {
		s.Offset += offset
		s.d = d
	}
	return s
}

func NewSemanticError(err error, field string) *SemanticError {
	var semantic *SemanticError
	if errors.As(err, &semantic) {
		return semantic.wrap(field)
	}
	return &SemanticError{Message: err.Error(), Fields: []string{field}}
}

func (s *SemanticError) wrap(field string) *SemanticError {
	s.Fields = append([]string{field}, s.Fields...)
	return s
}

func (s *SemanticError) Error() string {
	return fmt.Sprintf("semantic error at %s: %s", strings.Join(s.Fields, "."), s.Message)
}
