package directory_test

import (
	"encoding/json"
	"mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/providers/directory"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"testing"

	"github.com/stretchr/testify/require"
)

type eventLog interface {
	SetDuration(int64)
	SetActions([]*common.Action)
}

// finished sets duration and actions the way the handler does after the log has been pushed.
func finished[T eventLog](l T) T {
	l.SetDuration(12)
	l.SetActions([]*common.Action{{Duration: 2}})
	return l
}

func TestLog_MarshalJSON(t *testing.T) {
	const action = `[{"duration":2,"tags":null,"parameters":null,"logs":null,"error":null}]`
	response := &directory.Response{Status: "Success", MatchedDn: "cn=foo", Message: "done"}
	const responseJson = `{"status":"Success","matchedDn":"cn=foo","message":"done"}`

	testcases := []struct {
		name string
		log  any
		json string
	}{
		{
			name: "search",
			log: finished(&directory.SearchLog{
				Request: &directory.SearchRequest{Operation: "Search", BaseDN: "dc=foo", Scope: "WholeSubtree", Filter: "(cn=*)"},
				Response: &directory.SearchResponse{Status: "Success", Results: []directory.SearchResult{
					{Dn: "cn=foo", Attributes: map[string][]string{"mail": {"foo@bar"}}},
				}},
			}),
			json: `{"request":{"operation":"Search","baseDN":"dc=foo","scope":"WholeSubtree","dereferencePolicy":0,"sizeLimit":0,"timeLimit":0,"typesOnly":false,"filter":"(cn=*)","attributes":null,"controls":null},"response":{"status":"Success","results":[{"dn":"cn=foo","attributes":{"mail":["foo@bar"]}}]},"duration":12,"actions":` + action + `}`,
		},
		{
			name: "search without results",
			log: &directory.SearchLog{
				Request:  &directory.SearchRequest{Operation: "Search"},
				Response: &directory.SearchResponse{},
			},
			json: `{"request":{"operation":"Search","baseDN":"","scope":"","dereferencePolicy":0,"sizeLimit":0,"timeLimit":0,"typesOnly":false,"filter":"","attributes":null,"controls":null},"response":{"status":"","results":null},"duration":0,"actions":null}`,
		},
		{
			name: "bind",
			log: finished(&directory.BindLog{
				Request:  &directory.BindRequest{Operation: "Bind", Version: 3, Name: "cn=foo", Password: "secret", Auth: "Simple"},
				Response: response,
			}),
			json: `{"request":{"operation":"Bind","version":3,"name":"cn=foo","password":"secret","auth":"Simple"},"response":` + responseJson + `,"duration":12,"actions":` + action + `}`,
		},
		{
			name: "unbind",
			log: finished(&directory.UnbindLog{
				Request: &directory.UnbindRequest{Operation: "Unbind"},
			}),
			json: `{"request":{"operation":"Unbind"},"duration":12,"actions":` + action + `}`,
		},
		{
			name: "add",
			log: finished(&directory.AddLog{
				Request:  &directory.AddRequest{Operation: "Add", Dn: "cn=foo", Attributes: []directory.Attribute{{Type: "mail", Values: []string{"foo@bar"}}}},
				Response: response,
			}),
			json: `{"request":{"operation":"Add","dn":"cn=foo","attributes":[{"type":"mail","values":["foo@bar"]}]},"response":` + responseJson + `,"duration":12,"actions":` + action + `}`,
		},
		{
			name: "modify",
			log: finished(&directory.ModifyLog{
				Request: &directory.ModifyRequest{Operation: "Modify", Dn: "cn=foo", Items: []directory.Modify{
					{Modification: "Replace", Attribute: &directory.Attribute{Type: "mail", Values: []string{"foo@bar"}}},
				}},
				Response: response,
			}),
			json: `{"request":{"operation":"Modify","dn":"cn=foo","items":[{"modification":"Replace","attribute":{"type":"mail","values":["foo@bar"]}}]},"response":` + responseJson + `,"duration":12,"actions":` + action + `}`,
		},
		{
			name: "delete",
			log: finished(&directory.DeleteLog{
				Request:  &directory.DeleteRequest{Operation: "Delete", Dn: "cn=foo"},
				Response: response,
			}),
			json: `{"request":{"operation":"Delete","dn":"cn=foo"},"response":` + responseJson + `,"duration":12,"actions":` + action + `}`,
		},
		{
			name: "modifydn",
			log: finished(&directory.ModifyDNLog{
				Request:  &directory.ModifyDNRequest{Operation: "ModifyDN", Dn: "cn=foo", NewRdn: "cn=bar", DeleteOldDn: true, NewSuperiorDn: "dc=x"},
				Response: response,
			}),
			json: `{"request":{"operation":"ModifyDN","dn":"cn=foo","newRdn":"cn=bar","deleteOldDn":true,"newSuperiorDn":"dc=x"},"response":` + responseJson + `,"duration":12,"actions":` + action + `}`,
		},
		{
			name: "compare",
			log: finished(&directory.CompareLog{
				Request:  &directory.CompareRequest{Operation: "Compare", Dn: "cn=foo", Attribute: "mail", Value: "foo@bar"},
				Response: response,
			}),
			json: `{"request":{"operation":"Compare","dn":"cn=foo","attribute":"mail","value":"foo@bar"},"response":` + responseJson + `,"duration":12,"actions":` + action + `}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.log)
			require.NoError(t, err)
			require.Equal(t, tc.json, string(b))
		})
	}
}

func TestSearchLog_AddResultAndSetStatus(t *testing.T) {
	l := directory.NewSearchLogEvent(&ldap.SearchRequest{BaseDN: "dc=foo"}, &eventstest.Handler{}, events.NewTraits())

	l.AddResult(directory.SearchResult{Dn: "cn=foo"})
	l.SetStatus("Success")

	b, err := json.Marshal(l)
	require.NoError(t, err)
	require.Contains(t, string(b), `"response":{"status":"Success","results":[{"dn":"cn=foo","attributes":null}]}`)
}

func TestSearchLog_IndexFieldsDoesNotAliasResults(t *testing.T) {
	l := directory.NewSearchLogEvent(&ldap.SearchRequest{}, &eventstest.Handler{}, events.NewTraits())
	l.AddResult(directory.SearchResult{Dn: "cn=foo"})

	fields := l.IndexFields()
	l.AddResult(directory.SearchResult{Dn: "cn=bar"})
	l.SetStatus("Success")

	response := fields["response"].(*directory.SearchResponse)
	require.Equal(t, "", response.Status)
	require.Len(t, response.Results, 1)
	require.Equal(t, "cn=foo", response.Results[0].Dn)
}
