---
title: "Install Mokapi: Quick & Easy Setup Guide"
description: Learn how to install Mokapi effortlessly across Windows, macOS, and Linux. Follow the step-by-step guide for a smooth setup experience.
---
# Install Mokapi

Mokapi is an open-source tool designed to simplify API mocking and schema validation.
It enables developers to prototype, test, and demonstrate APIs with realistic data and
scenarios. This guide provides straightforward instructions to install Mokapi on various
platforms.

## Installation Options

Mokapi can be installed via direct download or through package managers on supported platforms.
Choose your preferred method below:

::: tabs

@tab "macOS"

### Homebrew

```bash
brew tap marle3003/tap 
brew install mokapi
```

### Direct Download

Download the latest macOS version from [GitHub](https://github.com/marle3003/mokapi/releases)

@tab "Windows"

### Chocolatey

```Powershell
choco install mokapi
```

### Direct Download

Download the latest Windows version from [GitHub](https://github.com/marle3003/mokapi/releases)

@tab "Linux"

### Direct Download

Download file appropriate for your Linux distribution and ARCH from the [release page](https://github.com/marle3003/mokapi/releases), then install with

```tab=deb
dpkg -i mokapi_{version}_linux_{arch}.deb
```

```tab=rpm
rpm -i mokapi_{version}_linux_{arch}.rpm
```

@tab "Docker"

To get started with Mokapi using Docker, visit [DockerHub](https://hub.docker.com/r/mokapi/mokapi/tags) for a list of available images.
You can also use a custom base Docker image as demonstrated in [these examples](/resources/examples/mokapi-with-custom-base-image.md).

```
docker pull mokapi/mokapi
```

@tab "NPM"

If you prefer to install Mokapi as a Node.js package, use the following command:

```bash
npm install go-mokapi
```

:::

### Mokapi Scripts Type Definitions

Mokapi allows you to write **custom scripts** to handle API events or modify responses.  
For full type safety and autocompletion in TypeScript, you can install the [`@types/mokapi`](https://www.npmjs.com/package/@types/mokapi`) package:

```bash
npm install --save-dev @types/mokapi
```

## Next steps

- [Create your first Mock](/docs/get-started/running.md)
- [Install @types/mokapi](https://www.npmjs.com/package/@types/mokapi)