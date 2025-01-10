package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"mokapi/try"
	"os"
	"testing"
	"time"
)

func TestConfig_v31(t *testing.T) {
	reader := &dynamictest.Reader{
		Data: map[string]*dynamic.Config{},
	}

	root := readJsonTestFile(t, "./test/petstore31.json", &openapi.Config{}, "https://petstore31.swagger.io/api/v31/openapi.json")
	reader.Data[root.Info.Url.String()] = root

	petDetails := readJsonTestFile(t, "./test/petdetails.json", &schema.Ref{}, "https://petstore31.swagger.io/api/v31/components/schemas/petdetails")
	reader.Data[petDetails.Info.Url.String()] = petDetails
	category := readJsonTestFile(t, "./test/category.json", &schema.Ref{}, "https://petstore31.swagger.io/api/v31/components/schemas/category")
	reader.Data[category.Info.Url.String()] = category
	tag := readJsonTestFile(t, "./test/tag.json", &schema.Ref{}, "https://petstore31.swagger.io/api/v31/components/schemas/tag")
	reader.Data[tag.Info.Url.String()] = tag

	c := root.Data.(*openapi.Config)
	err := c.Parse(root, reader)
	require.NoError(t, err)
}

func readJsonTestFile(t *testing.T, path string, data interface{}, url string) *dynamic.Config {
	b, err := os.ReadFile(path)
	require.NoError(t, err)

	return &dynamic.Config{
		Info: dynamic.ConfigInfo{
			Provider: "http",
			Url:      try.MustUrl(url),
			Checksum: nil,
			Time:     time.Time{},
		},
		Raw:   b,
		Data:  data,
		Scope: *dynamic.NewScope(url),
	}
}
