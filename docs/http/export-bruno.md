---
title: Export a Bruno Collection
description: Export a ready-to-use Bruno collection from the Mokapi Dashboard, with example parameters, headers and request bodies from your OpenAPI spec.
---

# Export a Bruno Collection

[Bruno](https://www.usebruno.com) is a free, open-source, Git-friendly API client. Collections are plain
files, so you can version them alongside your code.

Bruno can already import an OpenAPI specification, but the imported requests are mostly empty: you still
have to fill in every request body, header and parameter by hand. Mokapi closes this gap. It generates a
complete Bruno collection from your OpenAPI specification, including example values for every operation,
so you can send the first request right away.

## Export from the Dashboard

1. Open the [Dashboard](/docs/get-started/dashboard.md) and select an HTTP service.
2. Click the **⋮** button at the top right of the service card.
3. Select **Export Bruno Collection**.
4. Adjust the options if needed and click **Download**.

You can also copy the download URL from the dialog to use it in scripts or CI pipelines.

> The same menu contains **Export OpenAPI**, which downloads the specification of the service as JSON or YAML.

## Import into Bruno

1. Open Bruno and choose **Open Collection** or **Import Collection**.
2. Select the downloaded file.
3. Pick the environment in the top right corner of Bruno (see [Environments](#environments)).
4. Open a request and click **Send**.

## Options

| Option | Description |
|--------|-------------|
| **Base URL** | Host used to complete server URLs of the specification that have no host, for example `/api/books` or `http://:8080/api/books`. Server URLs that already define a host are not changed. The dialog pre-fills the address you use to open the Dashboard. Change it if Mokapi runs behind a proxy or in a container. |
| **Folder Arrangement** | Creates folders based on the **tags** or the **paths** of the specification. |
| **Item Name** | Preferred source for the name of each request: the **summary** or the **path**. If the chosen value is empty, Mokapi falls back to the `operationId` and finally to the path. |

## What Mokapi Generates

### Requests

Every operation becomes a request. The name starts with the HTTP method, followed by the item name, for
example `GET Find pet by ID`.

Mokapi fills in:

- **Path and query parameters**, serialized according to the `style` and `explode` settings of the parameter
- **Headers**, including `Content-Type` for requests with a body
- **Request body**, generated from the schema of the request body

Example values are chosen in this order:

1. `example` or `examples` defined on the parameter or media type
2. `example` or `default` defined in the schema
3. Random data generated from the schema, the same data generator Mokapi uses for mocked responses

### Folders

If tags are used for the folder arrangement, each tag becomes a folder. With OpenAPI 3.2, tags can have a
parent, which results in nested folders. Folders without any operation are not created, and operations without
a tag are placed at the root of the collection.

An operation with multiple tags appears only once, in the folder of its first tag.

Requests are sorted in the order they are defined in the specification.

### <a name="environments"></a>Environments

Mokapi creates one environment per server of the specification, including servers added by a
[patch](/docs/configuration/patching). Each environment defines the variable `baseUrl`, which every request
uses as prefix.

Server URLs without a host are completed with the Base URL from the export dialog. The port is kept if the
server URL defines one:

| Server URL in the specification | Base URL `localhost` | Resulting `baseUrl` |
|----------------------------------|----------------------|---------------------|
| `/mokapi/api/billing` | `localhost` | `http://localhost/mokapi/api/billing` |
| `http://:8080/mokapi/api/billing` | `localhost` | `http://localhost:8080/mokapi/api/billing` |

Environments are named after the `description` of the server. If a server has no description, Mokapi derives
a short name from the URL.

## Limitations

- The export is a snapshot. If the specification changes, download the collection again. Changes you made
  to exported requests, such as tests or assertions, are replaced when you overwrite the files.
- Bruno has no concept of OpenAPI parameter styles, so Mokapi resolves them into the final values. Path
  parameters of the styles `label` or `matrix` with several parameters in one path segment are written
  directly into the URL and cannot be edited as separate path parameters in Bruno.
- Request bodies are generated for the first suitable media type of operation, preferring
  `application/json`.

## Next Steps

- [Learn how Mokapi generates test data](/docs/get-started/test-data.md)
- [Mock HTTP APIs with OpenAPI](/docs/http/overview.md)