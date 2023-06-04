declare module 'mokapi/mail' {
    /**
     * Sends an email message to an SMTP server for delivery.
     * @param server Host to which the message is to be sent.
     * @param message a Message that contains the message to send.
     */
    function send(server: string, message: Message)
}

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