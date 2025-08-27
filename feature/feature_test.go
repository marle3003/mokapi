package feature_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/feature"
	"testing"
)

func TestFeature(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "no feature enabled",
			test: func(t *testing.T) {
				require.False(t, feature.IsEnabled("foo"), "feature is not enabled")
			},
		},
		{
			name: "feature enabled",
			test: func(t *testing.T) {
				feature.Enable([]string{"foo"})
				require.True(t, feature.IsEnabled("foo"), "feature is enabled")
			},
		},
		{
			name: "feature does not match",
			test: func(t *testing.T) {
				feature.Enable([]string{"foo"})
				require.False(t, feature.IsEnabled("bar"), "feature is not enabled")
			},
		},
		{
			name: "reset all features",
			test: func(t *testing.T) {
				feature.Enable([]string{"foo"})
				feature.Reset()
				require.False(t, feature.IsEnabled("foo"), "feature is not enabled")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
