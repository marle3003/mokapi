---
title: EventArgs
description: EventArgs is an object used to configure event handlers registered with the on function.
---
# EventArgs

`EventArgs` is an optional configuration object passed to the
[`on`](/docs/javascript-api/mokapi/on.md) function when registering an event handler.
It allows controlling how and when an event handler is executed.

<div class="table-responsive-sm">
<table>
  <thead>
    <tr>
      <th scope="col" class="col-2">Name</th>
      <th scope="col" class="col-2">Type</th>
      <th scope="col" class="col">Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>tags</td>
      <td><code>object</code></td>
      <td>Adds or overrides existing tags that are used in dashboard.</td>
    </tr>
    <tr>
      <td>track</td>
      <td><code>boolean | <br />(params) => boolean</code></td>
      <td>Controls whether this event handler is tracked in the dashboard.</td>
    </tr>
    <tr>
      <td>priority</td>
      <td><code>integer</code></td>
      <td>Defines the execution priority of the event handler. Handlers with a higher value are executed first. If no priority is specified, the default priority is <code>0</code>.</td>
    </tr>
  </tbody>
</table>
</div>

## Example: Adding custom tags

The following example registers an event handler and adds a custom tag.

```javascript
import { every } from 'mokapi'

export default function() {
    on('1m', function(request, response) {
        // handler logic
    }, { tags: { foo: 'bar' } })
}
```

## Example: Controlling whether an event handler is tracked in the dashboard

The track field controls whether executions of an event handler appear in the dashboard.

It supports two modes:
- true / false — always track or never track 
- a function — decide dynamically per request

### Static tracking

```javascript
on('http', handler, { track: true })
```

### Dynamic tracking (function)

```javascript
on('http', handler, {
    track: (request, response) => request.key !== '/health'
})
```

### Full example: Delay simulation with tracking

```javascript
import { on, sleep } from 'mokapi'

export default () => {
    let delay = undefined
    
    on('http', (request, response) => {
        if (request.key === '/simulations/delay') {
            switch (request.method) {
                case 'PUT': 
                    delay = request.query.duration;
                    break;
                case 'DELETE':
                    delay = undefined;
                    break;
            }
        }
    }, { track: true })
    on('http', (request, response) => {
        if (!delay) {
            sleep(delay);
        }
    }, { track: () => delay !== undefined } )
}
```

#### Explanation
- The first handler exposes a control endpoint to configure the delay. 
  - track: true ensures all configuration changes are visible in the dashboard. 
- The second handler applies the delay to incoming requests. 
  - Tracking is enabled only while a delay is active, keeping the dashboard clean when no simulation is running.

## Example: Controlling execution order with priority

When multiple handlers are registered for the same event, the priority
property controls the order in which they are executed.

```javascript
import { on } from 'mokapi'

export default function() {
    on('http', (req, res) => {
        res.data.stage = 'early'
    }, { priority: 10 })

    on('http', (req, res) => {
        res.data.stage = 'default'
    })

    on('http', (req, res) => {
        res.data.stage = 'late'
    }, { priority: -10 })
}
```

In this example:
- The handler with priority 10 is executed first 
- The handler without a priority is executed next (priority 0)
- The handler with priority -10 is executed last