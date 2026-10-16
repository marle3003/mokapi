package engine

import (
	"mokapi/engine/common"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScheduler(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s Scheduler)
	}{
		{
			name: "add one job run immediately",
			test: func(t *testing.T, s Scheduler) {
				count := 0
				_, err := s.Every("2s", func() {
					count++
				}, common.JobOptions{})
				require.NoError(t, err)
				time.Sleep(100 * time.Millisecond)
				require.Equal(t, 1, count)
			},
		},
		{
			name: "add one job run not immediately",
			test: func(t *testing.T, s Scheduler) {
				count := 0
				_, err := s.Every("2s", func() {
					count++
				}, common.JobOptions{SkipImmediateFirstRun: true})
				require.NoError(t, err)
				time.Sleep(100 * time.Millisecond)
				require.Equal(t, 0, count)
			},
		},
		{
			name: "add one job run only one time",
			test: func(t *testing.T, s Scheduler) {
				count := 0
				_, err := s.Every("100ms", func() {
					time.Sleep(200 * time.Millisecond)
					count++
				}, common.JobOptions{Times: 1})
				require.NoError(t, err)
				time.Sleep(500 * time.Millisecond)
				require.Equal(t, 1, count)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewDefaultScheduler()
			defer s.Close()
			s.Start()
			tc.test(t, s)
		})
	}
}
