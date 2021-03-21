package pipeline

import (
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/parser"
	"mokapi/providers/pipeline/lang/runtime"
	"mokapi/providers/pipeline/lang/types"
	"path/filepath"
	"time"
)

type Scheduler struct {
	cron *gocron.Scheduler
	jobs []*gocron.Job
}

func NewScheduler() *Scheduler {
	return &Scheduler{cron: gocron.NewScheduler(time.UTC)}
}

func (s *Scheduler) Start(config *mokapi.Config, options ...PipelineOptions) (err error) {
	if config == nil || len(config.Schedules) == 0 {
		return
	}

	for _, c := range config.Schedules {
		t := s.cron.Every(c.Every)
		if c.Iterations >= 1 {
			t.LimitRunsTo(c.Iterations)
		}
		j, err := t.Do(func() {
			do(c.Name, c.Pipeline, config, options...)
		})
		if err != nil {
			return err
		}
		s.jobs = append(s.jobs, j)
	}

	s.cron.StartAsync()

	return
}

func (s *Scheduler) Stop() {
	for _, j := range s.jobs {
		s.cron.RemoveByReference(j)
	}
}

func do(name string, pipeline string, config *mokapi.Config, options ...PipelineOptions) {
	scope := ast.NewScope(builtInFunctions)
	err := WithGlobalVars(map[types.Type]interface{}{
		runtime.EnvVarsType: runtime.NewEnvVars(
			runtime.FromOS(),
			runtime.With(map[string]string{
				"WORKING_DIRECTORY": filepath.Dir(config.ConfigPath),
			})),
	})(scope)
	if err != nil {
		log.Errorf("error in job %q: %v", name, err.Error())
	}

	for _, o := range options {
		err = o(scope)
		if err != nil {
			log.Errorf("error in job %q: %v", name, err.Error())
		}
	}

	var f *ast.File
	f, err = parser.ParseConfig(config, scope)
	if err != nil {
		return
	}

	if err := runtime.RunPipeline(f, pipeline); err != nil {
		log.Errorf("job %v, error in pipeline %q: %v", name, pipeline, err)
	}
}
