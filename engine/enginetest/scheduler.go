package enginetest

import (
	"mokapi/engine"
	"mokapi/engine/common"
)

type Scheduler struct {
	EveryFunc  func(every string, handler func(), opt common.JobOptions) (engine.Job, error)
	CronFunc   func(every string, handler func(), opt common.JobOptions) (engine.Job, error)
	RemoveFunc func(job engine.Job)
}

func (s *Scheduler) Every(every string, handler func(), opt common.JobOptions) (engine.Job, error) {
	if s.EveryFunc != nil {
		return s.EveryFunc(every, handler, opt)
	}
	return nil, nil
}

func (s *Scheduler) Cron(expr string, handler func(), opt common.JobOptions) (engine.Job, error) {
	if s.CronFunc != nil {
		return s.CronFunc(expr, handler, opt)
	}
	return nil, nil
}

func (s *Scheduler) Remove(job engine.Job) {
	if s.RemoveFunc != nil {
		s.RemoveFunc(job)
	}
}

func (s *Scheduler) Start() {}

func (s *Scheduler) Close() {}
