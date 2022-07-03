package service

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/kafkatest"
	"mokapi/try"
	"testing"
)

func TestKafkaBroker_Add(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "no handler",
			f: func(t *testing.T) {
				port, err := try.GetFreePort()
				require.NoError(t, err)
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				b := NewKafkaBroker(fmt.Sprintf("%v", port))
				b.Start()
				defer b.Stop()
				client := kafkatest.NewClient(addr, "test")
				r, err := client.ApiVersion(3, &apiVersion.Request{})
				require.Equal(t, kafka.UnknownServerError, r.ErrorCode)
			},
		},
		{
			name: "with handler",
			f: func(t *testing.T) {
				port, err := try.GetFreePort()
				require.NoError(t, err)
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				b := NewKafkaBroker(fmt.Sprintf("%v", port))
				b.Start()
				defer b.Stop()

				called := false
				b.Add(addr, kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {
					called = true
					rw.Write(&apiVersion.Response{})
				}))
				client := kafkatest.NewClient(addr, "test")
				r, err := client.ApiVersion(3, &apiVersion.Request{})
				require.Equal(t, kafka.None, r.ErrorCode)
				require.True(t, called, "handler should be called")
			},
		},
		{
			name: "with handler but wrong host",
			f: func(t *testing.T) {
				port, err := try.GetFreePort()
				require.NoError(t, err)
				addr := fmt.Sprintf("foo:%v", port)
				b := NewKafkaBroker(fmt.Sprintf("%v", port))
				b.Start()
				defer b.Stop()

				called := false
				b.Add(fmt.Sprintf("localhost:%v", port), kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {
					called = true
					rw.Write(&apiVersion.Response{})
				}))

				r := kafkatest.NewRecorder()
				b.ServeMessage(r, &kafka.Request{Message: &apiVersion.Request{}, Host: addr})

				res := r.Message.(*apiVersion.Response)
				require.NoError(t, err)
				require.Equal(t, kafka.UnknownServerError, res.ErrorCode)
				require.False(t, called, "handler should not be called")
			},
		},
		{
			name: "with handler any host",
			f: func(t *testing.T) {
				port, err := try.GetFreePort()
				require.NoError(t, err)
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				b := NewKafkaBroker(fmt.Sprintf("%v", port))
				b.Start()
				defer b.Stop()

				called := false
				b.Add(fmt.Sprintf(":%v", port), kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {
					called = true
					rw.Write(&apiVersion.Response{})
				}))
				client := kafkatest.NewClient(addr, "test")
				r, err := client.ApiVersion(3, &apiVersion.Request{})
				require.Equal(t, kafka.None, r.ErrorCode)
				require.True(t, called, "handler should be called")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}

func TestKafkaBroker_Remove(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "no handler",
			f: func(t *testing.T) {
				port, err := try.GetFreePort()
				require.NoError(t, err)
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				b := NewKafkaBroker(fmt.Sprintf("%v", port))
				defer b.Stop()
				b.Start()
				b.Remove("")
				client := kafkatest.NewClient(addr, "test")
				_, err = client.ApiVersion(3, &apiVersion.Request{})
				require.Error(t, err)
			},
		},
		{
			name: "remove handler",
			f: func(t *testing.T) {
				port, err := try.GetFreePort()
				require.NoError(t, err)
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				b := NewKafkaBroker(fmt.Sprintf("%v", port))
				defer b.Stop()
				b.Start()
				b.Add(addr, nil)
				b.Remove(addr)
				client := kafkatest.NewClient(addr, "test")
				_, err = client.ApiVersion(3, &apiVersion.Request{})
				require.Error(t, err)
			},
		},
		{
			name: "remove other handler",
			f: func(t *testing.T) {
				port, err := try.GetFreePort()
				require.NoError(t, err)
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				b := NewKafkaBroker(fmt.Sprintf("%v", port))
				defer b.Stop()
				b.Start()
				called := false
				b.Add(fmt.Sprintf(":%v", port), kafka.HandlerFunc(func(rw kafka.ResponseWriter, req *kafka.Request) {
					called = true
					rw.Write(&apiVersion.Response{})
				}))
				b.Remove(fmt.Sprintf("localhost:%v", port))
				client := kafkatest.NewClient(addr, "test")
				r, err := client.ApiVersion(3, &apiVersion.Request{})
				require.Equal(t, kafka.None, r.ErrorCode)
				require.True(t, called, "handler should be called")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}
