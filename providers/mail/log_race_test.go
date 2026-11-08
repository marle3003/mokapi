package mail_test

import (
	"context"
	"encoding/json"
	"io"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/mail"
	"mokapi/runtime/events"
	"mokapi/smtp"
	"mokapi/smtp/smtptest"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// The mail log is stored in the event store before the event handlers have run.
// Actions and Duration are attached afterwards, while the dashboard API may
// already serialize the event. Run with -race.
func TestHandler_LogIsReadWhileHandlerIsRunning(t *testing.T) {
	logrus.SetOutput(io.Discard)

	config := &mail.Config{Mailboxes: map[string]*mail.MailboxConfig{"alice@foo.bar": {}, "bob@foo.bar": {}}}
	sm := events.NewStoreManager(nil)
	sm.SetStore(10, events.NewTraits().WithNamespace("mail"))

	handlerRunning := make(chan struct{})
	emitter := enginetest.MailEventHandlerFunc(func(_ *smtp.Message, _ *smtp.Status) []*common.Action {
		close(handlerRunning)
		time.Sleep(10 * time.Millisecond)
		return []*common.Action{{Duration: 1}}
	})
	h := mail.NewHandler(config, mail.NewStore(config), emitter, sm)

	ctx := context.WithValue(smtp.NewClientContext(context.Background(), ""), "time", time.Now())
	sendMail(t, h, ctx)
	sendRcpt(t, h, ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-handlerRunning
		logs := sm.GetEvents(events.NewTraits().WithNamespace("mail"))
		require.Len(t, logs, 1)
		for i := 0; i < 200; i++ {
			_, err := json.Marshal(logs[0].Data)
			require.NoError(t, err)
			time.Sleep(100 * time.Microsecond)
		}
	}()

	rr := smtptest.NewRecorder()
	h.ServeSMTP(rr, smtp.NewDataRequest(&smtp.Message{
		From:    []smtp.Address{{Address: "alice@foo.bar"}},
		To:      []smtp.Address{{Address: "bob@foo.bar"}},
		Date:    time.Now(),
		Subject: "A mail message",
	}, ctx))
	wg.Wait()

	log := sm.GetEvents(events.NewTraits().WithNamespace("mail"))[0].Data.(*mail.Log)
	require.Len(t, log.GetActions(), 1)
}
