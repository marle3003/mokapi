package types

import (
	_ "embed"
)

//go:embed index.d.ts
var index string

//go:embed global.d.ts
var global string

var Mokapi = global + "\n" + index

//go:embed faker.d.ts
var Faker string

//go:embed http.d.ts
var Http string

//go:embed kafka.d.ts
var Kafka string

//go:embed mustache.d.ts
var Mustache string

//go:embed yaml.d.ts
var Yaml string
