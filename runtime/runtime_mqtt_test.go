package runtime_test

import (
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestMqttStore_MergedConfigs(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "removing one of two files leaves exactly the config of the other file",
			test: func(t *testing.T, app *runtime.App) {
				a, b := twoFilesWithSameTitle("mqtt")
				_, err := app.Mqtt.Add(a, enginetest.NewEngine())
				require.NoError(t, err)
				info, err := app.Mqtt.Add(b, enginetest.NewEngine())
				require.NoError(t, err)
				requireMergedWithB(t, info.Config())

				app.Mqtt.Remove(b)

				info = app.Mqtt.Get("foo")
				require.Len(t, info.Configs(), 1)
				requireOnlyFileA(t, "mqtt", info.Config())
			},
		},
		{
			name: "merging does not change the files",
			test: func(t *testing.T, app *runtime.App) {
				a, b := twoFilesWithSameTitle("mqtt")
				beforeA, beforeB := marshal(t, a.Data), marshal(t, b.Data)

				_, err := app.Mqtt.Add(a, enginetest.NewEngine())
				require.NoError(t, err)
				_, err = app.Mqtt.Add(b, enginetest.NewEngine())
				require.NoError(t, err)

				require.Equal(t, beforeA, marshal(t, a.Data))
				require.Equal(t, beforeB, marshal(t, b.Data))
			},
		},
		{
			name: "the default broker is not added to the file",
			test: func(t *testing.T, app *runtime.App) {
				c := asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo"), asyncapi3test.WithChannel("bar"))

				info, err := app.Mqtt.Add(getConfig("https://mokapi.io/a.yml", c), enginetest.NewEngine())

				require.NoError(t, err)
				require.NotNil(t, info.Config().Servers.Lookup("mokapi"))
				require.Equal(t, 0, c.Servers.Len())
			},
		},
		{
			name: "merging AsyncAPI 2.x files does not change them",
			test: func(t *testing.T, app *runtime.App) {
				a, b := twoAsyncApi2FilesWithSameTitle(t)

				_, err := app.Mqtt.Add(a, enginetest.NewEngine())
				require.NoError(t, err)
				info, err := app.Mqtt.Add(b, enginetest.NewEngine())
				require.NoError(t, err)

				requireExtraPropertyOnlyIn(t, info.Config(), a, b)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, runtime.New(&static.Config{}, &dynamictest.Reader{}))
		})
	}
}

func TestMqttStore_Remove(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Run("unknown config does not panic and keeps the store usable", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		unknown := getConfig("https://mokapi.io/unknown.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("unknown")))
		known := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))

		requireStoreUsableAfterRemoveOfUnknownConfig(t,
			func() { app.Mqtt.Remove(unknown) },
			func() { _, _ = app.Mqtt.Add(known, enginetest.NewEngine()) },
		)
	})

	t.Run("config that cannot be converted does not keep the lock", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		known := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo")))

		requireStoreUsableAfterRemoveOfUnknownConfig(t,
			func() { app.Mqtt.Remove(asyncApi2FileThatCannotBeConverted("foo")) },
			func() { _, _ = app.Mqtt.Add(known, enginetest.NewEngine()) },
		)
	})

	t.Run("two files removed at the same time", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a, b := twoFilesWithSameTitle("mqtt")
		_, err := app.Mqtt.Add(a, enginetest.NewEngine())
		require.NoError(t, err)
		_, err = app.Mqtt.Add(b, enginetest.NewEngine())
		require.NoError(t, err)

		removeConcurrently(app.Mqtt.Remove, a, b)

		require.Equal(t, 0, app.Mqtt.Len())
	})
}

