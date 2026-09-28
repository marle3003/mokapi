package acceptance

import (
	"encoding/json"
	"fmt"
	"mokapi/config/static"
	"mokapi/schema/json/generator"
	"mokapi/try"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type PetStoreSuite struct{ BaseSuite }

func (suite *PetStoreSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Health.Port = cfg.Api.Port
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./petstore"}}
	cfg.Api.Search.Enabled = true
	cfg.Api.Search.InMemory = true
	cfg.Api.Search.NumIndexWorker = 1
	suite.initCmd(cfg)
}

func (suite *PetStoreSuite) SetupTest() {
	generator.Seed(11)
}

func (suite *PetStoreSuite) TestApi() {
	suite.T().Run("CORS", func(t *testing.T) {
		try.GetRequest(t, fmt.Sprintf("http://127.0.0.1:%v", suite.cfg.Api.Port),
			nil,
			try.HasStatusCode(http.StatusOK),
			try.HasHeader("Access-Control-Allow-Origin", "*"))
	})

	suite.T().Run("get Swagger HTTP service", func(t *testing.T) {
		try.GetRequest(t, fmt.Sprintf("http://127.0.0.1:%v/api/services/http/Swagger%%20Petstore", suite.cfg.Api.Port),
			nil,
			try.HasStatusCode(http.StatusOK),
			try.BodyContains(`{"name":"Swagger Petstore","description":"This is a sample server Petstore server.  You can find out more about `),
			try.BodyMatch(`"configs":\[{"id":".*","url":".*\/acceptance\/petstore\/openapi\.yml","provider":"file","time":".*"}\]`),
		)
	})
}

func (suite *PetStoreSuite) TestJsHttpHandler() {
	// ensure scripts are executed
	time.Sleep(4 * time.Second)
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/2",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusNotFound),
		try.HasBody(""))

	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/3",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusNotFound),
		try.HasBody(""))

	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/4",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusInternalServerError),
		try.HasBody("HTTP body marshalling failed.\n\nBody: {}\n\nValidation error count 1:\n\t- #/required: required properties are missing: name, photoUrls\n"))

	// use generated data but change pet's name
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/5",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusOK),
		try.BodyContains(`"name":"Zoe"`))

	// test http metrics
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%d/api/metrics/http?path=/pet/{petId}", suite.cfg.Api.Port), nil,
		try.BodyContains(`http_requests_total{service=\"Swagger Petstore\",endpoint=\"/pet/{petId}\",method=\"GET\"}","value":4}`),
	)
}

func (suite *PetStoreSuite) TestNewJsHttpHandler() {
	time.Sleep(4 * time.Second)
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/findByStatus?status=sold",
		map[string]string{"Accept": "application/json", "api_key": "123"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`[{"name":"Zoe","photoUrls":[]}]`))
}

func (suite *PetStoreSuite) TestLuaFile() {
	// ensure scripts are executed
	time.Sleep(2 * time.Second)
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/pet/findByStatus?status=available&status=pending",
		map[string]string{"Accept": "application/json", "Authorization": "foo"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody("[{\"name\":\"Gidget\",\"photoUrls\":[\"http://www.pets.com/gidget.png\"],\"status\":\"pending\"},{\"name\":\"Max\",\"photoUrls\":[\"http://www.pets.com/max.png\"],\"status\":\"available\"}]"))
}

func (suite *PetStoreSuite) TestGetOrderById() {
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/store/order/1",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"petId":47057,"shipDate":"1995-11-16T10:19:08Z","status":"placed","complete":true}`))

	try.GetRequest(suite.T(), "https://localhost:18443/store/order/10",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		// properties like id or petId are optional
		try.HasBody(`{"id":50660,"petId":89957,"quantity":53,"shipDate":"1981-01-14T21:14:21Z","complete":true}`))
}

func (suite *PetStoreSuite) TestTls() {
	try.GetRequest(suite.T(), "https://localhost:18443/store/order/10",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK),
		try.IsTls("localhost"),
	)
}

func (suite *PetStoreSuite) TestEvents() {
	try.GetRequest(suite.T(), "http://127.0.0.1:18080/user/bob",
		map[string]string{"Accept": "application/json"},
		try.HasStatusCode(http.StatusOK))

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/events?namespace=http&name=Swagger%%20Petstore&path=/user/{username}", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.BodyContains(`"url":"http://127.0.0.1:18080/user/bob"`))

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/search/query?q=type:event%%20event.traits.namespace=http", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var data map[string]any
			err := json.Unmarshal([]byte(body), &data)
			assert.NoError(t, err)
			results := data["results"].([]any)
			assert.NotNil(t, results, "search result should contain results")
			assert.Greater(t, len(results), 0)
			evt, ok := results[0].(map[string]any)
			assert.True(t, ok, "event should be a map[string]any")
			assert.NotNil(t, evt)
			assert.Equal(t, "Event", evt["type"])
			assert.Equal(t, "http://127.0.0.1:18080/user/bob", evt["title"])
			assert.Equal(t, "Swagger Petstore", evt["domain"])
			params := evt["params"].(map[string]any)
			assert.Len(t, params, 6)
			assert.Equal(t, "event", params["type"])
			assert.Equal(t, "http", params["traits.namespace"])
			assert.Equal(t, "Swagger Petstore", params["traits.name"])
			assert.Equal(t, "/user/{username}", params["traits.path"])
			assert.Equal(t, "GET", params["traits.method"])
			assert.Equal(t, "Swagger Petstore", params["traits.name"])
		}),
	)
}

func (suite *PetStoreSuite) TestSearch_Paging() {
	time.Sleep(3 * time.Second)

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/search/query?q=api:%%22Swagger%%20Petstore%%22%%20type:http", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var data map[string]any
			err := json.Unmarshal([]byte(body), &data)
			assert.NoError(t, err)
			assert.NotNil(t, data)

			assert.Equal(t, float64(51), data["total"])

			items := data["results"].([]any)
			assert.Len(t, items, 10)
			evt := items[0].(map[string]any)
			assert.Equal(t, "HTTP", evt["type"])
			assert.Equal(t, "Swagger Petstore", evt["title"])
			assert.NotContains(t, evt, "domain")
		}),
	)

	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/api/search/query?q=api:%%22Swagger%%20Petstore%%22%%20type:http&index=1", suite.cfg.Api.Port),
		nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var data map[string]any
			err := json.Unmarshal([]byte(body), &data)
			assert.NoError(t, err)
			assert.NotNil(t, data)

			assert.Equal(t, float64(51), data["total"])

			items := data["results"].([]any)
			assert.Len(t, items, 10)
			evt := items[0].(map[string]any)
			assert.Equal(t, "HTTP", evt["type"])
			assert.Equal(t, "/pet/{petId}/uploadImage", evt["title"])
			assert.Equal(t, "Swagger Petstore", evt["domain"])
		}),
	)
}

func (suite *PetStoreSuite) TestHealth() {
	try.GetRequest(suite.T(), fmt.Sprintf("http://127.0.0.1:%v/health", suite.cfg.Api.Port), nil,
		try.HasStatusCode(http.StatusOK),
		try.HasBody(`{"status":"healthy"}`),
	)
}
