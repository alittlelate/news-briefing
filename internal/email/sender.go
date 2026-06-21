package email

import (
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"strings"
)

func Send(subject, htmlBody string) error {
	from := os.Getenv("EMAIL_FROM")
	to := os.Getenv("EMAIL_TO")
	password := os.Getenv("EMAIL_APP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	if from == "" || to == "" || password == "" {
		return fmt.Errorf("EMAIL_FROM, EMAIL_TO, EMAIL_APP_PASSWORD not set")
	}

	auth := smtp.PlainAuth("", from, password, smtpHost)

	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("UTF-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		htmlBody,
	}, "\r\n")

	return smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{to},
		[]byte(msg),
	)
}
