package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/mail"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
)

type SendEmailHandler struct {
	mailer       *mail.Mailer
	userDAO      iam.UserDAO
	templateName string
	baseURL      string
	urlKey       string
	subject      string
}

func NewSendEmailHandler(
	mailer *mail.Mailer,
	userDAO iam.UserDAO,
	templateName, baseURL, urlKey, subject string,
) SendEmailHandler {
	return SendEmailHandler{
		mailer:       mailer,
		userDAO:      userDAO,
		templateName: templateName,
		baseURL:      baseURL,
		urlKey:       urlKey,
		subject:      subject,
	}
}

func (h SendEmailHandler) Handle(ctx context.Context, t *asynq.Task) (err error) {
	var payload tasks.SendEmailPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	ctx = helper.ContextWithRequestID(ctx, payload.RequestID)

	defer func() {
		if err != nil {
			slog.Error("task failed", "task", t.Type(), "user_id", payload.UserID, "request_id", payload.RequestID, "error", err)
		}
	}()

	slog.Info("task started", "task", t.Type(), "user_id", payload.UserID, "request_id", payload.RequestID)

	u, err := h.userDAO.Find(ctx, payload.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if u == nil {
		slog.Warn("task skipped", "task", t.Type(), "user_id", payload.UserID, "request_id", payload.RequestID, "reason", "user_not_found")
		return nil
	}

	tmpl, err := template.ParseFS(mail.Templates, h.templateName)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]any{
		"FirstName": u.FirstName,
		"LastName":  u.LastName,
		h.urlKey:    fmt.Sprintf("%s/%s", h.baseURL, payload.Token),
	}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := h.mailer.Send(u.Email, h.subject, body.String()); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("task completed", "task", t.Type(), "user_id", payload.UserID, "request_id", payload.RequestID)

	return nil
}
