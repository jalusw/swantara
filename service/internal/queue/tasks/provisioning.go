package tasks

import "github.com/hibiken/asynq"

const TypeOrganizationProvisioning = "organization-provisioning"

type OrganizationProvisioningPayload struct {
	OrganizationID uint64 `json:"organization_id"`
	RequestID      string `json:"request_id,omitempty"`
}

func NewOrganizationProvisioningTask(organizationID uint64, requestID string) (*asynq.Task, error) {
	return newTask(TypeOrganizationProvisioning, OrganizationProvisioningPayload{
		OrganizationID: organizationID,
		RequestID:      requestID,
	})
}
