package mail_test

import (
	"context"
	"mokapi/engine/enginetest"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/providers/mail"
	"mokapi/runtime/events/eventstest"
	"mokapi/smtp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// lockedUpdateWriter is a thread-safe IDLE client, so that the race detector
// only reports races inside the mail package and not in the test double.
type lockedUpdateWriter struct {
	m sync.Mutex
	r imaptest.UpdateRecorder
}

func (w *lockedUpdateWriter) WriteNumMessages(n uint32) error {
	w.m.Lock()
	defer w.m.Unlock()
	return w.r.WriteNumMessages(n)
}

func (w *lockedUpdateWriter) WriteMessageFlags(msn uint32, flags []imap.Flag) error {
	w.m.Lock()
	defer w.m.Unlock()
	return w.r.WriteMessageFlags(msn, flags)
}

func (w *lockedUpdateWriter) WriteExpunge(msn uint32) error {
	w.m.Lock()
	defer w.m.Unlock()
	return w.r.WriteExpunge(msn)
}

// IMAP clients enter and leave IDLE while other connections store flags,
// expunge and while the SMTP handler appends incoming mail to the same
// folder. Run with -race.
func TestFolder_IdleClientsComeAndGoWhileFolderIsModified(t *testing.T) {
	cfg := &mail.Config{Mailboxes: map[string]*mail.MailboxConfig{
		"alice@mokapi.io": {Username: "alice", Password: "foo"},
	}}
	s := mail.NewStore(cfg)
	h := mail.NewHandler(cfg, s, enginetest.NewEngine(), &eventstest.Handler{})
	newClient := func() context.Context {
		ctx := imap.NewClientContext(context.Background(), "127.0.0.1:84793")
		require.NoError(t, h.Login("alice", "foo", ctx))
		_, err := h.Select("Inbox", false, ctx)
		require.NoError(t, err)
		return ctx
	}
	mb := s.Mailboxes["alice@mokapi.io"]
	mb.Append(&smtp.Message{MessageId: "seed"})

	const rounds = 200
	var wg sync.WaitGroup
	run := func(f func(i int)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < rounds; i++ {
				f(i)
			}
		}()
	}

	idleCtx := newClient()
	run(func(int) {
		done := make(chan struct{})
		require.NoError(t, h.Idle(&lockedUpdateWriter{}, done, idleCtx))
		close(done)
	})

	storeCtx := newClient()
	run(func(int) {
		_ = h.Store(&imap.StoreRequest{
			Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
			Action:   "add",
			Flags:    []imap.Flag{imap.FlagSeen},
		}, &imaptest.FetchRecorder{}, storeCtx)
	})

	expungeCtx := newClient()
	run(func(int) {
		_ = h.Expunge(nil, &imaptest.ExpungeRecorder{}, expungeCtx)
	})

	fetchCtx := newClient()
	run(func(int) {
		_ = h.Fetch(&imap.FetchRequest{
			Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
			Options:  imap.FetchOptions{Flags: true, UID: true},
		}, &imaptest.FetchRecorder{}, fetchCtx)
	})

	searchCtx := newClient()
	run(func(int) {
		_, _ = h.Search(&imap.SearchRequest{
			Criteria: &imap.SearchCriteria{Flag: []imap.Flag{imap.FlagSeen}},
		}, searchCtx)
	})

	// incoming SMTP mail
	run(func(int) {
		mb.Append(&smtp.Message{MessageId: "mail"})
	})

	wg.Wait()
}

func TestMailbox_Append_WithoutInbox(t *testing.T) {
	mb := &mail.Mailbox{Name: "alice@mokapi.io"}

	mb.Append(&smtp.Message{MessageId: "first"})

	require.Len(t, mb.Folders, 1)
	require.Len(t, mb.Folders["INBOX"].Messages, 1)
}

func TestMailbox_EnsureInbox_ConcurrentWithAppend(t *testing.T) {
	mb := &mail.Mailbox{Name: "alice@mokapi.io"}

	const clients = 20
	var wg sync.WaitGroup
	for i := 0; i < clients; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			mb.EnsureInbox()
		}()
		go func() {
			defer wg.Done()
			mb.Append(&smtp.Message{MessageId: "mail"})
		}()
	}
	wg.Wait()

	require.Len(t, mb.Folders, 1)
	require.Len(t, mb.Folders["INBOX"].Messages, clients)
}

func TestHandler_Idle_ClientIsNotWrittenToAfterDone(t *testing.T) {
	cfg := &mail.Config{Mailboxes: map[string]*mail.MailboxConfig{
		"alice@mokapi.io": {Username: "alice", Password: "foo"},
	}}
	s := mail.NewStore(cfg)
	h := mail.NewHandler(cfg, s, enginetest.NewEngine(), &eventstest.Handler{})
	ctx := imap.NewClientContext(context.Background(), "127.0.0.1:84793")
	require.NoError(t, h.Login("alice", "foo", ctx))
	_, err := h.Select("INBOX", false, ctx)
	require.NoError(t, err)
	mb := s.Mailboxes["alice@mokapi.io"]

	w := &lockedUpdateWriter{}
	done := make(chan struct{})
	require.NoError(t, h.Idle(w, done, ctx))

	mb.Append(&smtp.Message{MessageId: "while idle"})
	w.m.Lock()
	require.Len(t, w.r.Messages, 1)
	w.m.Unlock()

	// the writer is removed asynchronously after done is closed
	close(done)
	require.Eventually(t, func() bool {
		w.m.Lock()
		before := len(w.r.Messages)
		w.m.Unlock()
		mb.Append(&smtp.Message{MessageId: "after done"})
		w.m.Lock()
		defer w.m.Unlock()
		return len(w.r.Messages) == before
	}, 5*time.Second, time.Millisecond)
}
