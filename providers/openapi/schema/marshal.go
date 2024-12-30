package schema

import (
	"fmt"
	"mokapi/media"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
)

func (r *Ref) Marshal(i interface{}, contentType media.ContentType) ([]byte, error) {
	if contentType.IsXml() {
		p := parser.Parser{ConvertStringToNumber: true, ConvertToSortedMap: true, ValidateAdditionalProperties: false}
		i, err := p.ParseWith(i, ConvertToJsonSchema(r))
		if err == nil {
			var b []byte
			b, err = marshalXml(i, r)
			if err == nil {
				return b, nil
			}
		}

		if uw, ok := err.(interface{ Unwrap() []error }); ok {
			errs := uw.Unwrap()
			if len(errs) > 1 {
				return nil, fmt.Errorf("encoding data to '%v' failed:\n %w", contentType.String(), err)
			}
		}

		return nil, fmt.Errorf("encoding data to '%v' failed: %w", contentType, err)
	}

	e := encoding.NewEncoder(ConvertToJsonSchema(r))
	return e.Write(i, contentType)
}

func (s *Schema) Marshal(i interface{}, contentType media.ContentType) ([]byte, error) {
	r := &Ref{Value: s}
	return r.Marshal(i, contentType)
}
