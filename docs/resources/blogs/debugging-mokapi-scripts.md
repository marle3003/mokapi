---
title: Debugging Mokapi JavaScript
description: Learn how to debug your JavaScript code inside Mokapi using console.log, console.error, and event handler tracing.
---

# Debugging Mokapi JavaScript

Mokapi makes it easy not only to mock APIs, LDAP servers, Kafka topics, and mail servers, but also to 
customize behavior using JavaScript. To help you debug your JavaScript code inside Mokapi, it provides
simple but powerful tools: `console.log`, `console.error`, and advanced event handler tracing.

This article will walk you through how to effectively debug your JavaScript running in Mokapi.

## Console Output: `console.log` and `console.error`

Mokapi provides the familiar `console.log` and `console.error` functions that you know from the browser or Node.js. 
These functions write directly to the Mokapi console output.

Use `console.log` to output general debug information:
```javascript
console.log('Hello World');
```

Use `console.error` to report errors or unexpected behavior:
```javascript
console.error({ firstname:"John", lastname:"Doe" });
```

Everything you log will appear in the terminal or system where you run Mokapi, making 
it easy to follow what your scripts are doing behind the scenes.

``` box=tip
There is also a function console.warn and console.debug
```

## Event Handler Tracing

Sometimes you need even deeper insights into how Mokapi processes a request — especially when you 
are using event handlers to customize behavior.

If your event handler function returns `true`, Mokapi not only executes the event but also:

- Logs the event handler data such as file, event type, event parameters 
- Displays it in the Mokapi Dashboard under the request details

This makes it very easy to see which handlers were involved in a request and in what order.

Here's an example:
```javascript
import { on } from 'mokapi';

export default () => {
    on("http", (req, res) => {
        console.log("Handling request for:", req.path);

        if (req.path.startsWith("/admin")) {
            console.log("Admin path detected");
            return true; // Trace this handler
        }
        return false;
    });
}
```

In this case, because the handler returns true, Mokapi will:

- Log that the onRequest handler was triggered 
- List the handler in the dashboard under the request details

This visibility is incredibly helpful when debugging complex request flows, especially if you have 
multiple handlers reacting to a single request.

<img src="/dashboard-event-handler-debugging.png" alt="Mokapi dashboard displaying event handler logs for a specific HTTP request." class="image shadow">

## Best Practices for Debugging

- Use console.log freely during development to trace values and decisions inside your scripts. 
- Return true from important event handlers you want to trace in the dashboard. 
- Use console.error for anything unexpected to separate normal logs from problems. 
- Check the Dashboard for a clear overview of all event handlers triggered by a request.

## Conclusion

Debugging Mokapi JavaScript is straightforward with console.log, console.error, and event handler 
tracing. Whether you're building complex mock logic or just getting started, these tools will help you 
understand exactly what happens inside Mokapi when it processes your requests.

Give it a try — and happy debugging!