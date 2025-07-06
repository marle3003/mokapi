package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/version"
	"testing"
)

func TestNew(t *testing.T) {
	version.BuildVersion = "1.0"
	defer func() {
		version.BuildVersion = ""
	}()
	app := runtime.New(&static.Config{})
	require.NotNil(t, app.Monitor)
	require.Equal(t, "1.0", app.Version)
	require.Len(t, app.ListHttp(), 0)
	require.Len(t, app.Kafka.List(), 0)
	require.Len(t, app.Ldap.List(), 0)
	require.Len(t, app.Mail.List(), 0)
}
