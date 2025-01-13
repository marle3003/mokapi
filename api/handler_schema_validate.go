package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
	jsonSchema "mokapi/schema/json/schema"
	"net/http"
)

type validateRequest struct {
	Format string
	Schema interface{}
	Data   string
}

func (h *handler) validate(w http.ResponseWriter, r *http.Request) {
	ct, err := getValidationDataContentType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	valReq, err := parseValidationRequestBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch s := valReq.Schema.(type) {
	case *schema.Ref:
		err = parseByOpenApi([]byte(valReq.Data), s, ct)
	default:
		err = parseByJson([]byte(valReq.Data), s.(*jsonSchema.Schema), ct)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getValidationDataContentType(r *http.Request) (media.ContentType, error) {
	dataContentType := r.Header.Get("Data-Content-Type")
	if len(dataContentType) == 0 {
		dataContentType = "application/json"
	}
	ct := media.ParseContentType(dataContentType)
	if ct.IsAny() {
		ct = media.ParseContentType("application/json")
	}
	if ct.Subtype != "json" && ct.Subtype != "xml" {
		return media.Empty, fmt.Errorf("content-type %v not supported. Only json or xml are supported", ct)
	}
	return ct, nil
}

func parseValidationRequestBody(r *http.Request) (validateRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return validateRequest{}, err
	}

	validateData := validateRequest{}
	err = json.Unmarshal(body, &validateData)
	if err != nil {
		return validateData, err
	}

	return validateData, nil
}

func parseByOpenApi(data []byte, s *schema.Ref, ct media.ContentType) error {
	var v interface{}
	var err error
	if ct.IsXml() {
		v, err = schema.UnmarshalXML(bytes.NewReader(data), s)
	} else {
		v, err = encoding.Decode(data, encoding.WithContentType(ct))
	}
	if err != nil {
		return err
	}

	p := parser.Parser{ValidateAdditionalProperties: true}
	_, err = p.ParseWith(v, schema.ConvertToJsonSchema(s))
	return err
}

func parseByJson(data []byte, s *jsonSchema.Schema, ct media.ContentType) error {
	v, err := encoding.Decode(data, encoding.WithContentType(ct))
	if err != nil {
		return err
	}

	p := parser.Parser{ValidateAdditionalProperties: true}
	_, err = p.ParseWith(v, s)
	return err
}

func (r *validateRequest) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	t, err := d.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("unexpected token %s; expected '{'", t)
	}

	var raw json.RawMessage
	for {
		t, err = d.Token()
		if err != nil {
			return err
		}

		if delim, ok := t.(json.Delim); ok && delim == '}' {
			break
		}

		switch t.(string) {
		case "data":
			t, err = d.Token()
			if err != nil {
				return err
			}
			r.Data = t.(string)
		case "format":
			t, err = d.Token()
			if err != nil {
				return err
			}
			r.Format = t.(string)
		case "schema":
			err = d.Decode(&raw)
			if err != nil {
				return err
			}
		}
	}

	r.Schema, err = unmarshal(raw, r.Format)

	return err
}
