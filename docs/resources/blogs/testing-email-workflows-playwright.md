---
title: Testing Email Workflows with Playwright and Mokapi
description: A complete guide to end-to-end testing of email workflows using Mokapi and Playwright
icon: bi-envelope-paper
tech: mail
---

# Reliable Testing of Email Workflows with Playwright and Mokapi

Modern web applications rely on email for critical user flows — account registration, email verification,
and password reset are just a few examples. These flows are often automated through forms and background
logic, but testing them effectively can be tricky. Emails don’t just “send” — they need to arrive, include
the right content, and offer a smooth experience across different clients and devices.

In this post, I’ll share a practical way to test email workflows end-to-end using Playwright and Mokapi,
a mock server that supports both SMTP and IMAP protocols. The solution is lightweight, fast, and easy to
integrate into CI pipelines — while still letting developers verify the actual email content during local
development.

## Email Workflows Deserve Better Testing
If your application sends emails, you’ve probably tested it manually: signing up with a test email address,
checking your inbox, clicking a link. Maybe you even used a disposable inbox service.

But this approach doesn’t scale. Manual testing is slow and error-prone. External email providers introduce
delays or spam filtering. And verifying the content — like checking for typos or broken links — often falls
through the cracks.

What we need is a way to test these workflows programmatically — without relying on real email servers — and
ideally in a way that runs as part of our automated test suite.

## A Mock Mail Server, Built for Testing

Mokapi offers a mock mail server that behaves like a real SMTP server, but captures all sent emails for
inspection via a simple HTTP API. You can run Mokapi locally or in your CI environment, and connect your 
application to it instead of a real mail server.

When your backend sends an email (for example, after a user signs up), Mokapi intercepts the message.
Then, your test script can fetch that message via the API and assert its subject, content, or even inspect
a confirmation link inside.

This gives you full control over your email workflows — with no real emails being sent and no delays in
delivery.

```yaml
mail: '1.0'
info:
  title: Email Workflows
servers:
  smtp:
    host: :2525
    protocol: smtp
```

## Automating the Flow with Playwright

To tie everything together, we use Playwright to automate the user journey.

Let’s say your app includes a signup form. In your Playwright test, you’d fill in the form, submit it,
and wait for a success message. But instead of stopping there, the test also contacts Mokapi’s mail API
to retrieve the email that should have been sent to the user.

This allows the test to confirm that:
- The correct email address received a message 
- The subject is what you expect ("Verify your email address")
- The email contains the expected body content (like a verification link)

With this setup, your test covers the full flow — from user interaction to backend logic to outbound email
— without relying on any external systems.

```typescript
import { test, expect } from '@playwright/test';

test('Email verification after signup', async ({ page, request }) => {
  // use a random mail address to ensure test is isolated
  const recipient = randomEmail()

  // Step 1: Go to page and submit signup form
  await page.goto('');
  const form = page.getByRole('form', { name: 'Sign Up' })
  await form.getByLabel('Email').fill(recipient)
  await form.getByLabel('Password').fill('SuperSecure123!')
  await form.getByRole('button', { name: 'Sign Up' }).click()
  await expect(form.getByText('Check your inbox to verify your email')).toBeVisible()

  // Step 2: Wait and fetch latest mail from Mokapi (limit=1)
  const mails = await request.get(`http://localhost:8080/api/services/mail/Email%20Workflows/mailboxes/${recipient}/messages?limit=1`);
  const mailList = await mails.json();
  await expect(mailList.length).toBe(1)
  await expect(mailList[0]).toEqual(expect.objectContaining({
    subject: 'Verify your email address',
    from: expect.arrayContaining([expect.objectContaining({
      address: 'noreply@example.com'
    })]),
    to: expect.arrayContaining([expect.objectContaining({
      address: recipient
    })])
  }))

  // Step 3: Fetch the mail body
  const res = await request.get(`http://localhost:8080/api/services/mail/messages/${mailList[0].messageId}`);
  const mail = await res.json()
  await expect(mail).toEqual(expect.objectContaining({
    body: `<p>Thanks for signing up!</p><p><a href="http://example.com/verify?email=${encodeURIComponent(recipient)}">Verify your email</a></p>`
  }))
});
```

The Playwright test makes a real request to the frontend, just like a user would. Then it retrieves
the mocked email and verifies its content.

## Check Emails Visually with Your Favorite Mail Client

Mokapi doesn’t just support automated tests. It also offers IMAP access to the captured messages.

This means you can connect your local mail client — such as Apple Mail, Thunderbird, or even Outlook —
to Mokapi and view the emails exactly as a real user would. This is especially helpful during development
or QA when you want to see how your HTML emails render in real-world environments.

Being able to visually inspect the email helps catch formatting issues, broken links, or styling problems
that automated tests might not detect.

Add the IMAP server to your config:

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

Now you can connect Apple Mail, Thunderbird, or Outlook to localhost:1430 and preview your emails in real
time. It’s a great way to catch formatting issues, broken links, or rendering quirks that tests might miss.

## All Together: Easy to Run, Easy to Maintain

The full example is available on [GitHub](https://github.com/marle3003/mokapi-email-workflow), including:
- A Vue frontend with a simple signup form 
- An Express backend that sends emails using Nodemailer 
- A Mokapi configuration file for a mock SMTP server 
- A Playwright test that drives the signup and verifies the email 

To run the whole flow locally, you just need to start three services:
1. The Mokapi mail server (mokapi mocks/mail.yaml)
2. The Express backend (node backend/index.js)
3. The Playwright test (npx playwright test — which will also start the frontend)

The Playwright test handles everything from filling out the form to checking the email — 
making it easy to test real scenarios without complicated setups.

## Conclusion

Email is a critical part of the user experience, but it’s often the least tested. With Mokapi and
Playwright, you can bring email workflows into your automated tests — and get reliable results in
both development and CI.

Whether you're validating content, checking formatting, or verifying links, Mokapi gives you full
control over your mail layer. And with IMAP access, you can preview real-looking messages without
ever sending a real email.

Don’t leave your email experience to chance. Mock it, test it, and ship it with confidence.

---

Try the example on GitHub: [mokapi-email-workflow](https://github.com/marle3003/mokapi-email-workflow)\
Learn more about [Mokapi Mail](/docs/mail/overview.md)