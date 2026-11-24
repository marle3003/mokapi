package acceptance

import (
	"encoding/json"
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
          },
          "503": { "description": "Service Unavailable" }
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

// orderContract uses a header, nested objects, objects in arrays, dependentRequired and oneOf
const orderContract = `{
  "openapi": "3.1.0",
  "info": { "title": "Order API", "version": "1" },
  "servers": [{ "url": "http://localhost:%d/orders-api" }],
  "paths": {
    "/orders/1": {
      "get": {
        "responses": {
          "200": {
            "description": "ok",
            "headers": { "X-Request-Id": { "required": true, "schema": { "type": "string" } } },
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "required": ["id", "customer", "items"],
                  "properties": {
                    "id": { "type": "integer" },
                    "status": { "type": "string", "enum": ["open", "closed"] },
                    "customer": {
                      "type": "object",
                      "required": ["name", "email"],
                      "properties": { "name": { "type": "string" }, "email": { "type": "string" } }
                    },
                    "items": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "required": ["sku", "qty"],
                        "properties": { "sku": { "type": "string" }, "qty": { "type": "integer" } }
                      }
                    },
                    "card": {
                      "type": "object",
                      "dependentRequired": { "number": ["cvv"] },
                      "properties": { "number": { "type": "string" }, "cvv": { "type": "string" } }
                    },
                    "pet": {
                      "oneOf": [
                        {
                          "type": "object",
                          "required": ["bark"],
                          "additionalProperties": false,
                          "properties": { "kind": { "type": "string" }, "bark": { "type": "string" } }
                        },
                        {
                          "type": "object",
                          "required": ["meow"],
                          "additionalProperties": false,
                          "properties": { "kind": { "type": "string" }, "meow": { "type": "string" } }
                        }
                      ]
                    }
                  }
                }
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

	suite.writeFile(dir, "order.json", fmt.Sprintf(orderContract, try.GetFreePort()))

	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: dir}}
	suite.initCmd(cfg)

	suite.apiUrl = fmt.Sprintf("http://localhost:%d/api", cfg.Api.Port)
	try.GetRequest(suite.T(), suite.apiUrl+"/services?type=http", nil,
		try.BodyContains("Profile API v2"), try.BodyContains(`"Profile API"`), try.BodyContains("Order API"))
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

func partialProfilePayload(partial bool, query, body string) string {
	return fmt.Sprintf(`{
  "api": "Profile API",
  "partialResponse": %t,
  "request": { "method": "GET", "url": "/api/v1/profile%s" },
  "response": { "statusCode": 200, "headers": { "Content-Type": "application/json" }, "body": %q }
}`, partial, query, body)
}

func partialOrderPayload(partial bool, headers, body string) string {
	return fmt.Sprintf(`{
  "api": "Order API",
  "partialResponse": %t,
  "request": { "method": "GET", "url": "/orders-api/orders/1" },
  "response": { "statusCode": 200, "headers": %s, "body": %q }
}`, partial, headers, body)
}

const orderHeaders = `{ "Content-Type": "application/json", "X-Request-Id": "abc" }`

func (suite *HttpValidateSuite) assertErrors(valid bool, partialEcho bool, locationAndMessages ...string) []try.ResponseCondition {
	return []try.ResponseCondition{
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var result struct {
				Valid           bool `json:"valid"`
				PartialResponse *bool
				Errors          []struct{ Location, Message string }
			}
			assert.NoError(t, json.Unmarshal([]byte(body), &result))
			assert.Equal(t, valid, result.Valid, body)
			if partialEcho {
				if assert.NotNil(t, result.PartialResponse, body) {
					assert.True(t, *result.PartialResponse)
				}
			} else {
				assert.NotContains(t, body, "partialResponse")
			}
			var actual []string
			for _, e := range result.Errors {
				actual = append(actual, e.Location+": "+e.Message)
			}
			assert.Equal(t, len(locationAndMessages), len(actual), body)
			for _, expected := range locationAndMessages {
				assert.Condition(t, func() bool {
					for _, a := range actual {
						if strings.Contains(a, expected) {
							return true
						}
					}
					return false
				}, "expected error %q in %v", expected, actual)
			}
		}),
	}
}

func (suite *HttpValidateSuite) TestPartialResponseSkipsMissingRequiredProperties() {
	suite.validate(partialProfilePayload(true, "?id=1", `{"name":"Rex"}`), suite.assertErrors(true, true)...)
}

func (suite *HttpValidateSuite) TestStrictResponseStillReportsMissingRequiredProperties() {
	suite.validate(partialProfilePayload(false, "?id=1", `{"name":"Rex"}`),
		suite.assertErrors(false, false, "response.body: required properties are missing: email")...)
}

func (suite *HttpValidateSuite) TestPartialResponseStillReportsWrongType() {
	suite.validate(partialProfilePayload(true, "?id=1", `{"name":42}`),
		suite.assertErrors(false, true, "response.body")...)
}

func (suite *HttpValidateSuite) TestPartialResponseValidatesRequestStrictly() {
	suite.validate(partialProfilePayload(true, "", `{"name":"Rex"}`),
		suite.assertErrors(false, true, "request.query.id")...)
}

func (suite *HttpValidateSuite) TestPartialResponseWithoutMatchingOperationIsEchoed() {
	suite.validate(`{
  "api": "Profile API",
  "partialResponse": true,
  "request": { "method": "GET", "url": "/api/v1/unknown" },
  "response": { "statusCode": 200 }
}`, suite.assertErrors(false, true, "operation: no operation found")...)
}

func (suite *HttpValidateSuite) TestPartialResponseIsEchoedForAmbiguousOperation() {
	suite.validate(`{
  "partialResponse": true,
  "request": { "method": "GET", "url": "/api/v1/profile?id=1" },
  "response": { "statusCode": 200, "body": "{}" }
}`, suite.assertErrors(false, true, "operation: ambiguous operation")...)
}

func (suite *HttpValidateSuite) TestPartialResponseStillReportsMissingResponseHeader() {
	suite.validate(partialOrderPayload(true, `{ "Content-Type": "application/json" }`, `{"id":1}`),
		suite.assertErrors(false, true, "response.header.X-Request-Id")...)
}

func (suite *HttpValidateSuite) TestPartialResponseSkipsRequiredInNestedObjectsAndArrays() {
	suite.validate(partialOrderPayload(true, orderHeaders, `{"customer":{"name":"Jane"},"items":[{"sku":"a"},{"qty":2}]}`),
		suite.assertErrors(true, true)...)
	suite.validate(partialOrderPayload(false, orderHeaders, `{"customer":{"name":"Jane"},"items":[{"sku":"a"}]}`),
		suite.assertErrors(false, false,
			"missing: id", "missing: email", "missing: qty")...)
}

func (suite *HttpValidateSuite) TestPartialResponseStillReportsEnumAndNestedTypeErrors() {
	suite.validate(partialOrderPayload(true, orderHeaders, `{"status":"pending"}`),
		suite.assertErrors(false, true, "response.body")...)
	suite.validate(partialOrderPayload(true, orderHeaders, `{"items":[{"qty":"two"}]}`),
		suite.assertErrors(false, true, "response.body")...)
}

func (suite *HttpValidateSuite) TestPartialResponseSkipsDependentRequired() {
	suite.validate(partialOrderPayload(true, orderHeaders, `{"card":{"number":"4111"}}`), suite.assertErrors(true, true)...)
	suite.validate(partialOrderPayload(false, orderHeaders, `{"id":1,"customer":{"name":"a","email":"b"},"items":[],"card":{"number":"4111"}}`),
		suite.assertErrors(false, false, "response.body")...)
}

func (suite *HttpValidateSuite) TestPartialResponseAcceptsSeveralOneOfMatches() {
	suite.validate(partialOrderPayload(true, orderHeaders, `{"pet":{"kind":"dog"}}`), suite.assertErrors(true, true)...)
}

func (suite *HttpValidateSuite) TestPartialResponseRejectsBodyMatchingNoOneOfVariant() {
	suite.validate(partialOrderPayload(true, orderHeaders, `{"pet":{"bark":1}}`),
		suite.assertErrors(false, true, "response.body")...)
}

func serviceUnavailablePayload(body string) string {
	return fmt.Sprintf(`{
  "api": "Profile API",
  "request": { "method": "GET", "url": "/api/v1/profile?id=1" },
  "response": { "statusCode": 503, "headers": { "Content-Type": "application/json" }, "body": %q }
}`, body)
}

// The contract defines 503 with a description only, so the Content-Type of the mocked response
// (e.g. MSW HttpResponse.json) has no content definition to be checked against.
func (suite *HttpValidateSuite) TestResponseWithoutDefinedContentAcceptsJsonBody() {
	suite.validate(serviceUnavailablePayload(`{"error":"down"}`), suite.assertErrors(true, false)...)
}

func (suite *HttpValidateSuite) TestResponseWithoutDefinedContentAcceptsEmptyBodyWithContentType() {
	suite.validate(serviceUnavailablePayload(""), suite.assertErrors(true, false)...)
}
