package runtime_test

import (
	"io"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
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
				requireMergedWithB(t, info.Config)

				app.Mqtt.Remove(b)

				info = app.Mqtt.Get("foo")
				require.Len(t, info.Configs(), 1)
				requireOnlyFileA(t, "mqtt", info.Config)
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
				require.NotNil(t, info.Servers.Lookup("mokapi"))
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

				requireExtraPropertyOnlyIn(t, info.Config, a, b)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, runtime.New(&static.Config{}, &dynamictest.Reader{}))
		})
	}
}
