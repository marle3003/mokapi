export type MailTab = 'Body' | 'Attachments' | 'Details'
export type MailStatus = 'Accepted' | 'Rejected'

// A mail is rejected when the event carries an error. The backend logs rejections of MAIL FROM, RCPT and
// authentication with an error, a rule rejection in DATA once with and once without one.
export function mailStatus(data: SmtpEventData): MailStatus {
    return data.error ? 'Rejected' : 'Accepted'
}

// The sub tabs of the panel of a mail. A rejected mail has no content. A mail that is not stored anymore
// (`mail` is null) shows Body (with a message that the content is gone) and Details.
export function mailTabs(data: SmtpEventData, mail: Message | null | undefined): MailTab[] {
    if (mailStatus(data) === 'Rejected') {
        return ['Details']
    }
    const hasAttachments = (mail?.data.attachments?.length ?? 0) > 0
    return hasAttachments ? ['Body', 'Attachments', 'Details'] : ['Body', 'Details']
}

// The sub tab belongs to the panel, not to the mail: the next mail opens on the same tab if it has it.
export function activeMailTab(tabs: MailTab[], selected: MailTab): MailTab {
    return tabs.includes(selected) ? selected : tabs[0]
}

export function orDash(value: string | undefined): string {
    return value || '-'
}

export function recipients(to: string[] | null | undefined): string {
    return orDash(to?.join(', '))
}
