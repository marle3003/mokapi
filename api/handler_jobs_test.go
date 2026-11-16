package api

import (
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"strings"
	"testing"
	"time"
)

type jobsEngine struct {
	common.EventEmitter
	jobs []common.JobInfo
}

func (e *jobsEngine) Jobs() []common.JobInfo { return e.jobs }

func TestHandler_Jobs(t *testing.T) {
	testcases := []struct {
		name string
		app  func() *runtime.App
		test []try.ResponseCondition
	}{
		{
			name: "engine does not list jobs",
			app:  func() *runtime.App { return &runtime.App{} },
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(`[]`),
			},
		},
		{
			name: "no jobs",
			app:  func() *runtime.App { return &runtime.App{Engine: &jobsEngine{}} },
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasBody(`[]`),
			},
		},
		{
			name: "job that has not run: nextRun, no lastRun",
			app: func() *runtime.App {
				return &runtime.App{Engine: &jobsEngine{jobs: []common.JobInfo{{
					Id:       "key-0",
					Name:     "kafka.ts",
					Schedule: "5m",
					MaxRuns:  2,
					State:    common.JobStateScheduled,
					NextRun:  try.MustTime("2026-11-16T12:05:00Z"),
					Tags:     map[string]string{"name": "kafka.ts", "file": "/path/kafka.ts", "fileKey": "key"},
				}}}}
			},
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasBody(`[{"id":"key-0","name":"kafka.ts","schedule":"5m","maxRuns":2,"runs":0,"state":"scheduled","nextRun":"2026-11-16T12:05:00Z","tags":{"file":"/path/kafka.ts","fileKey":"key","name":"kafka.ts"}}]`),
			},
		},
		{
			name: "finished job: lastRun, no nextRun",
			app: func() *runtime.App {
				return &runtime.App{Engine: &jobsEngine{jobs: []common.JobInfo{{
					Id:       "key-0",
					Name:     "kafka.ts",
					Schedule: "5m",
					MaxRuns:  2,
					Runs:     2,
					State:    common.JobStateFinished,
					Tags:     map[string]string{"name": "kafka.ts"},
					LastRun: &common.JobRun{
						EventId:  "evt-1",
						Time:     time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC),
						Duration: 12,
						Status:   common.JobStatusSuccess,
					},
				}}}}
			},
			test: []try.ResponseCondition{
				try.HasStatusCode(http.StatusOK),
				try.HasBody(`[{"id":"key-0","name":"kafka.ts","schedule":"5m","maxRuns":2,"runs":2,"state":"finished","tags":{"name":"kafka.ts"},"lastRun":{"eventId":"evt-1","time":"2026-11-16T12:00:00Z","duration":12,"status":"success"}}]`),
			},
		},
		{
			name: "order of the engine is kept",
			app: func() *runtime.App {
				return &runtime.App{Engine: &jobsEngine{jobs: []common.JobInfo{
					{Id: "b", Name: "b"}, {Id: "a", Name: "a"},
				}}}
			},
			test: []try.ResponseCondition{
				try.AssertBody(func(t *testing.T, body string) {
					ia, ib := strings.Index(body, `"id":"a"`), strings.Index(body, `"id":"b"`)
					if ib < 0 || ia < 0 || ib > ia {
						t.Fatalf("expected b before a: %s", body)
					}
				}),
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app(), static.Api{})

			try.Handler(t, http.MethodGet, "http://foo.api/api/jobs", nil, "", h, tc.test...)
		})
	}
}
