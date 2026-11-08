package kafka_test

import (
	"errors"
	"mokapi/kafka"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Close may run while Serve is starting (e.g. a config reload removes the cluster right after it
// was added). Serve must then return ErrServerClosed. Run with -race.
func TestServer_CloseWhileServeIsStarting(t *testing.T) {
	for i := 0; i < 50; i++ {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		s := &kafka.Server{Handler: kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {})}
		done := make(chan error, 1)
		go func() { done <- s.Serve(l) }()
		s.Close()
		_ = l.Close()

		select {
		case err = <-done:
			require.True(t, errors.Is(err, kafka.ErrServerClosed), "Serve returned %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Serve did not return after Close")
		}
	}
}

// Close before Serve was called must not leave Serve running without a close signal. Run with -race.
func TestServer_ServeAfterClose(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	s := &kafka.Server{Handler: kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {})}
	s.Close()

	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()

	select {
	case err = <-done:
		require.True(t, errors.Is(err, kafka.ErrServerClosed), "Serve returned %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Serve keeps running although the server was closed before")
	}

	// the server owns the listener it was given, so the port is free again
	l2, err := net.Listen("tcp", l.Addr().String())
	require.NoError(t, err, "listener passed to Serve is still open")
	_ = l2.Close()
}

func TestServer_CloseTwice(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	s := &kafka.Server{Handler: kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {})}
	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()

	require.NotPanics(t, func() {
		s.Close()
		s.Close()
	})
	_ = l.Close()

	select {
	case err = <-done:
		require.ErrorIs(t, err, kafka.ErrServerClosed)
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return after Close")
	}
}

func TestServer_CloseWithoutServe(t *testing.T) {
	s := &kafka.Server{}
	require.NotPanics(t, func() {
		s.Close()
		s.Close()
	})
	require.ErrorIs(t, s.ListenAndServe(), kafka.ErrServerClosed)
}

func TestServer_Close_ClosesListenerPassedToServe(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()

	s := &kafka.Server{Handler: kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {})}
	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()

	// Serve registers the listener asynchronously, a successful dial shows it accepts
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	_ = conn.Close()

	s.Close()
	select {
	case err = <-done:
		require.ErrorIs(t, err, kafka.ErrServerClosed)
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

	s := &kafka.Server{Addr: l.Addr().String()}
	err = s.ListenAndServe()
	require.Error(t, err)
	require.NotErrorIs(t, err, kafka.ErrServerClosed)
}

func TestServer_ListenAndServe_ClosedServerDoesNotBind(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	s := &kafka.Server{Addr: l.Addr().String()}
	s.Close()
	require.ErrorIs(t, s.ListenAndServe(), kafka.ErrServerClosed)
}
