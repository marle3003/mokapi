---
title: How to mock LDAP with Mokapi
description: Integrate your App with a mock LDAP server
cards:
  items:
    - title: Run your first mocked LDAP
      href: /docs/guides/ldap/quick-start
      description: Learn how to quickly set up and run your first LDAP mock and use ldapsearch tool
    - title: Mock LDAP Authentication in Node.js
      href: /docs/examples/tutorials/mock-ldap-authentication-in-node
      description: Learn how to mock LDAP authentication using Mokapi and a Node.js backend. Step-by-step guide with code examples for testing LDAP login without a real server!
    - title: Mocking LDAP using Group Permission
      href: /docs/examples/tutorials/mock-ldap-group-permission-in-node
      description: Learn how to mock LDAP authentication and group permission using a Node.js backend.
---

# Mocking LDAP Server

## Overview

Mokapi provides a powerful and flexible solution for mocking LDAP servers, allowing developers to test authentication, 
directory services, and user management without relying on a real LDAP setup. This streamlines development, simplifies 
CI/CD workflows, and removes dependencies on external LDAP environments.

With Mokapi, you can quickly create an LDAP mock server that even supports write operations, to validate query accuracy, 
simulate user authentication, and test directory structures. Whether you're debugging LDAP interactions, experimenting 
with access control, or integrating LDAP into your applications, Mokapi offers an intuitive and developer-friendly 
experience.

## Why Use Mokapi for LDAP Mocking?

Mokapi smoothly integrates into your existing codebases and testing pipelines, enabling you to test different LDAP 
configurations without setting up a full LDAP infrastructure.

Key Features:

- <p><strong>Mock LDAP Authentication:</strong><br /> Simulate user authentication with customizable credentials.</p>
- <p><strong>LDIF Import Support:</strong><br /> Load bulk user and group data using LDIF files.</p>
- <p><strong>Hot-Reloading Configuration:</strong><br /> Update LDAP settings without restarting the server.</p>
- <p><strong>CI/CD Integration:</strong><br /> Incorporate into automated pipelines to validate authentication and directory-based workflows.</p>
- <p><strong>Write Operation Support:</strong><br />Making it a more dynamic testing solution.</p>

Supported Operations:

- Bind - Supports only simple authentication.
- Search - Includes paging for large datasets.
- Add - Create new LDAP entries.
- Modify - Update existing LDAP entries.
- Delete - Remove LDAP entries.
- ModifyDN - Rename, copy or move an LDAP entry.
- Compare - Check attribute values against directory entries.

{{ card-grid key="cards" }}