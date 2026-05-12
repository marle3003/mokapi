package runtime_test

import (
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/version"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	version.BuildVersion = "1.0"
	defer func() {
		version.BuildVersion = ""
	}()
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	require.NotNil(t, app.Monitor)
	require.Equal(t, "1.0", app.Version)
	require.Equal(t, 0, app.Http.Len())
	require.Equal(t, 0, app.Kafka.Len())
	require.Equal(t, 0, app.Ldap.Len())
	require.Equal(t, 0, app.Mail.Len())
}
