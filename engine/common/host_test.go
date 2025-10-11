package common

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewJobOptions(t *testing.T) {
	opts := NewJobOptions()
	require.NotNil(t, opts.Tags)
	require.Equal(t, -1, opts.Times)
	require.False(t, opts.SkipImmediateFirstRun)
}

func TestAction_String(t *testing.T) {
	a := &Action{
		Tags: map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
	}
	s := a.String()
	// loop over map results to random order
	require.True(t, s == "foo=bar, baz=qux" || s == "baz=qux, foo=bar")
}

func TestAction_AppendLog(t *testing.T) {
	a := &Action{}
	a.AppendLog("error", "foo")
	require.Equal(t, []Log{{Level: "error", Message: "foo"}}, a.Logs)
}

func TestAJobExecution_AppendLog(t *testing.T) {
	job := &JobExecution{}
	job.AppendLog("error", "foo")
	require.Equal(t, []Log{{Level: "error", Message: "foo"}}, job.Logs)
}

func TestAJobExecution_Title(t *testing.T) {
	job := &JobExecution{}
	require.Equal(t, "", job.Title())
	job.Tags = map[string]string{"name": "bar"}
	require.Equal(t, "bar", job.Title())
}
