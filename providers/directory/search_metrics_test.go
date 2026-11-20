package directory_test

import (
	"context"
	"io"
	"mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/providers/directory"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// statusEmitter simulates a JavaScript ldap event handler that sets the response status.
type statusEmitter struct {
	status *uint8
}

func (e *statusEmitter) EmitLdap(_ *ldap.SearchRequest, res *ldap.SearchResponse) []*common.Action {
	if e.status != nil {
		res.Status = *e.status
	}
	return nil
}

// userEntries returns two entries that match the filter (cn=user*)
func userEntries() *directory.EntryStore {
	return convertArray([]directory.Entry{
		{Dn: "cn=user1,dc=foo,dc=com", Attributes: map[string][]string{"cn": {"user1"}}},
		{Dn: "cn=user2,dc=foo,dc=com", Attributes: map[string][]string{"cn": {"user2"}}},
	})
}

func TestSearch_ErrorMetric(t *testing.T) {
	logrus.SetOutput(io.Discard)

	noSuchObject := uint8(ldap.NoSuchObject)
	testcases := []struct {
		name      string
		filter    string
		sizeLimit int64
		emitter   *statusEmitter
		errors    float64
	}{
		{name: "successful search is not an error", filter: "(cn=user*)", emitter: &statusEmitter{}, errors: 0},
		{name: "invalid filter", filter: "cn=user", emitter: &statusEmitter{}, errors: 1},
		{name: "unsupported filter", filter: "(cn=user))", emitter: &statusEmitter{}, errors: 1},
		{name: "server size limit exceeded", filter: "(cn=user*)", sizeLimit: 1, emitter: &statusEmitter{}, errors: 1},
		{name: "js handler sets noSuchObject", filter: "(cn=user*)", emitter: &statusEmitter{status: &noSuchObject}, errors: 1},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &directory.Config{
				Info:      directory.Info{Name: "foo"},
				SizeLimit: tc.sizeLimit,
				Entries:   userEntries(),
			}
			m := monitor.NewLdap()
			ctx := monitor.NewLdapContext(context.Background(), m)

			h := directory.NewHandler(cfg, tc.emitter, &eventstest.Handler{})
			h.ServeLDAP(ldaptest.NewRecorder(), ldaptest.NewRequestWithContext(1, &ldap.SearchRequest{
				Scope:  ldap.ScopeWholeSubtree,
				Filter: tc.filter,
			}, ctx))

			require.Equal(t, tc.errors, m.Errors.Sum(metrics.NewQuery()))
		})
	}
}

func TestSearch_UnknownScope(t *testing.T) {
	logrus.SetOutput(io.Discard)

	cfg := &directory.Config{Info: directory.Info{Name: "foo"}, Entries: userEntries()}
	m := monitor.NewLdap()
	ctx := monitor.NewLdapContext(context.Background(), m)
	eh := &eventstest.Handler{}

	h := directory.NewHandler(cfg, &statusEmitter{}, eh)
	rr := ldaptest.NewRecorder()
	h.ServeLDAP(rr, ldaptest.NewRequestWithContext(1, &ldap.SearchRequest{
		Scope:  3,
		Filter: "(cn=user*)",
	}, ctx))

	res := rr.Message.(*ldap.SearchResponse)
	require.Equal(t, ldap.ProtocolError, res.Status)
	require.Equal(t, "unsupported search scope", res.Message)
	require.Equal(t, 1.0, m.Errors.Sum(metrics.NewQuery()))
	require.Equal(t, 1.0, m.RequestCounter.Sum(metrics.NewQuery()))

	logs := eh.Events
	require.Len(t, logs, 1)
	require.Equal(t, "ProtocolError", logs[0].Data.(*directory.SearchLog).Response.Status)
}

