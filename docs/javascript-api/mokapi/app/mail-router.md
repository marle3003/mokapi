---
title: MailRouter
description: Handle mails accepted by a mail API per API and mailbox, and deliver mails straight into a mail API using the app API.
---
# MailRouter

`MailRouter` registers handlers for mails accepted by a mail API and delivers mails into the
store of a mail API. Get one via `app.mail()` or `app.api('My API').mail()`.

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Mail Server').mail()
        .mailbox('alice@example.com')
            .message((mail) => {
                console.log(`alice got: ${mail.subject}`)
            })

    app.api('Mail Server').mail().send({
        from: [{ address: 'bob@example.com' }],
        to: [{ address: 'alice@example.com' }],
        subject: 'Hello',
        body: 'Hi Alice'
    })
}
```

## Methods

| Method                                           | Returns                          | Description                                        |
|--------------------------------------------------|----------------------------------|----------------------------------------------------|
| [`mailbox( address )`](#mailbox)                 | [MailboxRouter](#mailboxrouter)  | Returns a router for a single mailbox              |
| [`message( handler, [opts] )`](#message)         | `MailRouter`                     | Registers a handler for all mails                  |
| [`send( message )`](#send)                       | `MailRouter`                     | Delivers a mail into a mail API                    |
| [`sendAsync( message )`](#send)                  | `Promise<MailRouter>`            | Delivers a mail asynchronously                     |

To scope the router to one API, use [`app.api( title ).mail()`](/docs/javascript-api/mokapi/app/api.md).
The title must match `info.title` of the mail configuration.

## mailbox

Returns a [MailboxRouter](#mailboxrouter) for the given mailbox address. Its handlers run
for mails whose envelope recipients (`to`, `cc` and `bcc`) contain the address. A mail to two
watched mailboxes runs the handler of each mailbox once.

| Parameter | Type   | Description                                      |
|-----------|--------|--------------------------------------------------|
| address   | string | Address of the mailbox, e.g. `alice@example.com` |

The address must match exactly, including upper and lower case, the same way Mokapi looks up
a mailbox. `Alice@example.com` doesn't match a handler registered for `alice@example.com`.
Wildcards and mailbox names are not supported.

## message

Registers a handler for **all** mails of this router. The handler runs for mails received over
SMTP and for mails delivered with [`send()`](#send). It runs after the delivery checks and
before the mail is stored.

| Parameter       | Type     | Description                                                                                           |
|-----------------|----------|-------------------------------------------------------------------------------------------------------|
| handler         | function | Function that receives a [MailEventMessage](#maileventmessage)                                        |
| opts (optional) | object   | [EventArgs](/docs/javascript-api/mokapi/eventhandler/eventargs.md) such as `priority`, `track`, `tags` |

```javascript
import { app } from 'mokapi'

export default function() {
    app.mail().message((mail) => {
        console.log(`[${mail.api}] ${mail.subject}`)
    }, { tags: { name: 'log-mails' } })
}
```

Handlers show up as actions in the mail log of the dashboard. [`on('smtp')`](/docs/javascript-api/mokapi/on.md)
handlers receive the same event, so both styles can be combined.

## send

Delivers a mail into the store of a mail API. `send()` runs synchronously and throws an error
if the mail is rejected. `sendAsync()` returns a `Promise` that resolves with the router and
rejects with the same error.

| Parameter | Type   | Description                      |
|-----------|--------|----------------------------------|
| message   | object | A [MailMessage](#mailmessage)    |

`send()` doesn't open a network connection. No SMTP server has to be configured and no
authentication applies. Besides that, the mail is treated like a mail received over SMTP: it is
appended to the INBOX of every recipient, IMAP clients and the dashboard mail log see it,
metrics are updated, and `message` and `on('smtp')` handlers run.

```javascript
import { app } from 'mokapi'

export default function() {
    try {
        app.api('Mail Server').mail().send({
            from: [{ address: 'bob@example.com' }],
            to: [{ address: 'alice@example.com' }],
            subject: 'Hello',
            body: 'Hi Alice'
        })
    } catch (err) {
        console.log(`mail rejected: ${err}`)
    }
}
```

``` box=warning
Use `sendAsync()` when you send from inside a handler. A synchronous `send()` runs the handlers
of the delivered mail right away, so handlers that send to each other, especially across scripts,
can block.
```

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Mail Server').mail()
        .mailbox('support@example.com')
        .message(async (mail) => {
            await app.api('Mail Server').mail().sendAsync({
                from: [{ address: 'support@example.com' }],
                to: mail.from,
                subject: `Re: ${mail.subject}`,
                body: 'We received your request.'
            })
        })
}
```

`sendAsync()` reads the message when you call it, so changing the object afterwards doesn't
affect the mail. Invalid arguments throw at the call. Errors from the delivery reject the
promise.

### Delivery checks

`send()` applies the checks a real SMTP client would hit. It checks the sender and all
recipients before it delivers anything. If one check fails, no recipient gets the mail and no
mailbox is created.

| Check                                              | Error message                                                    |
|----------------------------------------------------|------------------------------------------------------------------|
| [Rules](/docs/mail/rules.md) for sender, recipient, subject or body | The reject message of the rule |
| `allowUnknownSenders: false` and the sender has no mailbox | `Unknown mailbox bob@example.com`                        |
| Recipient has no mailbox and `autoCreateMailbox` is off | `Unknown mailbox alice@example.com`                         |
| More recipients than `maxRecipients`               | `Too many recipients of 2 reached`                               |

