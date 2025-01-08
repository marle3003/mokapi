package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
	"net/http"
)

type schemaInfo struct {
	Format string      `json:"format,omitempty"`
	Schema interface{} `json:"schema,omitempty"`
}

type requestExample struct {
	Name   string      `json:"name,omitempty"`
	Format string      `json:"format"`
	Schema interface{} `json:"schema"`
}

func (h *handler) getExampleData(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	if len(accept) == 0 {
		accept = "application/json"
	}
	ct := media.ParseContentType(accept)
	if ct.IsAny() {
		ct = media.ParseContentType("application/json")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	re := &requestExample{}
	err = json.Unmarshal(body, &re)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if re.Format == "" {
		re.Format = "application/schema+json;version=2020-12"
	}

	switch {
	case ct.Subtype == "json":
		break
	case ct.Subtype == "xml":
		break
	case ct.Key() == "text/plain":
		break
	case ct.Key() == "avro/binary" && isAvro(re.Format):
		break
	case ct.Key() == "application/octet-stream" && isAvro(re.Format):
		break
	default:
		http.Error(w, fmt.Sprintf("Content type %s with schema format %s is not supported", ct, re.Format), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", ct.String())

	var data []byte
	switch s := re.Schema.(type) {
	case *schema.Ref:
		data, err = getRandomByOpenApi(re.Name, s, ct)
	case *avro.Schema:
		data, err = getRandomByJson(re.Name, s.Convert(), ct)
	default:
		data, err = getRandomByJson(re.Name, s.(*jsonSchema.Schema), ct)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}
}

func getRandomByOpenApi(name string, r *schema.Ref, ct media.ContentType) ([]byte, error) {
	j := schema.ConvertToJsonSchema(r)
	data, err := generator.New(&generator.Request{
		Path: generator.Path{
			&generator.PathElement{Name: name, Schema: j},
		},
	})
	if err != nil {
		return nil, err
	}
	return r.Marshal(data, ct)
}

func getRandomByJson(name string, r *jsonSchema.Schema, ct media.ContentType) ([]byte, error) {
	data, err := generator.New(&generator.Request{
		Path: generator.Path{
			&generator.PathElement{Name: name, Schema: r},
		},
	})
	if err != nil {
		return nil, err
	}
	return encoding.NewEncoder(r).Write(data, ct)
}

func (s *schemaInfo) UnmarshalJSON(data []byte) error {
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
		case "format":
			t, err = d.Token()
			if err != nil {
				return err
			}
			s.Format = t.(string)
		case "schema":
			err = d.Decode(&raw)
			if err != nil {
				return err
			}
		}
	}

	s.Schema, err = unmarshal(raw, s.Format)

	return nil
}

func (r *requestExample) UnmarshalJSON(data []byte) error {
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
		case "name":
			t, err = d.Token()
			if err != nil {
				return err
			}
			r.Name = t.(string)
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

func unmarshal(raw json.RawMessage, format string) (interface{}, error) {
	if raw != nil {
		switch {
		case isOpenApi(format):
			var r *schema.Ref
			err := json.Unmarshal(raw, &r)
			return r, err
		case isAvro(format):
			var a *avro.Schema
			err := json.Unmarshal(raw, &a)
			return a, err
		default:
			var r *jsonSchema.Schema
			err := json.Unmarshal(raw, &r)
			return r, err
		}
	}
	return nil, nil
}

func isAvro(format string) bool {
	switch format {
	case "application/vnd.apache.avro;version=1.9.0",
		"application/vnd.apache.avro+json;version=1.9.0":
		return true
	default:
		return false
	}
}

func isOpenApi(format string) bool {
	switch format {
	case "application/vnd.oai.openapi+json;version=3.0.0",
		"application/vnd.oai.openapi;version=3.0.0":
		return true
	default:
		return false
	}
}
