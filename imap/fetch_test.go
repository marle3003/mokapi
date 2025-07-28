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

func TestServer_Fetch(t *testing.T) {
	date, err := time.Parse(time.RFC3339, "2025-03-01T13:07:04+01:00")
	require.NoError(t, err)

	testcases := []struct {
		name    string
		handler newHandler
		test    func(t *testing.T, c *imap.Client)
	}{
		{
			name: "fetch UID",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, uint32(1), request.Sequence.Ids[0].(*imap.Range).Start.Value)
						require.Equal(t, uint32(1), request.Sequence.Ids[0].(*imap.Range).End.Value)
						require.True(t, request.Options.UID, "UID is set")

						msg := response.NewMessage(1)
						msg.WriteUID(11)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.Fetch(num(1), imap.FetchOptions{UID: true})
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, uint32(11), cmd.Messages[0].UID)
			},
		},
		{
			name: "fetch all messages in sequence (1:*)",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, uint32(1), request.Sequence.Ids[0].(*imap.Range).Start.Value)
						require.True(t, request.Sequence.Ids[0].(*imap.Range).End.Star, "star is set")

						msg := response.NewMessage(1)
						msg.WriteUID(11)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.Fetch(all(), imap.FetchOptions{UID: true})
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, uint32(11), cmd.Messages[0].UID)
			},
		},
		{
			name: "fetch FLAGS",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.Flags, "FLAGS is set")

						msg := response.NewMessage(1)
						msg.WriteFlags(imap.FlagSeen)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.Fetch(num(1), imap.FetchOptions{Flags: true})
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, []imap.Flag{imap.FlagSeen}, cmd.Messages[0].Flags)
			},
		},
		{
			name: "fetch flags and internal date",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.Flags, "flags is set")
						require.True(t, request.Options.InternalDate, "internal date is set")

						msg := response.NewMessage(1)
						msg.WriteFlags(imap.FlagSeen)
						msg.WriteInternalDate(date)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.Fetch(num(1), imap.FetchOptions{InternalDate: true, Flags: true})
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, date, cmd.Messages[0].InternalDate)
				require.Equal(t, []imap.Flag{imap.FlagSeen}, cmd.Messages[0].Flags)
			},
		},
		{
			name: "fetch body structure",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.BodyStructure, "body structure is set")

						msg := response.NewMessage(1)
						msg.WriteBodyStructure(&imap.BodyStructure{
							Type:     "text",
							Subtype:  "plain",
							Params:   map[string]string{"CHARSET": "UTF-8"},
							Encoding: "7BIT",
							Size:     384,
						})
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.Fetch(num(1), imap.FetchOptions{BodyStructure: true})
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, imap.BodyStructure{
					Type:     "text",
					Subtype:  "plain",
					Params:   map[string]string{"CHARSET": "UTF-8"},
					Encoding: "7BIT",
					Size:     384,
				}, cmd.Messages[0].BodyStructure)
			},
		},
		{
			name: "BODY.PEEK[HEADER.FIELDS (date)]",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.NotNil(t, request.Options.Body, "body is set")
						require.NotNil(t, "header", request.Options.Body[0].Specifier)
						require.NotNil(t, []string{"date"}, request.Options.Body[0].Fields)
						require.True(t, request.Options.Body[0].Peek, "peek is set")

						msg := response.NewMessage(1)
						w := msg.WriteBody(request.Options.Body[0])
						w.WriteHeader("date", date.Format(imap.DateTimeLayout))
						w.Close()
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 (BODY.PEEK[HEADER.FIELDS (date)])")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, "date: 01-Mar-2025 13:07:04 +0100\r\n\r\n", cmd.Messages[0].Body[0].Data)
			},
		},
		{
			name: "BODY.PEEK[HEADER]",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, "header", request.Options.Body[0].Specifier)
						require.Len(t, request.Options.Body[0].Fields, 0)
						require.True(t, request.Options.Body[0].Peek, "peek is set")

						msg := response.NewMessage(1)
						w := msg.WriteBody(request.Options.Body[0])
						w.WriteHeader("date", date.Format(imap.DateTimeLayout))
						w.Close()
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 (BODY.PEEK[HEADER])")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, "date: 01-Mar-2025 13:07:04 +0100\r\n\r\n", cmd.Messages[0].Body[0].Data)
			},
		},
		{
			name: "ALL",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.Flags, "Flags are set")
						require.True(t, request.Options.InternalDate, "InternalDate are set")
						require.True(t, request.Options.RFC822Size, "RFC822Size are set")
						require.True(t, request.Options.Envelope, "Envelope are set")

						response.NewMessage(1)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 ALL")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
			},
		},
		{
			name: "FAST",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.Flags, "Flags are set")
						require.True(t, request.Options.InternalDate, "InternalDate are set")
						require.True(t, request.Options.RFC822Size, "RFC822Size are set")

						response.NewMessage(1)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 FAST")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
			},
		},
		{
			name: "FULL",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.Flags, "Flags are set")
						require.True(t, request.Options.InternalDate, "InternalDate are set")
						require.True(t, request.Options.RFC822Size, "RFC822Size are set")
						require.True(t, request.Options.Envelope, "Envelope are set")
						require.True(t, request.Options.BodyStructure, "BodyStructure are set")

						response.NewMessage(1)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 FULL")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
			},
		},
		{
			name: "BODY",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.BodyStructure, "BodyStructure are set")

						response.NewMessage(1)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 BODY")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
			},
		},
		{
			name: "(BODY)",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.True(t, request.Options.BodyStructure, "BodyStructure are set")

						response.NewMessage(1)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 (BODY)")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
			},
		},
		{
			name: "(BODY[TEXT])",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, "text", request.Options.Body[0].Specifier)

						response.NewMessage(1)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 (BODY[TEXT])")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
			},
		},
		{
			name: "(BODY[1])",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, "", request.Options.Body[0].Specifier)
						require.Equal(t, []int{1}, request.Options.Body[0].Parts)

						response.NewMessage(1)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 (BODY[1])")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
			},
		},
		{
			name: "(BODY[1.MIME])",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, "mime", request.Options.Body[0].Specifier)
						require.Equal(t, []int{1}, request.Options.Body[0].Parts)

						response.NewMessage(1)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 (BODY[1.MIME])")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
			},
		},
		{
			name: "(BODY[]<0.100)",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, "", request.Options.Body[0].Specifier)
						require.Nil(t, request.Options.Body[0].Parts)
						require.NotNil(t, request.Options.Body[0].Partially)
						require.Equal(t, uint32(0), request.Options.Body[0].Partially.Offset)
						require.Equal(t, uint32(100), request.Options.Body[0].Partially.Limit)

						response.NewMessage(1)
						return nil
					},
				}
				return h
			},
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				cmd, err := c.FetchRaw("FETCH 1 (BODY[]<0.100>)")
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := try.GetFreePort()
			s := &imap.Server{
				Addr:    fmt.Sprintf(":%v", p),
				Handler: tc.handler(t),
			}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imap.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}

