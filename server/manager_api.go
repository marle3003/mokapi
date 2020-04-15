package server

import (
	"mokapi/config"
	"mokapi/providers/data"
	"mokapi/server/handlers"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type ApiManager struct {
	api  *config.Api
	data data.DataProvider
}

func NewApiManager(api *config.Api, data data.DataProvider) *ApiManager {
	return &ApiManager{api: api, data: data}
}

func (m *ApiManager) Build() http.Handler {
	serviceHandler := handlers.NewServiceHandler()

	for path, endpoint := range m.api.EndPoints {
		endpointHandler := m.getEndPointHandler(endpoint)
		serviceHandler.AddHandler(path, endpointHandler)
	}

	return serviceHandler
}

func (m *ApiManager) getEndPointHandler(endpoint *config.Endpoint) http.Handler {
	handler := handlers.NewEndpointHandler(endpoint)

	if endpoint.Get != nil {
		operationHandler := m.getOperationHandler(endpoint.Get)
		handler.AddHandler("GET", operationHandler)
	}

	return handler
}

func (m *ApiManager) getOperationHandler(operation *config.Operation) http.Handler {
	operationHandler := handlers.NewOperationHandler()

	// todo error handling
	response := m.selectSuccessResponse(operation)
	if response != nil {
		for contentType, mediaType := range response.Content {
			schema := m.resolveSchema(mediaType.Schema)
			dataHandler := handlers.NewContentHandler(operation.Parameters, contentType, schema, m.data)
			operationHandler.AddHandler(contentType, dataHandler)
		}
	}

	return operationHandler
}

func (m *ApiManager) resolveSchema(schema *config.Schema) *config.Schema {
	// todo resolving schema: $ref
	if schema.Reference != "" {
		if strings.HasPrefix(schema.Reference, "#/components/schemas/") {
			key := strings.TrimPrefix(schema.Reference, "#/components/schemas/")
			return m.resolveSchema(m.api.Components.Schemas[key])
		} else {
			// todo
		}
	} else if schema.Items != nil {
		schema.Items = m.resolveSchema(schema.Items)
	}

	for i, p := range schema.Properties {
		schema.Properties[i] = m.resolveSchema(p)
	}

	return schema
}

func (m *ApiManager) selectSuccessResponse(operation *config.Operation) *config.Response {
	keys := make([]int, 0, len(operation.Responses))
	for k := range operation.Responses {
		key, error := strconv.Atoi(k)
		if error == nil {
			keys = append(keys, key)
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, key := range keys {
		if key >= 200 && key < 300 {
			return operation.Responses[strconv.Itoa(key)]
		}
	}

	return nil
}
