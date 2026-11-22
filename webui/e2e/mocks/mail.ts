import { app } from 'mokapi'

export default function () {
    app.http()
        .operation('serviceMail', (req, res) => {
            res.data = services.find(x => x.name === req.path['name']) ?? services[0]
        })
        .operation('mailboxes', (_, res) => {
            res.data = services[0].mailboxes
        })
        .operation('mailbox', (req, res) => {
             for (const mb of services[0].mailboxes) {
                if (mb.name === req.path.mailbox) {
                    res.data = {
                        ...mb,
                        folders: mb.folders.map(x => x.name)
                    }
                    return
                }
            }
            res.rebuild(404)
        })
        .operation('smtpMessages', (req, res) => {
            for (const mb of services[0].mailboxes) {
                if (mb.name === req.path.mailbox) {
                    const result = []
                    for (const f of mb.folders) {
                        result.push(...f.messages)
                    }
                    res.data = result
                    return
                }
            }
            res.rebuild(404)
        })
        .operation('smtpMessage', (req, res) => {
            const messageId = req.path.messageId
            const mail = getMail(messageId)
            if (mail == null) {
                res.rebuild(404)
                return
            }
            res.data = {
                service: services[0].name,
                data: mail
            }
        })
        .operation('smtpMessageAttachment', (req, res) => {
            const attachment = getAttachment(req.path.messageId, req.path.name)
            if (attachment == null) {
                res.rebuild(404)
                return
            }
            res.headers['Content-Type'] = attachment.contentType
            res.headers['Content-Disposition'] = `inline; filename="${attachment.name}"`
            res.data = attachment.data
        })
}

export const mails = [
    {
        from: [{name: 'Alice', address: 'alice@mokapi.io'}],
        to: [{name: 'Bob', address: 'bob@mokapi.io'}, {address: 'carol@mokapi.io'}],
        date: '2023-02-23T08:49:25+01:00',
        contentType: 'text/html',
        contentTransferEncoding: 'quoted-printable',
        messageId: '20230223-084925.763-4196@mokapi.io',
        inReplyTo: '20230222-084925.763-4196@mokapi.io',
        subject: 'A test mail',
        body: `<style>
                    .container {color: black; }
                    @media (prefers-color-scheme: dark ) {
                        .container {color: white; }
                    }
                    
                </style>
                <div class="container">
                    <h1>Hello</h1>
                    Mail message from Alice<img src="cid:icon.png" />
                </div>`,
        attachments: [
            {
                name: 'foo.txt',
                contentType: 'text/plain',
                disposition: 'attachment',
                size: 34056,
                data: 'foobar'
            },
            {
                name: 'icon.png',
                contentType: 'image/png',
                disposition: 'inline',
                contentId: 'icon.png',
                size: 372,
                data: (open as unknown as (path: string, args: { as: 'binary' }) => string)('icon.png', {as: 'binary'})
            }
        ],
        duration: 30016,
        actions: [
            {
                duration: 20,
                tags: {
                    name: "dashboard",
                    file: "/Users/maesi/GolandProjects/mokapi/examples/mokapi/http_handler.js",
                    event: "http"
                }
            }
        ]
    },
    {
        from: [{name: 'Alice', address: 'alice@mokapi.io'}],
        to: [{name: 'Bob', address: 'bob@mokapi.io'}],
        date: '2023-02-22T10:15:00+01:00',
        contentType: 'text/plain',
        contentTransferEncoding: '7bit',
        messageId: '20230222-101500.123-1111@mokapi.io',
        subject: 'A plain text mail',
        body: 'Plain text message from Alice',
        attachments: [],
        duration: 120,
    },
]

