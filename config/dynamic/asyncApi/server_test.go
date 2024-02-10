package asyncApi

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetHost(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "host and port",
			test: func(t *testing.T) {
				server := Server{Url: "demo:9092"}
				require.Equal(t, "demo", server.GetHost())
			},
		},
		{
			name: "only port",
			test: func(t *testing.T) {
				server := Server{Url: ":9092"}
				require.Equal(t, "", server.GetHost())
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

func TestGetPort(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "host and port",
			test: func(t *testing.T) {
				server := Server{Url: "demo:9092"}
				require.Equal(t, 9092, server.GetPort())
			},
		},
		{
			name: "only port",
			test: func(t *testing.T) {
				server := Server{Url: ":9092"}
				require.Equal(t, 9092, server.GetPort())
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
