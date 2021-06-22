package openapi

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/test"
	"testing"
)

type testDataEntry struct {
	name    string
	content string
	test    func(t *testing.T, config *Config)
	reader  dynamic.ConfigReader
}

type testReader struct {
	data map[string]string
}

func TestConfig(t *testing.T) {
	run(t, testLocalData)
	run(t, testExternalData)
}

func run(t *testing.T, data []testDataEntry) {
	for _, td := range data {
		t.Logf("testcase %q", td.name)

		var c *Config
		err := yaml.Unmarshal([]byte(td.content), &c)
		test.Ok(t, err)

		r := refResolver{reader: td.reader, path: "", config: c, eh: dynamic.NewEmptyEventHandler(c)}
		err = r.resolveConfig()
		test.Ok(t, err)

		td.test(t, c)
	}
}

func (tr *testReader) Read(path string, c dynamic.Config, h dynamic.ChangeEventHandler) error {
	if s, ok := tr.data[path]; ok {
		err := yaml.Unmarshal([]byte(s), c)
		return err
	}
	return fmt.Errorf("path %q not found", path)
}
