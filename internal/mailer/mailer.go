package mailer

import (
	"bytes"
	"html/template"

	"github.com/wneessen/go-mail"
)

type EmailService struct {
	Templates *template.Template
	Client    *mail.Client
	From      string
}

type VerificationEmailData struct {
	Email     string
	VerifyURL string
}

func NewEmailService(port int, host, usermame, password string) (*EmailService, error) {
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	client, err := mail.NewClient(host,
		mail.WithPassword(password),
		mail.WithUsername(usermame),
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithTLSPortPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return nil, err
	}

	return &EmailService{
		Templates: tmpl,
		Client:    client,
		From:      usermame,
	}, nil
}

func (s *EmailService) SendVerificationEmail(to string, data VerificationEmailData) error {
	var body bytes.Buffer

	err := s.Templates.ExecuteTemplate(&body, "verification.html", data)
	if err != nil {
		return err
	}

	m := mail.NewMsg()
	if err := m.From(s.From); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}

	m.Subject("Action Required. Verify Your Email")
	m.SetBodyString(mail.TypeTextHTML, body.String())

	if err := s.Client.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
