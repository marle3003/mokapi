---
title: Use the MCP Server with an AI Client
description: Configure Mokapi's MCP server with a flag, environment variable or config file, find its endpoint URL and connect AI clients such as Claude Code or GitHub Copilot, also when Mokapi runs in Docker.
---
# Use the MCP Server with an AI Client

Mokapi includes an MCP (Model Context Protocol) server. An AI client that supports MCP can connect to it, explore
your mocked APIs, inspect logs and events, and get the reference material it needs to write mock scripts. For
examples of what that looks like in practice, read the blog post
[Explore and Debug APIs with AI](/resources/blogs/mock-apis-with-mcp).

## The server is on by default

The MCP server is enabled by default and shares port `8080` with the dashboard. The log shows it at startup:

```
level=info msg="MCP server runs on binding :8080 on path /mcp"
```

So with a default setup, the endpoint is `http://localhost:8080/mcp`. Mokapi implements the streamable HTTP
transport of MCP, so clients connect over HTTP. There is nothing to install on the client side.

## Change or disable the server

Three options control the server. See [CLI Flags](/docs/configuration/static/cli-flags#mcp-server-enabled) for the
full reference.

| Flag | Environment variable | Config file | Default |
|------|----------------------|-------------|---------|
| `--mcp-server-enabled` | `MOKAPI_MCP_SERVER_ENABLED` | `mcp.server.enabled` | `true` |
| `--mcp-server-port` | `MOKAPI_MCP_SERVER_PORT` | `mcp.server.port` | `8080` |
| `--mcp-server-path` | `MOKAPI_MCP_SERVER_PATH` | `mcp.server.path` | `/mcp` |

Turn the server off if you don't need it:

```bash tab=CLI
mokapi --mcp-server-enabled=false ./mocks
```
```bash tab=Env
MOKAPI_MCP_SERVER_ENABLED=false mokapi ./mocks
```
```yaml tab=File
mcp:
  server:
    enabled: false
```

Move it to its own port and path:

```bash tab=CLI
mokapi --mcp-server-port 5000 --mcp-server-path /agent ./mocks
```
```bash tab=Env
MOKAPI_MCP_SERVER_PORT=5000 MOKAPI_MCP_SERVER_PATH=/agent mokapi ./mocks
```
```yaml tab=File
mcp:
  server:
    port: 5000
    path: /agent
```

The endpoint is then `http://localhost:5000/agent`. If the MCP port differs from `--api-port`, Mokapi starts a
separate HTTP listener for it. If both ports are the same, the MCP server is served by the same listener as the
dashboard and the REST API. In that case `--api-path` also applies: with `--api-path /mokapi` the endpoint is
`http://localhost:8080/mokapi/mcp`.

## Connect a client

Start Mokapi, then register the endpoint in your client. Mokapi runs MCP in stateless mode, so a client doesn't
need to keep a session open.

### Claude Code

```bash
claude mcp add --transport http mokapi http://localhost:8080/mcp
```

### Other clients

Most clients read a JSON file where each server has a `type` and a `url`. The exact file name and the top-level
key depend on the client. A common form looks like this:

```json
{
  "mcpServers": {
    "mokapi": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

Use the URL with your own port and path if you changed them. Check your client's documentation for where the file
lives and whether it expects `mcpServers` or another key.

## Run Mokapi in Docker

Inside the container, the MCP server listens on `8080` like the dashboard. Publish that port and use it in the
client configuration:

```bash
docker run --rm \
  -p 8080:8080 -p 80:80 \
  -v "$(pwd)/mocks:/data:ro" \
  mokapi/mokapi:latest /data
```

If you move the MCP server with `--mcp-server-port`, publish that port instead of (or in addition to) `8080`. See
[Docker and Docker Compose](/docs/configuration/use-cases/docker.md) for the other ports.

## What an agent can do

The server is named `mokapi-mcp` and offers three tools:

| Tool | What it does |
|------|--------------|
| `mokapi_get_automation_definitions` | Returns the TypeScript definitions of the automation API (`core`, `http`, `kafka`, `event`). |
| `mokapi_execute_code` | Runs JavaScript in a sandboxed Mokapi runtime to query live data: specs of OpenAPI, AsyncAPI, LDAP and mail services, schemas, logs and events. It can also invoke API operations. |
| `mokapi_get_mock_reference` | Returns reference material for writing mock scripts, either type definitions (`mokapi`, `http`, `kafka`, `faker`, `mustache`, `yaml`) or example scenarios such as `delay-latency`. |

The tool descriptions tell the agent to fetch the definitions before it executes code, so it doesn't have to guess
the API. The same material is available as MCP resources with the URIs `mokapi://lib/automation`,
`mokapi://lib/mocking/types` and `mokapi://lib/mocking/scenarios`.

## Security

The MCP server has no authentication of its own, and Mokapi binds it to all interfaces, like the dashboard.
Anyone who can reach the port can ask it to run code with `mokapi_execute_code`, including calls to your mocked
operations. Use it on a developer machine or in a trusted network. In shared environments, disable it with
`--mcp-server-enabled=false` or don't publish its port.
