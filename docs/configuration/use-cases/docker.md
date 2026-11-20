---
title: Running Mokapi in Docker and Docker Compose
description: Run the official Mokapi image with docker run or Docker Compose, mount specs and scripts, publish ports, configure Mokapi with environment variables and add a health check.
---
# Running Mokapi in Docker and Docker Compose

The official image `mokapi/mokapi` contains only the `mokapi` binary, which is the entrypoint. Everything you
put after the image name is passed to Mokapi as flags and positional arguments, exactly as in
[Running Mokapi](/docs/configuration/running-mokapi.md). Specs and scripts come from a mounted volume, a URL
or a Git repository, and options can be set with environment variables.

| Tag | Base image | Use |
|-----|------------|-----|
| `latest` | Alpine | Latest release |
| `<version>`, for example `0.57.0` | Alpine | A specific release, recommended for reproducible builds |
| `ubuntu` | Ubuntu | Latest release on an Ubuntu base |

If you need your own base image, see [Mokapi with custom base image](/resources/examples/mokapi-with-custom-base-image).

## Run with docker run

Mount a directory with your specs and scripts, publish the ports, and tell Mokapi where to read from:

```bash
docker run --rm \
  -p 8080:8080 -p 80:80 \
  -v "$(pwd)/mocks:/data:ro" \
  mokapi/mokapi:latest /data
```

`/data` is a positional argument. Mokapi reads all files in the directory and watches it for changes, see the
[file provider](/docs/configuration/dynamic/file.md). A single file works the same way:

```bash
docker run --rm -p 8080:8080 -p 80:80 \
  -v "$(pwd)/petstore.yaml:/data/petstore.yaml:ro" \
  -v "$(pwd)/petstore.js:/data/petstore.js:ro" \
  mokapi/mokapi:latest /data/petstore.yaml /data/petstore.js
```

``` box=warning title=Limitation
Mokapi detects changes through file system notifications. In a Linux container on a Windows WSL2 host, these
notifications can be missing and Mokapi doesn't reload changed files. Restart the container in that case.
```

## Ports

Mokapi doesn't have one port for everything. The ports depend on what you mock:

