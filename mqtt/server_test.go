package mqtt_test

import (
	"errors"
	"fmt"
	"mokapi/mqtt"
	"mokapi/mqtt/mqtttest"
	"mokapi/try"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	testcases := []struct {
		name    string
		handler func(t *testing.T) mqtt.HandlerFunc
		test    func(t *testing.T, c *mqtttest.Client)
	}{
		{
			name: "Ping",
			handler: func(t *testing.T) mqtt.HandlerFunc {
				return func(rw mqtt.MessageWriter, m *mqtt.Message) {
					err := rw.Write(&mqtt.Message{
						Header: &mqtt.Header{
							Type: mqtt.PINGRESP,
						},
						Payload: &mqtt.PingResponse{},
					})
					require.NoError(t, err)
				}
			},
			test: func(t *testing.T, c *mqtttest.Client) {
				res, err := c.Send(&mqtt.Message{
					Header: &mqtt.Header{
						Type: mqtt.PINGREQ,
					},
					Payload: &mqtt.PingRequest{},
				})
				require.NoError(t, err)
				require.NotNil(t, res)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			addr := fmt.Sprintf(":%d", try.GetFreePort())
			s := &mqtt.Server{
				Addr:    addr,
				Handler: tc.handler(t),
			}
			go func() {
				err := s.ListenAndServe()
				if err != nil && !errors.Is(err, mqtt.ErrServerClosed) {
					panic(err)
				}
			}()
			defer s.Close()

			c := mqtttest.NewClient(addr)
			tc.test(t, c)
		})
	}
}

func TestServer_Close(t *testing.T) {
	t.Run("close while serve starts is race-free", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			s := &mqtt.Server{Addr: "127.0.0.1:0"}

			done := make(chan error, 1)
			go func() { done <- s.ListenAndServe() }()
			go s.Close()

			select {
			case err := <-done:
				require.ErrorIs(t, err, mqtt.ErrServerClosed)
			case <-time.After(2 * time.Second):
				t.Fatal("ListenAndServe did not return after Close")
			}
		}
	})

	t.Run("close before listen", func(t *testing.T) {
		s := &mqtt.Server{Addr: "127.0.0.1:0"}
		s.Close()

		done := make(chan error, 1)
		go func() { done <- s.ListenAndServe() }()

		select {
		case err := <-done:
			require.ErrorIs(t, err, mqtt.ErrServerClosed)
		case <-time.After(2 * time.Second):
			t.Fatal("ListenAndServe did not return after Close")
		}
	})
}

func TestServer_Serve_ClosedServerClosesListener(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()

	s := &mqtt.Server{}
	s.Close()

	require.ErrorIs(t, s.Serve(l), mqtt.ErrServerClosed)

	l2, err := net.Listen("tcp", addr)
	require.NoError(t, err, "listener passed to Serve is still open")
	_ = l2.Close()
}

func TestServer_Close_ClosesListenerPassedToServe(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()

	s := &mqtt.Server{}
	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()

	// Serve registers the listener asynchronously, a successful dial shows it accepts
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	_ = conn.Close()

	s.Close()
	select {
	case err = <-done:
		require.ErrorIs(t, err, mqtt.ErrServerClosed)
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return after Close")
	}

	l2, err := net.Listen("tcp", addr)
	require.NoError(t, err, "Close did not close the listener passed to Serve")
	_ = l2.Close()
}

func TestServer_ListenAndServe_PortInUse(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	s := &mqtt.Server{Addr: l.Addr().String()}
	err = s.ListenAndServe()
	require.Error(t, err)
	require.NotErrorIs(t, err, mqtt.ErrServerClosed)
}

func TestServer_ListenAndServe_ClosedServerDoesNotBind(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	s := &mqtt.Server{Addr: l.Addr().String()}
	s.Close()
	require.ErrorIs(t, s.ListenAndServe(), mqtt.ErrServerClosed)
}