Mokapi creates the mailbox of a recipient when it delivers the mail, not before, and only if
`autoCreateMailbox` allows it. The limit `maxInboxMails` applies as for SMTP. See
[Mail Server](/docs/mail/overview.md) for the settings.

Other errors you can get:

| Error message                                   | Cause                                                              |
|-------------------------------------------------|--------------------------------------------------------------------|
| `mail API 'Mail Server' not found`              | `app.api( title )` names no mail API                               |
| `a sender or from address must be specified`    | The mail has neither `sender` nor `from`                           |
| `sender required if using multiple from addresses` | The mail has several `from` addresses and no `sender`          |

### Choosing the mail API

With `app.api( title ).mail()` the mail goes to the API with that title. Without a scope
(`app.mail().send()`), Mokapi chooses the API:

1. If two or more mail APIs have `autoCreateMailbox` enabled, `send()` throws an error.
2. Otherwise, the mail APIs that already have a mailbox for at least one recipient count.
   If exactly one does, it gets the mail. If several do, `send()` throws an error.
3. If no API has a mailbox for any recipient, the one API with `autoCreateMailbox` gets the
   mail. If there is none, `send()` throws an error.

Each error names the problem and the candidate titles, and tells you to use
`app.api('...').mail()`. For example, if `Mail A` has `autoCreateMailbox` enabled and `Mail B`
has a mailbox `bob@example.com`, a mail to `bob@example.com` goes to `Mail B` and a mail to
`new@example.com` goes to `Mail A`.

``` box=tip
Use `app.api( title ).mail()` for `send()` as soon as you have more than one mail API.
```

## Changing a mail

A handler receives a copy of the mail and can change it. Mokapi writes the changes back before
it stores the mail, so IMAP clients, the mail view and the dashboard mail log show the changed
mail.

```javascript
import { app } from 'mokapi'

export default function() {
    app.mail().message((mail) => {
        mail.subject = `[mock] ${mail.subject}`
    })
}
```

The envelope stays fixed. Changing `to`, `cc` or `bcc` in a handler only changes the headers
of the stored mail. Delivery uses the recipients from the SMTP transaction or from `send()`.
Handlers can't reject a mail. Use [mail rules](/docs/mail/rules.md) for that.

## MailboxRouter

Groups handlers for a single mailbox. `message()` behaves like the one of `MailRouter`. A
`MailboxRouter` has no `send()`, because the recipients belong in the mail.

| Method                       | Returns         | Description                                 |
|------------------------------|-----------------|---------------------------------------------|
| `message( handler, [opts] )` | `MailboxRouter` | Registers a handler for mails to this mailbox |

## MailEventMessage

The argument of a `message` handler. The [`on('smtp')`](/docs/javascript-api/mokapi/on.md) handler
receives the same object.

| Name                 | Type         | Description                                                    |
|----------------------|--------------|----------------------------------------------------------------|
| api                  | string       | Name of the mail API (`info.title`)                            |
| server               | string       | Name of the SMTP server that received the mail                 |
| sender (optional)    | Address      | Value of the `Sender` header                                   |
| from                 | Address[]    | Addresses of the `From` header                                 |
| to                   | Address[]    | Primary recipients                                             |
| replyTo (optional)   | Address[]    | Addresses of the `Reply-To` header                             |
| cc (optional)        | Address[]    | Carbon copy recipients                                         |
| bcc (optional)       | Address[]    | Blind carbon copy recipients                                   |
| messageId            | string       | Unique ID of the mail                                          |
| inReplyTo (optional) | string       | Message ID of the mail this mail replies to                    |
| time                 | Date         | Date of the mail                                               |
| subject              | string       | Subject of the mail                                            |
| contentType          | string       | MIME type of the body, e.g. `text/plain`                       |
| encoding             | string       | Content transfer encoding, e.g. `base64`                       |
| body                 | string       | Body of the mail                                               |
| attachments          | Attachment[] | Attachments of the mail                                        |

`api` is read-only. `Address` is an object with `address` and an optional `name`.

## MailMessage

The argument of `send()` and `sendAsync()`. It has the fields of
`Message` from `mokapi/smtp`, except `server`.

| Name                    | Type         | Description                                                                         |
|-------------------------|--------------|-------------------------------------------------------------------------------------|
| from                    | Address[]    | Addresses of the `From` header                                                      |
| sender (optional)       | Address      | Envelope sender. Required if `from` has more than one address. Default: the single `from` address |
| to                      | Address[]    | Primary recipients                                                                  |
| cc (optional)           | Address[]    | Carbon copy recipients                                                              |
| bcc (optional)          | Address[]    | Blind carbon copy recipients                                                        |
| replyTo (optional)      | Address[]    | Addresses of the `Reply-To` header                                                  |
| messageId (optional)    | string       | Unique ID of the mail. Default: generated                                           |
| inReplyTo (optional)    | string       | Message ID of the mail this mail replies to                                         |
| time (optional)         | Date         | Date of the mail. Default: the current time                                         |
| subject                 | string       | Subject of the mail                                                                 |
| body                    | string       | Body of the mail                                                                    |
| contentType (optional)  | string       | MIME type of the body. Without it, the mail has no content type                     |
| encoding (optional)     | string       | Content transfer encoding of the body, e.g. `base64`                                |
| attachments (optional)  | Attachment[] | Attachments. An attachment has `name`, `contentType` and `data` (`Uint8Array`)      |

The envelope recipients are `to`, `cc` and `bcc` without duplicates. A recipient that appears in
several lists gets one copy.
