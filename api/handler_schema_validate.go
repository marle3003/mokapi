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
	"net/http"
)

type validateRequest struct {
	Schema *schemaInfo
	Data   string
}

func (h *handler) validate(w http.ResponseWriter, r *http.Request) {
	ct, err := getValidationDataContentType(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s, data, err := parseValidationRequestBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var v interface{}
	if ct.IsXml() {
		v, err = schema.UnmarshalXML(bytes.NewReader(data), s)
	} else {
		v, err = encoding.Decode(data, encoding.WithContentType(ct))
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := parser.Parser{ValidateAdditionalProperties: true}
	v, err = p.ParseWith(v, schema.ConvertToJsonSchema(s))
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

func parseValidationRequestBody(r *http.Request) (*schema.Ref, []byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	validateData := validateRequest{}
	err = json.Unmarshal(body, &validateData)
	if err != nil {
		s := err.Error()
		_ = s
		return nil, nil, err
	}

	return &schema.Ref{Value: toSchema(validateData.Schema)}, []byte(validateData.Data), nil
}
