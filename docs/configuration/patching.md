---
title: Use Patches to add or override configurations.
description: Patch configuration changes for mocking needs, rather than changing the original contract.
---
# Configuration Patching

You use patch files to add or change dynamic configurations. 
Mokapi merges files to create the configuration that is used at runtime.
Mokapi merges configuration with the same *info-name*. 
Patches are processed in the alphabetical order only of the filenames.

This can be useful if you are retrieving an API definition from a library, and you need an additional
server URL.
In the example below Mokapi will merge these configurations

```yaml tab=petstore.yaml
openapi: 3.0.3
info:
  title: Swagger Petstore - OpenAPI 3.0
servers:
  - url: https://petstore3.swagger.io/api/v3
```
```yaml tab=petstore-patch.yaml
openapi: 3.0.3
info:
  title: Swagger Petstore - OpenAPI 3.0
servers:
  - url: http://localhost/petstore
```

When Mokapi applies the changes, it tries to match each element to an element
in the existing configuration. If there is a match, Mokapi updates the existing element. If there is no match, Mokapi inserts a new element.

``` box=tip
Mokapi patches any configuration file even if the source comes from a different provider.
```

## Patching Result:

After merging these configuration files, we get the following as result:

```yaml
openapi: 3.0.3
info:
  title: Swagger Petstore - OpenAPI 3.0
servers:
  - url: https://petstore3.swagger.io/api/v3
  - url: http://localhost/petstore
```

## Patching Discriminators

| Type     | Discriminator |
|----------|---------------|
| OpenAPI  | *info.title*  |
| AsyncAPI | *info.title*  |
| SMTP     | *info.title*  |
| LDAP     | *info.title*  |
