---
title: "Mokapi Documentation: Getting Started with API Mocking"
description: Learn how to set up Mokapi to mock APIs and validate requests using OpenAPI or AsyncAPI. No account needed—free, open-source, and easy to use.
cards:
  items:
    - title: Run your first mocked REST API
      href: /docs/get-started/running
      description: Quickly set up your first mock REST API and view simulated live traffic inside the Mokapi dashboard.
    - title: Using Mokapi
      href: /docs/configuration/overview
      description: Get a complete overview of core features, configuration options, and how to patch configuration providers effectively.
    - title: Mock Event-Driven APIs with Apache Kafka
      href: /docs/kafka/quick-start
      description: Learn how to mock Kafka topics and simulate event-driven architectures for realistic asynchronous testing.
    - title: Mock MQTT Broker & Topics
      href: /docs/mqtt/overview
      description: Learn how to mock MQTT brokers, topics, and pub/sub messages seamlessly from your AsyncAPI specifications.
    - title: Mokapi JavaScript API
      href: /docs/javascript-api/overview
      description: Discover how to control, script, and customize your mocked APIs programmatically using JavaScript.
    - title: Random Data Generator
      href: /docs/get-started/test-data
      description: Explore the built-in random data engine to dynamically generate realistic payloads for your API testing.
    - title: Mokapi Dashboard
      href: /docs/get-started/dashboard
      description: Dive into the real-time UI to analyze, monitor, and debug requests and responses instantly.
---

# Mocking APIs with Mokapi

**Welcome to Mokapi!**

[Mokapi](https://mokapi.io) makes it easy to build, test, and monitor API-driven applications without infrastructure
dependencies. The tool supports high-performance HTTP mocking and local Kafka simulation directly from your OpenAPI
and AsyncAPI specifications.

> *Think of Mokapi as your ever-reliable API contract guardian—lightweight, transparent, and specification-driven.*

## Build Better Software with Mokapi

In today's world, modern applications rely on multiple external APIs. When these dependencies are slow, unreliable,
or unavailable, your development velocity drops. Mokapi eliminates these obstacles, enabling you to:

- **Develop Faster:** Eliminate bottlenecks by coding independently of external upstream systems.
- **Test with Confidence:** Simulate realistic API behaviors, validation errors, and edge cases.
- **Automate Your Pipelines:** Enhance CI/CD reliability with consistent, zero-dependency mock responses.
- **Ensure Compliance:** Seamlessly integrate tools like Dependabot or Renovate for future-proof dependencies.

## Key Features

- **Spec-Driven Mocking:**  
  Quickly create OpenAPI or AsyncAPI mock servers for REST, SOAP, Kafka, and MQTT event-driven architectures
- **No-Code Configuration:**   
  Jump right in, no complex coding required!
- **Live Monitoring:**  
  Use the [Mokapi Dashboard](/docs/get-started/dashboard) to track requests and responses in real-time, simplifying your debugging process.
- **Dynamic Test Data:**  
  Utilize the built-in random data generator to create realistic payloads tailored to your schema rules
- **Local & Secure:**  
  An offline-first tool, Mokapi keeps your data on your machine—no accounts, no cloud syncs, and no telemetry tracking.
- **Open Source:**  
  Free and transparent, explore Mokapi's source code directly on [GitHub](https://github.com/marle3003/mokapi).</p>

## Data Privacy and Security

Mokapi prioritizes your privacy:

- **Local Control:** Your data remains entirely on your machine at all times.
- **No Account Required:** Enjoy a hassle-free setup without the need for user accounts, credentials, or authentication.
- **No Cloud Connectivity:** Mokapi operates fully offline. Your API specifications, mock configurations, and payloads never leave your local environment.

## Explore Mokapi Documentation Guides

Whether you're mocking APIs for local testing or validating async event-driven
systems, select a guide below to streamline your workflow:

{{ card-grid key="cards" }}
