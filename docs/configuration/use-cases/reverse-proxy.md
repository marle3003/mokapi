---
title: Running the Dashboard Behind a Reverse Proxy
description: Serve the Mokapi dashboard and API under a URL prefix behind a reverse proxy and learn when to use api-path, api-base or both.
---
# Running the Dashboard Behind a Reverse Proxy

When a reverse proxy publishes the Mokapi dashboard under a prefix such as `/mokapi`, two options come into
play: [`--api-path`](/docs/configuration/static/cli-flags#api-path) and
[`--api-base`](/docs/configuration/static/cli-flags#api-base). They are easy to mix up because both take a
path and often have the same value. They answer different questions.

| Option | Question it answers | What Mokapi does with it |
|--------|---------------------|--------------------------|
| `api-path` | Which prefix does Mokapi receive in the request? | Removes the prefix from the request path before routing. |
| `api-base` | Which prefix does the browser use? | Writes it into the `<base href>` of the dashboard page. |

Neither option changes the port. The mocked APIs of your specs are not affected, only the dashboard and the
REST API under `/api`.

## What the options do

`api-path` is about the requests that reach Mokapi. If a request path starts with the configured prefix,
Mokapi removes it and routes the rest as usual. `/mokapi/api/info` is handled like `/api/info`. A request
without the prefix is still served, so `/api/info` works too. Mokapi doesn't reject it.

`api-base` is about the browser. The dashboard is a single page application that builds every URL (assets,
images and calls to `/api/...`) relative to the `<base href>` of its `index.html`. Mokapi fills that tag with
a prefix when you set `api-path` or `api-base`. If both are set, `api-base` wins. If only `api-path` is set, the
path is used. `api-base` has no influence on how Mokapi routes incoming requests.

| `api-path` | `api-base` | `<base href>` of the dashboard | Prefix removed from requests |
|:----------:|:----------:|--------------------------------|:----------------------------:|
| not set | not set | `/` | no |
| `/mokapi` | not set | `/mokapi/` | `/mokapi` |
| not set | `/mokapi` | `/mokapi/` | no |
| `/mokapi/dashboard` | `/app/mokapi` | `/app/mokapi/` | `/mokapi/dashboard` |

Use values without a trailing slash. Mokapi appends it.

The question to ask is: what does the proxy send to Mokapi, and what does the browser see? Three setups cover
almost everything.

## The proxy forwards the prefix unchanged

The browser opens `https://example.com/mokapi/`, and the proxy passes the request to Mokapi as
`/mokapi/`. Mokapi receives the prefix, so it must remove it. Set `api-path`.

```nginx
location /mokapi/ {
    proxy_pass http://mokapi:8080;
}
```

`proxy_pass` without a URI part keeps the original request path. In Caddy, `handle /mokapi/*` does the same.

```yaml tab=File
api:
  path: /mokapi
```
```bash tab=CLI
--api-path /mokapi
```
```bash tab=Env
MOKAPI_API_PATH=/mokapi
```

You don't need `api-base` here. Mokapi uses `api-path` for the `<base href>` as well.

## The proxy strips the prefix

The browser opens `https://example.com/mokapi/`, and the proxy passes the request to Mokapi as `/`. Mokapi
never sees the prefix, so there is nothing to remove. It only has to tell the browser which prefix to use for
the next requests. Set `api-base`.

```nginx
location /mokapi/ {
    proxy_pass http://mokapi:8080/;
}
```

The URI part `/` in `proxy_pass` replaces the matched `/mokapi/`. In Caddy, use `handle_path /mokapi/*`, in
Traefik the `StripPrefix` middleware.

```yaml tab=File
api:
  base: /mokapi
```
```bash tab=CLI
--api-base /mokapi
```
```bash tab=Env
MOKAPI_API_BASE=/mokapi
```

Setting `api-path` to the same value would also work in this setup, because Mokapi serves requests with and
without the prefix. `api-base` describes the situation more precisely: Mokapi itself is served at the root.

## The proxy rewrites the prefix

The public prefix differs from the one Mokapi receives. The browser opens `https://example.com/app/mokapi/`,
and the proxy passes the request to Mokapi as `/mokapi/dashboard/`. Now both values are needed: `api-path` is
what arrives at Mokapi, `api-base` is what the browser uses.

```nginx
location /app/mokapi/ {
    proxy_pass http://mokapi:8080/mokapi/dashboard/;
}
```

```yaml tab=File
api:
  path: /mokapi/dashboard
  base: /app/mokapi
```
```bash tab=CLI
--api-path /mokapi/dashboard \
--api-base /app/mokapi
```
```bash tab=Env
MOKAPI_API_PATH=/mokapi/dashboard
MOKAPI_API_BASE=/app/mokapi
```

## Choosing the wrong option

The dashboard is served, but its calls go to the wrong address. The typical case is a proxy that forwards the
prefix unchanged while only `api-base` is set. Mokapi has no prefix to remove, so a request such as
`/mokapi/api/info` doesn't match any API route. It falls through to the dashboard and returns the HTML page
instead of JSON, and the dashboard shows no data.

If the page itself doesn't load correctly, check which URLs the browser requests in the developer tools of
your browser. They all start with the `<base href>` that Mokapi rendered. Compare it with the public URL in
your address bar: that is what `api-base` (or `api-path`, if `api-base` isn't set) must contain.

## Check your setup

Open the page source of the dashboard through the proxy and look at the `<base>` tag. It must contain the
public prefix with a trailing slash. Then call the API through the proxy:

```bash
curl https://example.com/mokapi/api/info
```

The response is a JSON document with the Mokapi version. An HTML response means that the request reached the
dashboard instead of the API, and the prefix handling doesn't match your proxy.
