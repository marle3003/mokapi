package api

import (
	"fmt"
	"mokapi/media"
	"mokapi/providers/openapi"
	openApiSchema "mokapi/providers/openapi/schema"
	"mokapi/schema/json/generator"
	"net/http"
	"strconv"
	"strings"
)

func getOpenApiExample(w http.ResponseWriter, r *http.Request, cfg *openapi.Config) {
	pathName := r.URL.Query().Get("path")
	opName := r.URL.Query().Get("operation")
	status := r.URL.Query().Get("status")
	contentType := r.URL.Query().Get("contentType")
	if contentType == "" {
		contentType = r.Header.Get("Accept")
	}

	var path *openapi.Path
	if pathName == "" {
		if len(cfg.Paths) > 1 {
			http.Error(w, fmt.Sprintf(toManyResults), http.StatusBadRequest)
			return
		}
		for _, p := range cfg.Paths {
			path = p.Value
			break
		}
	} else {
		if r := cfg.Paths[pathName]; r != nil {
			path = r.Value
		}
	}
	if path == nil {
		http.Error(w, fmt.Sprintf("No result found"), http.StatusBadRequest)
		return
	}

	var op *openapi.Operation
	if opName == "" {
		if len(path.Operations()) > 1 {
			http.Error(w, fmt.Sprintf(toManyResults), http.StatusBadRequest)
			return
		}
		for _, o := range path.Operations() {
			op = o
			break
		}
	} else {
		op = path.Operation(strings.ToUpper(opName))
	}
	if op == nil {
		http.Error(w, fmt.Sprintf("No result found"), http.StatusBadRequest)
		return
	}

	var res *openapi.Response
	if status == "" {
		val := op.Responses.Values()
		if len(val) > 1 {
			http.Error(w, fmt.Sprintf(toManyResults), http.StatusBadRequest)
			return
		}
		res = val[0].Value
	} else {
		i, err := strconv.Atoi(status)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid status value: %s", status), http.StatusBadRequest)
		}
		res = op.Responses.GetResponse(i)
	}
	if res == nil {
		http.Error(w, fmt.Sprintf("No result found"), http.StatusBadRequest)
		return
	}

	var s *openApiSchema.Schema
	var ct media.ContentType
	if len(contentType) == 0 {
		if len(res.Content) > 1 {
			http.Error(w, fmt.Sprintf(toManyResults), http.StatusBadRequest)
			return
		}
		for key, c := range res.Content {
			ct = media.ParseContentType(key)
			s = c.Schema
			break
		}
	} else {
		var mt *openapi.MediaType
		var err error
		ct, mt, err = openapi.NegotiateContentType(contentType, res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s = mt.Schema
	}
	if s == nil {
		http.Error(w, fmt.Sprintf("No result found"), http.StatusBadRequest)
		return
	}

	rnd, err := generator.New(&generator.Request{
		Path: generator.Path{
			&generator.PathElement{Schema: openApiSchema.ConvertToJsonSchema(s)},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var data []byte
	data, err = s.Marshal(rnd, ct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", ct.String())
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
