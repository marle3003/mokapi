# Configuration

Mokapi has two types of configuration:
- The startup configuration (referred as the static configuration)
- The dynamic configuration (e.g. OpenApi configuration)

Static configuration sets up the providers which reads dynamic 
configurations. Dynamic configurations contains mock service like 
a REST API or a Kafka Topic. This configuration can change without 
restarting of Mokapi

## Static Configuration

Elements in the static configuration don't often change and changes require a restart of the application.

There are three different ways to define a static configuration options.
1. Configuration file
2. Command-line arguments
3. Environment variables

These ways are evaluated in the order listed above.

### Configuration File

At first, Mokapi searches for static configuration in a file in:

1. ./mokapi.yaml
2. ./mokapi.yml
3. /etc/mokapi.yaml
4. /etc/mokapi.yml

You can override this:

```bash tab=CLI
mokapi --configFile=/foo/mokapi.yaml
```
```bash tab=Env
MOKAPI_CONFIGFILE=/foo/mokapi.yaml
```

### CLI Arguments

A list of available arguments can be found [here](/docs/configuration/reference.md)

### Environment Variables

A list of available environment variables can be found [here](/docs/configuration/reference.md)

## Dynamic Configuration

Mokapi reads the dynamic configuration using providers. You can find more information about providers in the next pages.

- [File](/docs/configuration/providers/file.md)
- [HTTP](/docs/configuration/providers/http.md)
- [GIT](/docs/configuration/providers/git.md)