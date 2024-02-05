package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
	"mime"
	"mime/multipart"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/json/ref"
	"mokapi/media"
	"net/http"
)

type RequestBodies map[string]*RequestBodyRef

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
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func BodyFromRequest(r *http.Request, op *Operation) (body *Body, err error) {
	if r.ContentLength == 0 && op.RequestBody.Value.Required {
		return nil, fmt.Errorf("request body is required")
	}

	contentType := media.ParseContentType(r.Header.Get("content-type"))
	switch {
	case contentType.IsEmpty():
		body, err = tryParseBody(r, op)
	default:
		body, err = readBody(r, op, contentType)
	}
	if err != nil {
		return body, err
	}

	return body, nil
}

func readBody(r *http.Request, op *Operation, contentType media.ContentType) (*Body, error) {
	_, mt := getMedia(contentType, op.RequestBody.Value)
	if mt == nil {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("read request body failed: %w", err)
		}
		return &Body{Raw: string(data)}, fmt.Errorf("read request body failed: no matching content type for '%v' defined", contentType.String())
	}

	if contentType.Type == "multipart" {
		return readMultipart(r, mt, contentType)
	} else {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("read request body failed: %w", err)
		}

		body, err := mt.Schema.Unmarshal(data, contentType)
		if err != nil {
			err = fmt.Errorf("read request body '%v' failed: %w", contentType, err)
		}
		return &Body{Value: body, Raw: string(data)}, err
	}
	return nil, fmt.Errorf("ERROR")
}

func tryParseBody(r *http.Request, o *Operation) (*Body, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body failed: %w", err)
	}
	for _, mt := range o.RequestBody.Value.Content {
		if b, err := mt.Schema.Unmarshal(data, mt.ContentType); err == nil {
			return &Body{Value: b, Raw: string(data)}, nil
		}
	}

	return &Body{Raw: string(data)}, nil
}

func readMultipart(r *http.Request, mt *MediaType, ct media.ContentType) (*Body, error) {
	s := mt.Schema.Value
	if s.Type != "object" {
		return nil, fmt.Errorf("schema %q not support for content type multipart/form-data, expected 'object'", s.Type)
	}
	if s.Properties == nil {
		// todo raw value
		return nil, nil
	}

	_, params, err := mime.ParseMediaType(ct.String())
	if err != nil {
		return nil, err
	}
	multipartReader := multipart.NewReader(r.Body, params["boundary"])
	o := make(map[string]interface{})
	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := part.FormName()
		p := s.Properties.Get(name)
		v, err := parsePart(part, p)
		if err != nil {
			return nil, err
		}
		o[name] = v
	}
	return &Body{Value: o, Raw: toString(o)}, nil
}

func parsePart(part *multipart.Part, p *schema.Ref) (interface{}, error) {
	defer part.Close()
	b, err := io.ReadAll(part)
	if err != nil {
		return nil, err
	}

	ct := media.ParseContentType(part.Header.Get("Content-Type"))
	return p.Unmarshal(b, ct)
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
	for _, mt := range body.Content {
		if contentType.Match(mt.ContentType) {
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
			bestMediaType = mt
		}
	}

	return best, bestMediaType
}

func (r RequestBodies) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	for name, body := range r {
		if err := body.parse(config, reader); err != nil {
			inner := errors.Unwrap(err)
			return fmt.Errorf("parse request body '%v' failed: %w", name, inner)
		}
	}

	return nil
}

func (r *RequestBodyRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return fmt.Errorf("parse request body failed: %w", err)
		}
		return nil
	}

	return r.Value.Content.parse(config, reader)
}

func (r RequestBodies) patch(patch RequestBodies) {
	for k, p := range patch {
		if p == nil || p.Value == nil {
			continue
		}
		if v, ok := r[k]; ok && v != nil {
			v.patch(p)
		} else {
			r[k] = p
		}
	}
}

func (r *RequestBodyRef) patch(patch *RequestBodyRef) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
	} else {
		r.Value.patch(patch.Value)
	}
}

func (r *RequestBody) patch(patch *RequestBody) {
	if len(patch.Description) > 0 {
		r.Description = patch.Description
	}
	r.Required = patch.Required

	if len(r.Content) == 0 {
		r.Content = patch.Content
		return
	}

	r.Content.patch(patch.Content)
}
