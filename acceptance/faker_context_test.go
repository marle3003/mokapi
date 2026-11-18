package acceptance

import (
	"encoding/json"
	"fmt"
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
	try.GetRequest(suite.T(), "http://localhost:18100/users/42", nil,
		try.HasStatusCode(http.StatusOK),
		try.AssertBody(func(t *testing.T, body string) {
			var user map[string]any
			require.NoError(t, json.Unmarshal([]byte(body), &user), body)

			require.Equal(t, float64(42), user["id"], body)
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
