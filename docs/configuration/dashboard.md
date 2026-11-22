---
title: Dashboard Configuration
description: Change the port and the path prefix of the Mokapi dashboard and its API, or serve it behind a reverse proxy.
---
# Dashboard Configuration

The dashboard and its [API](/docs/configuration/dashboard-api.md) share one server. It listens on
port 8080 and serves the dashboard at `/dashboard` and the API under `/api`. See
[Using Mokapi Dashboard](/docs/get-started/dashboard.md) for how to use the dashboard.

## Port

To use another port:

```bash tab=CLI
--api-port 5000
```
```bash tab=Env
MOKAPI_API_PORT=5000
```
```yaml tab=File (YAML)
api:
  port: 5000
```

## Path prefix

If the dashboard doesn't live at the root of its domain, for example at `example.com/mokapi`,
set the prefix. Mokapi then serves the dashboard at `/mokapi/dashboard` and its API at
`/mokapi/api`.

```bash tab=CLI
--api-path /mokapi
```
```bash tab=Env
MOKAPI_API_PATH=/mokapi
```
```yaml tab=File (YAML)
api:
  path: /mokapi
```

## Behind a reverse proxy

If a reverse proxy strips segments from the URL, set the base path instead. For example, the
browser opens `example.com/app/mokapi`, but the proxy forwards the request to Mokapi as `/`.

```bash tab=CLI
--api-base /app/mokapi
```
```bash tab=Env
MOKAPI_API_BASE=/app/mokapi
```
```yaml tab=File (YAML)
api:
  base: /app/mokapi
```

[Running the Dashboard Behind a Reverse Proxy](/docs/configuration/use-cases/reverse-proxy.md) explains when to use `--api-path`, `--api-base` or both.

The [CLI Flags](/docs/configuration/static/cli-flags#api-base) page lists further dashboard options, such as `--api-dashboard` and the search options.
