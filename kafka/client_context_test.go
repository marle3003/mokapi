package kafka

import (
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestClientContext_AddGroup(t *testing.T) {
	ctx := &ClientContext{}
	ctx.AddGroup("foo", "client")
	require.Contains(t, ctx.Member, "foo")
	require.Equal(t, "client", ctx.Member["foo"])
}

func TestClientContext_GetOrCreateMemberId(t *testing.T) {
	testcases := []struct {
		name      string
		clientId  string
		groupName string
		test      func(t *testing.T, ctx *ClientContext)
	}{
		{
			name:      "empty clientId",
			clientId:  "",
			groupName: "foo",
			test: func(t *testing.T, ctx *ClientContext) {
				const id = "0000671e-0cc1-48a9-8a3d-dfa854b62d07"
				SetUUIDGenerator(func() string {
					return id
				})
				memberId := ctx.GetOrCreateMemberId("foo")
				require.Equal(t, id, memberId)
			},
		},
		{
			name:      "with clientId",
			clientId:  "foo",
			groupName: "foo",
			test: func(t *testing.T, ctx *ClientContext) {
				const id = "16c38b3b-b354-4c26-bf2f-8570981f901a"
				SetUUIDGenerator(func() string {
					return id
				})
				memberId := ctx.GetOrCreateMemberId("foo")
				require.Equal(t, "foo-"+id, memberId)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := &ClientContext{clientId: tc.clientId}
			ctx.AddGroup(tc.groupName, "")
			tc.test(t, ctx)
		})
	}
}

func TestClientContext_ConcurrentAccess(t *testing.T) {
	ctx := &ClientContext{}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			ctx.SetHeartbeat(time.Now())
			ctx.SetClientId("foo")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = ctx.Heartbeat()
			_ = ctx.ClientId()
		}
	}()
	wg.Wait()

	require.Equal(t, "foo", ctx.ClientId())
	require.False(t, ctx.Heartbeat().IsZero())
}
