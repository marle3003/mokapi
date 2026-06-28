package engine

import (
	"mokapi/engine/common"
	"sync"
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

	jobDef := gocron.DurationJob(parseDuration(every))

	var jobOpts []gocron.JobOption
	if opt.Times > 0 {
		jobOpts = append(jobOpts, gocron.WithLimitedRuns(uint(opt.Times)))
	}
	if !opt.SkipImmediateFirstRun {
		jobOpts = append(jobOpts, gocron.WithStartAt(gocron.WithStartImmediately()))
	}

	j, err := s.scheduler.NewJob(jobDef, gocron.NewTask(handler), jobOpts...)
	if err != nil {
		return nil, err
	}

	return &jobWrapper{job: j}, nil
}

func (s *DefaultScheduler) Cron(expr string, handler func(), opt common.JobOptions) (Job, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if !opt.SkipImmediateFirstRun {
		handler()
	}

	jobDef := gocron.CronJob(expr, false)

	j, err := s.scheduler.NewJob(jobDef, gocron.NewTask(handler))
	if err != nil {
		return nil, err
	}

	return &jobWrapper{job: j}, nil
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
