package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventRequest_String(t *testing.T) {
	r := &EventRequest{
		Method: "GET",
		Url: Url{
			Scheme: "https",
			Host:   "foo.bar",
			Path:   "/yuh",
			Query:  "test=123",
		},
	}
	require.Equal(t, "GET https://foo.bar/yuh?test=123", r.String())
}

func TestEventResponse_HasBody(t *testing.T) {
	r := &EventResponse{}
	require.False(t, r.HasBody())
	r.Body = "foo"
	require.True(t, r.HasBody())
	r.Body = ""
	r.Data = 123
	require.True(t, r.HasBody())
}

func TestHttpResource(t *testing.T) {
	testcases := []struct {
		url      Url
		resource any
		test     func(t *testing.T, b bool, body any, err error)
	}{
		{
			url: Url{
				Path: "/foo",
			},
			resource: map[string]any{
				"foo": "bar",
			},
			test: func(t *testing.T, b bool, body any, err error) {
				require.NoError(t, err)
				require.True(t, b)
				require.Equal(t, "bar", body)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.url.String(), func(t *testing.T) {
			t.Parallel()
			req := &EventRequest{
				Url: tc.url,
			}
			res := &EventResponse{}
			b, err := HttpEventHandler(req, res, tc.resource)
			tc.test(t, b, res.Data, err)
		})
	}
}
