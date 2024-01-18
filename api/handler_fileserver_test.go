package api_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/api"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/try"
	"mokapi/webui"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestHandler_FileServer(t *testing.T) {
	testcases := []struct {
		name       string
		config     static.Api
		fn         func(t *testing.T, h http.Handler)
		fileServer http.Handler
	}{
		{
			name:   "request api info",
			config: static.Api{Path: "/mokapi", Dashboard: true},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":""}`))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/index.html" {
					writer.WriteHeader(404)
				}
			}),
		},
		{
			name:   "request web app",
			config: static.Api{Path: "/mokapi", Dashboard: true},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/index.html",
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/index.html" {
					writer.WriteHeader(404)
				}
			}),
		},
		{
			name:   "request web app",
			config: static.Api{Path: "/mokapi/dashboard", Dashboard: true},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/index.html",
					nil,
					"",
					h,
					try.HasStatusCode(200), try.BodyContains(`<base href="/mokapi/dashboard/" />`))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/index.html" {
					writer.WriteHeader(404)
				}
			}),
		},
		{
			name:   "request asset",
			config: static.Api{Path: "/mokapi/dashboard", Dashboard: true},
			fn: func(t *testing.T, h http.Handler) {
				// select an asset from fs
				dir, err := webui.App.ReadDir("dist/assets")
				require.NoError(t, err)
				// asset: DashboardView-7c070f1a.js
				asset := strings.TrimSuffix(dir[0].Name(), "dist/")

				try.Handler(t,
					http.MethodGet,
					fmt.Sprintf("http://foo.api/mokapi/dashboard/assets/%v", asset),
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/assets/index.js" {
					writer.WriteHeader(404)
				}
			}),
		},
		{
			name:   "request svg",
			config: static.Api{Path: "/mokapi/dashboard", Dashboard: true},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/logo.svg",
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/logo.svg" {
					writer.WriteHeader(404)
				}
			}),
		},
		{
			name:   "request png",
			config: static.Api{Path: "/mokapi/dashboard", Dashboard: true},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/mail.png",
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/mail.png" {
					writer.WriteHeader(404)
				}
			}),
		},
		{
			name:   "url rewrite (proxy)",
			config: static.Api{Path: "/mokapi/dashboard", Base: "/foo/mokapi/dashboard", Dashboard: true},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/foo/mokapi/dashboard/index.html",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.BodyContains(`<base href="/foo/mokapi/dashboard/" />`))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/index.html" {
					writer.WriteHeader(404)
				}
			}),
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := api.New(runtime.New(), tc.config)
			tc.fn(t, h)
		})
	}
}

