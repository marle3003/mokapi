package workflow

import (
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/workflow/runtime"
	"time"
)

type Scheduler struct {
	cron *gocron.Scheduler
	jobs map[string][]*gocron.Job
}

type task struct {
	workflow mokapi.Workflow
	options  []WorkflowOptions
}

func NewScheduler() *Scheduler {
	return &Scheduler{cron: gocron.NewScheduler(time.UTC), jobs: make(map[string][]*gocron.Job)}
}

func (s *Scheduler) AddOrUpdate(key string, workflows []mokapi.Workflow, options ...WorkflowOptions) error {
	jobs, ok := s.jobs[key]
	if ok {
		// remove all jobs
		for _, j := range jobs {
			s.cron.RemoveByReference(j)
		}
		jobs = nil
	} else {
		jobs = make([]*gocron.Job, 0)
		s.jobs[key] = jobs
	}

	for _, w := range workflows {
		for _, t := range w.On {
			if t.Schedule == nil || len(t.Schedule.Every) == 0 {
				continue
			}

			s.cron.Every(t.Schedule.Every)
			if t.Schedule.Iterations >= 1 {
				s.cron.LimitRunsTo(t.Schedule.Iterations)
			}

			t := task{
				workflow: w,
				options:  options,
			}

			j, err := s.cron.Do(func() {
				t.do()
			})
			if err != nil {
				return err
			}
			j.Tag(w.Name)
			jobs = append(jobs, j)
		}
	}

	if len(jobs) > 0 {
		s.jobs[key] = jobs
	}

	return nil
}

func (s *Scheduler) Start() {
	s.cron.StartAsync()
}

func (s *Scheduler) Stop() {
	for _, jobs := range s.jobs {
		for _, j := range jobs {
			s.cron.RemoveByReference(j)
		}
	}
}

func (t task) do() {
	ctx := runtime.NewWorkflowContext(actionCollection, fCollection)
	for _, opt := range t.options {
		opt(ctx)
	}

	s, _ := runtime.Run(t.workflow, ctx)
	log.WithField("action", s).Infof("executed action %v", t.workflow.Name)
}
