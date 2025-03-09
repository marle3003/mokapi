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
						require.Equal(t, uint32(1), request.Sequence.Ranges[0].Start.Value)
						require.Equal(t, uint32(1), request.Sequence.Ranges[0].End.Value)
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
						require.Equal(t, uint32(1), request.Sequence.Ranges[0].Start.Value)
						require.True(t, request.Sequence.Ranges[0].End.Star, "star is set")

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
						msg.WriteBodyStructure(imap.BodyStructure{
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
			name: "fetch body peek with one header field",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.NotNil(t, request.Options.Body, "body is set")

						msg := response.NewMessage(1)
						//msg.WriteBody(map[string]string{"date": date.Format(imap.DateTimeLayout)})
						w := msg.WriteBody2(request.Options.Body[0])
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
				cmd, err := c.Fetch(num(1), imap.FetchOptions{
					Body: []imap.FetchBodySection{
						{
							Type:   "header",
							Fields: []string{"date"},
							Peek:   true,
						},
					},
				})
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, "date: 01-Mar-2025 13:07:04 +0100\r\n\r\n", cmd.Messages[0].Body[0].Data)
			},
		},
		{
			name: "fetch body peek with one header field",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.NotNil(t, request.Options.Body, "body is set")
						require.NotNil(t, "header", request.Options.Body[0].Type)
						require.NotNil(t, []string{"date"}, request.Options.Body[0].Fields)
						require.True(t, request.Options.Body[0].Peek, "peek is set")

						msg := response.NewMessage(1)
						msg.WriteBody(map[string]string{"date": date.Format(imap.DateTimeLayout)})
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
				cmd, err := c.Fetch(num(1), imap.FetchOptions{
					Body: []imap.FetchBodySection{
						{
							Type:   "header",
							Fields: []string{"date"},
							Peek:   true,
						},
					},
				})
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, "date: 01-Mar-2025 13:07:04 +0100\r\n\r\n", cmd.Messages[0].Body[0].Data)
			},
		},
		{
			name: "fetch body peek",
			handler: func(t *testing.T) imap.Handler {
				h := &imaptest.Handler{
					FetchFunc: func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error {
						require.NotNil(t, "", request.Options.Body[0].Type)
						require.Len(t, request.Options.Body[0].Fields, 0)
						require.True(t, request.Options.Body[0].Peek, "peek is set")

						msg := response.NewMessage(1)
						msg.WriteBody(map[string]string{"date": date.Format(imap.DateTimeLayout)})
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
				cmd, err := c.Fetch(num(1), imap.FetchOptions{
					Body: []imap.FetchBodySection{
						{
							Type: "",
							Peek: true,
						},
					},
				})
				require.NoError(t, err)
				require.Equal(t, uint32(1), cmd.Messages[0].SeqNumber)
				require.Equal(t, "date: 01-Mar-2025 13:07:04 +0100\r\n\r\n", cmd.Messages[0].Body[0].Data)
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
						require.Equal(t, uint32(1), request.Sequence.Ranges[0].Start.Value)
						require.Equal(t, uint32(1), request.Sequence.Ranges[0].End.Value)
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
						require.Equal(t, uint32(1), request.Sequence.Ranges[0].Start.Value)
						require.True(t, request.Sequence.Ranges[0].End.Star)

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
		name     string
		request  string
		response []string
		handler  imap.Handler
	}{
		{
			name:    "fetch uid",
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
			name:    "fetch uid and body.peek",
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
					w := msg.WriteBody2(request.Options.Body[0])
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
		t.Run(tc.name, func(t *testing.T) {
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
			defer c.Close()

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
	s := imap.Range{}
	s.Start.Value = uint32(i)
	s.End.Value = uint32(i)
	return imap.IdSet{Ranges: []imap.Range{s}}
}

func seq(start, end int) imap.IdSet {
	s := imap.Range{}
	s.Start.Value = uint32(start)
	s.End.Value = uint32(end)
	return imap.IdSet{Ranges: []imap.Range{s}}
}

func all() imap.IdSet {
	s := imap.Range{}
	s.Start.Value = uint32(1)
	s.End.Star = true
	return imap.IdSet{Ranges: []imap.Range{s}}
}
