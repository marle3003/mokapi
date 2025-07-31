package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/search"
	"mokapi/try"
	"testing"
)

func TestIndex(t *testing.T) {
	cfg := &static.Config{}
	cfg.Api.Search.Enabled = true
	app := runtime.New(cfg)

	trait := events.NewTraits().WithNamespace("test")
	app.Events.SetStore(10, trait)
	err := app.Events.Push(&eventstest.Event{
		Name: "foo",
		Api:  "Swagger Petstore",
	}, trait)
	require.NoError(t, err)

	petstore := openapitest.NewConfig("3.0", openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server. You can find out more about Swagger at https://swagger.io"),
		openapitest.WithPath("/pet", openapitest.NewPath(
			openapitest.WithOperation("put", openapitest.NewOperation(
				openapitest.WithOperationInfo("Update an existing pet.", "Update an existing pet by Id.", "updatePet", false),
			)),
		)),
	)
	app.AddHttp(&dynamic.Config{
		Info: dynamic.ConfigInfo{Url: try.MustUrl("petstore.yaml"), Provider: "NPM"},
		Data: petstore,
	})

	result, err := app.Search(search.Request{
		QueryText: "api:Petstore",
		Index:     0,
		Limit:     10,
	})
	require.NoError(t, err)
	require.Len(t, result.Results, 4)
	require.Len(t, result.Facets, 1)
	require.Contains(t, result.Facets, "type")
	require.Equal(t, []string{"http", "event"}, result.Facets["type"])
}
