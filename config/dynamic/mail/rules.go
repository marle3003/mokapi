package mail

import (
	"fmt"
	"mokapi/smtp"
)

func (r Rules) RunSender(sender string) *RejectResponse {
	for _, rule := range r {
		if rule.Sender != nil {
			match := rule.Sender.Match(sender)
			if match && rule.Action == Deny {
				if rule.RejectResponse != nil {
					return rule.RejectResponse
				}
				return &RejectResponse{
					StatusCode:         smtp.AddressRejected.Code,
					EnhancedStatusCode: smtp.AddressRejected.Status,
					Text:               rule.formatText("sender %v does match deny rule: %v", sender, rule.Sender),
				}
			} else if !match && rule.Action == Allow {
				if rule.RejectResponse != nil {
					return rule.RejectResponse
				}
				return &RejectResponse{
					StatusCode:         smtp.AddressRejected.Code,
					EnhancedStatusCode: smtp.AddressRejected.Status,
					Text:               rule.formatText("sender %v does not match allow rule: %v", sender, rule.Sender),
				}
			}
		}
	}
	return nil
}

func (r Rules) runRcpt(to string) *RejectResponse {
	for _, rule := range r {
		if rule.Recipient != nil {
			match := rule.Recipient.Match(to)
			if match && rule.Action == Deny {
				if rule.RejectResponse != nil {
					return rule.RejectResponse
				}
				return &RejectResponse{
					StatusCode:         smtp.AddressRejected.Code,
					EnhancedStatusCode: smtp.AddressRejected.Status,
					Text:               rule.formatText("recipient %v does match deny rule: %v", to, rule.Recipient),
				}
			} else if !match && rule.Action == Allow {
				if rule.RejectResponse != nil {
					return rule.RejectResponse
				}
				return &RejectResponse{
					StatusCode:         smtp.AddressRejected.Code,
					EnhancedStatusCode: smtp.AddressRejected.Status,
					Text:               rule.formatText("recipient %v does not match allow rule: %v", to, rule.Recipient),
				}
			}
		}
	}
	return nil
}

func (r Rules) runMail(m *smtp.Message) *RejectResponse {
	for _, r := range r {
		if res := r.runSubject(m.Subject); res != nil {
			return res
		}
		if res := r.runBody(m.Body); res != nil {
			return res
		}
	}
	return nil
}

func (r Rule) runSubject(subject string) *RejectResponse {
	if r.Subject == nil {
		return nil
	}
	match := r.Subject.Match(subject)
	if match && r.Action == Deny {
		if r.RejectResponse != nil {
			return r.RejectResponse
		}
		return &RejectResponse{
			StatusCode:         smtp.MailReject.Code,
			EnhancedStatusCode: smtp.MailReject.Status,
			Text:               r.formatText("subject %v does match deny rule: %v", subject, r.Subject),
		}
	} else if !match && r.Action == Allow {
		if r.RejectResponse != nil {
			return r.RejectResponse
		}
		return &RejectResponse{
			StatusCode:         smtp.MailReject.Code,
			EnhancedStatusCode: smtp.MailReject.Status,
			Text:               r.formatText("subject %v does not match allow rule: %v", subject, r.Subject),
		}
	}
	return nil
}

func (r Rule) runBody(body string) *RejectResponse {
	if r.Body == nil {
		return nil
	}
	match := r.Body.Match(body)
	if match && r.Action == Deny {
		if r.RejectResponse != nil {
			return r.RejectResponse
		}
		return &RejectResponse{
			StatusCode:         smtp.MailReject.Code,
			EnhancedStatusCode: smtp.MailReject.Status,
			Text:               r.formatText("body %v does match deny rule: %v", body, r.Body),
		}
	} else if !match && r.Action == Allow {
		if r.RejectResponse != nil {
			return r.RejectResponse
		}
		return &RejectResponse{
			StatusCode:         smtp.MailReject.Code,
			EnhancedStatusCode: smtp.MailReject.Status,
			Text:               r.formatText("body %v does not match allow rule: %v", body, r.Body),
		}
	}
	return nil
}

func (r Rule) formatText(format string, a ...interface{}) string {
	s := fmt.Sprintf(format, a...)
	if len(r.Name) > 0 {
		return fmt.Sprintf("rule %v: %v", r.Name, s)
	}
	return s
}
