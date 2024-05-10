package openapi

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"mokapi/config/dynamic"
	"mokapi/media"
	"mokapi/providers/openapi/ref"
	"mokapi/providers/openapi/schema"
	"net/http"
	"strings"
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
	_, mt := getMedia(contentType, op.RequestBody.Value)
	if !contentType.IsEmpty() && mt == nil {
		return noMatch(r, contentType)
	}

	var b *Body
	if contentType.IsEmpty() {
		b, err = readBodyDetectContentType(r, op)
	} else {
		b, err = readBody(r, contentType, mt)
	}

	if err != nil {
		if !contentType.IsEmpty() {
			err = fmt.Errorf("read request body '%v' failed: %w", contentType, err)
		} else {
			err = fmt.Errorf("read request body failed: %w", err)
		}
	}
	return b, err
}

func noMatch(r *http.Request, contentType media.ContentType) (*Body, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body failed: %w", err)
	}
	return &Body{Raw: string(data)}, fmt.Errorf("read request body failed: no matching content type for '%v' defined", contentType.String())
}

func readBodyDetectContentType(r *http.Request, op *Operation) (*Body, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	for _, mt := range op.RequestBody.Value.Content {
		if b, err := parseBody(data, mt.ContentType, mt); err == nil {
			return b, err
		}
	}

	return &Body{Raw: string(data)}, fmt.Errorf("no matching content type defined")
}

func readBody(r *http.Request, contentType media.ContentType, mt *MediaType) (*Body, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return parseBody(data, contentType, mt)
}

func parseBody(body []byte, contentType media.ContentType, mt *MediaType) (*Body, error) {
	v, err := mt.Parse(body, contentType)
	return &Body{Value: v, Raw: string(body)}, err
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

func getParser(ct media.ContentType) schema.Parser {
	if ct.Type == "text" {
		return schema.Parser{ConvertStringToNumber: true}
	}
	if ct.String() == "application/x-www-form-urlencoded" {
		return schema.Parser{ConvertStringToNumber: true}
	}
	return schema.Parser{}
}

type urlValueDecoder struct {
	mt *MediaType
}

func (d urlValueDecoder) decode(propName string, val interface{}) (interface{}, error) {
	values := val.([]string)

	prop := d.mt.Schema.Value.Properties.Get(propName)
	switch {
	case prop.Value.Type.IsOneOf("integer", "number", "string"):
		return values[0], nil
	case prop.Value.Type.IsArray():
		return d.decodeArray(propName, values)
	default:
		return nil, fmt.Errorf("unsupported type %v", prop.Value.Type)
	}
}

func (d urlValueDecoder) decodeArray(propName string, values []string) (interface{}, error) {
	enc, ok := d.mt.Encoding[propName]
	if !ok || enc.Explode {
		return values, nil
	}
	switch enc.Style {
	case "spaceDelimited":
		return strings.Split(values[0], " "), nil
	case "pipeDelimited":
		return strings.Split(values[0], "|"), nil
	default:
		return strings.Split(values[0], ","), nil
	}
}
