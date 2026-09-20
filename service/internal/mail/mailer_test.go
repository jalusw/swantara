package mail

import (
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/config"
)

func TestTemplatesEmbedded(t *testing.T) {
	for _, name := range []string{EmailVerificationTemplate, PasswordResetTemplate} {
		if _, err := Templates.ReadFile(name); err != nil {
			t.Errorf("embedded template %q not found: %v", name, err)
		}
	}
}

func TestNew_EncryptionModes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		enc   string
		user  string
		valid bool
	}{
		{name: "no encryption", enc: "none"},
		{name: "ssl", enc: "ssl"},
		{name: "tls", enc: "tls"},
		{name: "unknown falls back to no tls", enc: "garbage"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				SMTPHost:       "localhost",
				SMTPPort:       1025,
				SMTPEncryption: tc.enc,
				SMTPFromEmail:  "noreply@swantara.com",
				SMTPFromName:   "Swantara",
			}
			m, err := New(cfg)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if m == nil {
				t.Fatal("New() returned nil mailer")
			}
		})
	}
}

func TestNew_WithAuth(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:       "localhost",
		SMTPPort:       1025,
		SMTPUsername:   "user",
		SMTPPassword:   "pass",
		SMTPEncryption: "tls",
		SMTPFromEmail:  "noreply@swantara.com",
		SMTPFromName:   "Swantara",
	}
	if _, err := New(cfg); err != nil {
		t.Fatalf("New() error = %v", err)
	}
}

func TestSend_InvalidFrom(t *testing.T) {
	m, err := New(&config.Config{
		SMTPHost:      "localhost",
		SMTPPort:      1025,
		SMTPFromEmail: "not-an-email",
		SMTPFromName:  "Swantara",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := m.Send("user@example.com", "subject", "<p>body</p>"); err == nil {
		t.Error("Send() expected error for invalid from address")
	}
}

func TestSend_InvalidTo(t *testing.T) {
	m, err := New(&config.Config{
		SMTPHost:      "localhost",
		SMTPPort:      1025,
		SMTPFromEmail: "noreply@swantara.com",
		SMTPFromName:  "Swantara",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := m.Send("not-an-email", "subject", "<p>body</p>"); err == nil {
		t.Error("Send() expected error for invalid to address")
	}
}

func TestSend_DialFailure(t *testing.T) {
	m, err := New(&config.Config{
		SMTPHost:      "127.0.0.1",
		SMTPPort:      1,
		SMTPFromEmail: "noreply@swantara.com",
		SMTPFromName:  "Swantara",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := m.Send("user@example.com", "subject", "<p>body</p>"); err == nil {
		t.Error("Send() expected error when SMTP server unreachable")
	}
}
