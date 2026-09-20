package mail

import (
	"fmt"

	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/wneessen/go-mail"
)

type Mailer struct {
	client    *mail.Client
	fromEmail string
	fromName  string
}

func New(config *config.Config) (*Mailer, error) {
	opts := []mail.Option{
		mail.WithPort(config.SMTPPort),
		mail.WithUsername(config.SMTPUsername),
		mail.WithPassword(config.SMTPPassword),
	}

	switch config.SMTPEncryption {
	case "ssl":
		opts = append(opts, mail.WithSSL())
	case "tls":
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	default:
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	}

	if config.SMTPUsername != "" {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover))
	} else {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthNoAuth))
	}

	client, err := mail.NewClient(config.SMTPHost, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	return &Mailer{
		client:    client,
		fromEmail: config.SMTPFromEmail,
		fromName:  config.SMTPFromName,
	}, nil
}

func (m *Mailer) Send(to, subject, body string) error {
	msg := mail.NewMsg()

	if err := msg.FromFormat(m.fromName, m.fromEmail); err != nil {
		return fmt.Errorf("failed to set from: %w", err)
	}

	if err := msg.To(to); err != nil {
		return fmt.Errorf("failed to set to: %w", err)
	}

	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body)

	if err := m.client.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