// Mails that are stored but not in the mailbox of bob, so the mailbox counts stay as they are.
const storedMails = [
    {
        from: [{name: 'Carol', address: 'carol@mokapi.io'}],
        to: [{address: 'bob@mokapi.io'}],
        cc: [{address: 'dave@mokapi.io'}],
        bcc: [{address: 'erin@mokapi.io'}],
        date: '2023-02-20T09:00:00+01:00',
        contentType: 'text/plain',
        contentTransferEncoding: '7bit',
        messageId: '20230220-090000.001-2001@mokapi.io',
        subject: 'Quarterly report',
        body: 'The report is attached.',
        attachments: [
            {
                name: 'report.pdf',
                contentType: 'application/pdf',
                disposition: 'attachment',
                size: 20480,
                data: 'report'
            }
        ],
        duration: 80,
    },
    {
        from: [{name: 'Alice', address: 'alice@mokapi.io'}],
        to: [{name: 'Bob', address: 'bob@mokapi.io'}],
        date: '2023-02-19T14:30:00+01:00',
        contentType: 'text/plain',
        contentTransferEncoding: '7bit',
        messageId: '20230219-143000.002-2002@mokapi.io',
        subject: 'Handled by a script',
        body: 'This mail was processed by an event handler.',
        attachments: [],
        duration: 45,
    },
    {
        from: [{name: 'Frank', address: 'frank@second.io'}],
        to: [{address: 'grace@second.io'}],
        date: '2023-02-17T16:00:00+01:00',
        contentType: 'text/plain',
        contentTransferEncoding: '7bit',
        messageId: '20230217-160000.003-2003@second.io',
        subject: 'Hello from the second server',
        body: 'Mail on the second server.',
        attachments: [],
        duration: 60,
    },
]

export const services = [
    {
        name: 'Mail Testserver',
        description: 'This is a sample smtp server',
        version: '1.0',
        servers: [
            {
                name: 'smtp',
                description: "Mokapi's SMTP Server",
                host: 'localhost:8025',
                protocol: 'smtp'
            },
            {
                name: 'imap',
                description: "Mokapi's IMAP Server",
                host: 'localhost:8030',
                protocol: 'imap'
            },
        ],
        mailboxes: [
            {
                name: 'bob@mokapi.io',
                username: 'bob',
                password: 'foo',
                numMessages: 2,
                folders: [
                    {
                        name: 'INBOX',
                        messages: mails
                    }
                ]
            },
            {
                name: 'alice@mokapi.io',
                username: 'alice',
                password: 'foo',
                description: 'a description using *markdown*',
                numMessages: 0,
                folders: []
            },
        ],
        rules: [
            {name: 'mokapi.io', sender: '.*@mokapi.io', action: 'allow'},
            {name: 'spam', recipient: '.*@foo.bar', subject: 'spam', body: 'spam', action: 'deny'}
        ],
        settings: { maxRecipients: 0, autoCreateMailbox: true },
    },
    {
        name: 'Mail Second Server',
        description: 'Another sample smtp server',
        version: '1.0',
        servers: [
            {
                name: 'smtp',
                description: "Mokapi's SMTP Server",
                host: 'localhost:2525',
                protocol: 'smtp'
            },
        ],
        mailboxes: [],
        rules: [],
        settings: { maxRecipients: 0, autoCreateMailbox: true },
    }
]

export const mailEvents = [
    {
        id: "273cd167-f5a5-4da1-969e-d44213686491",
        traits: {
            namespace: "mail",
            name: "Mail Testserver",
        },
        time:  '2023-02-23T08:49:25.482366+01:00',
        data: {
            from: 'alice@mokapi.io',
            to: ['bob@mokapi.io'],
            messageId: '20230223-084925.763-4196@mokapi.io',
            subject: 'A test mail',
            duration: 30016,
        },
    },
    {
        id: "0b0f1b9e-1e4c-4a5b-8d0e-6f1d2a7c9e11",
        traits: {
            namespace: "mail",
            name: "Mail Testserver",
        },
        time:  '2023-02-22T10:15:00.482366+01:00',
        data: {
            from: 'alice@mokapi.io',
            to: ['bob@mokapi.io'],
            messageId: '20230222-101500.123-1111@mokapi.io',
            subject: 'A plain text mail',
            duration: 120,
        },
    },
]

function mailEvent(id: string, time: string, data: object, service = 'Mail Testserver') {
    return {
        id,
        traits: { namespace: 'mail', name: service },
        time,
        data: { from: '', to: [], messageId: '', subject: '', duration: 0, error: '', actions: null, ...data },
    }
}

