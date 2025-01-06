---
title: send( host, mail, [auth] ) (mokapi/smtp)
description: Sends an email message to an SMTP server for delivery.
---
# send( host, mail, [auth] )

Sends an email message to an SMTP server for delivery.

| Parameter       | Type    | Description                                                                                                  |
|-----------------|---------|--------------------------------------------------------------------------------------------------------------|
| host            | string  | String that contains the URL, name or IP address of the host computer used for SMTP transactions.            |
| mail            | object  | A [MailMessage](/docs/javascript-api/mokapi-smtp/mailmessage.md) that contains the message to send.          |
| auth (optional) | object  | An optional [Auth](/docs/javascript-api/mokapi-smtp/auth.md) object that contains authentication information |

## Examples

```javascript
import { send } from 'mokapi/smtp'

export default function() {
    send('foo.bar', {
        sender: 'carol@mokapi.io',
        from: [{ name: 'Alice', address: 'alice@mokapi.io' },'charlie@mokapi.io'],
        to: [{ name: 'Frank', address: 'frank@mokapi.io' } ],
        cc: [{ address: 'sybil@mokapi.io'}],
        bcc: [ 'bob@mokapi.io' ],
        replyTo: 'charlie@mokapi.io',
        contentType: 'text/html',
        subject: 'A test mail',
        body: '<html><body><h1>Hello</h1>You can send an email messages from Mokapi very easily.</body></html>'
    })
}
```