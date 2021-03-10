package pipeline

import (
	"github.com/go-co-op/gocron"
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/parser"
	"mokapi/providers/pipeline/lang/runtime"
	"mokapi/providers/pipeline/lang/types"
	"path/filepath"
	"time"
)

type Scheduler struct {
	config *mokapi.Config
	cron   *gocron.Scheduler
	jobs   []*gocron.Job
}

func NewScheduler(config *mokapi.Config) *Scheduler {
	return &Scheduler{config: config, cron: gocron.NewScheduler(time.UTC)}
}

func (s *Scheduler) Start(options ...PipelineOptions) (err error) {
	scope := ast.NewScope(builtInFunctions)
	err = WithGlobalVars(map[types.Type]interface{}{
		runtime.EnvVarsType: runtime.NewEnvVars(
			runtime.FromOS(),
			runtime.With(map[string]string{
				"WORKING_DIRECTORY": filepath.Dir(s.config.ConfigPath),
			})),
	})(scope)
	if err != nil {
		return err
	}

	for _, o := range options {
		err = o(scope)
		if err != nil {
			return err
		}
	}

	var f *ast.File
	f, err = parser.ParseConfig(s.config, scope)
	if err != nil {
		return
	}

	for _, c := range s.config.Schedules {
		j, err := s.cron.Every(c.Every).Do(func() {
			runtime.RunPipeline(f, c.Pipeline)
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
		s.cron.Remove(j)
	}
}
