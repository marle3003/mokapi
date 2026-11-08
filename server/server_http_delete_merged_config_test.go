package server_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/server"
	"mokapi/server/cert"
	"mokapi/try"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Two files with the same OpenAPI title are merged into one config.
// Deleting one of them must remove only its contribution.
func TestHttpManager_DeleteOneOfTwoFilesWithSameTitle(t *testing.T) {
	logrus.SetOutput(io.Discard)

	portA, portB := try.GetFreePort(), try.GetFreePort()
	newFile := func(file string, port int, path string) dynamic.ConfigEvent {
		cfg := openapitest.NewConfig("3.1.0",
			openapitest.WithInfo("foo", "", ""),
			openapitest.WithServer(fmt.Sprintf("http://127.0.0.1:%d", port), ""),
			openapitest.WithPath(path, openapitest.WithOperation("GET", openapitest.WithResponse(200))),
		)
		return dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: try.MustUrl(file)}, Data: cfg}}
	}
	a := newFile("a.yml", portA, "/fromA")
	b := newFile("b.yml", portB, "/fromB")

	certStore, err := cert.NewStore(&static.Config{})
	require.NoError(t, err)
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	m := server.NewHttpManager(enginetest.NewEngine(), certStore, app)
	defer m.Stop()

	m.Update(a)
	m.Update(b)
	requirePortOpen(t, portB)

	b.Event = dynamic.Delete
	m.Update(b)

	info := app.Http.Get("foo")
	require.NotNil(t, info)
	require.Len(t, info.Configs(), 1, "only file a.yml must remain in the store")
	require.Equal(t, "a.yml", info.Configs()[0].Info.Url.String())

	requirePortClosed(t, portB)

	c := http.Client{}
	r, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/fromA", portA))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, r.StatusCode)
	r, err = c.Get(fmt.Sprintf("http://127.0.0.1:%d/fromB", portA))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, r.StatusCode)
	requireMergedConfigIsFileA(t, info.Config)
}

func requireMergedConfigIsFileA(t *testing.T, merged *openapi.Config) {
	t.Helper()
	require.NotNil(t, merged)
	require.Nil(t, merged.Paths.Lookup("/fromB"), "path of deleted file must not be in merged config")
	require.NotNil(t, merged.Paths.Lookup("/fromA"))
	require.Len(t, merged.Servers, 1)
}
