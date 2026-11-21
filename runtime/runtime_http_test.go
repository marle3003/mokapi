package runtime_test

import (
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	oaschema "mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	jsonSchema "mokapi/schema/json/schema"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestApp_AddHttp(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "event store available",
			test: func(t *testing.T, app *runtime.App) {
				app.Http.Add(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))))

				require.NotNil(t, app.Http.Get("foo"))
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("http").WithName("foo"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "event store for endpoint available",
			test: func(t *testing.T, app *runtime.App) {
				app.Http.Add(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("bar"))))

				require.NotNil(t, app.Http.Get("foo"))
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("http").WithName("foo").With("path", "bar"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "request is counted in monitor",
			test: func(t *testing.T, app *runtime.App) {
				info := app.Http.Add(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithResponse(http.StatusOK),
						))),
				))
				m := monitor.NewHttp()
				h := info.Handler(info.Config(), m, enginetest.NewEngine(), &events.StoreManager{})

				r := httptest.NewRequest(http.MethodGet, "https://mokapi.io/foo", nil)
				rr := httptest.NewRecorder()
				err := h.ServeHTTP(rr, r)
				require.Nil(t, err)

				require.Equal(t, float64(1), m.RequestCounter.Sum(metrics.NewQuery()))
			},
		},
		{
			name: "retrieve configs",
			test: func(t *testing.T, app *runtime.App) {
				info := app.Http.Add(newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("bar"))))

				configs := filesOf(info)
				require.Len(t, configs, 1)
				require.Equal(t, "https://mokapi.io", configs[0].Info.Url.String())
			},
		},
		{
			name: "webhook not specified",
			test: func(t *testing.T, app *runtime.App) {
				app.Http.Add(newConfig(
					openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "")),
				))

				w, err := app.Http.Webhook("foo", common.WebhookArgs{})
				require.EqualError(t, err, "webhook not found: foo")
				require.Nil(t, w)
			},
		},
		{
			name: "webhook is specified",
			test: func(t *testing.T, app *runtime.App) {
				app.Http.Add(newConfig(
					openapitest.NewConfig("3.0",
						openapitest.WithInfo("foo", "", ""),
						openapitest.WithWebhook("foo"),
					),
				))

				w, err := app.Http.Webhook("foo", common.WebhookArgs{})
				require.NoError(t, err)
				require.NotNil(t, w)
			},
		},
		{
			name: "webhook ambiguous",
			test: func(t *testing.T, app *runtime.App) {
				app.Http.Add(newConfig(
					openapitest.NewConfig("3.0",
						openapitest.WithInfo("foo", "", ""),
						openapitest.WithWebhook("foo"),
					),
				))
				app.Http.Add(newConfig(
					openapitest.NewConfig("3.0",
						openapitest.WithInfo("bar", "", ""),
						openapitest.WithWebhook("foo"),
					),
				))

				w, err := app.Http.Webhook("foo", common.WebhookArgs{})
				require.EqualError(t, err, "ambiguous webhook 'foo': use args.api to refine")
				require.Nil(t, w)
			},
		},
		{
			name: "api removed with its last file",
			test: func(t *testing.T, app *runtime.App) {
				file := newConfig(openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithWebhook("x"),
				))
				app.Http.Add(file)
				app.Http.Remove(file)

				w, err := app.Http.Webhook("x", common.WebhookArgs{Api: "foo"})
				require.EqualError(t, err, "api not found: foo")
				require.Nil(t, w)
			},
		},
		{
			name: "api without config snapshot is skipped",
			test: func(t *testing.T, app *runtime.App) {
				app.Http.Set("foo", runtime.NewHttpInfo(nil))

				var w *openapi.Path
				var err error
				require.NotPanics(t, func() { w, err = app.Http.Webhook("x", common.WebhookArgs{}) })
				require.EqualError(t, err, "webhook not found: x")
				require.Nil(t, w)
			},
		},
		{
			name: "api without config snapshot is not found",
			test: func(t *testing.T, app *runtime.App) {
				app.Http.Set("foo", runtime.NewHttpInfo(nil))

				w, err := app.Http.Webhook("x", common.WebhookArgs{Api: "foo"})
				require.EqualError(t, err, "api not found: foo")
				require.Nil(t, w)
			},
		},
		{
			name: "webhook with api",
			test: func(t *testing.T, app *runtime.App) {
				app.Http.Add(newConfig(
					openapitest.NewConfig("3.0",
						openapitest.WithInfo("foo", "", ""),
						openapitest.WithWebhook("foo", openapitest.WithOperation(http.MethodGet)),
					),
				))
				app.Http.Add(newConfig(
					openapitest.NewConfig("3.0",
						openapitest.WithInfo("bar", "", ""),
						openapitest.WithWebhook("foo"),
					),
				))

				w, err := app.Http.Webhook("foo", common.WebhookArgs{Api: "foo"})
				require.NoError(t, err)
				require.NotNil(t, w)
				require.NotNil(t, w.Get)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg, &dynamictest.Reader{})
			tc.test(t, app)
		})
	}
}

