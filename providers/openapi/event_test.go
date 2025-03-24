package openapi_test

import (
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEvent(t *testing.T) {
	testcases := []struct {
		name   string
		config *openapi.Config
		test   func(t *testing.T, h http.Handler)
	}{
		{
			name: "use response example (deprecated)",
			config: openapitest.NewConfig("3.1.0",
				openapitest.WithPath("/foo", openapitest.NewPath(
					openapitest.WithOperation("GET", openapitest.NewOperation(
						openapitest.WithResponse(200, openapitest.WithContent(
							"application/json", &openapi.MediaType{
								Schema:      nil,
								Example:     "foo",
								ContentType: media.ContentType{},
								Encoding:    nil,
							},
						)),
					)),
				)),
			),
			test: func(t *testing.T, h http.Handler) {
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"foo"`, rr.Body.String())
			},
		},
		{
			name: "use response examples",
			config: openapitest.NewConfig("3.1.0",
				openapitest.WithPath("/foo", openapitest.NewPath(
					openapitest.WithOperation("GET", openapitest.NewOperation(
						openapitest.WithResponse(200, openapitest.WithContent(
							"application/json", &openapi.MediaType{
								Schema:  nil,
								Example: nil,
								Examples: openapi.Examples{
									"foo": {
										Value: &openapi.Example{
											Summary:       "",
											Description:   "",
											Value:         "foo",
											ExternalValue: "",
										},
									},
								},
								ContentType: media.ContentType{},
								Encoding:    nil,
							},
						)),
					)),
				)),
			),
			test: func(t *testing.T, h http.Handler) {
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"foo"`, rr.Body.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			test.NewNullLogger()

			h := openapi.NewHandler(tc.config, enginetest.NewEngine())

			tc.test(t, h)
		})
	}
}
