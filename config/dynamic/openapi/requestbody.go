package openapi

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
	"mime/multipart"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/media"
	"net/http"
)

type RequestBodyRef struct {
	ref.Reference
	Value *RequestBody
}

type RequestBody struct {
	// A brief description of the request body. This could contain
	// examples of use. CommonMark syntax MAY be used for rich text representation.
	Description string

	// The content of the request body. The key is a media type or media type range
	// and the value describes it. For requests that match multiple keys, only the
	// most specific key is applicable. e.g. text/plain overrides text/*
	Content Content

	// Determines if the request body is required in the request. Defaults to false.
	Required bool
}

type Body struct {
	Value interface{}
	Raw   string
}

func (r *RequestBodyRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *RequestBodyRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func BodyFromRequest(r *http.Request, op *Operation) (*Body, error) {
	contentType := media.ParseContentType(r.Header.Get("content-type"))
	body, err := readBody(r, op, contentType)
	if err != nil {
		return nil, err
	}
	if op.RequestBody.Value.Required && body == nil {
		return nil, fmt.Errorf("request body expected")
	}

	return body, nil
}

func readBody(r *http.Request, op *Operation, contentType media.ContentType) (*Body, error) {
	if r.ContentLength == 0 {
		return nil, nil
	}

	_, media := getMedia(contentType, op.RequestBody.Value)
	if media == nil {
		return nil, fmt.Errorf("content type '%v' of request body is not defined. Check your service configuration", contentType.String())
	}
	if media.Schema == nil || media.Schema.Value == nil {
		return nil, fmt.Errorf("schema of request body %q is not defined", contentType.String())
	}

	s := media.Schema.Value

	if contentType.Key() == "multipart/form-data" {
		if s.Type != "object" {
			return nil, fmt.Errorf("schema %q not support for content type multipart/form-data, expected 'object'", s.Type)
		}
		if s.Properties == nil {
			// todo raw value
			return nil, nil
		}

		err := r.ParseMultipartForm(512) // maxMemory 32MB
		defer func() {
			err := r.MultipartForm.RemoveAll()
			if err != nil {
				log.Errorf("error on removing multipart form: %v", err)
			}
		}()
		if err != nil {
			return nil, err
		}

		o := make(map[string]interface{})

		for name, values := range r.MultipartForm.Value {
			p := s.Properties.Get(name)
			if p == nil || p.Value == nil {
				continue
			}
			if p.Value.Type == "array" {
				a := make([]interface{}, 0, len(values))
				for _, v := range values {
					i, err := schema.ParseString(v, p.Value.Items)
					if err != nil {
						return nil, err
					}
					a = append(a, i)
				}
				o[name] = a
			} else {
				i, err := schema.ParseString(values[0], p)
				if err != nil {
					return nil, err
				}
				o[name] = i
			}
		}

		for name, files := range r.MultipartForm.File {
			p := s.Properties.Get(name)
			if p == nil || p.Value == nil {
				continue
			}
			if p.Value.Type == "array" {
				a := make([]interface{}, 0, len(files))
				for _, file := range files {
					i, err := parseFormFile(file)
					if err != nil {
						return nil, err
					}
					a = append(a, i)
				}
				o[name] = a
			} else {
				i, err := parseFormFile(files[0])
				if err != nil {
					return nil, err
				}
				o[name] = i
			}
		}

		return &Body{Value: o, Raw: toString(o)}, nil
	} else {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		body, err := schema.Parse(data, contentType, media.Schema)
		return &Body{Value: body, Raw: string(data)}, err
	}
}

func parseFormFile(fh *multipart.FileHeader) (interface{}, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Errorf("unable to close file: %v", err)
		}
	}()

	var sniff [512]byte
	_, err = f.Read(sniff[:])
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"filename": fh.Filename,
		"type":     http.DetectContentType(sniff[:]),
		"size":     prettyByteCountIEC(fh.Size),
	}, nil
}

func prettyByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func toString(i interface{}) string {
	b, err := json.Marshal(i)
	if err != nil {
		log.Errorf("error in schema.toString(): %v", err)
	}
	return string(b)
}

func getMedia(contentType media.ContentType, body *RequestBody) (media.ContentType, *MediaType) {
	best := media.Empty
	var bestMediaType *MediaType
	bestQ := -1.0
	for _, mt := range body.Content {
		if contentType.Match(mt.ContentType) {
			if bestQ > contentType.Q {
				continue
			}

			// text/plain > */* and text/*
			if best.IsPrecise() && (mt.ContentType.IsAny() || mt.ContentType.IsRange()) {
				continue
			}

			// text/* > */*
			if best.IsRange() && mt.ContentType.IsAny() {
				continue
			}

			if !best.IsEmpty() && len(best.Parameters) > len(mt.ContentType.Parameters) {
				continue
			}

			best = mt.ContentType
			bestQ = mt.ContentType.Q
			bestMediaType = mt
		}
	}

	return best, bestMediaType
}
