package mail

import "embed"

//go:embed templates/*.html
var Templates embed.FS

const (
	EmailVerificationTemplate = "templates/email_verification.html"
	PasswordResetTemplate     = "templates/password_reset.html"
	ReminderReminderTemplate  = "templates/reminder_reminder.html"
)
