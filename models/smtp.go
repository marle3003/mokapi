package models

import "time"

type Mail struct {
	Id   string
	From string
	To   string
	Data string

	Time time.Time
}

func (m *Metrics) AddMail(mail *Mail) {
	mail.Id = newId(10)
	m.TotalMails += 1
	if len(m.LastMails) > 10 {
		m.LastMails = m.LastMails[1:]
	}
	mail.Time = time.Now()
	m.LastMails = append(m.LastMails, mail)
}
