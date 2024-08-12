---
title: Mokapi Guides
description: These guides will help you to run Mokapi in your environment
---

# Welcome to Mokapi guides

Mokapi is an open-source mocking tool for Agile, DevOps and Continuous Deployment teams to
**prototype**, **test** and **demonstrate** APIs in software solutions using realistic data and scenarios.

With Mokapi you can quickly and easily test various
scenarios with your APIs, such as delayed or failed responses without
having to rely on a fully functional API.

Mokapi helps you improve the quality of APIs and
reduces the risk of bugs or errors in production.

## Key features

- **Multiple Protocol support**

  Mokapi can mock APIs for [HTTP(S)](/docs/guides/http/overview.md), [Apache Kafka](/docs/guides/kafka/overview.md), [SMTP](/docs/guides/smtp/smtp.md) and [LDAP](/docs/guides/ldap/ldap.md)
  With Mokapi HTTP you can simulate any REST or SOAP API.

- **Everything as Code**

  Reusing, version control, consistency and integrate mocks with your CI/CD.

- **An embedded JavaScript engine**

  Write the behavior of your mock in JavaScript and control everything: status, headers, delays, errors or other edge cases.

- **Configuration Options**

  Use CLI arguments, environment variable or file configuration and take advantage from hot-reloading of dynamic configurations.

- **Patch Configuration**

  Mokapi supports configuration patch files so that adjustments required for the mock do not have to be made to the
  original contract. This makes updating the original contract a breeze.

- **Multiple Configuration Provider**

  Mokapi can read configuration and scripts from [File](/docs/configuration/dynamic/file.md), [HTTP](/docs/configuration/dynamic/http.md), [GIT](/docs/configuration/dynamic/git.md), [NPM](/docs/configuration/dynamic/npm.md). 
  Every provider can be customized.

- **Run Everywhere**
  
  Use the Docker image, Chocolatey on Windows, the NPM package or directly as an executable.

- **Dashboard**

  Quickly analyze and inspect all requests and responses in the dashboard to gather insights on how your mocked APIs are used.

## Get started

- [Installation](/docs/guides/get-started/installation.md)
- [Running](/docs/guides/get-started/running.md)
- [Test Data](/docs/guides/get-started/test-data.md)
- [Dashboard](/docs/guides/get-started/dashboard.md)