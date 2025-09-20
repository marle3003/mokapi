package search_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/runtime/search"
	"testing"
)

func TestErrNotEnabled_Error(t *testing.T) {
	require.EqualError(t, &search.ErrNotEnabled{}, "search is not enabled")
}
