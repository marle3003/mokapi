package directory_test

import (
	"io"
	"mokapi/config/static"
	"mokapi/providers/directory"
	"mokapi/server"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// A directory config includes LDIF files through the Reader. The ConfigWatcher uses the value
// given to Read only as a type hint, so the records must come from the config that Read returns.
func TestConfig_LdifIncludeReadByConfigWatcher(t *testing.T) {
	logrus.SetOutput(io.Discard)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "users.ldif"), []byte("dn: cn=user\ncn: user\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "directory.json"), []byte(`{"files": ["./users.ldif"]}`), 0o600))

	w := server.NewConfigWatcher(&static.Config{})
	c, err := w.Read(&url.URL{Scheme: "file", Path: filepath.ToSlash(filepath.Join(dir, "directory.json"))}, &directory.Config{})
	require.NoError(t, err)

	cfg := c.Data().(*directory.Config)
	entry, ok := cfg.Entries.Get("cn=user")
	require.True(t, ok, "record of the included LDIF file")
	require.Equal(t, []string{"user"}, entry.Attributes["cn"])
	require.Len(t, c.Refs.List(false), 1, "LDIF file is a reference of the directory")
}
