package asyncApi

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *Config
		test func(t *testing.T, err error)
	}{
		{
			name: "not supported",
			cfg:  &Config{AsyncApi: "1.0"},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "not supported version: 1.0")
			},
		},
		{
			name: "2.0.0",
			cfg:  &Config{AsyncApi: "2.0.0"},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "2.1.0",
			cfg:  &Config{AsyncApi: "2.2.0"},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "2.2.0",
			cfg:  &Config{AsyncApi: "2.2.0"},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "2.3.0",
			cfg:  &Config{AsyncApi: "2.2.0"},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "2.4.0",
			cfg:  &Config{AsyncApi: "2.4.0"},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "2.5.0",
			cfg:  &Config{AsyncApi: "2.5.0"},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "2.6.0",
			cfg:  &Config{AsyncApi: "2.6.0"},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.cfg.Validate()
			tc.test(t, err)
		})
	}
}
