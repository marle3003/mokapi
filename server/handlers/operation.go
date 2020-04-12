package handlers

import (
	"encoding/json"
	"fmt"
	"mokapi/config"
	"net/http"
	"strings"

	"github.com/brianvoe/gofakeit"
)

type OperationHandler struct {
	operation *config.Operation
}

func NewOperationHandler(operation *config.Operation) *OperationHandler {
	return &OperationHandler{operation: operation}
}

func (handler *OperationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := handler.GetResponse(Success)
	if response == nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "No response found for %s", Success.String())
		return
	}

	accept := r.Header.Get("accept")
	contentType, schema := handler.GetSchema(accept, response)
	if contentType == "" {
		w.WriteHeader(500)
		fmt.Fprintf(w, "No supporting content type found")
		return
	}

	w.Header().Add("Content-Type", contentType)
	obj := getRandomObject(schema)
	data, error := json.Marshal(obj)
	if error != nil {
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (handler *OperationHandler) GetSchema(accept string, response *config.Response) (string, *config.Schema) {
	for _, s := range strings.Split(accept, ",") {
		contentType := strings.TrimSpace(s)
		if mediaType, ok := response.Content[contentType]; ok {
			return contentType, mediaType.Schema
		}
	}
	return "", nil
}

func (handler *OperationHandler) GetResponse(description ResponseDescription) *config.Response {
	for _, response := range handler.operation.Responses {
		responseDescription := strings.ToLower(response.Description)
		if strings.HasPrefix(responseDescription, description.String()) {
			return response
		}
	}
	return nil
}

func getRandomObject(schema *config.Schema) interface{} {
	if schema.Type == "object" {
		obj := make(map[string]interface{})
		for name, propSchema := range schema.Properties {
			value := getRandomObject(propSchema)
			obj[name] = value
		}
		return obj
	} else {
		if len(schema.Faker) > 0 {
			switch schema.Faker {
			case "{numbers.uint32}":
				return gofakeit.Uint32()
			default:
				return gofakeit.Generate(schema.Faker)
			}
		} else if schema.Type == "integer" {
			return gofakeit.Int32()
		} else if schema.Type == "string" {
			return gofakeit.Lexify("???????????????")
		}
	}
	return nil
}

type ResponseDescription int

const (
	Success ResponseDescription = 0 + iota
)

var (
	descriptions = [...]string{
		"success",
	}
)

func (r ResponseDescription) String() string {
	return descriptions[r]
}
