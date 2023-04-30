declare interface SmtpService extends Service {
    server: string
    mailboxes: SmtpMailbox[]
    rules: SmtpRule[]
}

declare interface SmtpMailbox{
    name: string
    username: string
    password: string
}

declare interface SmtpRule {
    sender:    string    
	recipient: string    
	subject:   string    
	body:      string    
	action:   RuleAction 
}

declare enum RuleAction {
    allow = 'allow',
    deny = 'deny'
}

declare interface SmtpEventData {
    from: string
    to: string[]
    mail: Mail
    duration: number
    actions: Action[]
}

declare interface Mail {
    Sender: MailAddress    
	From: MailAddress[]
	To: MailAddress[] 
	ReplyTo: MailAddress[]  
	Cc: MailAddress[]    
	Bcc: MailAddress[]    
	MessageId: string       
	Date: number
	Subject: string      
	ContentType: string      
	Encoding: string      
	Body: string    
}

declare interface MailAddress {
    name: string
    address: string
}