func TestApp_AddHttp_Patching(t *testing.T) {
	newConfig := func(name string, c *openapi.Config) *dynamic.Config {
		cfg := &dynamic.Config{Data: c}
		u, _ := url.Parse(name)
		cfg.Info.Url = u
		return cfg
	}

	testcases := []struct {
		name    string
		configs []*dynamic.Config
		static  *static.Config
		test    func(t *testing.T, app *runtime.App)
	}{
		{
			name: "overwrite value",
			configs: []*dynamic.Config{
				newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
				newConfig("https://mokapi.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Http.Get("foo")
				require.Equal(t, "bar", info.Config().Info.Description)
				configs := filesOf(info)
				require.Len(t, configs, 2)
			},
		},
		{
			name: "a is patched with b",
			configs: []*dynamic.Config{
				newConfig("https://mokapi.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
				newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Http.Get("foo")
				require.Equal(t, "foo", info.Config().Info.Description)
			},
		},
		{
			name: "order only by filename",
			configs: []*dynamic.Config{
				newConfig("https://a.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
				newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Http.Get("foo")
				require.Equal(t, "foo", info.Config().Info.Description)
			},
		},
		{
			name: "patch does not reset events",
			configs: []*dynamic.Config{
				newConfig("https://a.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
			},
			test: func(t *testing.T, app *runtime.App) {
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("http").WithName("foo"))
				require.NoError(t, err)

				app.Http.Add(
					newConfig("https://mokapi.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "bar"))),
				)

				e := app.Events.GetEvents(events.NewTraits().WithNamespace("http").WithName("foo"))
				require.Len(t, e, 1)
			},
		},
		{
			name: "patch does not reset events",
			configs: []*dynamic.Config{
				newConfig("https://a.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "foo"))),
			},
			static: &static.Config{Event: static.Event{Store: map[string]static.Store{
				"foo": {Size: 104},
			}}},
			test: func(t *testing.T, app *runtime.App) {
				for i := 0; i < 105; i++ {
					err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("http").WithName("foo"))
					require.NoError(t, err)
				}

				e := app.Events.GetEvents(events.NewTraits().WithNamespace("http").WithName("foo"))
				require.Len(t, e, 104)
			},
		},
		{
			name: "patch schema in response using $ref",
			configs: []*dynamic.Config{
				newConfig("https://a.io/a", openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", "foo"),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithContent("application/json",
									openapitest.WithSchema(
										schematest.NewTypes(nil, schematest.WithRef("#/components/schemas/Foo")),
									),
								),
							),
						),
					),
					openapitest.WithComponentSchema("Foo", schematest.New("string")),
				),
				),
				newConfig("https://a.io/b", openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", "foo"),
					openapitest.WithComponentSchema("Foo", schematest.New("string", schematest.WithFormat("date-time"))),
				),
				),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Http.Get("foo")
				p := info.Config().Paths.Lookup("/foo")
				require.NotNil(t, p)
				op := p.Value.Get
				require.NotNil(t, op)
				res := op.Responses.GetResponse(http.StatusOK)
				require.NotNil(t, res)
				mt := res.Content["application/json"]
				require.NotNil(t, mt)
				require.Equal(t, "date-time", mt.Schema.Format)
			},
		},
		{
			name: "add webhook",
			configs: []*dynamic.Config{
				newConfig("https://a.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithWebhook("foo"),
				)),
				newConfig("https://a.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithWebhook("bar"),
				)),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Http.Get("foo")
				require.Len(t, info.Config().Webhooks, 2)
			},
		},
		{
			name: "update webhook",
			configs: []*dynamic.Config{
				newConfig("https://a.io/a", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithWebhook("foo"),
				)),
				newConfig("https://a.io/b", openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
					openapitest.WithWebhook("foo", openapitest.WithOperation(http.MethodGet)),
				)),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Http.Get("foo")
				require.Len(t, info.Config().Webhooks, 1)
				require.NotNil(t, info.Config().Webhooks["foo"].Value.Get)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.static
			if cfg == nil {
				cfg = &static.Config{}
			}
			app := runtime.New(cfg, &dynamictest.Reader{})
			for _, c := range tc.configs {
				app.Http.Add(c)
			}
			tc.test(t, app)
		})
	}
}

