---
title: "Install Mokapi: Quick & Easy Setup Guide"
description: Install Mokapi on Windows, macOS, Linux, Docker, or Node.js. Follow a simple step-by-step guide to get started quickly.
subtitle: Get Mokapi running in seconds on Windows, macOS, Linux, Docker, or Node.js.
cards:
  items:
    - title: Run Your First Mock
      href: /docs/get-started/running
      description: Learn how to start Mokapi and mock your first API
    - title: Explore Tutorials
      href: /resources
      description: Follow step-by-step guides for REST, Kafka, LDAP, and SMTP
    - title: Write Scripts
      href: /docs/javascript-api/overview
      description: Add dynamic behavior to your mocks with JavaScript
    - title: Configure Mokapi
      href: /docs/configuration/overview
      description: Customize ports, providers, and other settings
---
# Install Mokapi

## Overview

Mokapi is an open-source API mocking tool that helps you develop and test faster by simulating REST APIs, Kafka topics,
LDAP directories, and SMTP servers. This guide shows you how to install Mokapi on your platform.

Choose your preferred installation method based on your platform and workflow.

```` box=benefits title="Try Without Installing"
Test Mokapi instantly with npx (requires Node.js):

```bash style=simple
npx go-mokapi serve https://petstore3.swagger.io/api/v3/openapi.json
```


This starts a mock server immediately without permanent installation. Perfect for quick tests and demos.
````

## Installation Options

Mokapi can be installed via direct download or through package managers on supported platforms.
Choose your preferred method below:

::: tabs

@tab "NPM"

If you prefer to install Mokapi globally as a Node.js package, install [go-mokapi](https://www.npmjs.com/package/go-mokapi)
using:

```bash
npm install -g go-mokapi
```

After installation, the mokapi command is available globally.

@tab "macOS"

### Homebrew

Install via Homebrew for easy updates and management:

```bash
brew tap marle3003/tap 
brew install mokapi
```

### Direct Download

Download the latest macOS binary from [GitHub](https://github.com/marle3003/mokapi/releases). Extract 
archive and move the binary to your PATH.

@tab "Windows"

### Chocolatey

Install via Chocolatey for easy updates:

```Powershell
choco install mokapi
```

### Direct Download

Download the latest Windows version from [GitHub](https://github.com/marle3003/mokapi/releases)

@tab "Linux"

Download the .deb package from the releases page and install:

### Direct Download

Download file appropriate for your Linux distribution and ARCH from the [release page](https://github.com/marle3003/mokapi/releases), then install with

```tab=deb
dpkg -i mokapi_{version}_linux_{arch}.deb
```

```tab=rpm
rpm -i mokapi_{version}_linux_{arch}.rpm
```

@tab "Docker"

Mokapi provides official Docker images on [Docker Hub](https://hub.docker.com/r/mokapi/mokapi):

```
docker pull mokapi/mokapi
```

:::

```` box=info title="Verify Installation"
After installation, verify Mokapi is working:

```bash style=simple
mokapi --version
```

````

## TypeScript Support

For full type safety and autocompletion when writing Mokapi scripts in TypeScript, install the type definitions:

```bash
npm install --save-dev @types/mokapi
```

This enables IntelliSense and type checking in your IDE when writing custom event handlers and scripts.

### Example TypeScript Script

```typescript title=petstore.ts
import { on } from 'mokapi'

export default function() {
  on('http', (request, response) => {
    // TypeScript provides full autocompletion here
    if (request.path.petId === '999') {
      response.statusCode = 404
      response.data = { message: 'Pet not found' }
    }
  })
}
```
``` box=tip title="IDE Integration"
With @types/mokapi installed, editors like VS Code will provide autocompletion for Mokapi's API, making script development faster and less error-prone.
```

## What's Next?

Now that Mokapi is installed, here's what you can do:

{{ card-grid key="cards" }}