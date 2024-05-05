package engine

import (
	"github.com/go-co-op/gocron"
	"mokapi/engine/common"
	"time"
)

type Scheduler interface {
	Start()
	Close()

	Every(every string, handler func(), opt common.JobOptions) (Job, error)
	Cron(every string, handler func(), opt common.JobOptions) (Job, error)
	Remove(job Job)
}

type Job interface {
}

type DefaultScheduler struct {
	scheduler *gocron.Scheduler
}

func NewDefaultScheduler() Scheduler {
	return &DefaultScheduler{scheduler: gocron.NewScheduler(time.UTC)}
}

func (s *DefaultScheduler) Every(every string, handler func(), opt common.JobOptions) (Job, error) {
	s.scheduler.Every(every)

	if opt.Times > 0 {
		s.scheduler.LimitRunsTo(opt.Times)
	}
	if opt.SkipImmediateFirstRun {
		s.scheduler.WaitForSchedule()
	}

	return s.scheduler.Do(handler)
}

func (s *DefaultScheduler) Cron(expr string, handler func(), opt common.JobOptions) (Job, error) {
	s.scheduler.Cron(expr)

	if opt.Times > 0 {
		s.scheduler.LimitRunsTo(opt.Times)
	}
	if opt.SkipImmediateFirstRun {
		s.scheduler.WaitForSchedule()
	}

	return s.scheduler.Do(handler)
}

func (s *DefaultScheduler) Remove(job Job) {
	s.scheduler.RemoveByReference(job.(*gocron.Job))
}

func (s *DefaultScheduler) Start() {
	s.scheduler.StartAsync()
}

func (s *DefaultScheduler) Close() {
	s.scheduler.Stop()
}
