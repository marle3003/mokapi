package mail_test

import (
	"fmt"
	"io"
	"mokapi/providers/mail"
	"mokapi/smtp"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// The store is read by SMTP, IMAP, the dashboard API and the MCP tools while a reload updates
// it. Run with -race.

func newRaceConfig(password string, maxInboxMails int) *mail.Config {
	return &mail.Config{
		Settings: &mail.Settings{AutoCreateMailbox: true, MaxInboxMails: maxInboxMails},
		Mailboxes: map[string]*mail.MailboxConfig{
			"alice@foo.bar": {
				Username:    "alice",
				Password:    password,
				Description: password,
				Folders:     map[string]*mail.FolderConfig{"INBOX": {Flags: []string{password}}, "Archive": {}},
			},
		},
	}
}

// runConcurrently runs every function in a loop of its own until all have done their iterations.
func runConcurrently(iterations int, funcs ...func(i int)) {
	var wg sync.WaitGroup
	for _, f := range funcs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				f(i)
			}
		}()
	}
	wg.Wait()
}

func TestStore_CredentialsWhileUpdate(t *testing.T) {
	s := mail.NewStore(newRaceConfig("a", 10))
	mb, ok := s.Mailbox("alice@foo.bar")
	require.True(t, ok)

	runConcurrently(200,
		func(i int) { s.Update(newRaceConfig(fmt.Sprint(i), 10)) },
		func(int) {
			_, _ = mb.Credentials()
			_ = mb.Description()
		},
	)
}

func TestStore_EnsureMailboxWhileListingAndUpdate(t *testing.T) {
	logrus.SetOutput(io.Discard)
	s := mail.NewStore(newRaceConfig("a", 10))

	runConcurrently(200,
		func(i int) { _ = s.EnsureMailbox(fmt.Sprintf("auto%d@foo.bar", i%20), nil) },
		func(i int) { s.Update(newRaceConfig(fmt.Sprint(i), 10)) },
		func(int) {
			for _, mb := range s.Mailboxes() {
				_ = mb.NumMessages()
			}
		},
		func(int) { _, _ = s.Mailbox("alice@foo.bar") },
	)

	require.Len(t, s.Mailboxes(), 21)
}

func TestStore_UpdateWhileReading(t *testing.T) {
	s := mail.NewStore(newRaceConfig("a", 10))
	mb, _ := s.Mailbox("alice@foo.bar")
	mb.Append(&smtp.Message{MessageId: "m1"})

	runConcurrently(200,
		func(i int) {
			cfg := newRaceConfig(fmt.Sprint(i), 10)
			cfg.Mailboxes["alice@foo.bar"].Folders[fmt.Sprintf("f%d", i%5)] = &mail.FolderConfig{}
			s.Update(cfg)
		},
		func(int) {
			require.NotNil(t, s.GetMail("m1"))
			_ = mb.FolderNames()
			_ = mb.Messages("")
			mb.ReadFolders(func(folders map[string]*mail.Folder) { _ = len(folders) })
			_ = mb.NumMessages()
		},
	)
}

func TestStore_UpdateMaxInboxMailsWhileAppending(t *testing.T) {
	s := mail.NewStore(newRaceConfig("a", 10))
	mb, _ := s.Mailbox("alice@foo.bar")

	runConcurrently(200,
		func(i int) { s.Update(newRaceConfig("a", 1+i%10)) },
		func(i int) { mb.Append(&smtp.Message{MessageId: fmt.Sprint(i)}) },
	)

	require.LessOrEqual(t, mb.NumMessages(), 10)
}