// Appended after the first two events, in display order (the mock doesn't sort):
// rejections as the backend logs them, a mail with attachments, a mail with an event handler,
// a mail that isn't stored anymore and events of a second service.
mailEvents.push(
    // MAIL FROM rejected: no messageId, no subject
    mailEvent('a1000000-0000-4000-8000-000000000001', '2023-02-21T11:00:00.100000+01:00', {
        error: 'Sender spammer@foo.bar is not allowed',
    }),
    // A rule rejection in DATA: the event of processMail (mail is never stored) ...
    mailEvent('a1000000-0000-4000-8000-000000000002', '2023-02-21T10:30:00.200000+01:00', {
        from: 'mallory@foo.bar',
        to: ['bob@mokapi.io'],
        messageId: '20230221-103000.004-3001@foo.bar',
        subject: 'Cheap spam',
        duration: 15,
    }),
    // ... and the rejected event without messageId
    mailEvent('a1000000-0000-4000-8000-000000000003', '2023-02-21T10:30:00.100000+01:00', {
        from: 'mallory@foo.bar',
        to: ['bob@mokapi.io'],
        error: 'Mail rejected by rule spam',
    }),
    mailEvent('a1000000-0000-4000-8000-000000000004', '2023-02-20T09:00:00.300000+01:00', {
        from: 'carol@mokapi.io',
        to: ['bob@mokapi.io'],
        messageId: '20230220-090000.001-2001@mokapi.io',
        subject: 'Quarterly report',
        duration: 80,
    }),
    mailEvent('a1000000-0000-4000-8000-000000000005', '2023-02-19T14:30:00.400000+01:00', {
        from: 'alice@mokapi.io',
        to: ['bob@mokapi.io'],
        messageId: '20230219-143000.002-2002@mokapi.io',
        subject: 'Handled by a script',
        duration: 45,
        actions: [
            {
                duration: 20,
                tags: {
                    name: 'dashboard',
                    file: 'examples/mokapi/smtp_handler.js',
                    event: 'smtp',
                },
                parameters: [
                    JSON.stringify({
                        from: 'alice@mokapi.io',
                        to: ['bob@mokapi.io'],
                        subject: 'Handled by a script',
                    }),
                ],
            },
        ],
    }),
    // accepted, but the mail was deleted from the mailbox store
    mailEvent('a1000000-0000-4000-8000-000000000006', '2023-02-18T12:00:00.500000+01:00', {
        from: 'dave@mokapi.io',
        to: ['bob@mokapi.io'],
        messageId: '20230218-120000.005-4001@mokapi.io',
        subject: 'Deleted mail',
        duration: 25,
    }),
    mailEvent('a1000000-0000-4000-8000-000000000007', '2023-02-17T16:00:00.600000+01:00', {
        from: 'frank@second.io',
        to: ['grace@second.io'],
        messageId: '20230217-160000.003-2003@second.io',
        subject: 'Hello from the second server',
        duration: 60,
    }, 'Mail Second Server'),
    mailEvent('a1000000-0000-4000-8000-000000000008', '2023-02-17T15:00:00.700000+01:00', {
        from: 'frank@second.io',
        error: 'Recipient is not known',
    }, 'Mail Second Server'),
    mailEvent('a1000000-0000-4000-8000-000000000009', '2023-02-17T14:00:00.800000+01:00', {
        from: 'frank@second.io',
        to: ['grace@second.io'],
        messageId: '20230217-140000.006-5001@second.io',
        subject: 'Second server, gone',
        duration: 35,
    }, 'Mail Second Server'),
    mailEvent('a1000000-0000-4000-8000-000000000010', '2023-02-17T13:00:00.900000+01:00', {
        from: 'frank@second.io',
        to: ['grace@second.io'],
        messageId: '20230217-130000.007-5002@second.io',
        subject: 'Another one from the second server',
        duration: 40,
    }, 'Mail Second Server'),
)


export function getMail(messageId: string) {
    for (const m of [...mails, ...storedMails]) {
        if (m.messageId === messageId) {
            return m
        }
    }
    return null
}

export function getAttachment(messageId: string, attachmentName: string) {
    const mail = getMail(messageId)
    if (!mail) {
        return null
    }

    for (const a of mail.attachments ?? []) {
        if (a.name === attachmentName || (a as { contentId?: string }).contentId === attachmentName) {
            return a
        }
    }
    return null
}