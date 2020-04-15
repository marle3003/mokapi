package handlers

import (
	"encoding/json"
	"fmt"
	"mokapi/config"
	"mokapi/providers/data"
	"net/http"
)

type ResponseHandler struct {
	Parameters  []*config.Parameter
	ContentType string
	schema      *config.Schema
	Provider    data.DataProvider
}

func NewContentHandler(parameters []*config.Parameter, contentType string, schema *config.Schema, provider data.DataProvider) *ResponseHandler {
	return &ResponseHandler{Parameters: parameters, ContentType: contentType, schema: schema, Provider: provider}
}

func (handler *ResponseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// set content type
	w.Header().Add("Content-Type", handler.ContentType)

	// get content
	content, error := handler.Provider.Provide(nil, handler.schema)
	fmt.Printf("--- m:\n%v\n\n", content)
	if error != nil {
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}

	// write content
	data, error := handler.Encode(content)
	if error != nil {
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (handler *ResponseHandler) Encode(obj interface{}) ([]byte, error) {
	switch handler.ContentType {
	case "application/json":
		return json.Marshal(obj)
	default:
		return nil, fmt.Errorf("Unspupported content type")
	}
}