// Two files with the same title are merged. Deleting one of them must remove
// only its contribution and must not change the parsed config of the other.
func TestApp_RemoveHttp_DeleteOneOfTwoFilesWithSameTitle(t *testing.T) {
	newA := func() *openapi.Config {
		return openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
			openapitest.WithPath("/shared",
				openapitest.WithOperation(http.MethodGet,
					openapitest.WithOperationSummary("from a"),
					openapitest.WithResponse(http.StatusOK),
				),
			),
			openapitest.WithServer("http://127.0.0.1:8001", ""),
		)
	}
	newB := func() *openapi.Config {
		return openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
			openapitest.WithPath("/shared",
				openapitest.WithOperation(http.MethodGet,
					openapitest.WithOperationSummary("from b"),
					openapitest.WithResponse(http.StatusOK,
						openapitest.WithResponseHeader("X-From-B", "", schematest.New("string")),
					),
				),
			),
			openapitest.WithPath("/fromB", openapitest.WithOperation(http.MethodGet)),
			openapitest.WithServer("http://127.0.0.1:8002", ""),
			openapitest.WithComponentSchema("FromB", schematest.New("string")),
		)
	}
	file := func(name string, c *openapi.Config) *dynamic.Config {
		cfg := &dynamic.Config{Data: c}
		cfg.Info.Url, _ = url.Parse("file:///" + name)
		return cfg
	}

	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	a := file("a.yml", newA())
	b := file("b.yml", newB())
	app.Http.Add(a)
	app.Http.Add(b)

	app.Http.Remove(b)

	info := app.Http.Get("foo")
	require.NotNil(t, info)
	configs := filesOf(info)
	require.Len(t, configs, 1)
	require.Equal(t, "file:///a.yml", configs[0].Info.Url.String())

	require.Nil(t, info.Config().Paths.Lookup("/fromB"), "path of deleted file must not be in merged config")
	require.Len(t, info.Config().Servers, 1)
	require.Equal(t, "http://127.0.0.1:8001", info.Config().Servers[0].Url)
	require.Nil(t, info.Config().Components.Schemas.Get("FromB"))
	op := info.Config().Paths.Lookup("/shared").Value.Get
	require.Equal(t, "from a", op.Summary)
	require.NotContains(t, op.Responses.GetResponse(http.StatusOK).Headers, "X-From-B")

	require.Equal(t, newA(), a.Data, "parsed config of a.yml must not be changed by merging")
}

