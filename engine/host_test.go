package engine

import (
	"runtime"
	"testing"
	"time"

	r "github.com/stretchr/testify/require"
)

type hostTestLogger struct{}

func (*hostTestLogger) Info(...interface{})        {}
func (*hostTestLogger) Warn(...interface{})        {}
func (*hostTestLogger) Error(...interface{})       {}
func (*hostTestLogger) Debug(...interface{})       {}
func (*hostTestLogger) IsLevelEnabled(string) bool { return true }

func TestScriptHost_EventLogger(t *testing.T) {
	tests := []struct {
		name     string
		act      func(sh *scriptHost)
		expected []hostTestLog
	}{
		{
			name: "log before event handler ends",
			act: func(sh *scriptHost) {
				sh.Info("message")
				sh.endEventHandler()
			},
			expected: []hostTestLog{{level: "log", message: "message"}},
		},
		{
			name: "log after event handler ends",
			act: func(sh *scriptHost) {
				sh.endEventHandler()
				sh.Info("message")
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logs []hostTestLog
			sh := &scriptHost{engine: &Engine{logger: &hostTestLogger{}}}
			sh.startEventHandler(func(level, message string) {
				logs = append(logs, hostTestLog{level: level, message: message})
			})

			test.act(sh)

			r.Equal(t, test.expected, logs)
		})
	}
}

func TestScriptHost_EndEventHandlerWaitsForActiveLogger(t *testing.T) {
	sh := &scriptHost{engine: &Engine{logger: &hostTestLogger{}}}
	loggerStarted := make(chan struct{})
	releaseLogger := make(chan struct{})
	loggerDone := make(chan struct{})
	handlerEnded := make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-releaseLogger:
		default:
			close(releaseLogger)
		}
	})

	var logs []hostTestLog
	// Keep the logger active until endEventHandler is waiting for it.
	sh.startEventHandler(func(level, message string) {
		close(loggerStarted)
		<-releaseLogger
		logs = append(logs, hostTestLog{level: level, message: message})
	})

	go func() {
		sh.Info("message")
		close(loggerDone)
	}()
	<-loggerStarted

	go func() {
		sh.endEventHandler()
		close(handlerEnded)
	}()

	// A pending writer makes TryRLock fail, proving that endEventHandler
	// is waiting instead of clearing the logger while it is in use.
	writerWaiting := false
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for !writerWaiting {
		select {
		case <-handlerEnded:
			r.FailNow(t, "event handler ended while logger was active")
		case <-deadline.C:
			r.FailNow(t, "event handler did not attempt to acquire the logger lock")
		default:
		}

		if sh.eventLoggerMu.TryRLock() {
			sh.eventLoggerMu.RUnlock()
			runtime.Gosched()
			continue
		}
		writerWaiting = true
	}

	// Once the logger returns, both operations must finish and retain the log.
	close(releaseLogger)
	select {
	case <-loggerDone:
	case <-deadline.C:
		r.FailNow(t, "logger did not finish")
	}
	select {
	case <-handlerEnded:
	case <-deadline.C:
		r.FailNow(t, "event handler did not finish")
	}

	r.Equal(t, []hostTestLog{{level: "log", message: "message"}}, logs)
}

type hostTestLog struct {
	level   string
	message string
}
