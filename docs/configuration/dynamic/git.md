---
title: GIT Provider for dynamic configuration
description: Pull your dynamic configuration from a GIT repository.
---
# GIT

Pull your dynamic configuration from a GIT repository.

``` box=tip
After cloning a GIT repository Mokapi uses the <a href="/docs/configuration/dynamic/file">File Provider</a> to read the
containing files.
```

## Configuration Example

```yaml tab=File (YAML)
providers:
  git:
    url: https://github.com//PATH-TO/REPOSITORY
```
```bash tab=CLI
--providers-git-url https://github.com//PATH-TO/REPOSITORY
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_URL=https://github.com//PATH-TO/REPOSITORY
```

## Provider Configuration

A list of all options that can be used with the GIT provider, refer to
the [CLI Flags](/docs/configuration/static/cli-flags#providers-git) page.


### URL
Defines the URL to the GIT repository. With the query parameter `ref` you can specify an alternative
branch.

```yaml tab=File (YAML)
providers:
  git:
    url: https://github.com/PATH-TO/REPOSITORY?ref=branch-name
```
```bash tab=CLI
--providers-git-url="https://github.com/PATH-TO/REPOSITORY?ref=branch-name"
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_URL=https://github.com/PATH-TO/REPOSITORY?ref=branch-name
```

Mokapi also accepts the scheme `git+https://`. It clones the repository over HTTPS, so `git+` is
removed for cloning. The URLs of the repository's files keep the spelling you configured, which
matters if you reference files with `$ref` (see [Referencing Repository Files](#referencing-repository-files)).
This works for `url`, `urls` and `repositories[].url`.

```yaml tab=File (YAML)
providers:
  git:
    url: git+https://github.com/PATH-TO/REPOSITORY?ref=branch-name
```

### Pull Interval
Defines in which interval Mokapi pulls possible changes, default 3 minutes.
Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

```yaml tab=File (YAML)
providers:
  git:
    pullInterval: 3m30s
```
```bash tab=CLI
--providers-git-pull-interval 3m30s
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_PULL-INTERVAL=3m30s
```

## Advanced Repository Settings

```yaml tab=File (YAML)
providers:
  git:
    repositories:
      - url: https://github.com/PATH-TO/REPOSITORY
        pullInterval: 5m
```
```bash tab=CLI
--providers-git-repositories[0]-url="https://github.com/PATH-TO/REPOSITORY" --providers-git-repositories[0]-pull-interval="5m"
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_URL=https://github.com/PATH-TO/REPOSITORY
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_PULL_INTERVAL=5m
```

### Files
Specifies an allow list of files. Mokapi will only read these files.

```yaml tab=File (YAML)
providers:
  git:
    repositories:
      - url: https://github.com/PATH-TO/REPOSITORY
        files: ['mokapi/api.json']
```

### Include
Specifies an array of filenames or patterns. Mokapi will only read these files.

```yaml tab=File (YAML)
providers:
  git:
    repositories:
      - url: https://github.com/PATH-TO/REPOSITORY
        include: ['mokapi/**/*.json']
```

### Pull Interval
Specifies a specific pull interval for this repository

```yaml tab=File (YAML)
providers:
  git:
    repositories:
      - url: https://github.com/PATH-TO/REPOSITORY
        pullInterval: 3m30s
```

### TempDir

By default, Mokapi checks out all repositories to system's temp directory. You can use the tempDir option to override the default path.

Default path
- On Unix systems, it uses $TMPDIR if non-empty, else /tmp.
- On Windows, it uses the first non-empty value from %TMP%, %TEMP%, %USERPROFILE%, or the Windows directory.

```yaml tab=File (YAML)
providers:
  git:
    url: https://github.com/PATH-TO/REPOSITORY?ref=branch-name
    tempDir: /path-to/repositories
```

### GitHub App Authentication

```yaml tab=File (YAML)
providers:
  git:
    repositories:
      - url: https://github.com/PATH-TO/REPOSITORY
        auth:
          github:
            appId: 12345
            installationId: 123456789
            privateKey: 2024-2-25.private-key.pem
```

#### GitHub AppId

You can find your app's ID on the settings page for your GitHub App.
Navigate to your Settings > Developer Settings > GitHub Apps > Your GitHub App > Edit.

#### GitHub InstallationId

Your installation ID can be found in the organization or user that you have installed your GitHub App too.

For users, please navigate to your Settings > Developer Settings > GitHub Apps > Your GitHub App > Edit > Install App > Configure Installation. 
The installation ID can be found at the end of the URL - if the URL is https://github.com/settings/installations/123456789, 
your installation id is 123456789

#### GitHub PrivateKey

Set content of your private key or path to your private key file. You can download your private key by navigating to
Settings > Developer Settings > GitHub Apps > Your GitHub App > Edit

## Referencing Repository Files

You can use a file from a GIT repository as `$ref` in a spec, or read it with
[open()](/docs/javascript-api/open.md) in a script. The URL has this format:

```
git+https://<host>/<repo-path>?ref=<branch>&file=<path>#<json-pointer>
```

```yaml
$ref: 'git+https://github.com/PATH-TO/REPOSITORY?ref=main&file=schemas/pet.yaml#/components/schemas/Pet'
```

| Part              | Description                                                                                                 |
|-------------------|-------------------------------------------------------------------------------------------------------------|
| `ref`             | Optional. A branch name. Without `ref`, Mokapi uses the default branch. Tags and commits are not supported. |
| `file`            | Required. The path of the file relative to the repository root, with or without a leading `/`.              |
| `#<json-pointer>` | Optional. References a part of the file.                                                                    |

The query has to come before the `#`, otherwise `ref` and `file` become part of the JSON pointer.
If `file` is missing, Mokapi reports `git URL requires query parameter 'file'`. A path with `..`
or a path into `.git` is rejected. A `$ref` reads the file directly, so `files` and `include` of
a configured repository don't restrict it.

### Clones and Authentication

If the same repository and branch is configured under `providers.git` (with `https://` or
`git+https://`), a `$ref` reuses that clone and its authentication. A `$ref` without `ref` only
matches a configured repository without `ref`. The authentication of a configured repository also
applies to other branches of that repository.

Otherwise Mokapi makes a shallow clone on demand, once per repository and branch, in the
[tempDir](#tempdir). Mokapi removes these clones on shutdown. If the clone fails, for example
because of a wrong URL or branch, Mokapi logs the error and tries again the next time the
referencing file is parsed.

### Updates

Mokapi pulls an on-demand clone of a configured repository with that repository's `pullInterval`.
Otherwise it uses `providers.git.pullInterval` (default 3 minutes). When a pull changes the
referenced file, Mokapi parses the referencing file again and updates the mock. If the file is
removed from the branch, the referencing file is parsed again and reports the missing file.

### URL Spelling

For `git+https`, the order of the query parameters and a leading `/` in `file` don't matter:
`?ref=main&file=/pet.yaml` and `?file=pet.yaml&ref=main` are the same config entry. Mokapi normalizes only these two, so `./pet.yaml` or `a//pet.yaml` are
separate entries. An `https`
URL and a `git+https` URL of the same file are two separate config entries, so use one spelling
consistently. A `$ref` with the scheme `git://` is not supported and fails with `unsupported scheme`.
