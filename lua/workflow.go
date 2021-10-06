package lua

import (
	"fmt"
	"github.com/go-co-op/gocron"
)

type eventHandler func(workflow *workflow, event string, args ...interface{}) bool

type Scheduler interface {
	Every(every string, do func(), times int) (*gocron.Job, error)
	Cron(cron string, do func(), times int) (*gocron.Job, error)
	CancelJob(*gocron.Job)
}

type workflow struct {
	Name          string
	EventHandlers map[string]eventHandler
	scheduler     Scheduler
	job           *gocron.Job
}

func (w *workflow) Event(event string, handler eventHandler) {
	w.EventHandlers[event] = handler
}

func (w *workflow) Timer(every string, handler func(w *workflow), args ...interface{}) error {
	if w.job != nil {
		return fmt.Errorf("already scheduled, call cancel register a new scheduled handler")
	}
	times := -1
	if len(args) > 0 {
		times = int(args[0].(float64))
	}

	j, err := w.scheduler.Every(every, func() {
		handler(w)
	}, times)

	if err != nil {
		return err
	}

	w.job = j
	return nil
}

func (w *workflow) Cron(expression string, handler func(w *workflow), args ...interface{}) error {
	if w.job != nil {
		return fmt.Errorf("already scheduled, call cancel register a new scheduled handler")
	}
	times := -1
	if len(args) > 0 {
		times = int(args[0].(float64))
	}

	j, err := w.scheduler.Cron(expression, func() {
		handler(w)
	}, times)

	if err != nil {
		return err
	}

	w.job = j
	return nil
}

func (w *workflow) Cancel() {
	w.scheduler.CancelJob(w.job)
	w.job = nil
}

func newWorkflow(name string, scheduler Scheduler) *workflow {
	return &workflow{
		Name:          name,
		EventHandlers: make(map[string]eventHandler),
		scheduler:     scheduler,
	}
}
