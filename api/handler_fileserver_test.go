package api_test

import (
	"fmt"
	"io/fs"
	"mokapi/api"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/try"
	"mokapi/webui"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler_FileServer(t *testing.T) {
	testcases := []struct {
		name       string
		config     static.Api
		fn         func(t *testing.T, h http.Handler)
		fileServer http.Handler
	}{
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
		},
		{
			name:   "request png",
			config: static.Api{Path: "/mokapi/dashboard", Dashboard: true},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/icon-180.png",
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
		},
		{
			name:   "request docs image",
			config: static.Api{Path: "/mokapi/dashboard", Dashboard: true},
			fn: func(t *testing.T, h http.Handler) {
				// docs images are bundled with a hashed file name
				images, err := fs.Glob(webui.App, "dist/assets/*.png")
				require.NoError(t, err)
				require.NotEmpty(t, images)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/assets/"+path.Base(images[0]),
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
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
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &static.Config{}
			h := api.New(runtime.New(cfg, &dynamictest.Reader{}), tc.config)
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
				cfg := &static.Config{}
				app := runtime.New(cfg, &dynamictest.Reader{})
				app.Http.Add(dynamictest.NewConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server.")), dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustParse("https://foo.bar")})))
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
				cfg := &static.Config{}
				app := runtime.New(cfg, &dynamictest.Reader{})
				app.Http.Add(dynamictest.NewConfig(openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}"),
				), dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustParse("https://foo.bar")})),
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
				cfg := &static.Config{}
				app := runtime.New(cfg, &dynamictest.Reader{})
				app.Http.Add(dynamictest.NewConfig(openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}",
						openapitest.WithPathInfo("foo", "bar"),
					),
				), dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustParse("https://foo.bar")})))
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
				cfg := &static.Config{}
				app := runtime.New(cfg, &dynamictest.Reader{})
				app.Http.Add(dynamictest.NewConfig(openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}",
						openapitest.WithPathInfo("", "bar"),
					)), dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustParse("https://foo.bar")})))
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
				cfg := &static.Config{}
				app := runtime.New(cfg, &dynamictest.Reader{})
				app.Http.Add(dynamictest.NewConfig(openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}",
						openapitest.WithOperation("GET"),
					)), dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustParse("https://foo.bar")})))
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
				cfg := &static.Config{}
				app := runtime.New(cfg, &dynamictest.Reader{})
				app.Http.Add(dynamictest.NewConfig(openapitest.NewConfig("3.0",
					openapitest.WithInfo("Swagger Petstore", "1.0", "This is a sample server Petstore server."),
					openapitest.WithPath("/pet/{petId}",
						openapitest.WithOperation("GET"),
					),
					openapitest.WithPath("/pet/{petId}/foo",
						openapitest.WithOperation("GET"),
					),
				), dynamictest.WithInfo(dynamic.ConfigInfo{Url: mustParse("https://foo.bar")})))
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

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
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

// The dashboard can be served under a path prefix (<base href>). Every image the built
// entry chunk refers to must resolve below that prefix.
func TestHandler_FileServer_AssetsBelowPathPrefix(t *testing.T) {
	const prefix = "/mokapi/dashboard"
	h := api.New(runtime.New(&static.Config{}, &dynamictest.Reader{}), static.Api{Path: prefix, Dashboard: true})

	get := func(p string) (int, string) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://foo.api"+p, nil))
		return rr.Code, rr.Body.String()
	}

	_, index := get(prefix + "/")
	entry := regexp.MustCompile(`src="\./(assets/index-[^"]+\.js)"`).FindStringSubmatch(index)
	require.Len(t, entry, 2, "entry chunk not found in index.html")
	_, chunk := get(prefix + "/" + entry[1])

	t.Run("images referenced by the entry chunk", func(t *testing.T) {
		refs := regexp.MustCompile("new URL\\(`(\\.\\./[^`]+)`,import\\.meta\\.url\\)").FindAllStringSubmatch(chunk, -1)
		require.NotEmpty(t, refs)
		for _, ref := range refs {
			u, err := url.Parse("http://foo.api" + prefix + "/" + entry[1])
			require.NoError(t, err)
			resolved := u.ResolveReference(&url.URL{Path: ref[1]})
			code, _ := get(resolved.Path)
			require.Equal(t, 200, code, "%v -> %v", ref[1], resolved.Path)
		}
	})
}
