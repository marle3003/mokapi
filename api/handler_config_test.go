package api

import (
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestHandler_Config(t *testing.T) {
	mustTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t
	}
	mustUrl := func(s string) *url.URL {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		return u
	}

	testcases := []struct {
		name       string
		app        func() *runtime.App
		requestUrl string
		test       []try.ResponseCondition
	}{
		{
			name: "not found",
			app: func() *runtime.App {
				return &runtime.App{Configs: map[string]*dynamic.Config{}}
			},
			requestUrl: "http://foo.api/api/configs/foo",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusNotFound),
			},
		},
		{
			name: "yaml file empty",
			app: func() *runtime.App {
				return &runtime.App{Configs: map[string]*dynamic.Config{
					"foo": {
						Info: dynamic.ConfigInfo{
							Url:  mustUrl("https://foo.bar/foo.yaml"),
							Time: mustTime("2023-12-27T13:01:30+00:00"),
						},
						Raw: nil,
					},
				}}
			},
			requestUrl: "http://foo.api/api/configs/foo",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasHeader("Last-Modified", "Wed, 27 Dec 2023 13:01:30 GMT"),
				try.HasHeader("Content-Type", "text/plain"),
				try.HasBody(""),
			},
		},
		{
			name: "json file with content",
			app: func() *runtime.App {
				return &runtime.App{Configs: map[string]*dynamic.Config{
					"foo": {
						Info: dynamic.ConfigInfo{
							Url:  mustUrl("https://foo.bar/foo.json"),
							Time: mustTime("2023-12-22T13:01:30+00:00"),
						},
						Raw: []byte(`{"foo": "bar"}`),
					},
				}}
			},
			requestUrl: "http://foo.api/api/configs/foo",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasHeader("Last-Modified", "Fri, 22 Dec 2023 13:01:30 GMT"),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(`{"foo": "bar"}`),
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app(), static.Api{})

			try.Handler(t,
				http.MethodGet,
				tc.requestUrl,
				nil,
				"",
				h,
				tc.test...)
		})
	}
}