func TestServer_Fetch_Macros(t *testing.T) {
	testcases := []struct {
		name    string
		handler newHandler
		test    func(t *testing.T, c *imap.Client)
	}{
		{
			name: "fetch one FAST",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.Equal(t, imap.IdNum(1), request.Sequence.Ids[0].(imap.IdNum))
						require.True(t, request.Options.Flags, "flag is set")
						require.True(t, request.Options.InternalDate, "internal date is set")
						require.True(t, request.Options.RFC822Size, "RFC822Size is set")

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
			test: func(t *testing.T, c *imap.Client) {
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
						require.Equal(t, uint32(1), request.Sequence.Ids[0].(*imap.Range).Start.Value)
						require.True(t, request.Sequence.Ids[0].(*imap.Range).End.Star)

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
			test: func(t *testing.T, c *imap.Client) {
				_, err := c.Dial()
				require.NoError(t, err)
				err = c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				lines, err := c.Send("FETCH 1:* FAST")
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
						require.True(t, request.Options.Flags, "flag is set")
						require.True(t, request.Options.InternalDate, "internal date is set")
						require.True(t, request.Options.RFC822Size, "RFC822Size is set")
						require.True(t, request.Options.Envelope, "envelope is set")

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
			test: func(t *testing.T, c *imap.Client) {
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
						require.True(t, request.Options.Flags, "flag is set")
						require.True(t, request.Options.InternalDate, "internal date is set")
						require.True(t, request.Options.RFC822Size, "RFC822Size is set")
						require.True(t, request.Options.Envelope, "envelope is set")
						require.True(t, request.Options.BodyStructure, "body structure is set")

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
			test: func(t *testing.T, c *imap.Client) {
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
			p := try.GetFreePort()
			s := &imap.Server{
				Addr:    fmt.Sprintf(":%v", p),
				Handler: tc.handler(t),
			}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imap.NewClient(fmt.Sprintf("localhost:%v", p))

			tc.test(t, c)
		})
	}
}

func TestFetch_Response(t *testing.T) {
	testcases := []struct {
		request  string
		response []string
		handler  imap.Handler
	}{
		{
			request: "FETCH 1 (UID)",
			response: []string{
				"* 1 FETCH (UID 11)",
				"A0002 OK FETCH completed",
			},
			handler: &imaptest.Handler{
				FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
					msg := response.NewMessage(1)
					msg.WriteUID(11)
					return nil
				},
			},
		},
		{
			request: "FETCH 1 (UID BODY.PEEK[])",
			response: []string{
				"* 1 FETCH (UID 11 BODY[] {36}",
				"From: bob@foo.bar",
				"",
				"Hello World",
				"",
				")",
				"A0002 OK FETCH completed",
			},
			handler: &imaptest.Handler{
				FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
					msg := response.NewMessage(1)
					msg.WriteUID(11)
					w := msg.WriteBody(request.Options.Body[0])
					w.WriteHeader("From", "bob@foo.bar")
					w.WriteBody("Hello World")
					w.Close()
					return nil
				},
			},
		},
		{
			request: "FETCH 1 (UID BODY.PEEK[TEXT]<0.2048>)",
			response: []string{
				"* 1 FETCH (UID 11 BODY[] {36}",
				"From: bob@foo.bar",
				"",
				"Hello World",
				"",
				")",
				"A0002 OK FETCH completed",
			},
			handler: &imaptest.Handler{
				FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
					require.Equal(t, uint32(0), request.Options.Body[0].Partially.Offset)
					require.Equal(t, uint32(2048), request.Options.Body[0].Partially.Limit)

					msg := response.NewMessage(1)
					msg.WriteUID(11)
					w := msg.WriteBody(request.Options.Body[0])
					w.WriteHeader("From", "bob@foo.bar")
					w.WriteBody("Hello World")
					w.Close()
					return nil
				},
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.request, func(t *testing.T) {
			p := try.GetFreePort()
			s := &imap.Server{
				Addr:    fmt.Sprintf(":%v", p),
				Handler: tc.handler,
			}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imap.NewClient(fmt.Sprintf("localhost:%v", p))
			defer func() {
				_ = c.Close()
			}()

			_, err := c.Dial()
			require.NoError(t, err)

			err = c.PlainAuth("", "", "")
			require.NoError(t, err)

			res, err := c.Send(tc.request)
			require.NoError(t, err)
			require.Equal(t, tc.response, res)
		})
	}
}

func num(i int) imap.IdSet {
	s := &imap.Range{}
	s.Start.Value = uint32(i)
	s.End.Value = uint32(i)
	return imap.IdSet{Ids: []imap.Set{s}}
}

func all() imap.IdSet {
	s := &imap.Range{}
	s.Start.Value = uint32(1)
	s.End.Star = true
	return imap.IdSet{Ids: []imap.Set{s}}
}