func TestSearch_FilterError(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name   string
		filter string
	}{
		{name: "invalid filter", filter: "cn=user"},
		{name: "unsupported filter", filter: "(cn=user))"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &directory.Config{Info: directory.Info{Name: "foo"}, Entries: userEntries()}
			eh := &eventstest.Handler{}

			h := directory.NewHandler(cfg, &statusEmitter{}, eh)
			rr := ldaptest.NewRecorder()
			h.ServeLDAP(rr, ldaptest.NewRequestWithContext(1, &ldap.SearchRequest{
				Scope:  ldap.ScopeWholeSubtree,
				Filter: tc.filter,
			}, context.Background()))

			res := rr.Message.(*ldap.SearchResponse)
			require.Equal(t, uint8(ldap.OperationsError), res.Status)

			logs := eh.Events
			require.Len(t, logs, 1)
			require.Equal(t, ldap.StatusText[ldap.OperationsError], logs[0].Data.(*directory.SearchLog).Response.Status)
		})
	}
}

func TestSearch_ErrorMetric_ClientLimits(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name    string
		request func() *ldap.SearchRequest
	}{
		{
			name: "client size limit",
			request: func() *ldap.SearchRequest {
				return &ldap.SearchRequest{Scope: ldap.ScopeWholeSubtree, Filter: "(cn=user*)", SizeLimit: 1}
			},
		},
		{
			name: "paged search",
			request: func() *ldap.SearchRequest {
				return &ldap.SearchRequest{
					Scope:    ldap.ScopeWholeSubtree,
					Filter:   "(cn=user*)",
					Controls: []ldap.Control{&ldap.PagedResultsControl{PageSize: 1}},
				}
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &directory.Config{
				Info:    directory.Info{Name: "foo"},
				Entries: userEntries(),
			}
			m := monitor.NewLdap()
			ctx := ldap.NewPagingFromContext(monitor.NewLdapContext(context.Background(), m))
			rec := ldaptest.NewRecorder()

			h := directory.NewHandler(cfg, &statusEmitter{}, &eventstest.Handler{})
			h.ServeLDAP(rec, ldaptest.NewRequestWithContext(1, tc.request(), ctx))

			res := rec.Message.(*ldap.SearchResponse)
			require.Equal(t, uint8(ldap.Success), res.Status)
			require.Len(t, res.Results, 1)
			require.Equal(t, float64(0), m.Errors.Sum(metrics.NewQuery()))
		})
	}
}

func TestSearch_ErrorMetric_CountedPerService(t *testing.T) {
	logrus.SetOutput(io.Discard)

	m := monitor.NewLdap()
	ctx := monitor.NewLdapContext(context.Background(), m)
	for _, name := range []string{"foo", "foo", "bar"} {
		h := directory.NewHandler(&directory.Config{Info: directory.Info{Name: name}}, &statusEmitter{}, &eventstest.Handler{})
		h.ServeLDAP(ldaptest.NewRecorder(), ldaptest.NewRequestWithContext(1, &ldap.SearchRequest{
			Scope:  ldap.ScopeWholeSubtree,
			Filter: "cn=user",
		}, ctx))
	}

	require.Equal(t, float64(2), m.Errors.Sum(metrics.NewQuery(metrics.ByLabel("service", "foo"))))
	require.Equal(t, float64(1), m.Errors.Sum(metrics.NewQuery(metrics.ByLabel("service", "bar"))))
}

func TestSearch_ErrorMetric_WithoutMonitor(t *testing.T) {
	logrus.SetOutput(io.Discard)

	h := directory.NewHandler(&directory.Config{Info: directory.Info{Name: "foo"}}, &statusEmitter{}, &eventstest.Handler{})
	rec := ldaptest.NewRecorder()

	require.NotPanics(t, func() {
		h.ServeLDAP(rec, ldaptest.NewRequestWithContext(1, &ldap.SearchRequest{
			Scope:  ldap.ScopeWholeSubtree,
			Filter: "cn=user",
		}, context.Background()))
	})
	require.Equal(t, uint8(ldap.OperationsError), rec.Message.(*ldap.SearchResponse).Status)

	require.NotPanics(t, func() {
		h.ServeLDAP(rec, ldaptest.NewRequestWithContext(2, &ldap.SearchRequest{
			Scope:  ldap.ScopeWholeSubtree,
			Filter: "(cn=user*)",
		}, context.Background()))
	})
	require.Equal(t, uint8(ldap.Success), rec.Message.(*ldap.SearchResponse).Status)
}