func TestApp_AddHttp_SingleFile(t *testing.T) {
	newFoo := func() *openapi.Config {
		return openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
			openapitest.WithPath("/foo",
				openapitest.WithOperation(http.MethodGet, openapitest.WithResponse(http.StatusOK)),
				openapitest.WithOperation(http.MethodPost, openapitest.WithResponse(http.StatusOK)),
			),
		)
	}
	parsed := newFoo()
	// Parse sets Path and Method of the operations
	require.NoError(t, parsed.Parse(&dynamic.Config{Data: parsed}, &dynamictest.Reader{}))
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})

	info := app.Http.Add(newConfig(parsed))

	t.Run("default server is added to the merged config only", func(t *testing.T) {
		require.Len(t, info.Config().Servers, 1)
		require.Equal(t, "/", info.Config().Servers[0].Url)
		require.True(t, info.Config().Servers[0].IsDefault)
		require.Empty(t, parsed.Servers)
	})

	t.Run("input config is not changed", func(t *testing.T) {
		expected := newFoo()
		require.NoError(t, expected.Parse(&dynamic.Config{Data: expected}, &dynamictest.Reader{}))
		require.Equal(t, expected, parsed)
	})

	t.Run("operations point into the merged config", func(t *testing.T) {
		path := info.Config().Paths.Lookup("/foo").Value
		require.NotSame(t, parsed.Paths.Lookup("/foo").Value, path)
		require.Equal(t, "/foo", path.Path)
		for method, op := range path.Operations() {
			require.Same(t, path, op.Path)
			require.Equal(t, method, op.Method)
		}
	})
}

func TestApp_RemoveHttp_StoreAfterRemove(t *testing.T) {
	newFile := func(name string, c *openapi.Config) *dynamic.Config {
		cfg := &dynamic.Config{Data: c}
		cfg.Info.Url, _ = url.Parse("file:///" + name)
		return cfg
	}
	newA := func() *openapi.Config {
		return openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "from a"),
			openapitest.WithPath("/foo",
				openapitest.WithOperation(http.MethodGet,
					openapitest.WithOperationSummary("from a"),
					openapitest.WithResponse(http.StatusOK, openapitest.WithResponseDescription("from a")),
				),
			),
		)
	}
	newB := func() *openapi.Config {
		return openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "from b"),
			openapitest.WithPath("/foo",
				openapitest.WithOperation(http.MethodGet,
					openapitest.WithOperationSummary("from b"),
					openapitest.WithResponse(http.StatusOK, openapitest.WithResponseDescription("from b")),
				),
			),
		)
	}

	t.Run("two files to one is the remaining file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := newFile("a.yml", newA())
		b := newFile("b.yml", newB())
		app.Http.Add(a)
		app.Http.Add(b)
		require.Equal(t, "from b", app.Http.Get("foo").Config().Paths.Lookup("/foo").Value.Get.Summary)

		app.Http.Remove(b)

		expected := runtime.New(&static.Config{}, &dynamictest.Reader{})
		expected.Http.Add(newFile("a.yml", newA()))
		require.Equal(t, expected.Http.Get("foo").Config(), app.Http.Get("foo").Config())
		op := app.Http.Get("foo").Config().Paths.Lookup("/foo").Value.Get
		require.Equal(t, "from a", op.Summary)
		require.Equal(t, "from a", op.Responses.GetResponse(http.StatusOK).Description)
		require.Equal(t, "from a", app.Http.Get("foo").Config().Info.Description)
	})

	t.Run("one file to zero removes the api", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := newFile("a.yml", newA())
		app.Http.Add(a)

		app.Http.Remove(a)

		require.Nil(t, app.Http.Get("foo"))
		require.Equal(t, 0, app.Http.Len())
	})

	t.Run("files are applied ordered by file name", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		c := newFile("c.yml", newB())
		app.Http.Add(c)
		app.Http.Add(newFile("b.yml", newB()))
		app.Http.Add(newFile("a.yml", newA()))
		require.Equal(t, "from b", app.Http.Get("foo").Config().Paths.Lookup("/foo").Value.Get.Summary)

		app.Http.Remove(c)

		require.Equal(t, "from b", app.Http.Get("foo").Config().Paths.Lookup("/foo").Value.Get.Summary)
	})
}

func TestIsHttpConfig(t *testing.T) {
	_, b := runtime.IsHttpConfig(&dynamic.Config{Data: openapitest.NewConfig("3.0")})
	require.True(t, b)
	_, b = runtime.IsHttpConfig(&dynamic.Config{Data: "foo"})
	require.False(t, b)
}

func newConfig(config *openapi.Config) *dynamic.Config {
	c := &dynamic.Config{Data: config}
	u, _ := url.Parse("https://mokapi.io")
	c.Info.Url = u
	return c
}

