package runtime

import (
	"mokapi/config/dynamic/smtp"
)

type SmtpInfo struct {
	*smtp.Config
}
