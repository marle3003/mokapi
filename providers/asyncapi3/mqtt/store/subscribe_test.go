package store_test

import (
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"testing"
)

func TestSubscribe(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store)
	}{
		{},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine())
			tc.test(t, s)
		})
	}
}
