package ldap

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"mokapi/try"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T, rw ResponseWriter, req *Request)
		test    func(t *testing.T, c Client)
	}{
		{
			name: "SimpleBind",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				bind := req.Message.(*BindRequest)
				require.Equal(t, Simple, bind.Auth)
				require.Equal(t, "foo", bind.Name)
				require.Equal(t, "bar", bind.Password)
				err := rw.Write(&BindResponse{
					Result: Success,
				})
				require.NoError(t, err)
			},
			test: func(t *testing.T, c Client) {
				r, err := c.Bind("foo", "bar")
				require.NoError(t, err)
				require.Equal(t, Success, r.Result)
				require.Equal(t, "", r.MatchedDN)
				require.Equal(t, "", r.Message)
			},
		},
		{
			name: "unbind",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				panic("should not call handler")
			},
			test: func(t *testing.T, c Client) {
				err := c.Unbind()
				require.NoError(t, err)
			},
		},
		{
			name: "abandon",
			handler: func(t *testing.T, rw ResponseWriter, req *Request) {
				panic("should not call handler")
			},
			test: func(t *testing.T, c Client) {
				err := c.AbandonSearch(0)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s, c := newTestServer(HandlerFunc(func(rw ResponseWriter, req *Request) {
				tc.handler(t, rw, req)
			}))
			defer s.Close()
			require.NoError(t, c.Dial())
			tc.test(t, c)
		})
	}
}

func newTestServer(h Handler) (*Server, Client) {
	p := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", p)

	s := &Server{
		Addr:    addr,
		Handler: h,
	}

	go func() {
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, ErrServerClosed) {
			panic(err)
		}
	}()

	c := Client{Addr: addr}

	return s, c
}

func TestServer_Close(t *testing.T) {
	t.Run("close while serve starts is race-free", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			s := &Server{Addr: "127.0.0.1:0"}

			done := make(chan error, 1)
			go func() { done <- s.ListenAndServe() }()
			go s.Close()

			select {
			case err := <-done:
				require.ErrorIs(t, err, ErrServerClosed)
			case <-time.After(2 * time.Second):
				t.Fatal("ListenAndServe did not return after Close")
			}
		}
	})

	t.Run("close before listen", func(t *testing.T) {
		s := &Server{Addr: "127.0.0.1:0"}
		s.Close()

		done := make(chan error, 1)
		go func() { done <- s.ListenAndServe() }()

		select {
		case err := <-done:
			require.ErrorIs(t, err, ErrServerClosed)
		case <-time.After(2 * time.Second):
			t.Fatal("ListenAndServe did not return after Close")
		}
	})
}
