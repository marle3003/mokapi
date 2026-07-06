<p align="center">
<a href="https://mokapi.io">
<img src="https://raw.githubusercontent.com/marle3003/mokapi/main/logo.svg" alt="Mokapi - Open Source Mock API Server" width="300" />
</a>
</p>
<h2 align="center">The Open-Source Mock API Tool Across Protocols</h2>
<p align="center">
<a href="https://github.com/marle3003/mokapi/releases"><img src="https://img.shields.io/github/release/marle3003/mokapi.svg" alt="Github release"></a>
<a href="https://github.com/marle3003/mokapi/actions/workflows/test.yml"><img src="https://github.com/marle3003/mokapi/actions/workflows/build.yml/badge.svg" alt="Build status"></a>
<a href="https://codecov.io/gh/marle3003/mokapi"><img src="https://img.shields.io/codecov/c/gh/marle3003/mokapi/main.svg" alt="Codecov branch"></a>
<a href="https://github.com/marle3003/mokapi/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
</p>
<p align="center">
  <a href="https://mokapi.io/docs/welcome">Documentation</a> ·
  <a href="https://github.com/marle3003/mokapi/releases">Releases</a> ·
  <a href="https://mokapi.io/resources/tutorials">Tutorials</a> ·
  <a href="https://mokapi.io/resources/blogs">Blog</a>
</p>

## About

Mokapi is an open-source, local-first **mock API tool** to develop and test faster. Simulate complete environments driven
by OpenAPI and AsyncAPI specifications without external dependencies.

It supports HTTP/REST, Apache Kafka, MQTT, Websocket, LDAP, and SMTP from a single tool, making it useful across the
whole stack, not just for REST APIs.

Use it to:
- Build frontend UIs before the backend exists
- Test error states, timeouts, and edge cases safely
- Run reliable CI/CD pipelines without external dependencies
- Validate API contracts before writing implementation code

## Quick Start

Try the **mock API** instantly:

```
npx go-mokapi https://petstore3.swagger.io/api/v3/openapi.json
```

```
curl http://localhost/api/v3/pet/1 -H 'Accept: application/json'
```

A working mock from a public OpenAPI spec in under a minute. No installation required.

## Feature

| Feature              | Description                                             |
|----------------------|---------------------------------------------------------|
| Multi-protocol       | HTTP/HTTPS, Apache Kafka, MQTT, Websocket, LDAP, SMTP   |
| Spec-driven          | Uses OpenAPI and AsyncAPI as the source of truth        |
| JavaScript scripting | Customize responses, simulate errors, add delays        | 
| Everything as code   | Version-controlled, CI/CD-ready, no UI required         |
| Config patching      | Override parts of a spec without modifying the original |
| Built-in dashboard   | Real-time request and response viewer at localhost:8080 |
| Flexible providers   | Load specs from files, URLs, Git repos, or NPM packages |

## Usage

### Basic HTTP mock

```bash
npx go-mokapi ./openapi.yaml
```

### With a remote spec

```bash
npx go-mokapi https://example.com/api/openapi.json
```

### Docker

```bash
docker run -it -p 8080:8080 -p 80:80 \
  -v $(pwd):/data mokapi/mokapi /data/openapi.yaml
```

### Dashboard

Open http://localhost:8080 to view live requests, responses, and logs.

<img src="https://raw.githubusercontent.com/marle3003/mokapi/refs/heads/main/webui.png" alt="Mokapi Web UI" />

## Configuration

### JavaScript scripting

```javascript
import { on } from 'mokapi'

export default function() {
  on('http', (request, response) => {
    // Return 404 for specific IDs
    if (request.path.petId === '999') {
      response.statusCode = 404
      return
    }

    // Customize response data
    response.data.name = 'Custom Pet Name'
  })
}
```

### Spec patching
Override parts of your OpenAPI spec for specific test scenarios without touching the original file. See the
[configuration guide](https://mokapi.io/docs/configuration/overview) for details.

# Tutorials

Explore tutorials that walk you through mocking different protocols and scenarios:

- [Get started with REST API Mocking](https://mokapi.io/resources/tutorials/get-started-with-rest-api)  
  Deploy a REST **mock API** using OpenAPI specification

- [Mock Kafka with AsyncAPI](https://mokapi.io/resources/tutorials/get-started-with-kafka)  
  Simulate Kafka topics and validate message producers

- [Mock LDAP Authentication](https://mokapi.io/resources/tutorials/mock-ldap-authentication-in-node)\
  Test authentication flows without a real LDAP server

- [Mock SMTP Mail Servers](https://mokapi.io/resources/tutorials/mock-smtp-server-send-mail-using-node)\
  Test email workflows without sending real messages

- [CI/CD Integration with GitHub Actions](https://mokapi.io/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline)\
  Run Mokapi in automated test pipelines

More at [mokapi.io/resources](https://mokapi.io/resources)

## Documentation

- [Getting Started Guide](https://mokapi.io/docs/welcome)
- [HTTP/REST API Documentation](https://mokapi.io/docs/http/overview)
- [Kafka Documentation](https://mokapi.io/docs/kafka/overview)
- [LDAP Documentation](https://mokapi.io/docs/ldap/overview)
- [SMTP/Mail Documentation](https://mokapi.io/docs/mail/overview)
- [JavaScript API Reference](https://mokapi.io/docs/javascript-api/overview)
- [Configuration Guide](https://mokapi.io/docs/configuration/overview)

## Support the Project

If Mokapi saves you time, consider buying me a coffee. It helps keep the project going.

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/mokapi) <a href="https://www.buymeacoffee.com/mokapi" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 30px !important;width: 174px !important;" ></a>

## Merch

[![Merch](https://raw.githubusercontent.com/marle3003/mokapi/main/merch.png)](https://mokapi.myspreadshop.com/)

## License

MIT License - see [LICENSE](https://github.com/marle3003/mokapi/blob/main/LICENSE) for details.

<p align="center">
  <a href="https://mokapi.io">Website</a> ·
  <a href="https://npmjs.com/package/go-mokapi">NPM</a> ·
  <a href="https://hub.docker.com/r/mokapi/mokapi">Docker Hub</a>
</p>