package logs

type Logs struct {
	Http []HttpLog
}

func New() *Logs {
	return &Logs{
		Http: make([]HttpLog, 0, 10),
	}
}
