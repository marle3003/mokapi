package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/smtp"
	"mokapi/try"
	"testing"
)

func TestAppend(t *testing.T) {

	handler := &imaptest.Handler{
		AppendFunc: func(mailbox string, msg *smtp.Message, opt imap.AppendOptions) error {
			require.Equal(t, "1994-02-07 21:52:25 -0800 -0800", msg.Date.String())
			require.Equal(t, []smtp.Address{{Name: "Fred Foobar", Address: "foobar@Blurdybloop.COM"}}, msg.From)
			require.Equal(t, "afternoon meeting", msg.Subject)
			require.Equal(t, []smtp.Address{{Address: "mooch@owatagu.siam.edu"}}, msg.To)
			require.Equal(t, "<B27397-0100000@Blurdybloop.COM>", msg.MessageId)
			require.Equal(t, "TEXT/PLAIN; CHARSET=US-ASCII", msg.ContentType)
			require.Equal(t, "Hello Joe, do you think we can meet at 3:30 tomorrow?", msg.Body)
			return nil
		},
	}

	p := try.GetFreePort()
	s := &imap.Server{
		Addr:    fmt.Sprintf(":%v", p),
		Handler: handler,
	}
	defer s.Close()
	go func() {
		err := s.ListenAndServe()
		require.ErrorIs(t, err, imap.ErrServerClosed)
	}()

	c := imap.NewClient(fmt.Sprintf("localhost:%v", p))
	defer c.Close()

	_, err := c.Dial()
	require.NoError(t, err)

	err = c.PlainAuth("", "", "")
	require.NoError(t, err)

	res, err := c.SendRaw("A003 APPEND Sent {310}")
	require.NoError(t, err)
	require.Equal(t, "+ Ready for literal data", res)

	request := []string{
		"Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)",
		"From: Fred Foobar <foobar@Blurdybloop.COM>",
		"Subject: afternoon meeting",
		"To: mooch@owatagu.siam.edu",
		"Message-Id: <B27397-0100000@Blurdybloop.COM>",
		"MIME-Version: 1.0",
		"Content-Type: TEXT/PLAIN; CHARSET=US-ASCII",
		"",
		"Hello Joe, do you think we can meet at 3:30 tomorrow?",
		"",
	}

	res2, err := c.SendRawLines(request)
	require.NoError(t, err)
	require.Equal(t, "A003 OK APPEND completed", res2)

}

func TestAppend_Flags(t *testing.T) {

	handler := &imaptest.Handler{
		AppendFunc: func(mailbox string, msg *smtp.Message, opt imap.AppendOptions) error {
			require.Equal(t, []imap.Flag{imap.FlagSeen}, opt.Flags)
			return nil
		},
	}

	p := try.GetFreePort()
	s := &imap.Server{
		Addr:    fmt.Sprintf(":%v", p),
		Handler: handler,
	}
	defer s.Close()
	go func() {
		err := s.ListenAndServe()
		require.ErrorIs(t, err, imap.ErrServerClosed)
	}()

	c := imap.NewClient(fmt.Sprintf("localhost:%v", p))
	defer c.Close()

	_, err := c.Dial()
	require.NoError(t, err)

	err = c.PlainAuth("", "", "")
	require.NoError(t, err)

	res, err := c.SendRaw("A003 APPEND Sent (\\Seen) {310}")
	require.NoError(t, err)
	require.Equal(t, "+ Ready for literal data", res)

	request := []string{
		"Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)",
		"From: Fred Foobar <foobar@Blurdybloop.COM>",
		"Subject: afternoon meeting",
		"To: mooch@owatagu.siam.edu",
		"Message-Id: <B27397-0100000@Blurdybloop.COM>",
		"MIME-Version: 1.0",
		"Content-Type: TEXT/PLAIN; CHARSET=US-ASCII",
		"",
		"Hello Joe, do you think we can meet at 3:30 tomorrow?",
		"",
	}

	res2, err := c.SendRawLines(request)
	require.NoError(t, err)
	require.Equal(t, "A003 OK APPEND completed", res2)

}
