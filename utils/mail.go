package utils

import (
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/noirbizarre/gonja"
)

const mailSendTimeout = 15 * time.Second

type MailConfig struct {
	Host      string
	Port      string
	Username  string
	Password  string
	FromEmail string
	BaseUrl   string
}

func (config MailConfig) IsConfigured() bool {
	return config.Host != "" && config.Port != "" && config.Username != "" && config.Username != "xxxx"
}

func (config MailConfig) address() string {
	return fmt.Sprintf("%s:%s", config.Host, config.Port)
}

func renderEmail(vars gonja.Context, template string, baseUrl string) (string, error) {
	vars["scriptable_base_url"] = baseUrl

	view, err := gonja.Must(gonja.FromFile("templates/emails/" + template + ".jinja")).Execute(vars)
	if err != nil {
		return "", err
	}

	vars["view"] = view
	return gonja.Must(gonja.FromFile("templates/emails/master.jinja")).Execute(vars)
}

func buildMessage(from string, to string, subject string, body string) []byte {
	message := "From: " + from + "\r\n"
	message += "To: " + to + "\r\n"
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-version: 1.0;\r\n"
	message += "Content-Type: text/html; charset=\"UTF-8\";\r\n"
	message += "Content-Transfer-Encoding: 7bit;\r\n"
	message += "\r\n"
	message += body
	return []byte(message)
}

func envelopeAddress(from string) string {
	parsed, err := mail.ParseAddress(from)
	if err != nil {
		return from
	}

	return parsed.Address
}

func sendWithTimeout(config MailConfig, from string, recipients []string, message []byte) error {
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	from = envelopeAddress(from)

	done := make(chan error, 1)
	go func() {
		done <- smtp.SendMail(config.address(), auth, from, recipients, message)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(mailSendTimeout):
		return errors.New("timed out talking to " + config.address())
	}
}

func SendEmail(config MailConfig, subject string, from string, recipients []string, vars gonja.Context, template string) error {
	if !config.IsConfigured() {
		fmt.Println("Oops! SMTP mail is not configured. Skipping sending this email.")
		return errors.New("SMTP mail is not configured")
	}

	if from == "" {
		from = config.FromEmail
	}

	body, err := renderEmail(vars, template, config.BaseUrl)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = sendWithTimeout(config, from, recipients, buildMessage(from, recipients[0], subject, body))
	if err != nil {
		fmt.Println("Mail send failed:", err, "Subject:", subject, "Recipients:", recipients)
	}

	return err
}
