package runtime_test

import (
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/common"
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

type jobListingEngine struct {
	common.EventEmitter
	jobs []common.JobInfo
}

func (e *jobListingEngine) Jobs() []common.JobInfo { return e.jobs }

func TestApp_Jobs(t *testing.T) {
	t.Run("engine does not list jobs", func(t *testing.T) {
		app := &runtime.App{Engine: common.EventEmitter(nil)}
		require.Nil(t, app.Jobs())

		app = &runtime.App{Engine: struct{ common.EventEmitter }{}}
		require.Nil(t, app.Jobs())
	})

	t.Run("engine lists jobs", func(t *testing.T) {
		jobs := []common.JobInfo{{Id: "a-0", Name: "a.js"}}
		app := &runtime.App{Engine: &jobListingEngine{jobs: jobs}}

		require.Equal(t, jobs, app.Jobs())
	})
}
