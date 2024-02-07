package openapi_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"net/http/httptest"
	"testing"
)

func TestContentTypeFromRequest(t *testing.T) {
	testcases := []struct {
		accept   string
		response *openapi.Response
		validate func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error)
	}{
		{
			accept:   "",
			response: openapitest.NewResponse(),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				// empty response should not get error
				require.Equal(t, media.Empty, ct)
				require.Nil(t, mt)
				require.Nil(t, err)
			},
		},
		{
			accept:   "text/plain",
			response: openapitest.NewResponse(),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				// empty response should not get error
				require.Equal(t, media.Empty, ct)
				require.Nil(t, mt)
				require.Nil(t, err)
			},
		},
		{
			accept:   "",
			response: openapitest.NewResponse(openapitest.WithContent("text/plain", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "text", ct.Type)
				require.Equal(t, "plain", ct.Subtype)
			},
		},
		{
			accept:   "text/plain",
			response: openapitest.NewResponse(openapitest.WithContent("text/plain", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "text", ct.Type)
				require.Equal(t, "plain", ct.Subtype)
			},
		},
		{
			accept:   "*/*",
			response: openapitest.NewResponse(openapitest.WithContent("text/plain", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "text", ct.Type)
				require.Equal(t, "plain", ct.Subtype)
			},
		},
		{
			accept:   "text/*",
			response: openapitest.NewResponse(openapitest.WithContent("text/plain", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "text", ct.Type)
				require.Equal(t, "plain", ct.Subtype)
			},
		},
		{
			accept:   "image/*",
			response: openapitest.NewResponse(openapitest.WithContent("text/plain", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.EqualError(t, err, "none of requests content type(s) are supported: \"image/*\"")
			},
		},
		{
			accept:   "image/*",
			response: openapitest.NewResponse(openapitest.WithContent("image/png", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "image", ct.Type)
				require.Equal(t, "png", ct.Subtype)
			},
		},
		{
			accept:   "image/png",
			response: openapitest.NewResponse(openapitest.WithContent("image/*", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "image/png", ct.String())
			},
		},
		{
			accept:   "image/*",
			response: openapitest.NewResponse(openapitest.WithContent("image/*", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "image/x-xpixmap", ct.String())
			},
		},
		{
			accept:   "image/*",
			response: openapitest.NewResponse(openapitest.WithContent("*/*", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "image/x-xpixmap", ct.String())
			},
		},
		{
			accept:   "*/*",
			response: openapitest.NewResponse(openapitest.WithContent("*/*", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "application/json", ct.String())
				require.Equal(t, media.Default.String(), ct.String())
			},
		},
		{
			accept:   "",
			response: openapitest.NewResponse(openapitest.WithContent("*/*", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "application/json", ct.String())
				require.Equal(t, media.Default.String(), ct.String())
			},
		},
		{
			accept:   "image/png",
			response: openapitest.NewResponse(openapitest.WithContent("*/*", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "image/png", ct.String())
			},
		},
		{
			accept:   "application/json;odata=verbose",
			response: openapitest.NewResponse(openapitest.WithContent("application/json", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "application/json", ct.String())
			},
		},
		{
			accept:   "application/json",
			response: openapitest.NewResponse(openapitest.WithContent("application/json;odata=verbose", openapitest.NewContent())),
			validate: func(t *testing.T, ct media.ContentType, mt *openapi.MediaType, err error) {
				require.NotNil(t, ct)
				require.NotNil(t, mt)
				require.NoError(t, err)
				require.Equal(t, "application/json;odata=verbose", ct.String())
			},
		},
	}

	for _, testcase := range testcases {
		test := testcase
		t.Run(test.accept, func(t *testing.T) {
			media.SetFaker(11)
			gofakeit.SetGlobalFaker(gofakeit.New(11))

			r := httptest.NewRequest("GET", "http://foo", nil)
			r.Header.Add("accept", test.accept)
			ct, mt, err := openapi.ContentTypeFromRequest(r, test.response)
			test.validate(t, ct, mt, err)
		})
	}
}
