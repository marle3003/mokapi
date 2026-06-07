---
title: "Explore and Debug APIs with AI: Mokapi's MCP Server"
description: Learn how Mokapi's MCP server lets your AI agent explore OpenAPI specs, debug failed requests, and generate valid test data — all in plain English.
subtitle: API specs are meant to be understood, not just parsed. Mokapi's MCP server connects your AI agent directly to your mock server so you can ask plain questions and get real answers — about endpoints, errors, and test data. This article shows three concrete scenarios using GitHub Copilot and a live Mokapi instance.
---

# Your AI Agent Can Now Talk to Your API: Meet Mokapi's MCP Server

You know that moment when you're handed a new API spec and asked to integrate against it? Maybe it's an OpenAPI file with 50 endpoints, or an AsyncAPI document describing a Kafka setup you've never touched. And you're sitting there, scrolling through YAML, trying to figure out where to even start.

Or maybe you're on the other side of it: something's broken, your app is throwing errors, and you're clicking through a dashboard trying to find the one log entry that explains what went wrong.

Both of these feel harder than they should. And honestly, they are.

That's what makes Mokapi's new MCP server interesting. It doesn't just mock your APIs: it lets your AI agent actually *understand* them. Ask a plain question, get a useful answer. No scrolling, no hunting, no context switching.

Let me show you what that looks like in practice.

---

## First, a quick word on Mokapi

If you haven't used it before: Mokapi is an open-source tool for mocking APIs. HTTP, Kafka, SMTP, LDAP: you point it at a spec file and it spins up a mock server that behaves like the real thing. It's the kind of tool that makes local development and testing a lot less painful.

And now, alongside the mock server, it runs an MCP server too.

---

## What's MCP got to do with it?

MCP (Model Context Protocol) is a standard that lets AI agents connect to external tools. Think of it as a way for your AI assistant to reach outside the conversation and actually *do things*: query a database, check a calendar, or in this case, explore a running Mokapi instance.

What's clever about how Mokapi implements it is that the agent doesn't need to read your entire spec upfront. Instead of dumping thousands of lines of YAML into the AI's context (slow, expensive, often unnecessary), the agent queries only what it needs. It's surgical. You ask a question, the agent fetches just enough to answer it, and you get a response.

That's worth saying plainly: **it saves tokens, and it's faster**.

---

## Three things you can do right now

To make this concrete, I set up a simple Coffee Shop API (orders, payments, customers) and connected Copilot to a running Mokapi instance. Here's what happened.

### 1. Discover: "How do I pay for an order?"

This is the classic onboarding problem. You've got a spec, you need to call *something*, but you don't know what.

I asked Copilot: *"Ask Mokapi: how do I pay for an order?"*

Here's what came back:

> Mokapi shows a Coffee Shop API endpoint for paying an order: **POST /orders/{orderId}/pay**
>
> Send a JSON body with `method` (card, cash, or voucher), `amount`, and optionally `voucherCode` if paying by voucher.
>
> Example:
> ```bash
> curl -X POST "http://localhost/orders/ORDER_ID/pay" \
>   -H "Content-Type: application/json" \
>   -d '{"method":"card","amount":9.00}'
> ```
>
> Response: `200 { paidAt, success: true, transactionId }` or `400` on failure.

That's not just "here's the endpoint." That's a ready-to-run curl command with the exact payload shape. It pulled the accepted values for `method`, noted the optional `voucherCode`, showed the success response structure. All from a plain English question.

No scrolling. No guessing.

### 2. Observe: "What went wrong with my last request?"

I made a deliberate bad request: called `GET /orders` with `?date=foo` instead of a proper date format. Mokapi rejected it, as it should. Then I asked Copilot: *"Ask Mokapi: what went wrong with my last request?"*

The response:

> Last request failed: **GET /orders?date=foo** returned 400.
>
> Reason: query parameter `date` failed validation: `'foo'` is not a valid date (expected format YYYY-MM-DD).
>
> **Fix:** send a valid date, e.g. `?date=2026-06-05`

Think about what just happened there. The agent didn't just surface the error: it explained *why* it failed and told me exactly how to fix it. That's the kind of feedback that normally costs you five minutes of log hunting and doc reading.

And it works for Kafka too. If you're producing messages and something looks off, you can ask the same kind of question about your Kafka events. Same idea, different protocol.

### 3. Generate: "Give me a valid example payload for placing a new order"

This one's maybe the most immediately practical. You need test data, but you don't want to hand-craft JSON and guess at field formats.

I asked: *"Ask Mokapi: generate a valid example payload for placing a new order."*

```json
{
  "customerId": "92a964a8-7f2f-474b-b73c-b61aa417ea65",
  "items": [
    { "menuItemId": "67ba2612-b57b-431e-b864-bd1b29bb0126", "quantity": 66 },
    { "menuItemId": "05ce5a30-b949-4489-9a19-06546b6a28a9", "quantity": 74 }
  ],
  "notes": "J2KmLIwGukK"
}
```

Real UUIDs. Required fields all present. The `items` array has entries (the spec says `minItems: 1` and it respected that). The quantities are amusingly random, and the notes string is gibberish, but that's fine. It's *valid*. Copilot even added a note: adjust quantities and menuItemIds to realistic values if needed.

---

## Setting it up

Here's the thing: this takes about two minutes.

**Start Mokapi** pointing at your spec file or a URL. If you want to follow along with the examples in this article, you can use the Coffee Shop API directly:

```bash
npx go-mokapi https://raw.githubusercontent.com/marle3003/mokapi/refs/heads/main/examples/coffee-shop/api.yaml
```

That's it. Mokapi starts both the mock server and the MCP server together.

**Connect your agent.** In Copilot CLI, run `/mcp add` and follow the guided setup. It'll ask a few questions and write the config for you. The result looks like this:

```json
{
  "mcpServers": {
    "mokapi": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

Done.

**A note on permissions:** when Copilot first uses the Mokapi MCP server, it'll ask you to confirm before running any code. You'll see the tool name, the exact JavaScript it wants to execute, and three options: allow it once, allow it for all future calls to that tool in the current directory, or stop and tell Copilot to do something differently. It's a clean consent model and you always know exactly what's being run.

---

## The bigger idea

API mocking used to be about standing up a fake server and hoping for the best. You'd mock an endpoint, write some tests, and move on: never quite sure if your understanding of the spec matched reality.

What's shifting now is that the spec isn't just a document you read anymore. It's something you can *talk to*. Ask it questions. Poke at it. Get feedback from it in real time, in the same flow where you're already working.

That's a genuine change in how we explore and understand software systems. And the fact that you can get there with a single `npx` command and two minutes of setup: that's what keeps me motivated to develop Mokapi into a truly useful and easy-to-use mocking tool.

Try it out with your own specifications. You'll be surprised how Mokapi can help even in the development phase, and not just in CI pipelines.