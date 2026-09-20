package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
)

type OrganizationProvisioner interface {
	Provision(ctx context.Context, organizationID uint64) error
}

type OrganizationProvisioningHandler struct {
	organizations OrganizationProvisioner
}

func NewOrganizationProvisioningHandler(organizations OrganizationProvisioner) OrganizationProvisioningHandler {
	return OrganizationProvisioningHandler{organizations: organizations}
}

func (h OrganizationProvisioningHandler) Handle(ctx context.Context, t *asynq.Task) (err error) {
	var payload tasks.OrganizationProvisioningPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	ctx = helper.ContextWithRequestID(ctx, payload.RequestID)

	defer func() {
		if err != nil {
			slog.Error("task failed", "task", t.Type(), "organization_id", payload.OrganizationID, "request_id", payload.RequestID, "error", err)
		}
	}()

	slog.Info("task started", "task", t.Type(), "organization_id", payload.OrganizationID, "request_id", payload.RequestID)

	if err := h.organizations.Provision(ctx, payload.OrganizationID); err != nil {
		return fmt.Errorf("failed to provision organization: %w", err)
	}

	slog.Info("task completed", "task", t.Type(), "organization_id", payload.OrganizationID, "request_id", payload.RequestID)

	return nil
}
