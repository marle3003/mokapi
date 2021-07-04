package api

import (
	"mokapi/models"
	"time"
)

type mailSummary struct {
	Id   string    `json:"id"`
	From string    `json:"from"`
	To   string    `json:"to"`
	Time time.Time `json:"time"`
}

type mail struct {
	From string `json:"from"`
	To   string `json:"to"`
	Data string `json:"data"`
}

func newMailSummary(mail *models.Mail) mailSummary {
	return mailSummary{
		Id:   mail.Id,
		From: mail.From,
		To:   mail.To,
		Time: mail.Time,
	}
}

func newMail(m *models.Mail) mail {
	return mail{
		From: m.From,
		To:   m.To,
		Data: m.Data,
	}
}
