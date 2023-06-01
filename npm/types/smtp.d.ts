type SmtpEventHandler = (record: Message) => boolean

declare interface Message {
    server: string
    sender?: Address
    from: Address[]
    to: Address[]
    replyTo?: Address[]
    cc?: Address[]
    bcc?: Address[]
    messageId: string
    inReplyTo?: string
    time?: Date
    subject: string
    contentType: string
    encoding: string
    body: string
    attachments: Attachment[]
}

declare interface Address {
    name?: string
    address: string
}

declare interface Attachment {
    name: string
    contentType: string
    data: Uint8Array
}