func TestOpenGraphInDashboard(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "http service",
			test: func(t *testing.T) {
				app := runtime.New()
				app.AddHttp(&dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("https://foo.bar")}, Data: openapitest.NewConfig("3.0", openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."))})
				h := api.New(app, static.Api{Path: "/mokapi", Dashboard: true})
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/http/services/Swagger%20Petstore?refresh=20",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.BodyContains(`<title>Swagger Petstore | mokapi.io</title>`),
					try.BodyContains(`<meta name="description" content="This is a sample server Petstore server." />`),
					try.BodyContains(`<meta property="og:title" content="Swagger Petstore | mokapi.io">`),
					try.BodyContains(`<meta property="og:description" content="This is a sample server Petstore server.">`))
			},
		},
		{
			name: "http service path without summary and description",
			test: func(t *testing.T) {
				app := runtime.New()
				app.AddHttp(&dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("https://foo.bar")}, Data: openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}", openapitest.NewPath()),
				)},
				)
				h := api.New(app, static.Api{Path: "/mokapi", Dashboard: true})
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/http/services/Swagger%20Petstore/pet%2F%7BpetId%7D?refresh=20",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.BodyContains(`<title>/pet/{petId} - Swagger Petstore | mokapi.io</title>`),
					try.BodyContains(`<meta name="description" content="This is a sample server Petstore server." />`),
					try.BodyContains(`<meta property="og:title" content="/pet/{petId} - Swagger Petstore | mokapi.io">`),
					try.BodyContains(`<meta property="og:description" content="This is a sample server Petstore server.">`))
			},
		},
		{
			name: "http service path with summary and description",
			test: func(t *testing.T) {
				app := runtime.New()
				app.AddHttp(&dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("https://foo.bar")}, Data: openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}", openapitest.NewPath(
						openapitest.WithPathInfo("foo", "bar"),
					)),
				),
				})
				h := api.New(app, static.Api{Path: "/mokapi", Dashboard: true})
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/http/services/Swagger%20Petstore/pet%2F%7BpetId%7D?refresh=20",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.BodyContains(`<title>/pet/{petId} - Swagger Petstore | mokapi.io</title>`),
					try.BodyContains(`<meta name="description" content="foo" />`),
					try.BodyContains(`<meta property="og:title" content="/pet/{petId} - Swagger Petstore | mokapi.io">`),
					try.BodyContains(`<meta property="og:description" content="foo">`))
			},
		},
		{
			name: "http service path with no summary but description",
			test: func(t *testing.T) {
				app := runtime.New()
				app.AddHttp(&dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("https://foo.bar")}, Data: openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}", openapitest.NewPath(
						openapitest.WithPathInfo("", "bar"),
					))),
				})
				h := api.New(app, static.Api{Path: "/mokapi", Dashboard: true})
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/http/services/Swagger%20Petstore/pet%2F%7BpetId%7D?refresh=20",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.BodyContains(`<title>/pet/{petId} - Swagger Petstore | mokapi.io</title>`),
					try.BodyContains(`<meta name="description" content="bar" />`),
					try.BodyContains(`<meta property="og:title" content="/pet/{petId} - Swagger Petstore | mokapi.io">`),
					try.BodyContains(`<meta property="og:description" content="bar">`))
			},
		},
		{
			name: "http service endpoint no summary and no description",
			test: func(t *testing.T) {
				app := runtime.New()
				app.AddHttp(&dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("https://foo.bar")}, Data: openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}", openapitest.NewPath(
						openapitest.WithOperation("GET", openapitest.NewOperation()),
					))),
				})
				h := api.New(app, static.Api{Path: "/mokapi", Dashboard: true})
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/http/services/Swagger%20Petstore/pet%2F%7BpetId%7D/get?refresh=20",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.BodyContains(`<title>GET /pet/{petId} - Swagger Petstore | mokapi.io</title>`),
					try.BodyContains(`<meta name="description" content="This is a sample server Petstore server." />`),
					try.BodyContains(`<meta property="og:title" content="GET /pet/{petId} - Swagger Petstore | mokapi.io">`),
					try.BodyContains(`<meta property="og:description" content="This is a sample server Petstore server.">`))
			},
		},
		{
			name: "http service endpoint get right path",
			test: func(t *testing.T) {
				app := runtime.New()
				app.AddHttp(&dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("https://foo.bar")}, Data: openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}", openapitest.NewPath(
						openapitest.WithOperation("GET", openapitest.NewOperation()),
					)),
					openapitest.WithPath("/pet/{petId}/foo", openapitest.NewPath(
						openapitest.WithOperation("GET", openapitest.NewOperation()),
					))),
				})
				h := api.New(app, static.Api{Path: "/mokapi", Dashboard: true})
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/http/services/Swagger%20Petstore/pet%2F%7BpetId%7D%2Ffoo/get?refresh=20",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.BodyContains(`<title>GET /pet/{petId}/foo - Swagger Petstore | mokapi.io</title>`),
					try.BodyContains(`<meta name="description" content="This is a sample server Petstore server." />`),
					try.BodyContains(`<meta property="og:title" content="GET /pet/{petId}/foo - Swagger Petstore | mokapi.io">`),
					try.BodyContains(`<meta property="og:description" content="This is a sample server Petstore server.">`))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
