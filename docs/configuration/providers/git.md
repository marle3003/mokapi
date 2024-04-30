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
    url: https://github.com/PATH-TO/REPOSITORY?ref=branch-name
```
```bash tab=CLI
--providers.git.url="https://github.com/PATH-TO/REPOSITORY?ref=branch-name"
```
```bash tab=Env
MOKAPI_Providers_GIT_URL=https://github.com/PATH-TO/REPOSITORY?ref=branch-name
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

## Advanced Repository Settings

```yaml tab=File (YAML)
providers:
  repositories:
    - url: https://github.com/PATH-TO/REPOSITORY
      pullInterval: 5m
```
```bash tab=CLI
--providers.git.repositories[0].url="https://github.com/PATH-TO/REPOSITORY" --providers.git.repositories[0].pullInterval="5m"
```
```bash tab=Env
MOKAPI_Providers_GIT_Repositories[0]_Url=https://github.com/PATH-TO/REPOSITORY
MOKAPI_Providers_GIT_Repositories[0]_PullInterval=5m
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
        pullInterval: 2h45m
```

### TempDir

By default, Mokapi checkouts all repository to systems temp directory. You can use the tempDir option to override the default path.

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