declare interface SmtpService extends Service {
    server: string
    maxRecipients: number
    mailboxes: SmtpMailbox[]
    rules: SmtpRule[]
}

declare interface SmtpMailbox{
    name: string
    username: string
    password: string
}

declare interface SmtpRule {
    name: string
    sender: string    
	recipient: string    
	subject: string    
	body: string    
	action: RuleAction 
}

declare enum RuleAction {
    allow = 'allow',
    deny = 'deny'
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
	time: number
	subject: string      
	contentType: string      
	encoding: string      
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
}