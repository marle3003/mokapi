# GIT

Pull your dynamic configuration from a GIT repository.

``` box=tip
After cloning a GIT repository Mokapi uses the File Provider to read the
containing files.
```

## Configuration

A list of all options that can be used with the GIT provider, refer to
the [reference page](/docs/references/static-configuration.md).


### URL
Defines the URL to the GIT repository. With the query parameter `ref` you can specify an alternative
branch.

```bash tab=CLI
--providers.git.url="https://github.com/marle3003/mokapi-example?ref=main"
```
```bash tab=Env
MOKAPI_Providers_GIT_URL=https://github.com/marle3003/mokapi-example?ref=main
```
```yaml tab=File (YAML)
providers:
  git:
    url: https://github.com/marle3003/mokapi-example?ref=main
```

### Pull Interval
Defines in which interval Mokapi pulls possible changes, default 5 seconds.
Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

```bash tab=CLI
--providers.git.pullInterval="2h45m"
```
```bash tab=Env
MOKAPI_Providers_GIT_PullInterval=2h45m
```
```yaml tab=File (YAML)
providers:
  git:
    pullInterval: 2h45m
```