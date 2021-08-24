package server

import (
	"github.com/go-co-op/gocron"
	"mokapi/lua"
	"mokapi/lua/kafka"
	"mokapi/models"
)

func (s *Server) Run(event string, args ...interface{}) []*models.WorkflowLog {
	logs := make([]*models.WorkflowLog, 0)
	for _, s := range s.scripts {
		l := s.Run(event, args...)

		for _, e := range l {
			logs = append(logs, &models.WorkflowLog{
				Name:     e.Name,
				Logs:     e.Log,
				Duration: e.Duration,
			})
		}

	}

	return logs
}

func (s *Server) AddScript(key string, code string) {
	if script, ok := s.scripts[key]; ok {
		script.Close()
	}

	s.scripts[key] = lua.NewScript(key, code, kafka.NewKafka(s.writeKafkaMessage), s)
}

func (s *Server) NewJob(every string, do func(), times int) (*gocron.Job, error) {
	s.cron.Every(every)
	if times >= 0 {
		s.cron.LimitRunsTo(times)
	}
	return s.cron.Do(do)
}

func (s *Server) CancelJob(j *gocron.Job) {
	s.cron.RemoveByReference(j)
}
