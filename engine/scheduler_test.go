package engine

import (
	"context"
	"mokapi/engine/common"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jonboulle/clockwork"
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
				var count atomic.Int32
				_, err := s.Every("2s", func() {
					count.Add(1)
				}, common.JobOptions{})
				require.NoError(t, err)
				time.Sleep(100 * time.Millisecond)
				require.Equal(t, int32(1), count.Load())
			},
		},
		{
			name: "add one job run not immediately",
			test: func(t *testing.T, s Scheduler) {
				var count atomic.Int32
				_, err := s.Every("2s", func() {
					count.Add(1)
				}, common.JobOptions{SkipImmediateFirstRun: true})
				require.NoError(t, err)
				time.Sleep(100 * time.Millisecond)
				require.Equal(t, int32(0), count.Load())
			},
		},
		{
			name: "add one job run only one time",
			test: func(t *testing.T, s Scheduler) {
				var count atomic.Int32
				_, err := s.Every("100ms", func() {
					time.Sleep(200 * time.Millisecond)
					count.Add(1)
				}, common.JobOptions{Times: 1})
				require.NoError(t, err)
				require.Eventually(t, func() bool { return count.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
				require.Never(t, func() bool { return count.Load() > 1 }, 500*time.Millisecond, 10*time.Millisecond)
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

func TestLimited(t *testing.T) {
	t.Run("skips runs after times", func(t *testing.T) {
		var count int
		h := limited(func() { count++ }, 2)
		for i := 0; i < 3; i++ {
			h()
		}
		require.Equal(t, 2, count)
	})
	t.Run("unlimited job runs every time", func(t *testing.T) {
		var count int
		h := limited(func() { count++ }, 0)
		for i := 0; i < 5; i++ {
			h()
		}
		require.Equal(t, 5, count)
	})
}

// cronScheduler uses a fake clock because a cron expression has minute granularity.
func cronScheduler(t *testing.T) (*DefaultScheduler, *clockwork.FakeClock) {
	clock := clockwork.NewFakeClockAt(time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC))
	g, err := gocron.NewScheduler(gocron.WithLocation(time.UTC), gocron.WithClock(clock))
	require.NoError(t, err)
	s := &DefaultScheduler{scheduler: g}
	s.Start()
	t.Cleanup(s.Close)
	return s, clock
}

// advanceUntilRuns moves the fake clock minute by minute until the handler has run `want` times.
// Each step waits until the scheduler has registered its timer, so no due run is skipped.
func advanceUntilRuns(t *testing.T, clock *clockwork.FakeClock, count *atomic.Int32, want int32) {
	t.Helper()
	for i := 0; i < 10 && count.Load() < want; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		require.NoError(t, clock.BlockUntilContext(ctx, 1))
		cancel()
		clock.Advance(time.Minute)
		runs := count.Load() + 1
		require.Eventually(t, func() bool { return count.Load() >= runs }, 5*time.Second, time.Millisecond)
	}
	require.Equal(t, want, count.Load())
}

func TestScheduler_Cron_Times(t *testing.T) {
	testcases := []struct {
		name string
		opt  common.JobOptions
		// runs is the number of runs the job is limited to, 0 for a job that keeps running
		runs int32
	}{
		{name: "immediate first run counts as the only run", opt: common.JobOptions{Times: 1}, runs: 1},
		{name: "immediate first run counts as one of two runs", opt: common.JobOptions{Times: 2}, runs: 2},
		{name: "skipped first run does not count", opt: common.JobOptions{Times: 2, SkipImmediateFirstRun: true}, runs: 2},
		{name: "without times the job keeps running", opt: common.JobOptions{}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, clock := cronScheduler(t)
			var count atomic.Int32
			_, err := s.Cron("* * * * *", func() { count.Add(1) }, tc.opt)
			require.NoError(t, err)

			if tc.runs == 0 {
				advanceUntilRuns(t, clock, &count, 4)
				require.Len(t, s.scheduler.Jobs(), 1)
				return
			}
			advanceUntilRuns(t, clock, &count, tc.runs)
			// gocron removes a limited job after its last run, so no further run can follow
			require.Eventually(t, func() bool { return len(s.scheduler.Jobs()) == 0 }, 5*time.Second, time.Millisecond)
			require.Equal(t, tc.runs, count.Load())
		})
	}
}

func TestScheduler_Cron_NoRunLeft(t *testing.T) {
	t.Run("single immediate run creates no scheduler job", func(t *testing.T) {
		s, _ := cronScheduler(t)

		j, err := s.Cron("* * * * *", func() {}, common.JobOptions{Times: 1})

		require.NoError(t, err)
		require.Nil(t, j)
		require.Empty(t, s.scheduler.Jobs())
	})

	t.Run("removing the missing job is a no-op", func(t *testing.T) {
		s, _ := cronScheduler(t)
		j, err := s.Cron("* * * * *", func() {}, common.JobOptions{Times: 1})
		require.NoError(t, err)

		require.NotPanics(t, func() { s.Remove(j) })
	})

	t.Run("job with runs left is scheduled", func(t *testing.T) {
		s, _ := cronScheduler(t)

		j, err := s.Cron("* * * * *", func() {}, common.JobOptions{Times: 2})

		require.NoError(t, err)
		require.NotNil(t, j)
		require.Len(t, s.scheduler.Jobs(), 1)
	})

	t.Run("invalid expression is an error", func(t *testing.T) {
		s, _ := cronScheduler(t)

		_, err := s.Cron("invalid", func() {}, common.JobOptions{Times: 2})

		require.Error(t, err)
	})
}
