---
title: Building a Virtual Hardware Simulation with Mokapi and JavaScript
description: Dynamic mocks simulate reality. Here's how to use Mokapi's JavaScript API to replace physical hardware with a fully controllable simulation.
subtitle: This article uses a parcel terminal as a case study to show what's actually possible when you combine Mokapi's spec validation with its JavaScript API. The terminal is just the example. The point is what you can build.
---

# Building a Virtual Hardware Simulation with Mokapi and JavaScript

## The Problem With Hardware in Your Test Environment

Imagine you're building a mobile app that controls a parcel terminal. A user opens the app, selects a
compartment, and the terminal unlocks the door. They put a parcel inside, close the door, and a sensor
detects something is inside. Simple enough flow.

But testing it is a nightmare.

Take the most basic scenario: your test unlocks a compartment. Now the door is open. For the next test to
run cleanly, that door needs to be closed, but the terminal only allows one door open at a time. So if the
door stays open, every subsequent test that tries to unlock a different compartment will fail. And the door
won't close on its own. Someone has to physically stand in front of the terminal and close it between
test runs.

Or take the sensor. It detects whether something is inside the compartment. The mobile app uses that to warn
the user if they try to close an empty compartment: "Are you sure? It looks like nothing was put inside."
To test that flow, you need a compartment with nothing in it and one with something in it, on demand,
reproducibly. With real hardware, that means someone physically placing and removing items between every
test run.

And then there are the edge cases that are nearly impossible to stage reliably. What happens when a user
tries to unlock a compartment but another door is already open? What does the app show? Does it block the
action or just warn? To test that, you need to get the terminal into a specific state before the test runs,
every time, consistently. With real hardware, that's fragile at best.

So most teams end up testing the happy path and hoping the edge cases work out. Or they write tests that
only run in one specific environment with a carefully prepared terminal. Neither scales.

What you actually need is a virtual terminal: a simulation that behaves like the real hardware, that any
developer or tester can run locally or in CI, and that lets you set up any scenario you need without touching
a physical device.

---

## What the Simulation Needs to Do

The mobile app talks to a backend for the terminal, which in turn calls the terminal's own API. To simulate
the full system, you need two layers.

The first is the real terminal API, mocked and spec-validated. The mobile app and backend should talk to the
simulation exactly as they would talk to real hardware. No special test mode, no code changes.

The second is a simulation API: a set of additional endpoints that have no equivalent in the real terminal's
API, because they're replacing physical actions. There's no "close door" endpoint in the real API because
closing the door is something a person does physically. In the simulation, however, a tester must be able to
close the door. The simulation API is how you replace the person standing at the terminal.

Mokapi gives you the building blocks for both. Define your endpoints in OpenAPI specs, write JavaScript
handlers to control behavior, and Mokapi handles spec validation automatically. Every request and response
gets validated against the contract. You write the logic. Mokapi enforces the rules.

---

## Building the Terminal API Handler

The real terminal API is defined in an OpenAPI spec. The JavaScript handler intercepts incoming requests and returns
the right state based on what's being asked.

```typescript
import { on } from 'mokapi'

const terminals = [
  {
    terminalId: 'terminal-1',
    compartments: [
      { compartmentId: 'c1', doorState: 'closed', sensorEnabled: true, occupied: false },
      { compartmentId: 'c2', doorState: 'closed', sensorEnabled: true, occupied: false },
    ]
  }
]

export default function() {
  on('http', (request, response) => {
    switch (request.key) {
      case '/terminals/{terminalId}':
        response.data = terminals.find(
          x => x.terminalId === request.path.terminalId
        )
        break

      case '/terminals/{terminalId}/compartments':
        const terminal = terminals.find(
          x => x.terminalId === request.path.terminalId
        )
        response.data = terminal?.compartments ?? []
        break

      case '/terminals/{terminalId}/compartments/{compartmentId}/unlock':
        const t = terminals.find(x => x.terminalId === request.path.terminalId)
        const compartment = t?.compartments.find(
          x => x.compartmentId === request.path.compartmentId
        )

        // Only one door can be open at a time
        const anyOpen = t?.compartments.some(c => c.doorState === 'open')
        if (anyOpen) {
          response.statusCode = 409
          response.data = { message: 'Another compartment is already open' }
          return
        }

        if (!compartment?.sensorEnabled) {
          response.statusCode = 422
          response.data = { message: 'Sensor is disabled' }
          return
        }

        compartment.doorState = 'open'
        response.data = { doorState: 'open' }
        break
    }
  })
}
```

This is already more interesting than a static mock. The "only one door can be open at a time" rule is
real business logic, enforced in the simulation. A tester can now write a test that unlocks one compartment
and then tries to unlock another, and the simulation will return a 409 exactly as the real terminal would.

---

## Adding the Simulation API

The simulation API is defined in a separate OpenAPI spec. These endpoints don't exist in the real
terminal's API. They're purely for testing: a way for developers and testers to trigger the physical
actions that can't be called over HTTP in real life.

```typescript
import { on } from 'mokapi'

// terminals is defined in the same file or imported
// If split across files, use shared() instead (more on that below)

export default function() {
  on('http', (request, response) => {
    switch (request.key) {
      case '/simulation/terminals/{terminalId}/compartments/{compartmentId}':
        const terminal = terminals.find(
          x => x.terminalId === request.path.terminalId
        )
        const compartment = terminal?.compartments.find(
          x => x.compartmentId === request.path.compartmentId
        )

        if (!compartment) {
          response.statusCode = 404
          return
        }

        if (request.method === 'PATCH') {
          // Simulate physical actions: close the door, toggle the sensor
          if (request.body.doorState !== undefined) {
            compartment.doorState = request.body.doorState
          }
          if (request.body.sensorEnabled !== undefined) {
            compartment.sensorEnabled = request.body.sensorEnabled
          }
          if (request.body.occupied !== undefined) {
            compartment.occupied = request.body.occupied
          }
          response.data = compartment
        }
        break
    }
  })
}
```

