package common

type Host interface {
	Logger
	Every(every string, do func(), times int) (int, error)
	Cron(expr string, do func(), times int) (int, error)
	Cancel(jobId int) error

	OpenFile(file string) (string, error)

	On(event string, do func(args ...interface{}) (bool, error), tags map[string]string)
}

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}
