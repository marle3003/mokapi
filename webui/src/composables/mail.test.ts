import { describe, expect, it } from 'vitest'
import { activeMailTab, mailStatus, mailTabs, orDash, recipients } from './mail'

const accepted = { from: 'alice@mokapi.io', to: ['bob@mokapi.io'], messageId: 'm1', subject: 'Hi', duration: 1, error: '', actions: [] } as SmtpEventData
const rejected = { ...accepted, messageId: '', error: 'Sender not allowed' } as SmtpEventData

function mailWith(attachments: unknown[]) {
  return { service: 'Mail Testserver', data: { attachments } } as Message
}

describe('mailStatus', () => {
  it('is Accepted without an error', () => {
    expect(mailStatus(accepted)).toBe('Accepted')
  })

  it('is Rejected with an error', () => {
    expect(mailStatus(rejected)).toBe('Rejected')
  })

  it('is Accepted when the event has no error field', () => {
    expect(mailStatus({ ...accepted, error: undefined } as unknown as SmtpEventData)).toBe('Accepted')
  })
})

describe('mailTabs', () => {
  it('has Body and Details for a mail without attachments', () => {
    expect(mailTabs(accepted, mailWith([]))).toEqual(['Body', 'Details'])
  })

  it('has Attachments only when the mail has some', () => {
    expect(mailTabs(accepted, mailWith([{ name: 'foo.txt' }]))).toEqual(['Body', 'Attachments', 'Details'])
  })

  it('has only Details for a rejected mail', () => {
    expect(mailTabs(rejected, null)).toEqual(['Details'])
  })

  it('has Body and Details for a mail that is not stored anymore', () => {
    expect(mailTabs(accepted, null)).toEqual(['Body', 'Details'])
  })
})

describe('activeMailTab', () => {
  it('keeps the selected tab the mail has', () => {
    expect(activeMailTab(['Body', 'Attachments', 'Details'], 'Attachments')).toBe('Attachments')
  })

  it('falls back to the first tab when the mail has not the selected one', () => {
    expect(activeMailTab(['Body', 'Details'], 'Attachments')).toBe('Body')
    expect(activeMailTab(['Details'], 'Body')).toBe('Details')
  })
})

describe('display values', () => {
  it('shows - for an empty value', () => {
    expect(orDash('')).toBe('-')
    expect(orDash(undefined)).toBe('-')
    expect(orDash('Hi')).toBe('Hi')
  })

  it('joins recipients and shows - for none', () => {
    expect(recipients(['a@x.io', 'b@x.io'])).toBe('a@x.io, b@x.io')
    expect(recipients([])).toBe('-')
    expect(recipients(null)).toBe('-')
  })
})
