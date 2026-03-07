<p align="center">
<a href="https://mokapi.io">
<img src="https://raw.githubusercontent.com/marle3003/mokapi/refs/heads/main/logo.svg" alt="Mokapi" title="Mokapi" width="300" />
</a>
</p>
<h3 align="center">Mock APIs Across Protocols. Test Faster. Ship Better.</h3>
<p align="center">
  <a href="https://www.npmjs.com/package/go-mokapi"><img src="https://img.shields.io/npm/v/go-mokapi.svg" alt="npm version"></a>
  <a href="https://github.com/marle3003/mokapi/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/marle3003/mokapi"><img src="https://img.shields.io/github/stars/marle3003/mokapi?style=social" alt="GitHub stars"></a>
</p>

## What is Mokapi?

Mokapi is an open-source API mocking tool that lets you develop and test without waiting
for backends. Mock REST APIs, Kafka topics, LDAP directories, and SMTP servers using
OpenAPI and AsyncAPI specifications.

Perfect for:
- Frontend developers building UIs before backends exist
- QA teams testing edge cases, errors, and timeouts
- DevOps engineers running reliable CI/CD tests without external dependencies
- API designers prototyping and validating contracts early

## Quick Start

Try Instantly

```
npx go-mokapi https://petstore3.swagger.io/api/v3/openapi.json
```

Then test your mock:

```
curl http://localhost/api/v3/pet/1 -H 'Accept: application/json'
```

Install Globally

```
npm install -g go-mokapi
mokapi https://petstore3.swagger.io/api/v3/openapi.json
```

### Other Installation Methods

Check other installation methods [here](https://mokapi.io/docs/get-started/installation)

## Key Features

### Multi-Protocol Support
Mock HTTP/HTTPS, Apache Kafka, LDAP, and SMTP — all from a single tool.

### Specification-Driven
Uses OpenAPI and AsyncAPI specs as the source of truth. Your mocks stay aligned with your API contracts.

### Dynamic Behavior with JavaScript

Control responses, simulate errors, add delays, or create complex workflows using embedded JavaScript:
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

### Everything as Code
Version control your mocks alongside your code. Run them in CI/CD pipelines. No UI configuration required.

### Configuration Patching
Override parts of your OpenAPI spec without modifying the original file. Perfect for testing different scenarios.

### Built-in Dashboard
Visualize requests, responses, and logs in real-time at http://localhost:8080
<img src="https://raw.githubusercontent.com/marle3003/mokapi/refs/heads/main/webui.png" alt="Mokapi Web UI" title="Mokapi Web UI" />

### Multiple Providers
Load specs from local files, HTTP URLs, Git repositories, or NPM packages.

## Common Use Cases

### Frontend Development
Mock backend APIs while building UIs. Test loading states, errors, and edge cases without waiting for real endpoints.

### API Testing
Simulate timeouts, 500 errors, rate limits, and malformed responses. Test how your application handles failures.

### CI/CD Integration
Run fast, reliable tests without external dependencies. No flaky tests due to network issues or unavailable services.

### Contract Validation
Validate that your requests and responses match your OpenAPI specification. Catch breaking changes early.

# Example Tutorials

Explore tutorials that walk you through mocking different protocols and scenarios:

- [Get started with REST API](https://mokapi.io/resources/tutorials/get-started-with-rest-api)  
  Mock a REST API using OpenAPI specification

- [Mock Kafka with AsyncAPI](https://mokapi.io/resources/tutorials/get-started-with-kafka)  
  Simulate Kafka topics and validate message producers

- [Mock LDAP Authentication](https://mokapi.io/resources/tutorials/mock-ldap-authentication-in-node)\
  Test authentication flows without a real LDAP server

- [Mock SMTP Mail Servers](https://mokapi.io/resources/tutorials/mock-smtp-server-send-mail-using-node)\
  Test email workflows without sending real messages

- [CI/CD Integration with GitHub Actions](https://mokapi.io/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline)\
  Run Mokapi in automated test pipelines

> More examples [mokapi.io/resources](https://mokapi.io/resources)

## Documentation

- [Getting Started Guide](https://mokapi.io/docs/welcome)
- [HTTP/REST API Documentation](https://mokapi.io/docs/http/overview)
- [Kafka Documentation](https://mokapi.io/docs/kafka/overview)
- [LDAP Documentation](https://mokapi.io/docs/ldap/overview)
- [SMTP/Mail Documentation](https://mokapi.io/docs/mail/overview)
- [JavaScript API Reference](https://mokapi.io/docs/javascript-api/overview)
- [Configuration Guide](https://mokapi.io/docs/configuration/overview)

## Support the Project

If Mokapi helps your team ship faster, consider supporting development:

<a href="https://www.buymeacoffee.com/mokapi" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>

## License

MIT License - see [LICENSE](https://github.com/marle3003/mokapi/blob/main/LICENSE) for details.

## Links

- Website: [mokapi.io](https://mokapi.io)
- GitHub: [github.com/marle3003/mokapi](https://github.com/marle3003/mokapi)
- NPM Package: [npmjs.com/package/go-mokapi](https://npmjs.com/package/go-mokapi)
- Documentation: [mokapi.io/docs](https://mokapi.io/docs)
- Tutorials: [mokapi.io/resources/tutorials](https://mokapi.io/resources/tutorials)
- Blog: [mokapi.io/resources/blogs](https://mokapi.io/resources/blogs)