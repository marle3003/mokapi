---
title: SMTP Documentation
description: Test emails safely and no risk of spamming recipients' inbox
---
# SMTP
With Mokapi you can capture your emails, being sent from your application, without flooding
your mailbox. You can also use different email addresses for testing, not just your own.

## Example
This example demonstrates how to configure an SMTP server listen on port 587 on all available
ip address.

```yaml tab=smtp
smtp: '1.0'
info:
  title: Mokapi's Mail Server
server: smtp://:587
```

```yaml tab=smtps
smtp: '1.0'
info:
  title: Mokapi's Mail Server
server: smtps://:587
```