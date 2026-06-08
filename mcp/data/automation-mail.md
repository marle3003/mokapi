# Mail

Interfaces for exploring Mail mailboxes and mail messages

```typescript
interface Mail extends ApiSummary {
    servers: { host: string, protocol: 'smtp' | 'smtps' | 'imap' | 'imaps', description: string }

    /**
     * Returns all operations of this API.
     */
    getMailboxes(): MailboxSummary[];

    /**
     * Returns details about specific mailbox
     * @param name The name of the mailbox
     * @example foo@example.com
     */
    getMailbox(name: string): HttpOperation

    /**
     * Sends a mail message to the recipient to
     * @param to The recipient of the mail message
     * @param message The mail message to send
     */
    sendMail(to: string, message: MailMessage)
}

interface MailboxSummary {
    name: string
    description: string
}

interface Mailbox {
    name: string
    description: string
    username: string
    password: string
    folders: Record<string, MailboxFolder>
}

interface MailboxFolder {
    name: string
    flags: string[]
    folders: Record<string, MailboxFolder>
    mails: MailMessage[]
}

interface MailMessage {
    server: string
    sender?: MailAddress
    from: MailAddress[]
    to: MailAddress[]
    replyTo?: MailAddress[]
    cc?: MailAddress[]
    bcc?: MailAddress[]
    messageId: string
    inReplyTo?: MailAddress[]
    date: string
    subject: string
    contentType?: string
    contentTransferEncoding?: string
    body: string
    attachments?: MailAttachment[]
    size: number
}

interface MailAddress {
    name: string
    address: string
}

interface MailAttachment {
    name: string
    contentType: string
    contentTransferEncoding: string
    contentDescription: string
    disposition: string
    data: Uint8Array[]
    contentId: string
}
```