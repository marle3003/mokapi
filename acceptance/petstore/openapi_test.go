package petstore

import (
	"github.com/brianvoe/gofakeit/v6"
	"gopkg.in/yaml.v3"
	"mokapi/acceptance/cmd"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/static"
	"mokapi/server/web/webtest"
	"mokapi/test"
	"os"
	"testing"
)

func TestOpenAPI(t *testing.T) {
	gofakeit.Seed(11)
	cfg := static.NewConfig()
	cfg.Providers.File.Filename = "./openapi.yml"
	cmd, err := cmd.Start(cfg)
	test.Ok(t, err)
	defer cmd.Stop()

	petstore := &openapi.Config{}
	b, err := os.ReadFile("./openapi.yml")
	test.Ok(t, err)
	err = yaml.Unmarshal(b, &petstore)
	test.Ok(t, err)
	err = petstore.Parse(&common.File{Data: petstore}, nil)
	test.Ok(t, err)

	pet := petstore.Components.Schemas.Get("Pet")
	err = webtest.GetRequest("http://127.0.0.1/pet/1",
		map[string]string{"Accept": "application/json"},
		webtest.HasStatusCode(200),
		webtest.HasBody(
			`{"Id":-8379641344161477543,"Category":{"Id":-1799207740735652432,"Name":"RMaRxHkiJBPtapW"},"Name":"doggie","PhotoUrls":[],"Tags":[{"Id":-3430133205295092491,"Name":"nSMKgtlxwnqhqcl"},{"Id":-4360704630090834069,"Name":"YkWwfoRLOPxLIok"},{"Id":-9084870506124948944,"Name":"qanPAKaXSMQFpZy"}],"Status":"pending"}`,
			pet.Value))
	test.Ok(t, err)
}