func TestMqttStore_ConcurrentAccess(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Run("reader of the info does not race with reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		reload := func(channel string) *dynamic.Config {
			return getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(
				asyncapi3test.WithTitle("foo"),
				asyncapi3test.WithServer("a", "mqtt", ":1111"),
				asyncapi3test.WithChannel(channel),
			))
		}
		_, err := app.Mqtt.Add(reload("first"), enginetest.NewEngine())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if _, err := app.Mqtt.Add(reload("second"), enginetest.NewEngine()); err != nil {
					t.Error(err)
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			// what the dashboard API handler does: get the info, then read the config
			for i := 0; i < 200; i++ {
				info := app.Mqtt.Get("foo")
				cfg := info.Config()
				_ = cfg.Info.Name
				for range cfg.Channels {
				}
				_ = info.Configs()
			}
		}()
		wg.Wait()
	})

	t.Run("reader does not race with remove", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a, b := twoFilesWithSameTitle("mqtt")
		info, err := app.Mqtt.Add(a, enginetest.NewEngine())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if _, err := app.Mqtt.Add(b, enginetest.NewEngine()); err != nil {
					t.Error(err)
					return
				}
				app.Mqtt.Remove(b)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if cfg := info.Config(); cfg != nil {
					_ = cfg.Info.Name
				}
				_ = info.Configs()
			}
		}()
		wg.Wait()
	})
}

func TestMqttInfo_Snapshot(t *testing.T) {
	logrus.SetOutput(io.Discard)

	fileWithChannel := func(channel string) *dynamic.Config {
		return getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(
			asyncapi3test.WithTitle("foo"),
			asyncapi3test.WithServer("a", "mqtt", ":1111"),
			asyncapi3test.WithChannel(channel),
		))
	}

	t.Run("snapshot is not changed by a reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		info, err := app.Mqtt.Add(fileWithChannel("first"), enginetest.NewEngine())
		require.NoError(t, err)
		snapshot := info.Config()

		_, err = app.Mqtt.Add(fileWithChannel("second"), enginetest.NewEngine())
		require.NoError(t, err)

		require.Contains(t, snapshot.Channels, "first")
		require.NotContains(t, snapshot.Channels, "second")
		require.NotSame(t, snapshot, info.Config())
		require.Contains(t, info.Config().Channels, "second")
	})

	t.Run("published config is not the config of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(asyncapi3test.WithTitle("foo"), asyncapi3test.WithChannel("bar")))

		info, err := app.Mqtt.Add(file, enginetest.NewEngine())

		require.NoError(t, err)
		require.NotSame(t, file.Data, info.Config())
		require.NotNil(t, info.Config().Servers.Lookup("mokapi"), "default broker is added to the published config")
		require.Equal(t, 0, file.Data.(*asyncapi3.Config).Servers.Len(), "but not to the file")
	})

	t.Run("config is nil after the last file is removed", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := fileWithChannel("first")
		info, err := app.Mqtt.Add(file, enginetest.NewEngine())
		require.NoError(t, err)

		app.Mqtt.Remove(file)

		require.Nil(t, info.Config())
		require.NotNil(t, info.Store, "store is not removed with the last file")
		require.Nil(t, app.Mqtt.Get("foo"))
	})

	t.Run("Configs returns a copy", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a, b := twoFilesWithSameTitle("mqtt")
		_, err := app.Mqtt.Add(a, enginetest.NewEngine())
		require.NoError(t, err)
		info, err := app.Mqtt.Add(b, enginetest.NewEngine())
		require.NoError(t, err)

		configs := info.Configs()
		require.Len(t, configs, 2)
		configs[0] = nil

		require.NotContains(t, info.Configs(), (*dynamic.Config)(nil))
		require.Len(t, info.Configs(), 2)
	})

	t.Run("Configs does not race with Add", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a, b := twoFilesWithSameTitle("mqtt")
		info, err := app.Mqtt.Add(a, enginetest.NewEngine())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if _, err := app.Mqtt.Add(b, enginetest.NewEngine()); err != nil {
					t.Error(err)
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_ = info.Configs()
			}
		}()
		wg.Wait()
	})
}
