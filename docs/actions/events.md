# Events

With events, you can configure your workflows to run when specific activity happens or at a scheduled time. A workflow can be run for one or more events using the *on* syntax.

## Using a single event

```yaml
# Triggered when any HTTP request is received to any url
on: http
```

## Using a single HTTP event with HTTP method configuration

```yaml
# Triggered when any HTTP POST request is received to any url
on: 
  http: post
```

## Using multiple events
```yaml
on: 
  http:
    post: /example
    get: [/example, /health]
```

## Using multiple events with wildcard
```yaml
on: 
  http:
    post: /example
    get: [/example, /health/*]
```

## Scheduled events
```yaml
# Triggered every 5 seconds
on:
  schedule:
    every: 5s
```

## Scheduled events with limited iterations
```yaml
# Triggered every 5 seconds but only 10 times
on:
  schedule:
    every: 5s
    iterations: 10
```

## Using SMTP events
```yaml
# Triggered every 5 seconds
on:
  smtp: login, logout, received
```

## Using SMTP events on specified host
```yaml
# Triggered every 5 seconds
on:
  smtp:
    received: true
    address: localhost:25
```

## Filter pattern

You can use special characters in path filters

- `*`: Matches zero or more characters, but does not match the `/` character
- `**`: Matches zero or more of any character