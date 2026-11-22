---
title: Load Mocks from a Private GitHub Repository
description: Pull OpenAPI and AsyncAPI specs and Mokapi scripts from a private GitHub repository with a GitHub App that has read-only access to the repository contents.
---
# Load Mocks from a Private GitHub Repository

Your team keeps its OpenAPI and AsyncAPI specs and Mokapi scripts in a private GitHub repository, and you want
Mokapi on a laptop or in CI to pull them from there. The [GIT provider](/docs/configuration/dynamic/git.md)
can do that. For a private repository it authenticates as a GitHub App, so you don't need a personal access
token tied to a user account.

This page walks through the setup. The options themselves are described in the
[GIT provider](/docs/configuration/dynamic/git.md#github-app-authentication) reference.

## Create the GitHub App

1. In GitHub, open Settings > Developer settings > GitHub Apps and select **New GitHub App**. For an
   organization, use the settings of the organization instead.
2. Enter a name and a homepage URL (any URL works). Clear **Active** under Webhook, Mokapi doesn't need one.
3. Under Repository permissions, set **Contents** to **Read-only**. Leave everything else at **No access**.
4. Under Where can this GitHub App be installed, choose **Only on this account**, then select
   **Create GitHub App**.

The permission **Contents: Read-only** is all Mokapi needs to clone and pull. Metadata read access is added by
GitHub automatically.

## Install the app and collect the values

Open the app settings and select **Install App**, then install it on your account or organization. Choose
**Only select repositories** and pick the repository with your mocks.

You need three values:

| Value | Where to find it |
|-------|------------------|
| `appId` | App ID at the top of the app's settings page (General). |
| `installationId` | The number at the end of the installation URL, for example `https://github.com/settings/installations/123456789`. Open it from Install App > Configure. |
| `privateKey` | On the app's settings page, scroll to Private keys and select **Generate a private key**. GitHub downloads a `.pem` file. |

Store the `.pem` file outside the repository that holds your Mokapi configuration, and don't commit it.

## Configure Mokapi

Set the options on the repository entry. The authentication belongs to one repository, it isn't a global
setting, so each private repository needs its own `auth` block.

```yaml tab=File
providers:
  git:
    repositories:
      - url: https://github.com/my-org/my-mocks
        auth:
          github:
            appId: 12345
            installationId: 123456789
            privateKey: /etc/mokapi/github-app.pem
```
```bash tab=CLI
--providers-git-repositories[0]-url https://github.com/my-org/my-mocks \
--providers-git-repositories[0]-auth-github appId=12345,installationId=123456789,privateKey=/etc/mokapi/github-app.pem
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_URL=https://github.com/my-org/my-mocks
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_AUTH_GITHUB=appId=12345,installationId=123456789,privateKey=/etc/mokapi/github-app.pem
```

`privateKey` accepts a path to the `.pem` file or the PEM content itself. Mokapi checks whether the value is an
existing file and otherwise treats it as content. A relative path is resolved against the working directory of
the Mokapi process, not against the location of the config file. An absolute path avoids surprises.

Mokapi clones the repository when it starts and pulls changes every 3 minutes. Use `pullInterval` on the
repository to change that. To load only part of the repository, or another branch, add `include`, `files` or
the `ref` query parameter to the URL:

```yaml
providers:
  git:
    repositories:
      - url: https://github.com/my-org/my-mocks?ref=develop
        include: ['mocks/**/*.yaml', 'mocks/**/*.ts']
        pullInterval: 1m
        auth:
          github:
            appId: 12345
            installationId: 123456789
            privateKey: /etc/mokapi/github-app.pem
```

## Run Mokapi in Docker

Mount the configuration file and the key read-only. This keeps the key out of the image and out of the process
list.

```yaml
# mokapi.yaml
providers:
  git:
    repositories:
      - url: https://github.com/my-org/my-mocks
        auth:
          github:
            appId: 12345
            installationId: 123456789
            privateKey: /run/secrets/github-app.pem
```

```bash
docker run --rm -p 8080:8080 \
  -v $(pwd)/mokapi.yaml:/etc/mokapi/mokapi.yaml:ro \
  -v $(pwd)/github-app.pem:/run/secrets/github-app.pem:ro \
  mokapi/mokapi
```

In CI, write the key from a secret store to a file at runtime and mount that file the same way. If you pass the
PEM content through an environment variable, remember that it ends up in the environment of the container.

## When it doesn't work

The GIT provider writes these errors to the Mokapi log:

```
init git repository 'https://github.com/my-org/my-mocks' failed: <reason>
```

| Reason | Check |
|--------|-------|
| `private key is empty` | `privateKey` isn't set, or points to an empty file. |
| `could not parse private key` | The value is neither an existing file nor valid PEM content. Check the path, and remember that a relative path depends on the working directory. |
| `unable to clone git ...` with `could not refresh installation id ...'s token` | GitHub rejected the token request. The message ends with the HTTP status. Check `appId`, `installationId` and that the private key belongs to this app. |
| `unable to clone git ...` with another error | The token was issued, but GitHub didn't let the app read the repository. Check that the app is installed on this repository and that **Contents** is set to read-only. |

Later pull errors are logged as `git fetch error` or `unable to pull`.
