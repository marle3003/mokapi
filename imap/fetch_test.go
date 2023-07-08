package imap_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
	"time"
)

type newHandler func(t *testing.T) imap.Handler

func TestServer_Fetch_Macros(t *testing.T) {
	testcases := []struct {
		name    string
		handler newHandler
		test    func(t *testing.T, c *imaptest.Client)
	}{
		{
			name: "fetch one FAST",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, 1, request.Sequence[0].Start)
						require.Equal(t, 1, request.Sequence[0].End)
						require.True(t, request.Options.Has(imap.FetchFlags), "has FetchFlags")
						require.True(t, request.Options.Has(imap.FetchInternalDate), "has FetchInternalDate")
						require.True(t, request.Options.Has(imap.FetchRFC822Size), "has FetchRFC822Size")

						date, err := time.Parse(time.RFC3339, "2023-03-16T13:07:04+01:00")
						require.NoError(t, err)
						msg := response.NewMessage(1)
						msg.WriteInternalDate(date)
						msg.WriteRFC822Size(43078)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("FETCH 1 FAST")
				require.NoError(t, err)
				require.Equal(t, "* 1 FETCH (INTERNALDATE \"16-Mar-2023 13:07:04 +0100\" RFC822.SIZE 43078)", lines[0])
			},
		},
		{
			name: "fetch two FAST",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, 1, request.Sequence[0].Start)
						require.Equal(t, 2, request.Sequence[0].End)

						date, err := time.Parse(time.RFC3339, "2023-03-16T13:07:04+01:00")
						require.NoError(t, err)
						msg := response.NewMessage(1)
						msg.WriteInternalDate(date)
						msg.WriteRFC822Size(43078)
						msg = response.NewMessage(2)
						msg.WriteInternalDate(date)
						msg.WriteRFC822Size(5000)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("FETCH 1:2 FAST")
				require.NoError(t, err)
				require.Equal(t, "* 1 FETCH (INTERNALDATE \"16-Mar-2023 13:07:04 +0100\" RFC822.SIZE 43078)", lines[0])
				require.Equal(t, "* 2 FETCH (INTERNALDATE \"16-Mar-2023 13:07:04 +0100\" RFC822.SIZE 5000)", lines[1])
			},
		},
		{
			name: "fetch one ALL",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.Has(imap.FetchFlags), "has FetchFlags")
						require.True(t, request.Options.Has(imap.FetchInternalDate), "has FetchInternalDate")
						require.True(t, request.Options.Has(imap.FetchRFC822Size), "has FetchRFC822Size")
						require.True(t, request.Options.Has(imap.FetchEnvelope), "has FetchEnvelope")

						date, err := time.Parse(time.RFC3339, "2023-03-16T13:07:04+01:00")
						require.NoError(t, err)
						msg := response.NewMessage(1)
						msg.WriteInternalDate(date)
						msg.WriteRFC822Size(43078)
						msg.WriteEnvelope(&imap.Envelope{
							Date:    date.Add(time.Hour * 2),
							Subject: "A Test Mail",
							From: []imap.Address{{
								Name:    "Bob",
								Mailbox: "bob",
								Host:    "mokapi.io",
							}},
							To: []imap.Address{{
								Name:    "Alice",
								Mailbox: "alice",
								Host:    "mokapi.io",
							}},
							MessageId: "123456",
						})
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("FETCH 1 ALL")
				require.NoError(t, err)

				date, err := imaptest.ParseInternalDate(lines[0])
				require.NoError(t, err)
				internalDate, _ := time.Parse(time.RFC3339, "2023-03-16T13:07:04+01:00")
				require.Equal(t, internalDate, date)

				size, err := imaptest.ParseSize(lines[0])
				require.NoError(t, err)
				require.Equal(t, int64(43078), size)

				envelope, err := imaptest.ParseEnvelope(lines[0])
				require.NoError(t, err)
				date, _ = time.Parse(time.RFC3339, "2023-03-16T15:07:04+01:00")
				require.Equal(t, date, envelope.Date)
				require.Equal(t, "A Test Mail", envelope.Subject)
				require.Equal(t, []imap.Address{{Name: "Bob", Mailbox: "bob", Host: "mokapi.io"}}, envelope.From)
				require.Equal(t, envelope.From, envelope.Sender)
				require.Equal(t, envelope.From, envelope.ReplyTo)
				require.Equal(t, "", envelope.InReplyTo)
				require.Equal(t, "123456", envelope.MessageId)
			},
		},
		{
			name: "fetch one FULL",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.Has(imap.FetchFlags), "has FetchFlags")
						require.True(t, request.Options.Has(imap.FetchInternalDate), "has FetchInternalDate")
						require.True(t, request.Options.Has(imap.FetchRFC822Size), "has FetchRFC822Size")
						require.True(t, request.Options.Has(imap.FetchEnvelope), "has FetchEnvelope")
						require.True(t, request.Options.Has(imap.FetchBodyStructure), "has FetchBodyStructure")

						date, err := time.Parse(time.RFC3339, "2023-03-16T13:07:04+01:00")
						require.NoError(t, err)
						msg := response.NewMessage(1)
						msg.WriteInternalDate(date)
						msg.WriteRFC822Size(43078)
						msg.WriteEnvelope(&imap.Envelope{
							Date:    date.Add(time.Hour * 2),
							Subject: "A Test Mail",
							From: []imap.Address{{
								Name:    "Bob",
								Mailbox: "bob",
								Host:    "mokapi.io",
							}},
							To: []imap.Address{{
								Name:    "Alice",
								Mailbox: "alice",
								Host:    "mokapi.io",
							}},
							MessageId: "123456",
						})
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imaptest.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("FETCH 1 FULL")
				require.NoError(t, err)
				_ = lines
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, err := try.GetFreePort()
			require.NoError(t, err)
			s := &imap.Server{
				Addr:    fmt.Sprintf(":%v", p),
				Handler: tc.handler(t),
			}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imaptest.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}
