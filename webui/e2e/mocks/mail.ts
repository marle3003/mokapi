import { app } from 'mokapi'

export default function () {
    app.http()
        .operation('serviceMail', (_, res) => {
            res.data = services[0]
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
}

export let mails = [
    {
        from: [{name: 'Alice', address: 'alice@mokapi.io'}],
        to: [{name: 'Bob', address: 'bob@mokapi.io'}, {address: 'carol@mokapi.io'}],
        date: '2023-02-23T08:49:25+01:00',
        contentType: 'text/html',
        encoding: 'quoted-printable',
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
                data: open('icon.png', {as: 'binary'})
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
]

export let services = [
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
                name: 'alice@mokapi.io',
                username: 'alice',
                password: 'foo',
                description: 'a description using *markdown*',
                numMessages: 0,
                folders: []
            },
            {
                name: 'bob@mokapi.io',
                username: 'bob',
                password: 'foo',
                numMessages: 1,
                folders: [
                    {
                        name: 'INBOX',
                        messages: mails
                    }
                ]
            },
        ],
        rules: [
            {name: 'mokapi.io', sender: '.*@mokapi.io', action: 'allow'},
            {name: 'spam', recipient: '.*@foo.bar', subject: 'spam', body: 'spam', action: 'deny'}
        ],
        settings: { maxRecipients: 0, autoCreateMailbox: true },
    }
]

export let mailEvents = [
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
]

export function getMail(messageId: string) {
    for (let m of mails) {
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

    for (let a of mail.attachments) {
        if (a.name === attachmentName) {
            return a
        }
    }
    return null
}