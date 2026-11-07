package kafka

import (
	"github.com/stretchr/testify/require"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestClientContext_AddGroup(t *testing.T) {
	ctx := &ClientContext{}
	ctx.AddGroup("foo", "client")
	memberId, ok := ctx.MemberId("foo")
	require.True(t, ok)
	require.Equal(t, "client", memberId)
	require.True(t, ctx.HasGroup("foo"))
}

func TestClientContext_UnknownGroup(t *testing.T) {
	ctx := &ClientContext{}

	memberId, ok := ctx.MemberId("foo")
	require.False(t, ok)
	require.Empty(t, memberId)
	require.False(t, ctx.HasGroup("foo"))
	require.Empty(t, ctx.Groups())
}

func TestClientContext_GroupsReturnsCopy(t *testing.T) {
	ctx := &ClientContext{}
	ctx.AddGroup("foo", "m1")

	groups := ctx.Groups()
	require.Equal(t, map[string]string{"foo": "m1"}, groups)

	groups["bar"] = "m2"
	groups["foo"] = "changed"
	require.Equal(t, map[string]string{"foo": "m1"}, ctx.Groups())
}

func TestClientContext_GetOrCreateMemberId_ReusesExistingId(t *testing.T) {
	SetUUIDGenerator(func() string { return "first" })
	ctx := &ClientContext{}
	require.Equal(t, "first", ctx.GetOrCreateMemberId("foo"))

	SetUUIDGenerator(func() string { return "second" })
	require.Equal(t, "first", ctx.GetOrCreateMemberId("foo"))

	memberId, ok := ctx.MemberId("foo")
	require.True(t, ok)
	require.Equal(t, "first", memberId)
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
			ctx.AddGroup("group"+strconv.Itoa(i), "member")
			_ = ctx.GetOrCreateMemberId("other")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = ctx.Heartbeat()
			_ = ctx.ClientId()
			_ = ctx.Groups()
			_ = ctx.HasGroup("group1")
		}
	}()
	wg.Wait()

	require.Equal(t, "foo", ctx.ClientId())
	require.False(t, ctx.Heartbeat().IsZero())
}
