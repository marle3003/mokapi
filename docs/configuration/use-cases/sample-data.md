---
title: Generate Test Data with mokapi sample-data
description: Use the mokapi sample-data command to generate random JSON, XML, YAML or Avro data from a JSON Schema, OpenAPI schema or Avro schema, for test fixtures, seed data and CI pipelines.
---
# Generate Test Data with mokapi sample-data

`mokapi sample-data` generates random example data from a schema and prints it to stdout. It doesn't start a server
and doesn't need a running Mokapi instance. You can use it to create fixtures for unit tests, fill a database before
an integration test or check what Mokapi would return for a schema. The generator is the same one Mokapi uses for
mocked responses, see [Generate Realistic Test Data](/docs/get-started/test-data.md).

```bash
mokapi sample-data ./user.json
```

```json
{"id":150,"email":"jschuster4@rose.org","name":"TerraTrove","role":"user","createdAt":"2021-12-27T10:32:16Z"}
```

The schema in this example is:

```json title=user.json
{
  "type": "object",
  "required": ["id", "email", "name", "createdAt"],
  "properties": {
    "id": { "type": "integer", "minimum": 1, "maximum": 1000 },
    "email": { "type": "string", "format": "email" },
    "name": { "type": "string" },
    "role": { "type": "string", "enum": ["admin", "user", "guest"] },
    "createdAt": { "type": "string", "format": "date-time" }
  }
}
```

The values are random, so your output differs. Mokapi picks values that fit the property name and format, which is why
`email` looks like an email address. Optional properties such as `role` are not always present.

## Choose the input

The command takes one argument. Without an argument it reads the schema from stdin. Mokapi detects an inline schema
by checking whether the argument is a JSON or YAML object, otherwise it treats the argument as a file path or URL.

```bash tab=File
mokapi sample-data ./user.json
```
```bash tab=URL
mokapi sample-data https://example.com/schema.yaml
```
```bash tab=Stdin
cat user.json | mokapi sample-data
```
```bash tab=Inline
mokapi sample-data '{"type":"object","properties":{"id":{"type":"integer"}}}'
```

Files and inline schemas can be JSON or YAML. For files, the extension (`.json`, `.yaml`, `.yml`) decides the
format. If there is no extension, Mokapi tries JSON first and then YAML.

### Select the schema type

By default, the input is a JSON Schema. Use `--input-type` for other formats:

| Value        | Input                    |
|--------------|--------------------------|
| `jsonschema` | JSON Schema (default)    |
| `openapi`    | OpenAPI schema object    |
| `avro`       | Avro schema              |

```bash
mokapi sample-data --input-type avro ./user.avsc
```

```json
{"id":92594,"name":"MirageStream"}
```

### Pick a schema from an OpenAPI document

Append a JSON pointer as URL fragment to point at a schema inside a larger file. This works for any schema location,
for example a component:

```bash
mokapi sample-data './openapi.yaml#/components/schemas/Order'
```

```json
{"id":92414,"status":"paid"}
```

or the request body of an operation. Escape `/` inside a path key as `~1`:

```bash
mokapi sample-data './openapi.yaml#/paths/~1orders/post/requestBody/content/application~1json/schema'
```

Quote the argument, because `#` and `~` have a meaning in some shells. If the pointer leads into the `paths` section
of an OpenAPI document, Mokapi switches to `--input-type openapi` on its own.

## Generate several samples

`--count` (or `-n`) sets how many samples to generate. Each sample is generated independently.

```bash
mokapi sample-data -n 3 ./user.json
```

```
{"id":652,"email":"joshb@daniel.name","name":"LumenHaven","role":"admin","createdAt":"2021-12-22T01:10:07Z"}
{"id":264,"email":"micaela.hahn@ritchie.biz","name":"UtopiaUnfurl","role":"guest","createdAt":"2021-01-25T16:45:26Z"}
{"id":416,"email":"alessandro-fuller@herzog.com","name":"Eon","createdAt":"2021-11-13T04:21:52Z"}
```

With the default JSON output, you get one document per line (JSON Lines), which is easy to process with `jq`, `head`
or a `while read` loop.

## Choose the output format

`--output` (or `-o`) takes `json` (default), `xml`, `yaml`, `binary` or a full content type such as
`application/json`. Mokapi always writes to stdout, so use a shell redirect to write a file.

```bash
mokapi sample-data -o xml -n 2 --input-type openapi './openapi.yaml#/components/schemas/Order'
```

```xml
<samples>
<data><id>30840</id></data>
<data><id>19840</id><status>open</status></data>
</samples>
```

With `xml` and more than one sample, Mokapi wraps the documents in a `<samples>` element so that the output is a
single valid XML document. For a single sample there is no wrapper. Element names come from the `xml` keyword of the
schema, as in OpenAPI.

`binary` is meant for Avro: with `--input-type avro --output binary` the command writes the Avro binary encoding of
the value.

With `yaml`, each sample is a YAML document. Several samples are separated by a `---` line, so the output is a
multi-document YAML stream.

```bash
mokapi sample-data -o yaml -n 2 --input-type openapi './openapi.yaml#/components/schemas/Order'
```

```yaml
id: 30840
---
id: 19840
status: open
```

## Examples

### Create a fixture file for tests

```bash
mokapi sample-data -n 20 ./schemas/user.json > test/fixtures/users.jsonl
```

Generate the file once and commit it, if your tests need stable data. Every run produces different values (see
[Random data](#random-data)).

### Process the output with jq

```bash
mokapi sample-data -n 2 ./user.json | jq -r .email
```

```
cbrown@pena.biz
mgoldner@martinez.io
```

### Seed a database in CI

Because the output is one JSON document per line, you can feed it into any client that accepts JSON. This example
posts 50 users to an API under test:

```bash
mokapi sample-data -n 50 ./schemas/user.json | while read -r user; do
  curl -fs -X POST http://localhost:8080/users \
    -H 'Content-Type: application/json' \
    -d "$user" > /dev/null || exit 1
done
```

In a pipeline, run it after the service under test is ready. If Mokapi itself runs in CI, see
[Start Mokapi in CI and Wait for the Health Check](/docs/configuration/use-cases/ci-health-check.md).

### Use the command in Docker

The image's entrypoint is `mokapi`, so the subcommand goes after the image name. Mount the folder with your schemas:

```bash
docker run --rm -v "$(pwd)/schemas:/schemas:ro" \
  mokapi/mokapi:latest sample-data -n 3 /schemas/user.json
```

## Random data

Each run generates new values. The command has no flag to set a seed, so you can't reproduce the same output.
If a test needs the same data every time, generate it once and commit the file.

The data generator settings of the Mokapi server, such as `--data-gen-optional-properties`, are not available for
this command. Passing the flag fails with `unknown flag`. To control what is generated, constrain the schema itself,
for example with `required`, `enum`, `minimum`, `maximum`, `pattern` or `format`.

## Errors

On any error, the command prints a message to stderr and exits with code `1`.

| Situation | Message |
|-----------|---------|
| File doesn't exist | `failed to read nofile.json: open nofile.json: no such file or directory` |
| No argument and nothing piped to stdin | `failed to read : no input provided (expected file, URL, or stdin)` |
| Unknown `--input-type` | `failed to read <input>: unsupported input type: foo` |
| Unsupported output for the schema | `failed to marshal data: ...` |

An inline schema that isn't valid JSON or YAML is treated as a file path, so you get a `failed to read` error instead
of a parse error. If you see it for an inline schema, check the quotes and braces first.

## Reference

See [CLI Flags](/docs/configuration/static/cli-flags#sample-data) for the complete list of flags of
`mokapi sample-data`.
