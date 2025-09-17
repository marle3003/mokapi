package common

import (
	"github.com/stretchr/testify/require"
	"testing"
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
