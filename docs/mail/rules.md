---
title: SMTP Mail Rules
description: Use mail rules to make Mokapi's SMTP server accept or reject mails by sender, recipient, subject or body, and return a custom SMTP error response.
---
# Mail Rules

Rules let your mocked SMTP server reject mails the way a real server would, for example mails to
a foreign domain or from a blocked sender. Each rule has a name and checks the sender, recipient,
subject or body of a mail with a regular expression. The `action` decides what a match means.

```yaml
mail: '1.0'
info:
  title: Mokapi's Mail Server
servers:
  smtp:
    host: :25
    protocol: smtp
rules:
  onlyCompanyRecipients:
    description: Accept mails to example.com only
    recipient: '@example\.com$'
    action: allow
  blockSpamSender:
    sender: '@spam\.example$'
    action: deny
    rejectResponse:
      statusCode: 550
      enhancedStatusCode: 5.7.1
      message: Sender is blocked
```

The rule name must match `^[A-Za-z0-9_-]+$`, otherwise Mokapi rejects the configuration. The
name appears in the default error message and in the [dashboard](/docs/mail/dashboard.md).

``` box=info
This page uses the `mail` format. The older format with `smtp: '1.0'` and `server:` is deprecated.
Mokapi converts it when it loads the file and logs a warning.
```

## Rule Properties

| Property       | Description                                                                          |
|----------------|--------------------------------------------------------------------------------------|
| description    | Optional text for the rule. It is shown in the dashboard and supports CommonMark.    |
| sender         | Regular expression for the sender address (SMTP `MAIL FROM`)                         |
| recipient      | Regular expression for a recipient address (SMTP `RCPT TO`)                          |
| subject        | Regular expression for the `Subject` header                                          |
| body           | Regular expression for the body of the mail                                          |
| action         | `allow` or `deny`                                                                    |
| rejectResponse | Response Mokapi sends when the rule rejects a mail. See [Reject Response](#reject-response). |

A rule needs at least one of `sender`, `recipient`, `subject` or `body`, and an `action`. A rule
without an `action` (or with another value) never rejects anything.

## How Rules Work

With `action: deny`, Mokapi rejects a mail if the value matches. With `action: allow`, Mokapi
rejects a mail if the value does not match. So `allow` doesn't mean "accept whatever matches and
skip the other rules". It means "reject everything that doesn't match".

Mokapi evaluates every rule, and a mail is accepted only if no rule rejects it. If there are no
rules, or none of them applies, the mail is accepted. A few consequences:

- A `deny` rule doesn't override an `allow` rule, and the order of rules in the file doesn't
  matter. Mokapi doesn't define in which order it checks them. If several rules would reject a
  mail, you can't predict which response the client gets.
- Two `allow` rules on the same field must both match. To accept two domains, use one rule with an
  alternation such as `@(example\.com|example\.org)$`.
- If a rule has several fields, Mokapi checks each field on its own. A `deny` rule with `sender`
  and `subject` rejects a mail if the sender matches or the subject matches, not only if both do.
  Use separate rules to keep this easy to read.

### When a Mail Is Rejected

Mokapi checks the fields at different steps of the SMTP conversation:

| Field     | Checked at   | Value                                                          |
|-----------|--------------|----------------------------------------------------------------|
| sender    | `MAIL FROM`  | The address of the SMTP envelope, without angle brackets       |
| recipient | `RCPT TO`    | The address of the SMTP envelope, checked once per recipient   |
| subject   | `DATA`       | The `Subject` header, after the client sent the whole mail     |
| body      | `DATA`       | The body of the mail, after the client sent the whole mail     |

The sender and recipient rules check the addresses the client sends in `MAIL FROM` and `RCPT TO`,
not the `From` and `To` headers of the message. A rejected recipient doesn't stop the other
recipients of the same mail. A mail rejected at `DATA` isn't stored in any mailbox.

The [dashboard](/docs/mail/dashboard.md#rejected-mails) lists rejected mails with the error
message, and it has a Rules tab that shows the rules of a service.

### Regular Expressions

Mokapi uses the Go regular expression syntax (RE2). The expression isn't anchored, so
`@example\.com` also matches `bob@example.com.evil.test`. Add `^` and `$` to match the whole
value. The match is case-sensitive. Start the expression with `(?i)` to ignore the case.

Quote the expression in YAML. A value that starts with `@` or `*` is not valid unquoted, and one that starts with `[` is read as a list.
Remember that `.` matches any character, so escape it (`\.`) in domain names. In single quotes,
a backslash is taken literally, which is what you want for a regular expression.

## Reject Response

If a rule rejects a mail and has no `rejectResponse`, Mokapi sends one of these responses:

| Checked field       | Status | Enhanced status | Example message                                                         |
|---------------------|--------|-----------------|-------------------------------------------------------------------------|
| sender, recipient   | 550    | 5.1.0           | `Rule onlyCompanyRecipients: Recipient bob@foo.bar does not match allow rule: @example\.com$` |
| subject, body       | 550    | 5.7.1           | `Rule noHello: Subject Hello World does match deny rule: ^Hello`        |

Use `rejectResponse` to return your own response, for example to test how your application
handles a specific error. Mokapi sends the response as you define it, so set all three fields.

| Property           | Description                                                    |
|--------------------|----------------------------------------------------------------|
| statusCode         | The SMTP status code, for example `550`                        |
| enhancedStatusCode | The enhanced status code in the format `x.y.z`, e.g. `5.7.1`   |
| message            | The text of the response                                       |

The `rejectResponse` is used for `allow` and for `deny` rules.

## Examples

### Accept Mails to Your Domain Only

```yaml
rules:
  onlyCompanyRecipients:
    recipient: '@example\.com$'
    action: allow
```

A client that sends a mail to `bob@foo.bar` gets a `550` response at `RCPT TO`. Mails to
`alice@example.com` are accepted.

### Block a Sender with a Custom Response

```yaml
rules:
  blockSpamSender:
    sender: '(?i)@spam\.example$'
    action: deny
    rejectResponse:
      statusCode: 550
      enhancedStatusCode: 5.7.1
      message: Sender is blocked
```

### Block a Mail by Subject

```yaml
rules:
  noTestMails:
    subject: '^\[TEST\]'
    action: deny
```

The client first gets the `354` reply to `DATA` and sends the whole mail. Mokapi then rejects it
with `550`.

## Patching Rules

If several files have the same `info.title`, Mokapi [patches](/docs/configuration/patching.md)
the configuration and merges rules with the same name. A property set in the second file replaces the one from
the first file, and the other properties stay. This lets you change a single rule per environment:

```yaml tab="Base Config"
mail: '1.0'
info:
  title: Mokapi's Mail Server
rules:
  spam:
    subject: '.*spam.*'
    action: allow
```

```yaml tab="Patch Config"
mail: '1.0'
info:
  title: Mokapi's Mail Server
rules:
  spam:
    action: deny
```

Mokapi applies changes to rules while it runs, you don't need to restart it.
