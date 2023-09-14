package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/sortedmap"
	"strconv"
)

type Responses struct {
	sortedmap.LinkedHashMap[int, *ResponseRef]
} // map[HttpStatus]*ResponseRef

type ResponseRef struct {
	ref.Reference
	Value *Response
}

type Response struct {
	// A short description of the response. CommonMark syntax
	// MAY be used for rich text representation.
	Description string

	// A map containing descriptions of potential response payloads.
	// The key is a media type or media type range and the value describes
	// it. For responses that match multiple keys, only the most specific
	// key is applicable. e.g. text/plain overrides text/*
	Content Content

	// Maps a header name to its definition. RFC7230 states header names are
	// case-insensitive. If a response header is defined with the name
	// "Content-Type", it SHALL be ignored.
	Headers map[string]*HeaderRef
}

func (r *Responses) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("unexpected token %s; expected '{'", token)
	}
	r.LinkedHashMap = sortedmap.LinkedHashMap[int, *ResponseRef]{}
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		val := &ResponseRef{}
		err = dec.Decode(&val)
		if err != nil {
			return err
		}

		if key == "default" {
			r.Set(0, val)
		} else {
			key, err := strconv.Atoi(key)
			if err != nil {
				log.Errorf("unable to parse http status %v", key)
				continue
			}
			r.Set(key, val)
		}
	}
}

func (r *Responses) GetResponse(httpStatus int) *Response {
	res := r.Get(httpStatus)
	if res == nil {
		// 0 as default
		res = r.Get(0)
	}

	if res == nil {
		return nil
	}

	return res.Value
}