// Two files with the same title define the same recursive schema (Node.properties.child
// refers to Node). Merging them must not overflow the stack.
func TestApp_AddHttp_SameCyclicSchemaInTwoFiles(t *testing.T) {
	newNode := func() *oaschema.Schema {
		props := &oaschema.Schemas{}
		props.Set("name", schematest.New("string"))
		node := schematest.New("object", schematest.WithTitle("Node"))
		node.Properties = props
		props.Set("child", &oaschema.Schema{Type: jsonSchema.Types{"object"}, Properties: props})
		return node
	}
	file := func(name string) *dynamic.Config {
		c := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
			openapitest.WithComponentSchema("Node", newNode()))
		cfg := &dynamic.Config{Data: c}
		cfg.Info.Url, _ = url.Parse("file:///" + name)
		return cfg
	}

	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	app.Http.Add(file("a.yml"))
	app.Http.Add(file("b.yml"))

	info := app.Http.Get("foo")
	require.NotNil(t, info)
	require.NotNil(t, info.Config().Components.Schemas.Get("Node"))
}

func TestHttpStore_Remove(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newFile := func(fileUrl, path string) *dynamic.Config {
		c := newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
			openapitest.WithPath(path, openapitest.WithOperation(http.MethodGet)),
		))
		c.Info.Url, _ = url.Parse(fileUrl)
		return c
	}

	t.Run("unknown config does not panic and keeps the store usable", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		unknown := newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("unknown", "", "")))

		requireStoreUsableAfterRemoveOfUnknownConfig(t,
			func() { app.Http.Remove(unknown) },
			func() { app.Http.Add(newFile("https://mokapi.io/a.yml", "/a")) },
		)
	})

	t.Run("two files removed at the same time", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := newFile("https://mokapi.io/a.yml", "/a")
		b := newFile("https://mokapi.io/b.yml", "/b")
		app.Http.Add(a)
		app.Http.Add(b)

		removeConcurrently(app.Http.Remove, a, b)

		require.Equal(t, 0, app.Http.Len())
	})
}

func TestHttpStore_ConcurrentAccess(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newFile := func(fileUrl, path string) *dynamic.Config {
		c := newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
			openapitest.WithPath(path, openapitest.WithOperation(http.MethodGet)),
		))
		c.Info.Url, _ = url.Parse(fileUrl)
		return c
	}

	t.Run("reader of the info does not race with reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		app.Http.Add(newFile("https://mokapi.io/a.yml", "/first"))

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				app.Http.Add(newFile("https://mokapi.io/a.yml", "/second"))
			}
		}()
		go func() {
			defer wg.Done()
			// what the dashboard API handler does: get the info, then read the config
			for i := 0; i < 200; i++ {
				info := app.Http.Get("foo")
				_ = info.Config().Info.Name
				_ = info.Config().Paths.Len()
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})

	t.Run("reader does not race with remove", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := newFile("https://mokapi.io/a.yml", "/a")
		b := newFile("https://mokapi.io/b.yml", "/b")
		info := app.Http.Add(a)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				app.Http.Add(b)
				app.Http.Remove(b)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if info.Config() != nil {
					_ = info.Config().Info.Name
				}
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})

	t.Run("Webhook does not race with Add and Remove", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		app.Http.Add(newFile("https://mokapi.io/a.yml", "/a"))
		b := newFile("https://mokapi.io/b.yml", "/b")

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				app.Http.Add(b)
				app.Http.Remove(b)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_, _ = app.Http.Webhook("unknown", common.WebhookArgs{})
				_, _ = app.Http.Webhook("unknown", common.WebhookArgs{Api: "foo"})
			}
		}()
		wg.Wait()
	})

	t.Run("Add racing with Remove of the same file does not register an orphaned info", func(t *testing.T) {
		// the red proof of this test was done with 100000 iterations
		for i := 0; i < 5000; i++ {
			app := runtime.New(&static.Config{}, &dynamictest.Reader{})
			a := newFile("https://mokapi.io/a.yml", "/a")
			app.Http.Add(a)

			var added *runtime.HttpInfo
			var wg sync.WaitGroup
			wg.Add(2)
			go func() { defer wg.Done(); added = app.Http.Add(a) }()
			go func() { defer wg.Done(); app.Http.Remove(a) }()
			wg.Wait()

			if added.Config() != nil {
				require.Same(t, added, app.Http.Get("foo"), "file was added but the info is not registered (iteration %d)", i)
			}
		}
	})
}

