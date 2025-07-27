declare interface MailService extends Service {
    servers?: MailServer[]
    mailboxes?: SmtpMailbox[]
    rules?: SmtpRule[]
    settings: MailSettings
}

declare interface MailServer {
    host: string
    protocol: string
    name: string
    description: string
}

declare interface SmtpMailbox {
    name: string
    username: string
    password: string
    description: string
    numMessages: number
    folders?: string[]
}

declare interface SmtpRule {
    name: string
    description?: string
    sender: string
	recipient: string
	subject: string
	body: string
	action: RuleAction
    rejectResponse?: RejectResponse
}

declare enum RuleAction {
    allow = 'allow',
    deny = 'deny'
}

declare interface RejectResponse {
    statusCode: string
    enhancedStatusCode: string
    message: string
}

declare interface SmtpEventData {
    from: string
    to: string[]
    messageId: string
    subject: string
    duration: number
    error: string
    actions: Action[]
}

declare interface Mail {
    sender: MailAddress
    from: MailAddress[]
	to: MailAddress[]
	replyTo: MailAddress[]
	cc: MailAddress[]
	bcc: MailAddress[]
	messageId: string
    inReplyTo: string
	date: number
	subject: string
	contentType: string
	contentTransferEncoding: string
	body: string
    attachments: Attachment[]
}

declare interface MailAddress {
    name: string
    address: string
}

declare interface Attachment {
    name: string
    contentType: string
    size: number
    disposition: string
    contentId?: string
}

declare interface MailSettings {
    maxRecipients: number
    autoCreateMailbox: boolean
}