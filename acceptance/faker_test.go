package acceptance

import (
	"encoding/json"
	"fmt"
	"io"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"net/url"
	"strings"
)

type fakerApiNode struct {
	Name        string         `json:"name"`
	Path        string         `json:"path"`
	Custom      bool           `json:"custom"`
	HasFakeFunc bool           `json:"hasFakeFunc"`
	Children    []fakerApiNode `json:"children"`
}

type fakerApiMatch struct {
	PropertyPath string   `json:"propertyPath"`
	Node         string   `json:"node"`
	Tokens       []string `json:"tokens"`
	Source       string   `json:"source"`
	Value        any      `json:"value"`
}

type fakerApiResult struct {
	Value   any             `json:"value"`
	Matches []fakerApiMatch `json:"matches"`
}

type FakerSuite struct{ BaseSuite }

func (suite *FakerSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./faker"}}
	suite.initCmd(cfg)
}

func (suite *FakerSuite) apiUrl(path string) string {
	return fmt.Sprintf("http://127.0.0.1:%v%v", suite.cfg.Api.Port, path)
}

func (suite *FakerSuite) getTree() fakerApiNode {
	res, err := http.Get(suite.apiUrl("/api/faker/tree"))
	suite.Require().NoError(err)
	defer res.Body.Close()
	suite.Require().Equal(http.StatusOK, res.StatusCode)
	var root fakerApiNode
	suite.Require().NoError(json.NewDecoder(res.Body).Decode(&root))
	return root
}

// waitForCustomNodes waits until the script has added its nodes. Scripts are loaded asynchronously.
func (suite *FakerSuite) waitForCustomNodes() fakerApiNode {
	suite.T().Helper()
	var root fakerApiNode
	try.WaitFor(suite.T(), func() bool {
		res, err := http.Get(suite.apiUrl("/api/faker/tree"))
		if err != nil {
			return false
		}
		defer res.Body.Close()
		if json.NewDecoder(res.Body).Decode(&root) != nil {
			return false
		}
		for _, c := range root.Children {
			if c.Name == "shipping" {
				return true
			}
		}
		return false
	})
	return root
}

func (suite *FakerSuite) explain(body string) (int, string) {
	res, err := http.Post(suite.apiUrl("/api/faker/fake"), "application/json", strings.NewReader(body))
	suite.Require().NoError(err)
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	suite.Require().NoError(err)
	return res.StatusCode, string(b)
}

func (suite *FakerSuite) mustExplain(body string) fakerApiResult {
	suite.T().Helper()
	suite.waitForCustomNodes()
	status, b := suite.explain(body)
	suite.Require().Equal(http.StatusOK, status, b)
	var result fakerApiResult
	suite.Require().NoError(json.Unmarshal([]byte(b), &result))
	return result
}

func (suite *FakerSuite) generateNode(path string) (int, string) {
	res, err := http.Get(suite.apiUrl("/api/faker/node/fake?path=" + url.QueryEscape(path)))
	suite.Require().NoError(err)
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	suite.Require().NoError(err)
	return res.StatusCode, string(b)
}

func (suite *FakerSuite) TestCustomNodeIsListedInTheTreeAsCustom() {
	root := suite.waitForCustomNodes()

	suite.Equal("root", root.Name)
	suite.Equal("/", root.Path)
	var billing, frequency *fakerApiNode
	for i, c := range root.Children {
		switch c.Name {
		case "billing":
			billing = &root.Children[i]
		case "frequency":
			frequency = &root.Children[i]
		}
	}
	suite.Require().NotNil(frequency)
	suite.True(frequency.Custom)
	suite.True(frequency.HasFakeFunc)
	suite.Equal("/frequency", frequency.Path)

	suite.Require().NotNil(billing)
	suite.True(billing.Custom)
	suite.False(billing.HasFakeFunc, "a grouping node has no fake function")
	suite.Require().Len(billing.Children, 1)
	suite.Equal("/billing/code", billing.Children[0].Path)
}

