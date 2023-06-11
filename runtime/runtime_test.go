package runtime

import (
	"github.com/stretchr/testify/require"
	"mokapi/version"
	"testing"
)

func TestNew(t *testing.T) {
	version.BuildVersion = "1.0"
	defer func() {
		version.BuildVersion = ""
	}()
	app := New()
	require.NotNil(t, app.Monitor)
	require.Equal(t, "1.0", app.Version)
	require.Len(t, app.Http, 0)
	require.Len(t, app.Kafka, 0)
	require.Len(t, app.Ldap, 0)
	require.Len(t, app.Smtp, 0)
}
