---
title: "Mock SMTP Server with Mokapi: Send Emails Using a Node.js Client"
description: Learn how to mock an SMTP server using Mokapi and send emails with a Node.js client. Perfect for testing email workflows without real mail servers.
icon: bi-envelope-at-fill
---

# Mocking an SMTP Server with Mokapi and Sending Emails Using a Node.js Client

## Introduction

Email communication is a crucial part of many applications, whether for user registration, notifications, or transactional messages. However, testing email workflows in a development environment can be challenging. Using a real SMTP server for testing carries several risks:

- <p><strong>Accidental email delivery</strong>: You might unintentionally send test emails to real users, causing confusion or compliance issues.</p>
- <p><strong>Spam filters</strong>: Test emails can be flagged as spam, affecting your domain's reputation.</p>
- <p><strong>Credential security</strong>: Using real SMTP credentials in a development environment can expose sensitive information.</p>
- <p><strong>Slow debugging</strong>: Waiting for email propagation can make troubleshooting slow and inefficient.</p>

By mocking an SMTP server with Mokapi, developers can simulate email sending without using a real mail server. This ensures that test emails are never sent to actual recipients while still allowing developers to review the email content and workflow.

In this tutorial, we’ll set up a mock SMTP server using Mokapi and send emails using a Node.js client. This setup provides a safe, controlled environment for email testing, making debugging faster and more reliable.

## Step 1: Setting Up Mokapi as a Mock SMTP Server

First, create a Mokapi configuration file (smtp.yaml) to define a mock SMTP server.

```yaml
smtp: '1.0'
info:
  title: Mokapi's Mail Server
servers:
  - url: smtp://127.0.0.1:2525
```

Start Mokapi using:

```bash
mokapi smtp.yaml
```

Mokapi will now listen for SMTP connections on port 2525.

## Step 2: Install Nodemailer in Node.js

We’ll use Nodemailer, a popular library for sending emails in Node.js.
Install it with:

```bash
npm install nodemailer
```

## Step 3: Create a Node.js Client

Now, let’s write a simple script to send an email using Mokapi’s mock SMTP server.

```javascript tab=sendEmail.js
const nodemailer = require("nodemailer");

async function sendEmail() {
  let transporter = nodemailer.createTransport({
    host: "localhost",
    port: 2525, // Mokapi's SMTP server
    secure: false, // Mokapi does not use SSL
    auth: {
      user: "", // No authentication needed for this mock
      pass: "",
    },
  });

  let info = await transporter.sendMail({
    from: '"Test Sender" <test@example.com>',
    to: "recipient@example.com",
    subject: "Hello from Mokapi",
    text: "This is a test email sent via Mokapi's mock SMTP server.",
  });

  console.log("Message sent: %s", info.messageId);
}

sendEmail().catch(console.error);
```

Run the script:

```bash
node sendEmail.js
```

## Step 4: Viewing the Sent Email in the Mokapi Dashboard

Once the email is sent, Mokapi captures and logs it in its web-based dashboard. This allows you to inspect the email’s details, including the sender, recipient, subject, content and attachments

To access the dashboard, open your browser and navigate to http://localhost:8080

Here’s an example of what the email looks like in the Mokapi Dashboard:

<img src="./dashboard-smtp.png" alt="Screenshot of Dashboard showing a received message." />

With this setup, you can safely test your email workflows in a controlled environment without modifying real email configurations or using external mail services.

## Step 5: Enabling SSL and Authentication in Mokapi’s Mock SMTP Server

In this section, we will enhance our mock SMTP server by enabling SSL/TLS encryption and requiring authentication 
for sending emails. This is useful for testing secure email workflows, ensuring that only authorized users can send 
emails.

### Configuring Mokapi for Secure SMTP (SMTPS) with Authentication

To enable SSL/TLS and authentication, update your Mokapi configuration file (smtp.yaml) as follows:

```yaml
smtp: '1.0'
info:
  title: Mokapi's Secure Mail Server
servers:
  - url: smtps://127.0.0.1:2525 # Enable SMTPS (SMTP over SSL)
mailboxes:
  - name: test@example.com
    username: bob
    password: secret

```

- smtps://127.0.0.1:2525 → Enables SMTP over SSL (SMTPS) on port 2525.
- mailboxes → Defines mailboxes with authentication. (By default autoCreateMailbox is true)
  - The email test@example.com is accessible by user bob with password secret.

By default, Mokapi automatically creates mailboxes that are not in the configuration (autoCreateMailbox=true). This means users can still send emails even if their mailbox is not predefined.

#### Hot-Reloading: No Restart Needed

If Mokapi is already running, you don’t need to restart it after modifying the configuration. Mokapi will detect the changes and update the mock server at runtime.

Now, let’s update our Node.js client to support authentication and SSL.

### Updating the Node.js Client for SSL and Authentication

Now, modify your sendEmail.js script to use TLS encryption and authentication.
Since Mokapi uses a self-signed certificate by default, we need to adjust our Node.js client to accept self-signed certificates when using SSL.

```javascript
const nodemailer = require("nodemailer");

async function sendEmail() {
    let transporter = nodemailer.createTransport({
        host: "localhost",
        port: 2525, // Mokapi's SMTP server
        secure: true,
        auth: {
            user: "bob",
            pass: "secret",
        },
        tls: {
            rejectUnauthorized: false
        }
    });

    let info = await transporter.sendMail({
        from: '"Test Sender" <test@example.com>',
        to: "recipient@example.com",
        subject: "Hello from Mokapi",
        text: "This is a test email sent via Mokapi's mock SMTP server.",
    });

    console.log("Message sent: %s", info.messageId);
}

sendEmail().catch(console.error);
```

### Testing Secure Email Sending

Run the script:

```bash
node sendEmail.js
```

If authentication is successful, you should see:

```
Message sent: <unique-message-id>
```

If authentication fails, Mokapi will respond with an error:

```
535 [5.7.8] Authentication credentials invalid
```