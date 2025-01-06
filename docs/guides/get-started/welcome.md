---
title: Mocking APIs with Mokapi
description: These guides covers everything you need to know about Mokapi and mocking APIs.
cards:
  items:
    - title: Run your first mocked REST API
      href: /docs/guides/get-started/running
      description: Learn how to run your first mock and view results in the dashboard.
    - title: Using Mokapi
      href: /docs/configuration/configuration/introduction
      description: Learn about Mokapi options, patching and the concept of configuration providers.
    - title: Mock event-driven API based on Apache Kafka
      href: /docs/guides/kafka/quick-start
      description: Discover how you can mock a Kafka topic with Mokapi and simulate realistic events.
    - title: Mokapi JavaScript API
      href: /docs/javascript-api/javascript-api/overview
      description: Learn how you can control the behavior of your mocked APIs with the Mokapi APIs
    - title: Random Data Generator
      href: /docs/guides/get-started/test-data
      description: Explore how Mokapi generates random test data and how you can influence it.
    - title: Mokapi Dashboard
      href: /docs/guides/get-started/dashboard
      description: Discover how to analyze and review APIs and all requests and their responses.
---

# Mocking APIs with Mokapi

Mokapi is an open-source and developer-friendly mocking API tool. Mokapi allows you to
**prototype**, **test** and **demonstrate** APIs in software solutions using realistic data and scenarios.

## Overview

Using Mokapi, you can quickly and easily test various scenarios of your APIs without
having to rely on a fully functional API.

Mokapi helps Agile, DevOps and Continuous Deployment teams prevent bugs or errors in production. 
This enables them to build resilient, reliable and error-free communication between applications.

Mokapi provides you with the following key features:

- ### Multiple Protocol support

  Mokapi can mock APIs for [HTTP(S)](/docs/guides/http/overview.md), [Apache Kafka](/docs/guides/kafka/overview.md), [SMTP](/docs/guides/smtp/overview.md) and [LDAP](/docs/guides/ldap/overview.md)
  With Mokapi HTTP you can mock any REST or SOAP API.

- ### Everything as Code

  Reusing, version control, consistency and integrate mocks with your CI/CD.

- ### An embedded JavaScript engine

  Write the behavior of your mock in JavaScript and control everything: status, headers, delays, errors or other edge cases.

- ### Configuration Options

  Use CLI arguments, environment variable or file configuration and take advantage from hot-reloading of dynamic configurations.

- ### Patch Configuration

  Mokapi supports configuration patch files so that adjustments required for the mock do not have to be made to the
  original contract. This makes updating the original contract a breeze.

- ### Multiple Configuration Provider

  Mokapi can read configuration and scripts from [File](/docs/configuration/dynamic/file.md), [HTTP](/docs/configuration/dynamic/http.md), [GIT](/docs/configuration/dynamic/git.md), [NPM](/docs/configuration/dynamic/npm.md). 
  Every provider can be customized.

- ### Run Everywhere

  Mokapi can be installed via packages, used in a Docker container or as a standalone binary.

- ### Dashboard

  Quickly analyze and inspect all requests and responses in the dashboard to gather insights on how your mocked APIs are used.

## Explore how you can mock your APIs with Mokapi

{{ card-grid key="cards" }}
