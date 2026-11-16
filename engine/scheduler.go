package engine

import (
	"mokapi/engine/common"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type Scheduler interface {
	Start()
	Close()

	Every(every string, handler func(), opt common.JobOptions) (Job, error)
	Cron(every string, handler func(), opt common.JobOptions) (Job, error)
	Remove(job Job)
}

type Job interface {
	NextRun() time.Time
}

type DefaultScheduler struct {
	scheduler gocron.Scheduler
	m         sync.Mutex
}

type jobWrapper struct {
	job gocron.Job
}

func NewDefaultScheduler() Scheduler {
	s, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		panic(err)
	}
	return &DefaultScheduler{scheduler: s}
}

func (s *DefaultScheduler) Every(every string, handler func(), opt common.JobOptions) (Job, error) {
	s.m.Lock()
	defer s.m.Unlock()

	var jobOpts []gocron.JobOption
	if !opt.SkipImmediateFirstRun {
		jobOpts = append(jobOpts, gocron.WithStartAt(gocron.WithStartImmediately()))
	}

	return s.newJob(gocron.DurationJob(parseDuration(every)), handler, opt.Times, jobOpts)
}

// Cron runs the first run inline, so it counts toward opt.Times. If no run is left, there is
// nothing to schedule and the returned Job is nil.
func (s *DefaultScheduler) Cron(expr string, handler func(), opt common.JobOptions) (Job, error) {
	s.m.Lock()
	defer s.m.Unlock()

	times := opt.Times
	if !opt.SkipImmediateFirstRun {
		handler()
		if times > 0 {
			times--
			if times == 0 {
				return nil, nil
			}
		}
	}

	return s.newJob(gocron.CronJob(expr, false), handler, times, nil)
}

// newJob registers the job. times <= 0 means unlimited.
func (s *DefaultScheduler) newJob(def gocron.JobDefinition, handler func(), times int, opts []gocron.JobOption) (Job, error) {
	opts = append(opts, limitRuns(times)...)
	j, err := s.scheduler.NewJob(def, gocron.NewTask(limited(handler, times)), opts...)
	if err != nil {
		return nil, err
	}
	return &jobWrapper{job: j}, nil
}

// limited makes the handler skip every run after the first `times` runs. gocron applies
// WithLimitedRuns asynchronously: it removes the job only after the last run has completed, and a
// rescheduled tick can start one more run before that. The job is not removed from inside the
// handler, because gocron would cancel the context of the still running task; WithLimitedRuns
// removes it afterwards. An unlimited job (times <= 0) gets the handler unchanged.
func limited(handler func(), times int) func() {
	if times <= 0 {
		return handler
	}
	var runs atomic.Int64
	return func() {
		if runs.Add(1) > int64(times) {
			return
		}
		handler()
	}
}

// limitRuns returns no options for an unlimited job (times <= 0).
func limitRuns(times int) []gocron.JobOption {
	if times <= 0 {
		return nil
	}
	return []gocron.JobOption{
		gocron.WithLimitedRuns(uint(times)),
		// Ensure WithLimitedRuns works correctly if job execution outlasts the schedule interval.
		// Without SingletonMode, gocron might trigger concurrent runs before the previous run
		// finishes and decrements the remaining execution counter, leading to too many runs.
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	}
}

func (s *DefaultScheduler) Remove(job Job) {
	s.m.Lock()
	defer s.m.Unlock()

	if jw, ok := job.(*jobWrapper); ok {
		_ = s.scheduler.RemoveJob(jw.job.ID())
	}
}

func (s *DefaultScheduler) Start() {
	s.m.Lock()
	defer s.m.Unlock()

	s.scheduler.Start()
}

func (s *DefaultScheduler) Close() {
	s.m.Lock()
	defer s.m.Unlock()

	_ = s.scheduler.Shutdown()
}

func (j *jobWrapper) NextRun() time.Time {
	t, _ := j.job.NextRun()
	return t
}

func parseDuration(every string) time.Duration {
	d, err := time.ParseDuration(every)
	if err != nil {
		panic(err)
	}
	return d
}
