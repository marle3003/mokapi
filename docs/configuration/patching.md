---
title: Use Patches to add or override configurations.
description: Patch configuration changes for mocking needs, rather than changing the original contract.
---
# Configuration Patching

Mokapi allows you to modify configurations without altering the original files. This process is achieved through 
the use of patch files, which can add or modify configurations like server URLs or authentication tokens.
Patches are applied in alphabetical order based only on the filenames, not the full file path.

## Example Scenario

If you retrieve an OpenAPI spec from a library, but need to add a staging server, you can create a patch file 
that only contains the server URL change.

```yaml tab="Base Config"
openapi: 3.0.3
info:
  title: Swagger Petstore - OpenAPI 3.0
servers:
  - url: https://petstore3.swagger.io/api/v3
```

```yaml tab="Patch Config"
openapi: 3.0.3
info:
  title: Swagger Petstore - OpenAPI 3.0
servers:
  - url: http://localhost/petstore
```

After merging these configuration files, we get the following as result:

```yaml tab=Result
openapi: 3.0.3
info:
  title: Swagger Petstore - OpenAPI 3.0
servers:
  - url: https://petstore3.swagger.io/api/v3
  - url: http://localhost/petstore
```

This approach ensures that you can update configurations without modifying the original, keeping your integrations clean and easy to manage.

When Mokapi applies the changes, it tries to match each element to an element
in the existing configuration. If there is a match, Mokapi updates the existing element. If there is no match, Mokapi inserts a new element.

``` box=tip
Mokapi patches any configuration file even if the source comes from a different provider.
```

## Patching Discriminators

| Type     | Discriminator |
|----------|---------------|
| OpenAPI  | *info.title*  |
| AsyncAPI | *info.title*  |
| SMTP     | *info.title*  |
| LDAP     | *info.title*  |

## Benefits

- <p><strong>Update Without Disruption:</strong><br />Easily upgrade to new versions of configurations while preserving your custom changes.</p>
- <p><strong>Version Control Friendly:</strong><br />Simplifies collaboration by allowing multiple users to apply patches without conflicts.