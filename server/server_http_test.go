package server_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/server"
	"mokapi/server/cert"
	"mokapi/try"
	"net/http"
	"testing"
	"time"
)

func TestHttp(t *testing.T) {
	port := server.DefaultHttpPort
	defer func() { server.DefaultHttpPort = port }()
	server.DefaultHttpPort = try.GetFreePort()
	waitStartup := func() {
		time.Sleep(100 * time.Millisecond)
	}

	testcases := []struct {
		name string
		test func(t *testing.T, m *server.HttpManager)
	}{
		{
			name: "new service but no paths defined",
			test: func(t *testing.T, m *server.HttpManager) {
				m.Update(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: openapitest.NewConfig("3.1.0"),
					},
				})
				waitStartup()

				c := http.Client{}
				r, err := c.Get(fmt.Sprintf("http://127.0.0.1:%v", server.DefaultHttpPort))
				require.NoError(t, err)
				require.Equal(t, http.StatusNotFound, r.StatusCode)
			},
		},
		{
			name: "new service with path",
			test: func(t *testing.T, m *server.HttpManager) {
				m.Update(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: openapitest.NewConfig("3.1.0",
							openapitest.WithPath("/", openapitest.NewPath(
								openapitest.WithOperation("GET", openapitest.NewOperation(
									openapitest.WithResponse(200))),
							)),
						),
					},
				})
				waitStartup()

				c := http.Client{}
				r, err := c.Get(fmt.Sprintf("http://127.0.0.1:%v", server.DefaultHttpPort))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, r.StatusCode)
			},
		},
		{
			name: "add server url should remove default url",
			test: func(t *testing.T, m *server.HttpManager) {
				m.Update(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: openapitest.NewConfig("3.1.0",
							openapitest.WithPath("/", openapitest.NewPath(
								openapitest.WithOperation("GET", openapitest.NewOperation(
									openapitest.WithResponse(200))),
							)),
						),
					},
				})
				m.Update(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						// Info must be different to other config event
						// otherwise no paths would be defined
						Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")},
						Data: openapitest.NewConfig("3.1.0",
							openapitest.WithServer("/foo", ""),
						),
					},
				})
				waitStartup()

				c := http.Client{}
				// default should be removed => 404
				r, err := c.Get(fmt.Sprintf("http://127.0.0.1:%v", server.DefaultHttpPort))
				require.NoError(t, err)
				require.Equal(t, http.StatusNotFound, r.StatusCode)

				r, err = c.Get(fmt.Sprintf("http://127.0.0.1:%v/foo", server.DefaultHttpPort))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, r.StatusCode)
			},
		},
		{
			name: "update server url",
			test: func(t *testing.T, m *server.HttpManager) {
				m.Update(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: openapitest.NewConfig("3.1.0",
							openapitest.WithPath("/", openapitest.NewPath(
								openapitest.WithOperation("GET", openapitest.NewOperation(
									openapitest.WithResponse(200))),
							)),
							openapitest.WithServer("/", ""),
						),
					},
				})
				m.Update(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: openapitest.NewConfig("3.1.0",
							openapitest.WithPath("/", openapitest.NewPath(
								openapitest.WithOperation("GET", openapitest.NewOperation(
									openapitest.WithResponse(200))),
							)),
							openapitest.WithServer("/foo", ""),
						),
					},
				})
				waitStartup()

				c := http.Client{}
				// default should be removed => 404
				r, err := c.Get(fmt.Sprintf("http://127.0.0.1:%v", server.DefaultHttpPort))
				require.NoError(t, err)
				require.Equal(t, http.StatusNotFound, r.StatusCode)

				r, err = c.Get(fmt.Sprintf("http://127.0.0.1:%v/foo", server.DefaultHttpPort))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, r.StatusCode)
			},
		},
		{
			name: "update server url on different port",
			test: func(t *testing.T, m *server.HttpManager) {
				m.Update(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: openapitest.NewConfig("3.1.0",
							openapitest.WithPath("/", openapitest.NewPath(
								openapitest.WithOperation("GET", openapitest.NewOperation(
									openapitest.WithResponse(200))),
							)),
							openapitest.WithServer("/foo", ""),
						),
					},
				})
				port := try.GetFreePort()
				m.Update(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: openapitest.NewConfig("3.1.0",
							openapitest.WithPath("/", openapitest.NewPath(
								openapitest.WithOperation("GET", openapitest.NewOperation(
									openapitest.WithResponse(200))),
							)),
							openapitest.WithServer(fmt.Sprintf("http://:%v/foo", port), ""),
						),
					},
				})
				waitStartup()

				c := http.Client{}
				r, err := c.Get(fmt.Sprintf("http://127.0.0.1:%v/foo", port))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, r.StatusCode)
			},
		},
		{
			name: "delete config event",
			test: func(t *testing.T, m *server.HttpManager) {
				m.Update(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: openapitest.NewConfig("3.1.0",
							openapitest.WithPath("/", openapitest.NewPath(
								openapitest.WithOperation("GET", openapitest.NewOperation(
									openapitest.WithResponse(200))),
							)),
							openapitest.WithServer("/foo", ""),
						),
					},
				})
				m.Update(dynamic.ConfigEvent{
					Event: dynamic.Delete,
					Config: &dynamic.Config{
						Info: dynamictest.NewConfigInfo(),
						Data: openapitest.NewConfig("3.1.0",
							openapitest.WithPath("/", openapitest.NewPath(
								openapitest.WithOperation("GET", openapitest.NewOperation(
									openapitest.WithResponse(200))),
							)),
							openapitest.WithServer("/foo", ""),
						),
					},
				})
				waitStartup()

				c := http.Client{}
				_, err := c.Get(fmt.Sprintf("http://127.0.0.1:%v/foo", server.DefaultHttpPort))
				require.EqualError(t, err, fmt.Sprintf(`Get "http://127.0.0.1:%v/foo": dial tcp 127.0.0.1:%v: connectex: No connection could be made because the target machine actively refused it.`, server.DefaultHttpPort, server.DefaultHttpPort))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			certStore, err := cert.NewStore(&static.Config{})
			require.NoError(t, err)

			m := server.NewHttpManager(enginetest.NewEngine(), certStore, runtime.New())
			defer m.Stop()

			tc.test(t, m)
		})
	}
}
