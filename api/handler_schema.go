package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mokapi/media"
	openApiSchema "mokapi/providers/openapi/schema"
	"mokapi/runtime"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
	"net/http"
)

const toManyResults = "Your request matches multiple results. Please refine your parameters for a more precise selection."

type schemaInfo struct {
	Format string      `json:"format,omitempty"`
	Schema interface{} `json:"schema,omitempty"`
}

type requestExample struct {
	Name         string
	Format       string
	Schema       interface{}
	ContentTypes []string
}

type example struct {
	ContentType string      `json:"contentType"`
	Value       interface{} `json:"value"`
	Error       string      `json:"error,omitempty"`
}

func (h *handler) getExampleData(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if len(r.URL.RawQuery) > 0 || len(body) == 0 {
		h.getExampleByQuery(w, r)
		return
	}

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
	if len(re.ContentTypes) == 0 {
		re.ContentTypes = []string{"application/json"}
	}

	var s *jsonSchema.Schema
	switch t := re.Schema.(type) {
	case *openApiSchema.Schema:
		s = openApiSchema.ConvertToJsonSchema(t)
	case *avro.Schema:
		s = t.Convert()
	default:
		var ok bool
		s, ok = t.(*jsonSchema.Schema)
		if !ok {
			http.Error(w, fmt.Sprintf("unsupported schema type: %T", t), http.StatusBadRequest)
			return
		}
	}

	rnd, err := generator.New(&generator.Request{
		Path:   []string{re.Name},
		Schema: s,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	examples, err := encodeExample(rnd, re.Schema, re.Format, re.ContentTypes)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	writeJsonBody(w, examples)
}

func (h *handler) getExampleByQuery(w http.ResponseWriter, r *http.Request) {

	spec := r.URL.Query().Get("spec")
	name := r.URL.Query().Get("name")

	specs := getSpecs(h.app, spec, name)
	if len(specs) > 1 {
		http.Error(w, fmt.Sprintf(toManyResults), http.StatusBadRequest)
		return
	} else if len(specs) == 0 {
		http.Error(w, fmt.Sprintf("No result found"), http.StatusBadRequest)
		return
	}

	switch v := specs[0].(type) {
	case *runtime.HttpInfo:
		getOpenApiExample(w, r, v.Config)
	case *runtime.KafkaInfo:
		getAsyncApiExample(w, r, v.Config)
	default:
		http.Error(w, fmt.Sprintf("No result found"), http.StatusBadRequest)
	}
}

func encodeExample(v interface{}, schema interface{}, schemaFormat string, contentTypes []string) ([]example, error) {
	var examples []example
	for _, str := range contentTypes {
		ct := media.ParseContentType(str)

		switch {
		case ct.Subtype == "json":
		case ct.Subtype == "xml":
		case ct.Key() == "text/plain":
		case ct.Key() == "avro/binary" && isAvro(schemaFormat):
		case ct.Key() == "application/octet-stream" && isAvro(schemaFormat):
		case ct.Key() == "application/avro" && isAvro(schemaFormat):
		case ct.IsAny():
			ct = media.ParseContentType("application/json")
		default:
			examples = append(examples, example{
				ContentType: ct.String(),
				Error:       fmt.Sprintf("Content type %s with schema format %s is not supported", ct, schemaFormat),
			})
			continue
		}

		var data []byte
		var err error
		switch t := schema.(type) {
		case *openApiSchema.Schema:
			data, err = t.Marshal(v, ct)
		case *avro.Schema:
			switch {
			case ct.Subtype == "json":
				data, err = encoding.NewEncoder(t.Convert()).Write(v, ct)
			case ct.Key() == "avro/binary" || ct.Key() == "application/avro" || ct.Key() == "application/octet-stream":
				data, err = t.Marshal(v)
			default:
				examples = append(examples, example{
					ContentType: ct.String(),
					Error:       fmt.Sprintf("unsupported schema type: %T", t),
				})
				continue
			}
		default:
			s, ok := schema.(*jsonSchema.Schema)
			if !ok {
				return nil, fmt.Errorf("unsupported schema type: %T", t)
			}
			data, err = encoding.NewEncoder(s).Write(v, ct)

		}

		if err != nil {
			examples = append(examples, example{
				ContentType: ct.String(),
				Error:       err.Error(),
			})
		} else {
			examples = append(examples, example{
				ContentType: ct.String(),
				Value:       data,
			})
		}
	}
	return examples, nil
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
		case "contentTypes":
			err = d.Decode(&r.ContentTypes)
			if err != nil {
				return err
			}
		}
	}

	r.Schema, err = unmarshal(raw, r.Format)

	return err
}

func unmarshal(raw json.RawMessage, format string) (interface{}, error) {
	var err error
	if raw != nil {
		switch {
		case isOpenApi(format):
			var r *openApiSchema.Schema
			err = json.Unmarshal(raw, &r)
			//err = r.Parse(&dynamic.Config{Data: r}, &dynamic.EmptyReader{})
			return r, err
		case isAvro(format):
			var a *avro.Schema
			err = json.Unmarshal(raw, &a)
			if err != nil {
				return nil, err
			}
			//err = a.Parse(&dynamic.Config{Data: a}, &dynamic.EmptyReader{})
			return a, err
		default:
			var r *jsonSchema.Schema
			err = json.Unmarshal(raw, &r)
			if err != nil {
				return nil, err
			}
			//err = r.Parse(&dynamic.Config{Data: r}, &dynamic.EmptyReader{})
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

func getSpecs(app *runtime.App, spec, name string) []interface{} {
	var results []interface{}

	switch spec {
	case "openapi":
		if name == "" {
			for _, s := range app.ListHttp() {
				results = append(results, s)
			}
		}
		s := app.GetHttp(name)
		if s != nil {
			results = append(results, s)
		}
	case "asyncapi":
		if name == "" {
			for _, s := range app.Kafka.List() {
				results = append(results, s)
			}
		} else if s := app.Kafka.Get(name); s != nil {
			results = append(results, s)
		}
	default:
		if name == "" {
			for _, s := range app.ListHttp() {
				results = append(results, s)
			}
		} else if s := app.GetHttp(name); s != nil {
			results = append(results, s)
		}
		if name == "" {
			for _, s := range app.Kafka.List() {
				results = append(results, s)
			}
		} else if s := app.Kafka.Get(name); s != nil {
			results = append(results, s)
		}
	}

	return results
}