| Port | What listens there | Source of the port |
|------|--------------------|--------------------|
| 8080 | Dashboard, REST API, [health check](/docs/configuration/health-check) and MCP server | [`--api-port`](/docs/configuration/static/cli-flags#api-port) |
| 80 | HTTP mocks whose server URL in the OpenAPI spec has no port | `servers` of the spec |
| other | Any port that a spec defines, for example `8081` for `http://localhost:8081`, `9092` for a Kafka broker | `servers` of the spec |

Mokapi listens on all interfaces inside the container, so you only need to publish the ports that you reach from
outside the container. Publish exactly the ports of your specs. If a spec defines `http://localhost:8081/petstore`,
add `-p 8081:8081`.

You can move the dashboard with `--api-port`. Health check and MCP server use the same port as the dashboard
unless you change [`--health-port`](/docs/configuration/health-check) or `--mcp-server-port`.

## Configure with environment variables

Every flag has an environment variable. Its name is `MOKAPI_` followed by the flag name in upper case, with `-`
replaced by `_`. The [CLI Flags](/docs/configuration/static/cli-flags) page lists the name for every option.

```bash
docker run --rm \
  -p 9090:9090 -p 80:80 \
  -v "$(pwd)/mocks:/data:ro" \
  -e MOKAPI_PROVIDERS_FILE_DIRECTORY=/data \
  -e MOKAPI_API_PORT=9090 \
  -e MOKAPI_LOG_LEVEL=debug \
  -e MOKAPI_LOG_FORMAT=json \
  mokapi/mokapi:latest
```

This is the same as starting Mokapi with `--providers-file-directory /data --api-port 9090 --log-level debug
--log-format json`. Environment variable names are not case-sensitive, so `MOKAPI_Log_Level` also works.

Some details to keep in mind:

- List options take values separated by spaces, for example
  `MOKAPI_PROVIDERS_FILE_INCLUDE="*.json *.yaml"`. For file directories you can also use the path separator of the
  system, `:` on Linux.
- Nested list items use an index, for example `MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_URL`. Complex
  options like this are easier to read in a [configuration file](#use-a-configuration-file).
- Mokapi rejects an unknown variable that starts with `MOKAPI_` and exits with
  `unknown environment variable`. This includes typos, so check the name if a container stops right after the
  start.
- A flag on the command line wins over an environment variable, see
  [Precedence](/docs/configuration/overview.md#precedence).

### Use a configuration file

For larger setups, mount a [configuration file](/docs/configuration/overview.md#configuration-file) instead of
setting many variables. Mokapi looks for `mokapi.yaml` in the working directory and in `/etc/mokapi`, so mount it
there:

```yaml tab=mokapi.yaml
providers:
  file:
    directory: /data
api:
  port: 9090
log:
  level: debug
```
```bash tab=Docker
docker run --rm -p 9090:9090 -p 80:80 \
  -v "$(pwd)/mocks:/data:ro" \
  -v "$(pwd)/mokapi.yaml:/etc/mokapi/mokapi.yaml:ro" \
  mokapi/mokapi:latest
```

## Load mocks from a URL or a Git repository

You don't need a volume if your specs live on a server or in a repository. Pass the URL as an argument or use
the provider options:

```bash
docker run --rm -p 8080:8080 -p 80:80 \
  mokapi/mokapi:latest https://petstore3.swagger.io/api/v3/openapi.json
```
```bash
docker run --rm -p 8080:8080 -p 80:80 \
  -e MOKAPI_PROVIDERS_GIT_URL=https://github.com/foo/bar.git \
  mokapi/mokapi:latest
```

See the [HTTP provider](/docs/configuration/dynamic/http.md) and the [GIT provider](/docs/configuration/dynamic/git.md).
For a private repository, follow [Load Mocks from a Private GitHub Repository](/docs/configuration/use-cases/private-github-repository.md)
and mount the private key file into the container. Behind a proxy or with an internal CA, see
[Corporate Proxy](/docs/configuration/use-cases/corporate-proxy.md).

## Run with Docker Compose

A `docker-compose.yml` with the same setup as above:

```yaml
services:
  mokapi:
    image: mokapi/mokapi:latest
    command: /data
    ports:
      - "8080:8080"
      - "80:80"
    volumes:
      - ./mocks:/data:ro
    environment:
      MOKAPI_LOG_LEVEL: debug
```

`command` replaces the default arguments of the image, so it holds the positional arguments (files,
directories and URLs) and flags. Options that have an environment variable can go under `environment`, which is
easier to override per environment.

Start it with `docker compose up -d` and open the dashboard at `http://localhost:8080`.

### Reach the mock from another service

Containers in the same Compose project share a network, and every service is reachable by its name. Another
service calls Mokapi with the service name as host and the port of the spec:

```yaml
services:
  mokapi:
    image: mokapi/mokapi:latest
    command: /data
    volumes:
      - ./mocks:/data:ro

  app:
    image: my-app
    environment:
      PETSTORE_URL: http://mokapi:80/api/v3
    depends_on:
      mokapi:
        condition: service_healthy
```

Inside the network you don't need `ports`. They are only required to reach Mokapi from your host, for example to
open the dashboard. Use the container port in the URL (`mokapi:80`), not the published port.

If the spec lists a different hostname in `servers`, for example `http://localhost/api/v3`, requests to
`http://mokapi/api/v3` still work. Mokapi first looks for a service that matches the `Host` header of the request,
then for a service without a host, and finally tries all services.

Kafka clients are different. The broker answers the metadata request with the host defined in the AsyncAPI
`servers`, and the client connects to that address afterwards. Set the host in the spec to the Compose service
name, for example `mokapi:9092`.

### Health check

Mokapi serves `/health` on the dashboard port by default, see [Health Check](/docs/configuration/health-check).
The Alpine-based image includes BusyBox `wget`, so no extra tool is needed:

```yaml
services:
  mokapi:
    image: mokapi/mokapi:latest
    command: /data
    volumes:
      - ./mocks:/data:ro
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 5s
      timeout: 2s
      retries: 10
```

The check passes when Mokapi's HTTP server answers with `200`. It doesn't verify that your specs loaded without
errors. If you changed `--api-port`, `--health-port` or `--health-path`, use the same values in the URL.

The example uses the Alpine-based tags (`latest` and version tags). The `ubuntu` image ships neither `wget` nor
`curl`, so this health check fails there. If you need the Ubuntu base, build your own image with one of them
installed (see [Mokapi with custom base image](/resources/examples/mokapi-with-custom-base-image)).

## Check your setup

Look at the container log. Mokapi logs every host it adds and the binding it uses, which shows whether your specs
were loaded and on which ports:

```bash
docker compose logs mokapi
```

Then call the health endpoint and one of your mocked endpoints from the host:

```bash
curl http://localhost:8080/health
curl http://localhost/api/v3/pet/12
```
