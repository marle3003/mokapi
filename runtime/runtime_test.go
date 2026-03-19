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
	require.Len(t, app.ListHttp(), 0)
	require.Len(t, app.Kafka.List(), 0)
	require.Len(t, app.Ldap.List(), 0)
	require.Len(t, app.Mail.List(), 0)
}
