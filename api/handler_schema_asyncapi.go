package api

import (
	"fmt"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	openApiSchema "mokapi/providers/openapi/schema"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
	"net/http"
)

func getAsyncApiExample(w http.ResponseWriter, r *http.Request, cfg *asyncapi3.Config) {
	channelName := r.URL.Query().Get("channel")
	msgName := r.URL.Query().Get("message")

	var channel *asyncapi3.Channel
	if channelName == "" {
		if len(cfg.Channels) > 1 {
			http.Error(w, fmt.Sprintf(toManyResults), http.StatusBadRequest)
			return
		}
		for _, p := range cfg.Channels {
			channel = p.Value
			break
		}
	} else {
		if r := cfg.Channels[channelName]; r != nil {
			channel = r.Value
		}
	}
	if channel == nil {
		http.Error(w, fmt.Sprintf("No result found"), http.StatusBadRequest)
		return
	}

	var msg *asyncapi3.Message
	if msgName == "" {
		if len(channel.Messages) > 1 {
			http.Error(w, fmt.Sprintf(toManyResults), http.StatusBadRequest)
			return
		}
		for _, m := range channel.Messages {
			msg = m.Value
			break
		}
	} else {
		msg = channel.Messages[msgName].Value
	}
	if msg == nil || msg.Payload == nil {
		http.Error(w, fmt.Sprintf("No result found"), http.StatusBadRequest)
		return
	}

	var s *jsonSchema.Schema
	switch t := msg.Payload.Value.Schema.(type) {
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
		Path: generator.Path{
			&generator.PathElement{Schema: s},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ct := media.ParseContentType(msg.ContentType)
	var data []byte
	switch t := msg.Payload.Value.Schema.(type) {
	case *openApiSchema.Schema:
		data, err = t.Marshal(rnd, ct)
	case *avro.Schema:
		switch {
		case ct.Subtype == "json":
			data, err = encoding.NewEncoder(t.Convert()).Write(rnd, ct)
		case ct.Key() == "avro/binary" || ct.Key() == "application/octet-stream":
			data, err = t.Marshal(rnd)
		default:
			http.Error(w, fmt.Sprintf("unsupported schema type: %T", t), http.StatusBadRequest)
			return
		}
	default:
		data, err = encoding.NewEncoder(s).Write(rnd, ct)

	}

	w.Header().Set("Content-Type", ct.String())
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
