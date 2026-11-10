package acceptance

import (
	"fmt"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// HttpValidateSuite uses the validate endpoint like a client such as the Storybook addon: it
// sends request/response pairs that were produced elsewhere and expects the documented result.
type HttpValidateSuite struct {
	BaseSuite
	apiUrl string
}

const profileContract = `{
  "openapi": "3.0.3",
  "info": { "title": "Profile API", "version": "1" },
  "servers": [{ "url": "http://localhost:%d/api/v1" }],
  "paths": {
    "/profile": {
      "get": {
        "parameters": [{ "name": "id", "in": "query", "required": true, "schema": { "type": "integer" } }],
        "responses": {
          "200": {
            "description": "ok",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "required": ["name", "email"],
                  "properties": { "name": { "type": "string" }, "email": { "type": "string" } }
                }
              }
            }
          }
        }
      }
    }
  }
}`

// profileV2Contract shares GET /api/v1/profile with the Profile API
const profileV2Contract = `{
  "openapi": "3.0.3",
  "info": { "title": "Profile API v2", "version": "2" },
  "servers": [{ "url": "http://localhost:%d/api/v1" }],
  "paths": {
    "/profile": {
      "get": {
        "responses": {
          "200": {
            "description": "ok",
            "content": {
              "application/json": {
                "schema": { "type": "object", "required": ["name"], "properties": { "name": { "type": "string" } } }
              }
            }
          }
        }
      }
    }
  }
}`

const validPair = `{
  "request": { "method": "GET", "url": "/api/v1/profile?id=1", "headers": { "Accept": "application/json" } },
  "response": { "statusCode": 200, "headers": { "Content-Type": "application/json" }, "body": "{\"name\":\"Jane Doe\",\"email\":\"jane@foo.bar\"}" }
}`

func (suite *HttpValidateSuite) SetupSuite() {
	dir := suite.T().TempDir()
	suite.writeFile(dir, "profile.json", fmt.Sprintf(profileContract, try.GetFreePort()))
	suite.writeFile(dir, "profile-v2.json", fmt.Sprintf(profileV2Contract, try.GetFreePort()))

	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: dir}}
	suite.initCmd(cfg)

	suite.apiUrl = fmt.Sprintf("http://localhost:%d/api", cfg.Api.Port)
	try.GetRequest(suite.T(), suite.apiUrl+"/services?type=http", nil,
		try.BodyContains("Profile API v2"), try.BodyContains(`"Profile API"`))
}

func (suite *HttpValidateSuite) writeFile(dir, name, content string) {
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}

func (suite *HttpValidateSuite) validate(payload string, conditions ...try.ResponseCondition) {
	try.Request(suite.T(), http.MethodPost, suite.apiUrl+"/services/http/validate",
		map[string]string{"Content-Type": "application/json"}, strings.NewReader(payload), conditions...)
}

func (suite *HttpValidateSuite) TestValidPairWithBasePath() {
	suite.validate(`{ "api": "Profile API",`+strings.TrimPrefix(validPair, "{"),
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.JSONEq(t, `{
				"valid": true,
				"operation": { "api": "Profile API", "method": "GET", "path": "/profile" },
				"errors": []
			}`, body)
		}))
}

func (suite *HttpValidateSuite) TestInvalidResponseBody() {
	suite.validate(`{
  "api": "Profile API",
  "request": { "method": "GET", "url": "/api/v1/profile?id=1" },
  "response": { "statusCode": 200, "headers": { "Content-Type": "application/json" }, "body": "{\"name\":\"Jane Doe\"}" }
}`,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.JSONEq(t, `{
				"valid": false,
				"operation": { "api": "Profile API", "method": "GET", "path": "/profile" },
				"errors": [
					{ "location": "response.body", "schema": "#/required", "message": "required properties are missing: email" }
				]
			}`, body)
		}))
}

func (suite *HttpValidateSuite) TestNoOperation() {
	suite.validate(`{
  "api": "Profile API",
  "request": { "method": "GET", "url": "/api/v1/unknown" },
  "response": { "statusCode": 200 }
}`,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.JSONEq(t, `{
				"valid": false,
				"errors": [{ "location": "operation", "message": "no operation found for GET /api/v1/unknown" }]
			}`, body)
		}))
}

func (suite *HttpValidateSuite) TestAmbiguousWithoutApiIsResolvedWithApi() {
	suite.validate(validPair,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.Contains(t, body, `"valid":false`)
			assert.Contains(t, body, `"location":"operation"`)
			assert.Contains(t, body, "Profile API")
			assert.Contains(t, body, "Profile API v2")
		}))

	suite.validate(`{ "api": "Profile API v2",`+strings.TrimPrefix(validPair, "{"),
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			assert.Contains(t, body, `"valid":true`)
			assert.Contains(t, body, `"api":"Profile API v2"`)
		}))
}

func (suite *HttpValidateSuite) TestUnknownApi() {
	suite.validate(`{ "api": "Unknown API",`+strings.TrimPrefix(validPair, "{"),
		try.HasStatusCode(http.StatusNotFound),
		try.AssertBody(func(t *testing.T, body string) {
			assert.JSONEq(t, `{"message":"api 'Unknown API' not found"}`, body)
		}))
}

func (suite *HttpValidateSuite) TestMalformedPayload() {
	suite.validate(`{ "request": { "method": "GET" } }`,
		try.HasStatusCode(http.StatusBadRequest),
		try.AssertBody(func(t *testing.T, body string) {
			assert.Contains(t, body, `{"message":`)
		}))
}

func (suite *HttpValidateSuite) TestValidationCreatesNoEvent() {
	var before string
	try.GetRequest(suite.T(), suite.apiUrl+"/events", nil,
		try.AssertBody(func(t *testing.T, body string) { before = body }))

	suite.validate(`{ "api": "Profile API",`+strings.TrimPrefix(validPair, "{"), try.HasStatusCode(http.StatusOK))

	try.GetRequest(suite.T(), suite.apiUrl+"/events", nil,
		try.AssertBody(func(t *testing.T, body string) { assert.Equal(t, before, body) }))
}

func (suite *HttpValidateSuite) TestPreflight() {
	try.Request(suite.T(), http.MethodOptions, suite.apiUrl+"/services/http/validate", nil, nil,
		try.HasStatusCode(http.StatusNoContent),
		try.HasHeader("Access-Control-Allow-Origin", "*"),
		try.HasHeader("Access-Control-Allow-Methods", "GET, POST"),
		try.HasHeader("Access-Control-Allow-Headers", "Content-Type"))
}
