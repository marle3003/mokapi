package imap_test

import (
	"fmt"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/try"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIdle(t *testing.T) {
	testcases := []struct {
		name    string
		handler newHandler
		test    func(t *testing.T, c *imap.Client)
	}{
		{
			name: "idle but not authenticated",
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{}
			},
			test: func(t *testing.T, c *imap.Client) {
				res, err := c.SendRaw("A01 IDLE")
				require.NoError(t, err)
				require.Equal(t, "A01 BAD Command is only valid in selected state", res)
			},
		},
		{
			name: "idle and done",
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
						return &imap.Selected{}, nil
					},
					IdleFunc: func(w imap.UpdateWriter, done chan struct{}, session map[string]interface{}) error {
						session["idle"] = done
						return nil
					},
					UnselectFunc: func(session map[string]interface{}) error {
						done := session["idle"].(chan struct{})
						doneClosed := false
						select {
						case <-done:
							doneClosed = true
						case <-time.After(time.Second):
						}

						require.True(t, doneClosed, "done is closed")
						return nil
					},
				}
			},
			test: func(t *testing.T, c *imap.Client) {
				err := c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				_, err = c.Select("INBOX", false)
				require.NoError(t, err)

				res, err := c.SendRaw("A01 IDLE")
				require.NoError(t, err)
				require.Equal(t, "+ idling", res)

				res, err = c.SendRaw("DONE")
				require.NoError(t, err)
				require.Equal(t, "A01 OK IDLE terminated", res)
			},
		},
		{
			name: "idle and done with lower case",
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
						return &imap.Selected{}, nil
					},
					IdleFunc: func(w imap.UpdateWriter, done chan struct{}, session map[string]interface{}) error {
						session["idle"] = done
						return nil
					},
					UnselectFunc: func(session map[string]interface{}) error {
						done := session["idle"].(chan struct{})
						doneClosed := false
						select {
						case <-done:
							doneClosed = true
						case <-time.After(time.Second):
						}

						require.True(t, doneClosed, "done is closed")
						return nil
					},
				}
			},
			test: func(t *testing.T, c *imap.Client) {
				err := c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				_, err = c.Select("INBOX", false)
				require.NoError(t, err)

				res, err := c.SendRaw("A01 idle")
				require.NoError(t, err)
				require.Equal(t, "+ idling", res)

				res, err = c.SendRaw("done")
				require.NoError(t, err)
				require.Equal(t, "A01 OK IDLE terminated", res)
			},
		},
		{
			name: "send keyword other than done",
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
						return &imap.Selected{}, nil
					},
					IdleFunc: func(w imap.UpdateWriter, done chan struct{}, session map[string]interface{}) error {
						session["idle"] = done
						return nil
					},
					UnselectFunc: func(session map[string]interface{}) error {
						done := session["idle"].(chan struct{})
						doneClosed := false
						select {
						case <-done:
							doneClosed = true
						case <-time.After(time.Second):
						}

						require.True(t, doneClosed, "done is closed")
						return nil
					},
				}
			},
			test: func(t *testing.T, c *imap.Client) {
				err := c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				_, err = c.Select("INBOX", false)
				require.NoError(t, err)

				res, err := c.SendRaw("A01 IDLE")
				require.NoError(t, err)
				require.Equal(t, "+ idling", res)

				res, err = c.SendRaw("A01 FINISHED")
				require.NoError(t, err)
				require.Equal(t, "A01 BAD Expected DONE to end IDLE", res)
			},
		},
		{
			name: "send updates while idle",
			handler: func(t *testing.T) imap.Handler {
				return &imaptest.Handler{
					SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
						return &imap.Selected{}, nil
					},
					IdleFunc: func(w imap.UpdateWriter, done chan struct{}, session map[string]interface{}) error {
						session["idle"] = done
						go func() {
							err := w.WriteNumMessages(10)
							require.NoError(t, err)
							err = w.WriteMessageFlags(20, []imap.Flag{imap.FlagSeen})
							require.NoError(t, err)
							err = w.WriteExpunge(1)
						}()
						return nil
					},
				}
			},
			test: func(t *testing.T, c *imap.Client) {
				err := c.PlainAuth("", "bob", "password")
				require.NoError(t, err)
				_, err = c.Select("INBOX", false)
				require.NoError(t, err)

				res, err := c.SendRaw("A01 IDLE")
				require.NoError(t, err)
				require.Equal(t, "+ idling", res)

				time.Sleep(4 * time.Second)

				res, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "* 10 EXISTS", res)

				res, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "* 20 FETCH (\\Seen)", res)

				res, err = c.ReadLine()
				require.NoError(t, err)
				require.Equal(t, "* 1 EXPUNGE", res)

				res, err = c.SendRaw("A01 FINISHED")
				require.NoError(t, err)
				require.Equal(t, "A01 BAD Expected DONE to end IDLE", res)
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
				Handler: tc.handler(t),
			}
			defer s.Close()
			go func() {
				err := s.ListenAndServe()
				require.ErrorIs(t, err, imap.ErrServerClosed)
			}()

			c := imap.NewClient(fmt.Sprintf("localhost:%v", p))
			defer func() { _ = c.Close() }()

			_, err := c.Dial()
			require.NoError(t, err)

			tc.test(t, c)
		})
	}
}

func TestIdle_DisconnectWithoutDone(t *testing.T) {
	var doneCh chan struct{}
	h := &imaptest.Handler{
		SelectFunc: func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error) {
			return &imap.Selected{}, nil
		},
		IdleFunc: func(w imap.UpdateWriter, done chan struct{}, session map[string]interface{}) error {
			doneCh = done
			return nil
		},
	}

	p := try.GetFreePort()
	s := &imap.Server{
		Addr:    fmt.Sprintf(":%v", p),
		Handler: h,
	}
	defer s.Close()
	go func() {
		err := s.ListenAndServe()
		require.ErrorIs(t, err, imap.ErrServerClosed)
	}()

	c := imap.NewClient(fmt.Sprintf("localhost:%v", p))

	_, err := c.Dial()
	require.NoError(t, err)

	err = c.PlainAuth("", "bob", "password")
	require.NoError(t, err)
	_, err = c.Select("INBOX", false)
	require.NoError(t, err)

	_, err = c.SendRaw("A01 IDLE")
	require.NoError(t, err)

	err = c.Disconnect()
	require.NoError(t, err)

	doneClosed := false
	select {
	case <-doneCh:
		doneClosed = true
	case <-time.After(time.Second):
	}

	require.True(t, doneClosed, "done is closed")
}
