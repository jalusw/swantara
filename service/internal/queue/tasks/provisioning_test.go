package tasks

import (
	"encoding/json"
	"testing"
)

func TestNewOrganizationProvisioningTask(t *testing.T) {
	task, err := NewOrganizationProvisioningTask(11, "req-1")
	if err != nil {
		t.Fatalf("constructor error = %v", err)
	}
	if task.Type() != TypeOrganizationProvisioning {
		t.Errorf("Type() = %q, want %q", task.Type(), TypeOrganizationProvisioning)
	}
	var payload OrganizationProvisioningPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if payload.OrganizationID != 11 || payload.RequestID != "req-1" {
		t.Errorf("payload = %+v", payload)
	}
}
