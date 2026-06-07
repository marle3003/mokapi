package events_test

import (
	"mokapi/runtime/events"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTraits(t *testing.T) {
	testcases := []struct {
		name   string
		traits events.Traits
		test   func(*testing.T, events.Traits)
	}{
		{
			name:   "empty",
			traits: events.Traits{},
			test: func(t *testing.T, traits events.Traits) {
				require.True(t, traits.IsEmpty())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.test(t, tc.traits)
		})
	}
}
