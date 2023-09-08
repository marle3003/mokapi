---
title: GIT Provider for dynamic configuration
description: Pull your dynamic configuration from a GIT repository.
---
# GIT

Pull your dynamic configuration from a GIT repository.

``` box=tip
After cloning a GIT repository Mokapi uses the <a href="/docs/configuration/providers/file">File Provider</a> to read the
containing files.
```

## Configuration Example

```yaml tab=File (YAML)
providers:
  git:
    url: https://github.com//PATH-TO/REPOSITORY
```
```bash tab=CLI
--providers.git.url="https://github.com//PATH-TO/REPOSITORY"
```
```bash tab=Env
MOKAPI_Providers_GIT_URL=https://github.com//PATH-TO/REPOSITORY
```

## Provider Configuration

A list of all options that can be used with the GIT provider, refer to
the [reference page](/docs/configuration/reference.md).


### URL
Defines the URL to the GIT repository. With the query parameter `ref` you can specify an alternative
branch.

```yaml tab=File (YAML)
providers:
  git:
    url: https://github.com//PATH-TO/REPOSITORY?ref=branch-name
```
```bash tab=CLI
--providers.git.url="https://github.com//PATH-TO/REPOSITORY?ref=branch-name"
```
```bash tab=Env
MOKAPI_Providers_GIT_URL=https://github.com//PATH-TO/REPOSITORY?ref=branch-name
```

### Pull Interval
Defines in which interval Mokapi pulls possible changes, default 5 seconds.
Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

```yaml tab=File (YAML)
providers:
  git:
    pullInterval: 2h45m
```
```bash tab=CLI
--providers.git.pullInterval="2h45m"
```
```bash tab=Env
MOKAPI_Providers_GIT_PullInterval=2h45m
```