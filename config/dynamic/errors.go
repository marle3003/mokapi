package dynamic

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type SemanticError struct {
	Fields  []string
	Value   string
	Message string
	Offset  int64
	d       *json.Decoder
}

func NewSemanticError(err error, offset int64, d *json.Decoder) error {
	var errType *json.UnmarshalTypeError
	if errors.As(err, &errType) {
		if len(errType.Field) == 0 {
			return err
		}
		return &SemanticError{Value: errType.Value, Fields: []string{errType.Field}, Offset: offset, d: d}
	}
	var semantic *SemanticError
	if errors.As(err, &semantic) {
		return semantic
	}

	return &SemanticError{Message: err.Error(), Offset: offset, d: d}
}

func NewSemanticErrorWithField(err error, offset int64, d *json.Decoder, field string) error {
	var errType *json.UnmarshalTypeError
	if errors.As(err, &errType) {
		return &SemanticError{Value: errType.Value, Fields: []string{field, errType.Field}, d: d}
	}
	var semantic *SemanticError
	if errors.As(err, &semantic) {
		return semantic.Wrap(field, offset, d)
	}

	return &SemanticError{Fields: []string{field}, Message: err.Error(), Offset: d.InputOffset(), d: d}
}

func (s *SemanticError) Error() string {
	if len(s.Value) > 0 {
		return fmt.Sprintf("semantic error at %s: %s", strings.Join(s.Fields, "."), s.Value)
	} else if len(s.Message) > 0 {
		return fmt.Sprintf("semantic error at %s: %s", strings.Join(s.Fields, "."), s.Message)
	}
	return fmt.Sprintf("semantic error at %s", strings.Join(s.Fields, "."))
}

func (s *SemanticError) Wrap(field string, offset int64, d *json.Decoder) *SemanticError {
	s.Fields = append([]string{field}, s.Fields...)
	if s.d != d {
		s.Offset += offset
		s.d = d
	}
	return s
}
