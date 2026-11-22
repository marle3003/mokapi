package runtime_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/directory"
	"mokapi/providers/mail"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/version"
	"net/url"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const petstoreV1 = `openapi: 3.0.0
info:
  title: pets
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          description: list of pets
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string
`

const petstoreV2 = `openapi: 3.0.0
info:
  title: pets
  version: "2.0"
paths:
  /animals:
    post:
      responses:
        "201":
          description: created
components:
  schemas:
    Pet:
      type: object
      properties:
        age:
          type: integer
`

// reparseCase describes one protocol: a file in two versions and how its store adds the file.
type reparseCase struct {
	name string
	// header and config type as Mokapi registers them at startup (pkg/cmd/mokapi)
	header     string
	configType any
	v1, v2     string
	add        func(app *runtime.App, file *dynamic.Config)
}

const asyncApiHead = "asyncapi: 3.0.0\ninfo:\n  title: foo\n  version: '%s'\nservers:\n  dev:\n    host: ':1234'\n    protocol: %s\n"

func asyncApi3(version, protocol string) string {
	return fmt.Sprintf(asyncApiHead, version, protocol)
}

var reparseCases = []reparseCase{
	{
		name: "http", header: "openapi", configType: &openapi.Config{},
		v1: petstoreV1, v2: petstoreV2,
		add: func(app *runtime.App, file *dynamic.Config) { app.Http.Add(file) },
	},
	{
		name: "kafka", header: "asyncapi", configType: &asyncapi3.Config{},
		v1: asyncApi3("1.0", "kafka"), v2: asyncApi3("2.0", "kafka") + "channels:\n  bar:\n    address: bar\n",
		add: func(app *runtime.App, file *dynamic.Config) { _, _ = app.Kafka.Add(file, enginetest.NewEngine()) },
	},
	{
		name: "mqtt", header: "asyncapi", configType: &asyncapi3.Config{},
		v1: asyncApi3("1.0", "mqtt"), v2: asyncApi3("2.0", "mqtt") + "channels:\n  bar:\n    address: bar\n",
		add: func(app *runtime.App, file *dynamic.Config) { _, _ = app.Mqtt.Add(file, enginetest.NewEngine()) },
	},
	{
		name: "websocket", header: "asyncapi", configType: &asyncapi3.Config{},
		v1: asyncApi3("1.0", "ws"), v2: asyncApi3("2.0", "ws") + "channels:\n  bar:\n    address: bar\n",
		add: func(app *runtime.App, file *dynamic.Config) { _, _ = app.Websocket.Add(file, enginetest.NewEngine()) },
	},
	{
		name: "ldap", header: "ldap", configType: &directory.Config{},
		v1:  "ldap: 1.0\ninfo:\n  title: foo\n  version: '1.0'\nentries:\n  - dn: cn=a,dc=example,dc=com\n    cn: a\n",
		v2:  "ldap: 1.0\ninfo:\n  title: foo\n  version: '2.0'\nentries:\n  - dn: cn=b,dc=example,dc=com\n    cn: b\n",
		add: func(app *runtime.App, file *dynamic.Config) { app.Ldap.Add(file, enginetest.NewEngine()) },
	},
	{
		name: "mail", header: "mail", configType: &mail.Config{},
		v1:  "mail: 1.0\ninfo:\n  title: foo\n  version: '1.0'\nmailboxes:\n  a@example.com: {}\n",
		v2:  "mail: 1.0\ninfo:\n  title: foo\n  version: '2.0'\nmailboxes:\n  b@example.com: {}\n",
		add: func(app *runtime.App, file *dynamic.Config) { app.Mail.Add(file) },
	},
}

// parsedFile returns a file as the ConfigWatcher holds it: Raw parsed with dynamic.Parse.
func (tc reparseCase) parsedFile(t *testing.T, filename string) *dynamic.Config {
	t.Helper()
	// Mokapi registers its config types at startup, dynamic.Parse needs them.
	registerReparseTypes.Do(func() {
		for _, c := range reparseCases {
			dynamic.Register(c.header, func(v version.Version) bool { return true }, c.configType)
		}
	})
	u, _ := url.Parse("file:///" + filename)
	c := &dynamic.Config{Raw: []byte(tc.v1)}
	c.Info.Url = u
	require.NoError(t, dynamic.Parse(c, &dynamictest.Reader{}))
	return c
}

// reparse does what ConfigWatcher.configChanged does: new Raw, parsed in place.
func reparse(file *dynamic.Config, raw string) error {
	file.Raw = []byte(raw)
	return dynamic.Parse(file, &dynamictest.Reader{})
}

var registerReparseTypes sync.Once

// ConfigWatcher.configChanged reparses a file in place with dynamic.Parse. A snapshot that was
// published before must stay as it was, because handlers and the dashboard still read it.
func TestApp_PublishedHttpSnapshotIsNotChangedByReparseOfFile(t *testing.T) {
	logrus.SetOutput(io.Discard)
	tc := reparseCases[0]
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	file := tc.parsedFile(t, "foo.yml")
	snapshot := app.Http.Add(file).Config()
	before := marshal(t, snapshot)

	require.NoError(t, reparse(file, tc.v2))

	require.JSONEq(t, before, marshal(t, snapshot), "published snapshot must be immutable")
}

// ConfigWatcher runs the store listener (Store.Add) in its own goroutine, while the next
// file change reparses the same file in place. The store reads file.Data() without a lock.
func TestApp_AddDoesNotRaceWithReparseOfFile(t *testing.T) {
	logrus.SetOutput(io.Discard)
	for _, tc := range reparseCases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtime.New(&static.Config{}, &dynamictest.Reader{})
			file := tc.parsedFile(t, "foo.yml")

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				for i := 0; i < 500; i++ {
					tc.add(app, file)
				}
			}()
			go func() {
				defer wg.Done()
				for i := 0; i < 500; i++ {
					raw := tc.v1
					if i%2 == 0 {
						raw = tc.v2
					}
					_ = reparse(file, raw)
				}
			}()
			wg.Wait()
		})
	}
}

// A listener that got file.Data() from an event keeps working with this object. A reparse must
// not change it.
func TestApp_ParsedDataOfFileIsNotChangedByReparse(t *testing.T) {
	for _, tc := range reparseCases {
		t.Run(tc.name, func(t *testing.T) {
			file := tc.parsedFile(t, "foo.yml")
			held := file.Data()
			before := marshal(t, held)

			require.NoError(t, reparse(file, tc.v2))

			require.JSONEq(t, before, marshal(t, held), "object held by a reader must stay as it was")
		})
	}
}

// A store with several files merges them into one config, while ConfigWatcher reparses one of
// the files in place. The merge reads the data of every file (publishMerged runs outside the
// store lock), so it must work on a consistent object of each file.
func TestApp_MergeOfFilesDoesNotRaceWithReparseOfFile(t *testing.T) {
	logrus.SetOutput(io.Discard)
	for _, tc := range reparseCases {
		if tc.header != "asyncapi" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			app := runtime.New(&static.Config{}, &dynamictest.Reader{})
			a := tc.parsedFile(t, "a.yml")
			b := tc.parsedFile(t, "b.yml")

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				for i := 0; i < 500; i++ {
					tc.add(app, a)
					tc.add(app, b)
				}
			}()
			go func() {
				defer wg.Done()
				for i := 0; i < 500; i++ {
					raw := tc.v1
					if i%2 == 0 {
						raw = tc.v2
					}
					_ = reparse(b, raw)
				}
			}()
			wg.Wait()
		})
	}
}
