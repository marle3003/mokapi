---
title: on( event, handler, [args] )
description: Attaches an event handler for the given event.
---
# on( event, handler, [args] )

Attaches an event handler for the given event.

| Parameter        | Type     | Description                                                                                                        |
|------------------|----------|--------------------------------------------------------------------------------------------------------------------|
| event            | string   | Event type such as `http                                                                                           |
| handler          | function | An [EventHandler](/docs/javascript-api/mokapi/eventhandler/eventhandler.md) to execute when the event is triggered |
| args (optional)  | object   | [EventArgs](/docs/javascript-api/mokapi/eventhandler/eventargs.md) object contains additional event arguments.     | 

``` box=tip url=[See app](/docs/javascript-api/mokapi/app/overview.md)
To handle specific paths, operations, topics or channels, the `app` routers are often shorter:
`app.http().get('/echo', handler)` instead of checking `request.url.path` inside an `on('http')` handler.
```

## Event smtp

`on('smtp', handler)` runs for every mail a mail API accepts, whether it was received over SMTP or
delivered with [`send()`](/docs/javascript-api/mokapi/app/mail-router.md#send). The handler runs
before the mail is stored and can change it. The mail has these fields:

| Name                 | Type         | Description                                    |
|----------------------|--------------|------------------------------------------------|
| api                  | string       | Name of the mail API (`info.title`)            |
| server               | string       | Name of the SMTP server that received the mail |
| sender (optional)    | Address      | Value of the `Sender` header                   |
| from                 | Address[]    | Addresses of the `From` header                 |
| to                   | Address[]    | Primary recipients                             |
| replyTo (optional)   | Address[]    | Addresses of the `Reply-To` header             |
| cc (optional)        | Address[]    | Carbon copy recipients                         |
| bcc (optional)       | Address[]    | Blind carbon copy recipients                   |
| messageId            | string       | Unique ID of the mail                          |
| inReplyTo (optional) | string       | Message ID of the mail this mail replies to    |
| time                 | Date         | Date of the mail                               |
| subject              | string       | Subject of the mail                            |
| contentType          | string       | MIME type of the body                          |
| encoding             | string       | Content transfer encoding                      |
| body                 | string       | Body of the mail                               |
| attachments          | Attachment[] | Attachments of the mail                        |

To handle the mails of one mail API or one mailbox, use the
[MailRouter](/docs/javascript-api/mokapi/app/mail-router.md): `app.api('Mail Server').mail().mailbox('alice@example.com').message(handler)`.

## Example Echo Server

```javascript tab=echo.js
import { on } from 'mokapi'

export default function() {
    on('http', function(request, response) {
        if (request.url.path === '/echo') {
            response.data = {
                url: request.url.toString(),
                body: request.body,
            }
        }
    })
}
```

```yaml tab=echo.yaml
openapi: 3.0.0
info:
  title: Echo Server
  version: '1.0'
servers:
 - url: /echo
paths:
  /:
    post:
      summary: Echoes the contents of the request body.
      requestBody: 
       content: 
        '*/*':
         schema:
          type: string
      responses:
        "200":
          description: the echo response
          content:
            application/json:
              schema:
                type: object
```