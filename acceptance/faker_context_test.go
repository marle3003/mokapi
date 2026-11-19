package acceptance

import (
	"encoding/json"
	"fmt"
	"io"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type FakerContextSuite struct{ BaseSuite }

func (suite *FakerContextSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./faker-context"}}
	suite.initCmd(cfg)
}

// TestEmailMatchesTheNameInAResponseWithNestedFriends verifies that the user and each friend
// get an email built from their own name.
func (suite *FakerContextSuite) TestEmailMatchesTheNameInAResponseWithNestedFriends() {
	try.GetRequest(suite.T(), "http://localhost:18100/users/7000001", nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var user map[string]any
			require.NoError(t, json.Unmarshal([]byte(body), &user), body)

			require.Equal(t, float64(7000001), user["id"], body)
			requireEmailMatchesName(t, user)

			friends, ok := user["friends"].([]any)
			require.True(t, ok, body)
			require.Len(t, friends, 3, body)

			names := map[string]bool{fullName(user): true}
			for _, f := range friends {
				friend := f.(map[string]any)
				requireEmailMatchesName(t, friend)
				names[fullName(friend)] = true
			}
			require.Len(t, names, 4, "user and friends should have different names: %v", body)

			// the path id describes the requested user, not the friends
			for _, f := range friends {
				require.NotEqual(t, float64(7000001), f.(map[string]any)["id"], body)
			}
		}))
}

func fullName(person map[string]any) string {
	return fmt.Sprintf("%v %v", person["firstName"], person["lastName"])
}

func requireEmailMatchesName(t *testing.T, person map[string]any) {
	t.Helper()
	prefix := strings.ToLower(fmt.Sprintf("%v.%v@", person["firstName"], person["lastName"]))
	require.True(t, strings.HasPrefix(strings.ToLower(person["email"].(string)), prefix), "%v", person)
}

// TestPathIdIsOnlyEchoedIntoTheWrappedUser verifies that the path id ends up in data.id and not in
// the friends of the user.
func (suite *FakerContextSuite) TestPathIdIsOnlyEchoedIntoTheWrappedUser() {
	getJson(suite, "/wrapped/users/7000001", nil, func(t *testing.T, body string, v map[string]any) {
		data := v["data"].(map[string]any)
		require.Equal(t, float64(7000001), data["id"], body)
		for _, f := range data["friends"].([]any) {
			require.NotEqual(t, float64(7000001), f.(map[string]any)["id"], body)
		}
	})
}

// TestQueryFilterIsEchoedIntoEveryItemOfARootArray verifies that a filter parameter reaches all
// items of an array response.
func (suite *FakerContextSuite) TestQueryFilterIsEchoedIntoEveryItemOfARootArray() {
	try.GetRequest(suite.T(), "http://localhost:18100/pets?status=available", nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var pets []map[string]any
			require.NoError(t, json.Unmarshal([]byte(body), &pets), body)
			require.Len(t, pets, 3, body)
			for _, pet := range pets {
				require.Equal(t, "available", pet["status"], body)
			}
		}))
}

// TestQueryFilterIsEchoedIntoEveryItemOfAWrapper verifies that a filter parameter reaches the
// items of a list wrapped in an object.
func (suite *FakerContextSuite) TestQueryFilterIsEchoedIntoEveryItemOfAWrapper() {
	getJson(suite, "/orders?status=active", nil, func(t *testing.T, body string, v map[string]any) {
		items := v["items"].([]any)
		require.Len(t, items, 3, body)
		for _, item := range items {
			require.Equal(t, "active", item.(map[string]any)["status"], body)
		}
	})
}

// TestPathIdIsEchoedIntoSiblingObjects verifies that properties with the same name at the same
// depth all get the parameter.
func (suite *FakerContextSuite) TestPathIdIsEchoedIntoSiblingObjects() {
	getJson(suite, "/accounts/7000001", nil, func(t *testing.T, body string, v map[string]any) {
		require.Equal(t, float64(7000001), v["billing"].(map[string]any)["id"], body)
		require.Equal(t, float64(7000001), v["shipping"].(map[string]any)["id"], body)
	})
}

// TestParameterThatDoesNotMatchTheRootTypeIsNotEchoedDeeper verifies that a string parameter is
// neither used for the integer root id nor moved to the nested string id.
func (suite *FakerContextSuite) TestParameterThatDoesNotMatchTheRootTypeIsNotEchoedDeeper() {
	getJson(suite, "/mismatch/abc", nil, func(t *testing.T, body string, v map[string]any) {
		require.IsType(t, float64(0), v["id"], body)
		for _, f := range v["friends"].([]any) {
			require.NotEqual(t, "abc", f.(map[string]any)["id"], body)
		}
	})
}

// TestHeaderAndCookieAreOnlyEchoedIntoTheRootObject verifies that header and cookie parameters
// follow the same rule as path and query parameters.
func (suite *FakerContextSuite) TestHeaderAndCookieAreOnlyEchoedIntoTheRootObject() {
	headers := map[string]string{"tenant": "acme", "Cookie": "session=s-42"}
	getJson(suite, "/tenants", headers, func(t *testing.T, body string, v map[string]any) {
		require.Equal(t, "acme", v["tenant"], body)
		require.Equal(t, "s-42", v["session"], body)
		child := v["child"].(map[string]any)
		require.NotEqual(t, "acme", child["tenant"], body)
		require.NotEqual(t, "s-42", child["session"], body)
	})
}

// TestScriptReadsPathIdInNestedObjects verifies that a custom node still sees the path id in
// context.values inside friends[].
func (suite *FakerContextSuite) TestScriptReadsPathIdInNestedObjects() {
	suite.waitForScriptNode("ownerId")
	getJson(suite, "/owners/7000001", nil, func(t *testing.T, body string, v map[string]any) {
		friends := v["friends"].([]any)
		require.Len(t, friends, 3, body)
		for _, f := range friends {
			require.Equal(t, float64(7000001), f.(map[string]any)["ownerId"], body)
		}
	})
}

// waitForScriptNode waits until a script has added the node. Scripts are loaded asynchronously.
func (suite *FakerContextSuite) waitForScriptNode(name string) {
	suite.T().Helper()
	try.WaitFor(suite.T(), func() bool {
		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%v/api/faker/tree", suite.cfg.Api.Port))
		if err != nil {
			return false
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return strings.Contains(string(b), fmt.Sprintf(`"name":"%v"`, name))
	})
}

func getJson(suite *FakerContextSuite, path string, headers map[string]string, assert func(t *testing.T, body string, v map[string]any)) {
	try.GetRequest(suite.T(), "http://localhost:18100"+path, headers,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var v map[string]any
			require.NoError(t, json.Unmarshal([]byte(body), &v), body)
			assert(t, body, v)
		}))
}
