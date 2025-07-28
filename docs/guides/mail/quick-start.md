---
title: SMTP Quick Start
description: Send and inspect fake emails locally using Mokapi as a mock SMTP server.
---
# SMTP Quick Start

Quickly get up and running with Mokapi as a fake SMTP server for testing and development — no real emails are sent,
and no risk of spamming real users.

## 1. Define a Mail Server Configuration

Start by creating a configuration file named `smtp.yaml`. This file defines a basic mock SMTP server.

```yaml
mail: '1.0'
info:
  title: Mokapi's Mail Server
servers:
  smtp:
    host: localhost:25
    protocol: smtp
```

This configuration instructs Mokapi to start an SMTP server on port 25.

``` box=tip
You can use any available port if 25 requires elevated privileges on your OS, e.g., :2525.
```

## 2. Start Mokapi

Run Mokapi with your configuration:

```
mokapi smtp.yaml
```

Mokapi will start both:
- The fake SMTP server (as defined in your config)
- The Dashboard at [http://localhost:8080](http://localhost:8080)

You can now send test emails and inspect them in real time through the web interface.

## 3. Send a Test Email from Your App

```c#
using System.Net.Mail;

string to = "alice@mokapi.io";
string from = "bob@mokapi.io";
string subject = "Using the new SMTP client.";
string body = "Using Mokapi SMTP server, you can send an email message from any application very easily.";

MailMessage message = new(from, to, subject, body);

using SmtpClient client = new SmtpClient("127.0.0.1", 25);
client.Send(message);
```

## 4. View the Message in the Dashboard

After sending the email, visit [http://localhost:8080](http://localhost:8080).
You’ll see the received message under Mail → Messages, including its full content, headers, and metadata.

## What's Next?

- [Test email workflows with Playwright and Mokapi](/docs/resources/blogs/testing-email-workflows-with-playwright-and-mokapi)
- [Add recipient rules](/docs/guides/mail/rules.md) to allow or deny specific domains
- [Patch the config](/docs/configuration/patching.md) to test different scenarios

> Mokapi gives you full control over your mail simulation environment — ideal for CI pipelines, 
> development, or demos.