func TestHttpStore_RemoveLastFileWithSearchEnabled(t *testing.T) {
	logrus.SetOutput(io.Discard)

	cfg := &static.Config{}
	cfg.Api.Search.Enabled = true
	app := runtime.New(cfg, &dynamictest.Reader{})
	file := newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
		openapitest.WithPath("/a", openapitest.WithOperation(http.MethodGet))))
	app.Http.Add(file)

	require.NotPanics(t, func() { app.Http.Remove(file) })
	require.Nil(t, app.Http.Get("foo"))
}

func TestHttpInfo_Snapshot(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newFile := func(fileUrl, path string) *dynamic.Config {
		c := newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""),
			openapitest.WithPath(path, openapitest.WithOperation(http.MethodGet)),
		))
		c.Info.Url, _ = url.Parse(fileUrl)
		return c
	}

	t.Run("snapshot is not changed by a reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		info := app.Http.Add(newFile("https://mokapi.io/a.yml", "/first"))
		snapshot := info.Config()

		app.Http.Add(newFile("https://mokapi.io/a.yml", "/second"))

		require.NotNil(t, snapshot.Paths.Lookup("/first"))
		require.Nil(t, snapshot.Paths.Lookup("/second"))
		require.NotSame(t, snapshot, info.Config())
		require.NotNil(t, info.Config().Paths.Lookup("/second"))
	})

	t.Run("published config is not the config of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := newFile("https://mokapi.io/a.yml", "/a")

		info := app.Http.Add(file)

		require.NotSame(t, file.Data, info.Config())
		require.Len(t, info.Config().Servers, 1, "default server is added to the published config")
		require.Equal(t, "/", info.Config().Servers[0].Url)
		require.Empty(t, file.Data.(*openapi.Config).Servers, "but not to the file")
	})

	t.Run("config is nil after the last file is removed", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := newFile("https://mokapi.io/a.yml", "/a")
		info := app.Http.Add(file)

		app.Http.Remove(file)

		require.Nil(t, info.Config())
		require.Nil(t, app.Http.Get("foo"))
	})

	t.Run("Configs returns a copy", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		app.Http.Add(newFile("https://mokapi.io/a.yml", "/a"))
		info := app.Http.Add(newFile("https://mokapi.io/b.yml", "/b"))

		configs := filesOf(info)
		require.Len(t, configs, 2)
		configs[0] = nil

		require.NotContains(t, filesOf(info), (*dynamic.Config)(nil))
		require.Len(t, filesOf(info), 2)
	})

	t.Run("Configs does not race with Add", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		info := app.Http.Add(newFile("https://mokapi.io/a.yml", "/a"))
		b := newFile("https://mokapi.io/b.yml", "/b")

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				app.Http.Add(b)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})
}

func TestHttpStore_RemoveInfoWithoutConfigWithSearchEnabled(t *testing.T) {
	logrus.SetOutput(io.Discard)

	cfg := &static.Config{}
	cfg.Api.Search.Enabled = true
	app := runtime.New(cfg, &dynamictest.Reader{})
	app.Http.Set("foo", runtime.NewHttpInfo(nil))
	file := newConfig(openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "")))

	require.NotPanics(t, func() { app.Http.Remove(file) })
	require.Nil(t, app.Http.Get("foo"))
}

func TestHttpStore_Set(t *testing.T) {
	t.Run("Set registers an info that Get and List return", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
		info := runtime.NewHttpInfo(cfg)

		app.Http.Set("foo", info)

		require.Same(t, info, app.Http.Get("foo"))
		require.Same(t, cfg, info.Config())
		require.Len(t, app.Http.List(), 1)
	})

	t.Run("info without config has no snapshot", func(t *testing.T) {
		require.Nil(t, runtime.NewHttpInfo(nil).Config())
	})
}
