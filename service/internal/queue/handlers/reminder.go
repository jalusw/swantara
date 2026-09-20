package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/mail"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
)

type SendReminderHandler struct {
	mailer       *mail.Mailer
	actions      accounting.ReminderActionDAO
	invoices     accounting.InvoiceDAO
	levels       accounting.ReminderLevelDAO
	contacts     contacts.ContactDAO
	templateName string
	subject      string
	now          func() time.Time
}

func NewSendReminderHandler(
	mailer *mail.Mailer,
	actions accounting.ReminderActionDAO,
	invoices accounting.InvoiceDAO,
	levels accounting.ReminderLevelDAO,
	contacts contacts.ContactDAO,
	templateName, subject string,
) SendReminderHandler {
	return SendReminderHandler{
		mailer:       mailer,
		actions:      actions,
		invoices:     invoices,
		levels:       levels,
		contacts:     contacts,
		templateName: templateName,
		subject:      subject,
		now:          time.Now,
	}
}

func (h SendReminderHandler) Handle(ctx context.Context, t *asynq.Task) (err error) {
	var payload tasks.SendReminderPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	defer func() {
		if err != nil {
			slog.Error("task failed", "task", t.Type(), "action_id", payload.ActionID, "error", err)
		}
	}()

	action, err := h.actions.Find(ctx, payload.ActionID)
	if err != nil {
		return fmt.Errorf("failed to find reminder action: %w", err)
	}
	if action == nil || action.SentAt != nil {
		return nil
	}
	invoice, err := h.invoices.Find(ctx, payload.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to find invoice: %w", err)
	}
	if invoice == nil || !invoice.AmountResidual.IsPositive() {
		return nil
	}
	contact, err := h.contacts.Find(ctx, payload.ContactID)
	if err != nil {
		return fmt.Errorf("failed to find contact: %w", err)
	}
	if contact == nil || contact.Email == nil || *contact.Email == "" {
		slog.Warn("task skipped", "task", t.Type(), "action_id", payload.ActionID, "reason", "contact_email_missing")
		return nil
	}
	level, err := h.levels.Find(ctx, action.LevelID)
	if err != nil {
		return fmt.Errorf("failed to find reminder level: %w", err)
	}

	daysOverdue := 0
	due := ""
	if invoice.DueDate != nil {
		now := h.now().UTC()
		daysOverdue = int(now.Sub(*invoice.DueDate).Hours() / 24)
		due = invoice.DueDate.Format("2006-01-02")
	}
	levelName := ""
	if level != nil {
		levelName = level.Name
	}
	tmpl, err := template.ParseFS(mail.Templates, h.templateName)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	invoiceName := fmt.Sprintf("#%d", invoice.ID)
	if invoice.Name != nil && *invoice.Name != "" {
		invoiceName = *invoice.Name
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]any{
		"ContactName": contact.Name,
		"InvoiceName": invoiceName,
		"Amount":      invoice.AmountResidual.Float64(),
		"DueDate":     due,
		"DaysOverdue": daysOverdue,
		"LevelName":   levelName,
	}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	if err := h.mailer.Send(*contact.Email, h.subject, body.String()); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	sentAt := h.now().UTC()
	action.SentAt = &sentAt
	if _, err := h.actions.Update(ctx, action); err != nil {
		return fmt.Errorf("failed to stamp sent at: %w", err)
	}
	return nil
}
