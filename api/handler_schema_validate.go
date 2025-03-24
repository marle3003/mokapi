package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
	jsonSchema "mokapi/schema/json/schema"
	"net/http"
)

type validateRequest struct {
	Format      string
	Schema      interface{}
	Data        []byte
	ContentType media.ContentType
}

func (h *handler) validate(w http.ResponseWriter, r *http.Request) {
	valReq, err := parseValidationRequestBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !(valReq.ContentType.Subtype == "json" || valReq.ContentType.Subtype == "xml" || valReq.ContentType.Key() == "avro/binary") {
		http.Error(w, fmt.Sprintf("content-type %v not supported. Only json or xml are supported", valReq.ContentType), http.StatusBadRequest)
		return
	}

	var v interface{}
	switch s := valReq.Schema.(type) {
	case *schema.Schema:
		v, err = parseByOpenApi(valReq.Data, s, valReq.ContentType)
	case *avro.Schema:
		p := &avro.Parser{Schema: s}
		v, err = encoding.Decode(valReq.Data, encoding.WithContentType(valReq.ContentType), encoding.WithParser(p))
	default:
		v, err = parseByJson(valReq.Data, s.(*jsonSchema.Schema), valReq.ContentType)
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		data := parser.Marshal(err)
		w.Write([]byte(data))
		return
	}

	formats := r.URL.Query()["outputFormat"]
	examples, err := encodeExample(v, valReq.Schema, valReq.Format, formats)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(examples) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.Header().Set("Content-Type", "application/json")
		writeJsonBody(w, examples)
	}
}

func parseValidationRequestBody(r *http.Request) (validateRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return validateRequest{}, err
	}

	validateData := validateRequest{ContentType: media.ParseContentType("application/json")}
	err = json.Unmarshal(body, &validateData)
	if err != nil {
		return validateData, err
	}

	return validateData, nil
}

func parseByOpenApi(data []byte, s *schema.Schema, ct media.ContentType) (interface{}, error) {
	p := &parser.Parser{ValidateAdditionalProperties: true, Schema: schema.ConvertToJsonSchema(s)}
	var v interface{}
	var err error
	if ct.IsXml() {
		v, err = schema.UnmarshalXML(bytes.NewReader(data), s)
		if err != nil {
			return v, err
		}
		_, err = p.ParseWith(v, schema.ConvertToJsonSchema(s))
	} else {
		_, err = encoding.Decode(data, encoding.WithContentType(ct), encoding.WithParser(p))
	}
	return v, err
}

func parseByJson(data []byte, s *jsonSchema.Schema, ct media.ContentType) (interface{}, error) {
	p := &parser.Parser{ValidateAdditionalProperties: true, Schema: s}
	return encoding.Decode(data, encoding.WithContentType(ct), encoding.WithParser(p))
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
			var b []byte
			b, err = base64.StdEncoding.DecodeString(t.(string))
			if err == nil {
				r.Data = b
			} else {
				r.Data = []byte(t.(string))
			}
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
		case "contentType":
			t, err = d.Token()
			if err != nil {
				return err
			}
			r.ContentType = media.ParseContentType(t.(string))
		}
	}

	r.Schema, err = unmarshal(raw, r.Format)

	return err
}
