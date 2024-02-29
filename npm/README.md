<p align="center">
<img src="https://github.com/marle3003/mokapi/raw/v0.5.0/logo.svg" alt="Mokapi" title="Mokapi" width="300" />
</p>

<h3 align="center">Easy and flexible API mocking</h3>

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

&nbsp;
<p align="center">
<a href="https://www.buymeacoffee.com/mokapi" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>
<p align="center">


## Example

<img src="https://mokapi.io/everythingcode.png" alt="Example">

## Web UI

<img src="https://github.com/marle3003/mokapi/raw/v0.5.0/webui.png" alt="Mokapi Web UI" title="Mokapi Web UI" />

## Usage

```shell
npx mokapi --Providers.File.Directory=./mokapi
```

## Documentation

- [Get Started](https://mokapi.io/docs/guides/get-started/welcome)
- [HTTP](https://mokapi.io/docs/guides/http/overview)
- [Kafka](https://mokapi.io/docs/guides/kafka/overview)
- [LDAP](https://mokapi.io/docs/guides/ldap/overview)
- [SMTP](https://mokapi.io/docs/guides/smtp/overview)
- [Javascript API](https://mokapi.io/docs/javascript-api)
- [Examples](https://mokapi.io/docs/examples)