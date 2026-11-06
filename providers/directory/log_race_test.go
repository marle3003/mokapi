package directory_test

import (
	"context"
	"encoding/json"
	"io"
	"mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/providers/directory"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type slowEmitter struct{}

func (e *slowEmitter) EmitLdap(_ *ldap.SearchRequest, _ *ldap.SearchResponse) []*common.Action {
	time.Sleep(10 * time.Millisecond)
	return []*common.Action{{Duration: 1}}
}

func newLdapStoreManager(size int) *events.StoreManager {
	sm := events.NewStoreManager(nil)
	sm.SetStore(size, events.NewTraits().WithNamespace("ldap"))
	return sm
}

// The LDAP logs are stored in the event store before the request is handled completely.
// Results, Status, Actions and Duration are written afterwards, while the dashboard API
// may already serialize the event. Run with -race.
func TestDirectory_SearchLogIsReadWhileHandlerIsRunning(t *testing.T) {
	logrus.SetOutput(io.Discard)

	sm := newLdapStoreManager(10)
	emitter := &slowEmitter{}
	h := directory.NewHandler(testConfig, emitter, sm)

	ctx := context.WithValue(context.Background(), "time", time.Now())
	rr := ldaptest.NewRecorder()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logs := sm.GetEvents(events.NewTraits().WithNamespace("ldap"))
		for len(logs) == 0 {
			time.Sleep(10 * time.Microsecond)
			logs = sm.GetEvents(events.NewTraits().WithNamespace("ldap"))
		}
		for i := 0; i < 200; i++ {
			_, err := json.Marshal(logs[0].Data)
			require.NoError(t, err)
			time.Sleep(100 * time.Microsecond)
		}
	}()

	h.ServeLDAP(rr, ldaptest.NewRequestWithContext(1, &ldap.SearchRequest{
		BaseDN: "dc=foo,dc=com",
		Scope:  ldap.ScopeWholeSubtree,
		Filter: "(objectclass=*)",
	}, ctx))
	wg.Wait()

	log := sm.GetEvents(events.NewTraits().WithNamespace("ldap"))[0].Data.(*directory.SearchLog)
	require.Len(t, log.GetActions(), 1)
}

// The Duration of all other operations is written in a deferred function after the log has been pushed.
func TestDirectory_LogDurationIsWrittenAfterPush(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name    string
		request ldap.Message
	}{
		{name: "bind", request: &ldap.BindRequest{Version: 3, Name: "cn=foo", Auth: ldap.Simple}},
		{name: "add", request: &ldap.AddRequest{Dn: "cn=new,dc=foo,dc=com"}},
		{name: "modify", request: &ldap.ModifyRequest{Dn: "cn=user1,dc=foo,dc=com"}},
		{name: "delete", request: &ldap.DeleteRequest{Dn: "cn=user2,dc=foo,dc=com"}},
		{name: "modifydn", request: &ldap.ModifyDNRequest{Dn: "cn=user1,dc=foo,dc=com", NewRdn: "cn=other"}},
		{name: "compare", request: &ldap.CompareRequest{Dn: "cn=user1,dc=foo,dc=com", Attribute: "mail", Value: "x"}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &directory.Config{Info: directory.Info{Name: "foo"}, Entries: convert(map[string]directory.Entry{
				"user1": {Dn: "cn=user1,dc=foo,dc=com", Attributes: map[string][]string{"mail": {"user1@foo.bar"}}},
				"user2": {Dn: "cn=user2,dc=foo,dc=com"},
			})}
			sm := newLdapStoreManager(10)
			h := directory.NewHandler(cfg, nil, sm)

			ctx := monitor.NewLdapContext(context.WithValue(context.Background(), "time", time.Now().Add(-time.Second)), monitor.NewLdap())
			rw := &responseReader{ResponseRecorder: ldaptest.NewRecorder(), sm: sm}
			h.ServeLDAP(rw, ldaptest.NewRequestWithContext(1, tc.request, ctx))
			rw.wg.Wait()
			require.True(t, rw.read, "log must be in the event store while the response is written")
		})
	}
}

// Unbind has no response: its Duration is written right after the log has been pushed.
func TestDirectory_UnbindLogDurationIsWrittenAfterPush(t *testing.T) {
	logrus.SetOutput(io.Discard)

	sm := newLdapStoreManager(1000)
	h := directory.NewHandler(testConfig, nil, sm)

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
			}
			for _, e := range sm.GetEvents(events.NewTraits().WithNamespace("ldap")) {
				_, err := json.Marshal(e.Data)
				require.NoError(t, err)
			}
		}
	}()

	for i := 0; i < 500; i++ {
		ctx := monitor.NewLdapContext(context.WithValue(context.Background(), "time", time.Now().Add(-time.Second)), monitor.NewLdap())
		h.Unbind(ctx)
	}
	close(done)
	wg.Wait()
}

// responseReader starts reading the stored log while the response is written,
// i.e. before the deferred Duration assignment.
type responseReader struct {
	*ldaptest.ResponseRecorder
	sm   *events.StoreManager
	wg   sync.WaitGroup
	read bool
}

func (r *responseReader) Write(msg ldap.Message) error {
	logs := r.sm.GetEvents(events.NewTraits().WithNamespace("ldap"))
	if len(logs) == 1 {
		r.read = true
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for i := 0; i < 50; i++ {
				_, _ = json.Marshal(logs[0].Data)
				time.Sleep(100 * time.Microsecond)
			}
		}()
	}
	return r.ResponseRecorder.Write(msg)
}
