---
title: How to Test Email Workflows End-to-End Without a Real Mail Server
description: Mock SMTP and IMAP with Mokapi and Playwright to test signup, verification, and password reset flows reliably.
subtitle: "Email is at the heart of most user flows: signup, verification, password reset. These flows are easy to break and hard to test. This guide shows how to mock a full SMTP and IMAP server with Mokapi and drive it end-to-end with Playwright, so your email logic gets the same test coverage as everything else."
icon: bi-envelope-paper
tech: mail
---

# How to Test Email Workflows End-to-End Without a Real Mail Server

## The Problem With Testing Email

Email is one of those things that's genuinely hard to test well. You can't just assert on a return value.
The email goes out through a third-party SMTP server, lands in a real inbox somewhere, and you have no
programmatic way to check what arrived.

So most teams end up doing one of two things. They test manually: sign up with a throwaway address, check the
inbox, click the link. Or they skip testing the email entirely and just verify that the send function was called.
Neither is great. Manual testing doesn't scale and doesn't belong in CI. And asserting that `sendEmail()`
was called tells you nothing about whether the subject line was right, the link was valid, or the HTML
rendered correctly.

That's why I built email support into Mokapi. The idea is the same as with every other protocol Mokapi supports:
replace the real infrastructure with a mock that your test can inspect. Your backend connects to Mokapi's SMTP
server exactly as it would a real mail server. Mokapi captures the message. Your test fetches it over HTTP and
asserts on the content.

Unlike Kafka or HTTP, there's no standard specification format for email configuration. So Mokapi uses a simple
YAML config file to define the mail servers you need. It's minimal by design, because there's not much to
configure: just declare the protocol, the host, and the port.

---

## The Scenario We're Testing

The workflow we're going to test is a classic signup flow:

1. A user fills in a signup form and submits it
2. The backend creates the account and sends a verification email via SMTP
3. The test retrieves the email from Mokapi and verifies the subject, sender, recipient, and body content including the verification link

This is a real end-to-end test. The frontend runs, the backend runs, the email goes through Mokapi's SMTP server,
and the test reads it back. No mocking inside the backend, no shortcuts.

---

## Step 1: Configure the Mock Mail Server

Because email doesn't have a specification format like AsyncAPI or OpenAPI, Mokapi uses a simple YAML config file.
Here's all you need to spin up a mock SMTP server. You can find the full config on [GitHub](https://github.com/marle3003/mokapi-email-workflow/blob/main/mocks/mail.yaml).

```yaml
mail: '1.0'
info:
  title: Email Workflows
servers:
  smtp:
    host: :2525
    protocol: smtp
```

That's it. Point your backend at `localhost:2525` instead of your real SMTP server, and Mokapi starts capturing everything.

If you also want IMAP access so you can preview emails in a real mail client during development, add it in the same file:

```yaml
mail: '1.0'
info:
  title: Email Workflows
servers:
  smtp:
    host: :2525
    protocol: smtp
  imap:
    host: :1430
    protocol: imap
```

More on that in a moment.

---

## Step 2: The Express Backend

The backend is a simple Express app using Nodemailer. Nothing special here: it connects to `localhost:2525` the same
way it would connect to any SMTP server.

```javascript
import nodemailer from 'nodemailer';

const transporter = nodemailer.createTransport({
  host: 'localhost',
  port: 2525,
  secure: false
});

app.post('/signup', async (req, res) => {
  const { email, password } = req.body;

  // Create account...

  await transporter.sendMail({
    from: 'noreply@example.com',
    to: email,
    subject: 'Verify your email address',
    html: `<p>Thanks for signing up!</p><p><a href="http://example.com/verify?email=${encodeURIComponent(email)}">Verify your email</a></p>`
  });

  res.json({ message: 'Check your inbox to verify your email' });
});
```

The backend has no idea Mokapi is there. It just sends mail to an SMTP server on port 2525. That transparency was
a deliberate design goal: you shouldn't need to change your application code to make it testable.

---

## Step 3: Writing the Playwright Test

This is where it all comes together. The test does three things:

1. Fills in the signup form and submits it, just like a real user
2. Fetches the captured email from Mokapi's HTTP API
3. Asserts on the subject, sender, recipient, and body content

You can find the full test file on [GitHub](https://github.com/marle3003/mokapi-email-workflow/blob/main/tests/email-workflows.spec.ts).

```typescript
import { test, expect } from '@playwright/test';

test('Email verification after signup', async ({ page, request }) => {
  const recipient = randomEmail();

  await test.step('Submit signup form', async () => {
    await page.goto('');
    const form = page.getByRole('form', { name: 'Sign Up' });
    await form.getByLabel('Email').fill(recipient);
    await form.getByLabel('Password').fill('SuperSecure123!');
    await form.getByRole('button', { name: 'Sign Up' }).click();
    await expect(form.getByText('Check your inbox to verify your email')).toBeVisible();
  });

  await test.step('Fetch email from Mokapi and verify contents', async () => {
    const mails = await request.get(
      `http://localhost:8080/api/services/mail/Email%20Workflows/mailboxes/${recipient}/messages?limit=1`
    );
    const mailList = await mails.json();
    expect(mailList.length).toBe(1);
    expect(mailList[0]).toEqual(expect.objectContaining({
      subject: 'Verify your email address',
      from: expect.arrayContaining([expect.objectContaining({
        address: 'noreply@example.com'
      })]),
      to: expect.arrayContaining([expect.objectContaining({
        address: recipient
      })])
    }));
  });

  await test.step('Verify email body and link', async () => {
    const res = await request.get(
      `http://localhost:8080/api/services/mail/messages/${mailList[0].messageId}`
    );
    const mail = await res.json();
    expect(mail).toEqual(expect.objectContaining({
      body: `<p>Thanks for signing up!</p><p><a href="http://example.com/verify?email=${encodeURIComponent(recipient)}">Verify your email</a></p>`
    }));
  });
});
```

A few things worth pointing out.

Using a random email address per test run is important. It keeps tests isolated from each other, so a message
from a previous run can't interfere with the current one. Same idea as the `Date.now()` document ID in the Kafka
example.

Because Mokapi supports IMAP, you can connect any real mail client to it during development and see exactly what
your users will see. Point Thunderbird, Apple Mail, or Outlook at localhost:1430 and your captured emails show
up just like they would in a real inbox.

This isn't part of the automated test. It's a development tool. But it's the kind of thing that catches the visual
bugs your assertions miss, and it costs nothing to set up since you already have Mokapi running.

---

## Why This Pattern Works

The backend sends email exactly as it does in production. Playwright drives a real browser against a real
frontend. Mokapi sits in the middle capturing everything and making it inspectable. No real emails are sent.
No external dependencies. No inbox polling.

And because the whole thing runs locally or in CI without any special infrastructure, you can add email assertions to
your test suite the same way you'd add any other assertion. It's not a separate manual process anymore.

The full working example is on GitHub: [mokapi-email-workflow](https://github.com/marle3003/mokapi-email-workflow). It includes a Vue frontend, an Express backend, the
Mokapi config, and the Playwright test. Clone it and have it running in a few minutes.

If your email flows have been living outside your test suite because testing them felt too complicated,
this is the setup that changes that.
