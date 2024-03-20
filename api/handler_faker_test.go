package api

import (
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Faker(t *testing.T) {
	testcases := []struct {
		name       string
		app        func() *runtime.App
		requestUrl string
		test       []try.ResponseCondition
	}{
		{
			name: "get tree",
			app: func() *runtime.App {
				return &runtime.App{}
			},
			requestUrl: "http://foo.api/api/faker/tree",
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.BodyContains(`{"name":"Faker","custom":false,"nodes":[{"name":"Generic"`),
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
