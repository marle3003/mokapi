package mokapi_test

import (
	"mokapi/pkg/cmd/mokapi"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenCliDoc(t *testing.T) {
	dir := t.TempDir()
	cmd := mokapi.NewGenCliDocCmd()
	cmd.SetArgs([]string{"-o", dir})

	require.NoError(t, cmd.Execute())

	_, err := os.Stat(filepath.Join(dir, "mokapi.md"))
	require.NoError(t, err)
}