A tester can now `PATCH /simulation/terminals/terminal-1/compartments/c1` with `{ "doorState": "closed" }` 
to simulate someone physically close the door. Or `{ "sensorEnabled": false }` to simulate a faulty
sensor. These aren't real API calls that would ever happen in production. They're test controls that replace
physical actions.

This is what makes the simulation really useful for edge cases. Want to test what the mobile app does
when a door is already open? PATCH it open before the test runs. Want to test the disabled sensor path?
PATCH it disabled. No physical terminal needed. No carefully staged hardware setup.

---

## Sharing State Across Files

If you keep everything in one file, a plain JavaScript variable holds state just fine. The terminals array
in the examples above is shared naturally because it's all in the same module.

But as your simulation grows, you'll probably want to split the real API handler and the simulation API
handler into separate files for readability. That's where Mokapi's `shared` API comes in. Each JavaScript
file in Mokapi runs in its own runtime, so a variable defined in one file isn't visible in another.
`shared` gives you a way to store values in Mokapi's internal shared storage so all your handlers can
access the same state.

```typescript
import { on, shared } from 'mokapi'

// Initialize shared state, or use the existing value if already set
const terminals = shared.update('terminals', (current) => current || [
  {
    terminalId: 'terminal-1',
    compartments: [
      { compartmentId: 'c1', doorState: 'closed', sensorEnabled: true, occupied: false },
      { compartmentId: 'c2', doorState: 'closed', sensorEnabled: true, occupied: false },
    ]
  }
])

export default function() {
  on('http', (request, response) => {
    // handler code here, same as before
    // terminals now refers to the shared state
  })
}
```

Both the terminal API handler and the simulation API handler can import from `shared`, and they'll be
working with the same underlying state. When the simulation handler closes a door, the terminal API
handler immediately sees the updated state.

---

## Going Further: A Visual Frontend for the Simulation

Here's where it gets interesting beyond automated tests.

A developer can build a simulation frontend using any web framework on top of Mokapi. Instead of calling
PATCH endpoints directly, testers and stakeholders get a visual interface: buttons to open and close
compartments, toggles to disable sensors, a live view of which compartments are occupied. Non-technical
users can interact with the simulation without knowing anything about HTTP requests.

Think about what that unlocks for a team. A stakeholder installs the mobile app on their own device and
configures it to point at the virtual terminal instead of a real one. A product manager can run through
a demo independently without a developer present. A UX review happens against realistic behavior, not a
static prototype. A tester can set up a specific scenario visually and then hand it off to someone else
to verify the app behavior.

The simulation frontend and Mokapi run together in the same Docker container, published to Kubernetes.
Everyone on the team gets a URL. The simulation is always available, always in a known state, and anyone
can reset it by triggering the right simulation endpoint.

And because the simulation frontend talks to Mokapi through the same API, its requests get validated too.
The simulation frontend is held to the same contract as the mobile app. If the frontend sends a malformed
request, Mokapi catches it. Nothing bypasses the spec.

Multiple users can run the simulation independently by scoping state to a session or user identifier.
Each stakeholder gets their own terminal state. One person's demo doesn't interfere with another person's
review. A developer can handle this in the JavaScript handlers by keying state off a request header or
query parameter.

This is what "own imagination is the only limit" actually means in practice. Mokapi provides the spec
validation and the JavaScript API. What you build on top of it, whether that's an automated test suite,
a visual simulation frontend, or a full stakeholder demo environment, is entirely up to you.

---

## Why Not Just Build Your Own Simulation Server?

Here's the thing about building a simulation with Mokapi versus writing one from scratch.

When you write a custom simulation server, you're essentially reimplementing the terminal's API in your
own code. Every endpoint is an HTTP controller you wrote. Every parameter is something you parsed and
validated yourself. When the terminal backend team ships a new API version, you open up the diff, compare
it against your controllers, find every endpoint that changed, update your handlers, and hope you didn't
miss anything.

With Mokapi, you swap the OpenAPI spec. That's it. Mokapi's validation immediately tells you what changed
and what's now broken. If a new required field was added to a request body, any test that doesn't send it
will fail with a clear validation error. If a response schema changed, any handler that returns the old
shape will fail before it reaches the client. You don't need to audit your code against the documentation.
The spec is the documentation, and Mokapi enforces it.

---

## What This Unlocks

The parcel terminal is one example, but the pattern applies anywhere physical reality is part of your system.
Payment terminals that need a card to be physically tapped. IoT sensors that need environmental conditions
to trigger. Medical devices that need a patient to be connected. Industrial equipment where certain states
can only be reached through a sequence of physical operations.

In all of these cases, you have the same problem: you can't reliably stage the conditions you need to test
in an automated environment. And in all of these cases, the same approach works. Define the real API in
a spec. Write JavaScript handlers that simulate the behavior. Add simulation endpoints for the physical
actions. Let Mokapi enforce the contract across all of it.

You're not building a perfect replica of the hardware. You're building enough of a simulation to make your
tests meaningful, fast, and reproducible. The spec keeps the simulation honest as the real API evolves.

That's what Mokapi's JavaScript API is actually for. Not to return a user object, but to develop a solution that
allows your entire team to develop and test without requiring the physical world.