# SMTP
With Mokapi you can capture your emails, being sent from your application, without flooding
your mailbox. You also can use different email addresses for testing and not just your own.

## Example
This example demonstrates how to configure an SMTP server listen on port 587 on all available
ip address.

```yaml
smtp: 1.0
address: "0.0.0.0:587"
```