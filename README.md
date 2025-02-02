<p align="center">
<a href="https://mokapi.io">
<img src="logo.svg" alt="Mokapi" title="Mokapi" width="300" />
</a>
</p>

<h3 align="center">Easy and flexible API mocking</h3>

<p align="center">
<a href="https://github.com/marle3003/mokapi/releases"><img src="https://img.shields.io/github/release/marle3003/mokapi.svg" alt="Github release"></a>
<a href="https://github.com/marle3003/mokapi/actions/workflows/test.yml"><img src="https://github.com/marle3003/mokapi/actions/workflows/build.yml/badge.svg" alt="Build status"></a>
<a href="https://codecov.io/gh/marle3003/mokapi"><img src="https://img.shields.io/codecov/c/gh/marle3003/mokapi/main.svg" alt="Codecov branch"></a>
<a href="https://github.com/marle3003/mokapi/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
</p>
<p align="center">
    <a href="https://github.com/marle3003/mokapi/releases">Download</a> ·
    <a href="https://mokapi.io/docs/guides/get-started/welcome">Documentation</a>
</p>

**Mokapi** is an open-source tool that allows Agile, DevOps and Continuous Deployment teams
to create and test API designs before actually building them.

With Mokapi you can quickly and easily test various
scenarios, such as delayed or failed responses without
having to rely on a fully functional API.

Mokapi helps you improve the quality of APIs and
reduces the risk of bugs or errors in production.

Its core feature are:

- **Multiple Protocol support**: HTTP, HTTPS, Apache Kafka, SMTP, LDAP
- **Everything as Code**: Reusing, version control, consistency and integrate mocks with your CI/CD.
- **An embedded JavaScript engine** to control everything - status, headers, delays, errors or other edge cases.
- **Patch Configuration** changes for mocking needs, rather than changing the original contract
- **Multiple Provider support**: File, HTTP, GIT, NPM to gather configurations and scripts.
- **Dashboard** to see what's going on.

## Usage

Windows
```shell
choco install mokapi
mokapi --providers-http-url https://petstore31.swagger.io/api/v31/openapi.json
```

MacOS
```shell
brew tap marle3003/tap 
brew install mokapi
mokapi --providers-http-url https://petstore31.swagger.io/api/v31/openapi.json
```

Docker
```shell
docker run --env 'MOKAPI_Providers_Http_URL'='https://petstore31.swagger.io/api/v31/openapi.json' -p 80:80 -p 8080:8080 mokapi/mokapi:latest
```

Run a request
```shell
curl http://localhost/api/v31/pet/2 -H 'Accept: application/json'
```

## Example

<img src="webui/public/control-mock-api-everything.png" alt="Mokapi Web UI" title="Mokapi Web UI" />

## Dashboard

<img src="webui.png" alt="Mokapi Web UI" title="Mokapi Web UI" />

## Documentation

- [Get Started](https://mokapi.io/docs/guides/get-started/welcome)
- [HTTP](https://mokapi.io/docs/guides/http)
- [Kafka](https://mokapi.io/docs/guides/kafka/overview)
- [LDAP](https://mokapi.io/docs/guides/ldap/overview)
- [SMTP](https://mokapi.io/docs/guides/smtp/overview)
- [Javascript API](https://mokapi.io/docs/javascript-api)
- [Examples & Tutorials](https://mokapi.io/docs/examples)
- [Blogs](https://mokapi.io/docs/blogs)

&nbsp;
<p align="center">
<a href="https://www.buymeacoffee.com/mokapi" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>
<p align="center">