func (suite *FakerSuite) TestPropertyPathReportsCustomNodeTokensAndValue() {
	result := suite.mustExplain(`{"path": ["customerFrequency"]}`)

	suite.Equal("weekly", result.Value)
	suite.Equal([]fakerApiMatch{{
		PropertyPath: "customerFrequency",
		Node:         "/frequency",
		Tokens:       []string{"frequency"},
		Source:       "node",
		Value:        "weekly",
	}}, result.Matches)
}

func (suite *FakerSuite) TestObjectSchemaReportsOneRowPerNestedProperty() {
	result := suite.mustExplain(`{
		"schema": {
			"type": "object",
			"required": ["frequency", "address"],
			"properties": {
				"frequency": {},
				"address": {
					"type": "object",
					"required": ["city"],
					"properties": { "city": { "type": "string" } }
				}
			}
		}
	}`)

	suite.Require().IsType(map[string]any{}, result.Value)
	suite.Require().Len(result.Matches, 2)
	suite.Equal("frequency", result.Matches[0].PropertyPath)
	suite.Equal("/frequency", result.Matches[0].Node)
	suite.Equal("weekly", result.Matches[0].Value)
	suite.Equal("address.city", result.Matches[1].PropertyPath)
	suite.NotEmpty(result.Matches[1].Node)
	suite.Equal([]string{"city"}, result.Matches[1].Tokens)
	suite.Equal(result.Value.(map[string]any)["address"].(map[string]any)["city"], result.Matches[1].Value)
}

func (suite *FakerSuite) TestSchemaWithoutMatchingNodeReportsFakeBySchema() {
	result := suite.mustExplain(`{"schema": {"type": "integer", "minimum": 1, "maximum": 3}}`)

	suite.Require().Len(result.Matches, 1)
	suite.Equal("(value)", result.Matches[0].PropertyPath)
	suite.Equal("fakeBySchema", result.Matches[0].Node)
	suite.Equal(result.Value, result.Matches[0].Value)
}

func (suite *FakerSuite) TestEnumPropertyIsNotAttributedToANode() {
	result := suite.mustExplain(`{
		"schema": {
			"type": "object",
			"required": ["status"],
			"properties": { "status": { "enum": ["open", "closed"] } }
		}
	}`)

	suite.Require().Len(result.Matches, 1)
	suite.Equal("status", result.Matches[0].PropertyPath)
	suite.Equal("-", result.Matches[0].Node)
	suite.Equal("enum", result.Matches[0].Source)
}

func (suite *FakerSuite) TestInvalidSchemaIsRejectedWithMessage() {
	suite.waitForCustomNodes()

	status, body := suite.explain(`{"schema": {"type": 123}}`)

	suite.Equal(http.StatusBadRequest, status)
	suite.Equal("schema error at field 'type': expected type string or array, got number", body)
}

func (suite *FakerSuite) TestGenerateByPathTellsNodesWithTheSameNameApart() {
	suite.waitForCustomNodes()

	tests := map[string]string{
		"/billing/code":  `"BILL-1"`,
		"/shipping/code": `"SHIP-1"`,
	}
	for path, expected := range tests {
		status, body := suite.generateNode(path)

		suite.Equal(http.StatusOK, status, path)
		suite.Equal(expected, body, path)
	}
}

func (suite *FakerSuite) TestGenerateByPathRejectsUnknownAndNonGeneratingNodes() {
	suite.waitForCustomNodes()

	tests := []struct {
		path   string
		status int
	}{
		{"/billing/unknown", http.StatusNotFound},
		{"/", http.StatusBadRequest},
		{"/billing", http.StatusBadRequest},
	}
	for _, tc := range tests {
		status, _ := suite.generateNode(tc.path)

		suite.Equal(tc.status, status, tc.path)
	}
}
