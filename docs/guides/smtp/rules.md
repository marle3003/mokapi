---
title: SMTP Rules
description: Define rules to allow or deny emails
---
# Rules

You can declaratively write a rule whether the mail server accepts or blocks a mail.

```yaml
smtp: '1.0'
info:
  title: Mokapi's Mail Server
server: smtp://127.0.0.1:25
rules:
  - name: Recipient's domain is mokapi.io
    recipient: '@mokapi.io'
    action: allow
```

## Rule Properties

The following table describes the rule properties that are available

| Property       | Description                                |
|----------------|--------------------------------------------|
| name           | Name of the rule                           |
| sender         | A regex expression to validate sender      |
| recipient      | A regex expression to validate recipient   |
| subject        | A regex expression to validate subject     |
| body           | A regex expression to validate body        |
| action         | `deny` or `allow` a matched mail           |
| rejectResponse | Define a custom response if action is deny |

```yaml
smtp: 1.0
server: smtp://localhost:8025
rules:
  - sender: @foo.bar
    action: deny
    rejectResponse:
      statusCode: 550
      enhancedStatusCode: 5.7.1
      text: Your error message
```