---
title: Configure your Mokapi with powerful configuration providers
description: Learn how Mokapi gathers configurations and scripts using powerful providers. Customize your API mocks dynamically for flexible testing and development.
---

# What Is Dynamic Configuration?

Dynamic configuration in Mokapi allows you to modify mock services, API behaviors, and test scenarios 
in real-time without restarting the application. This capability is essential for agile 
development and continuous integration workflows, enabling rapid iteration and testing of various 
API responses.

## Features

- **Live Updates:** Modify configurations on-the-fly to simulate different API behaviors. 
- **Multi-Source Support:** Load configurations from various sources such as files, HTTP endpoints, Git repositories, or NPM packages. 
- **JavaScript Integration:** Utilize embedded JavaScript to define dynamic behaviors, including conditional responses and simulated delays. 
- **Patch-Based Changes:** Apply changes using patch configurations, preserving the original contract while customizing behaviors. 
- **Dashboard Monitoring:** Visualize and manage configurations through an intuitive web interface.​

## Configuration Sources

Mokapi supports multiple configuration sources, providing flexibility in how you manage and deploy your mock services:

- [**File System:**](/docs/configuration/dynamic/file.md) Store configurations locally for quick access and version control.
- [**HTTP:**]((/docs/configuration/dynamic/http.md)) Fetch configurations from remote servers, facilitating centralized management. 
- [**Git:**](/docs/configuration/dynamic/git.md) Integrate with Git to leverage version control and collaborative workflows. 
- [**NPM Packages:**]((/docs/configuration/dynamic/npm.md)) Distribute and manage configurations as NPM packages for consistency across projects.

## Best Practices

- **Use Version Control:** Store your configuration files in a version control system like Git to track changes and collaborate effectively. 
- **Modularize Configurations:** Break down configurations into modular components for reusability and easier maintenance. 
- **Validate Configurations:** Regularly validate your configurations to ensure they meet the expected schema and behavior. 
- **Monitor Changes:** Utilize Mokapi's dashboard to monitor configuration changes and their impact on mock services.