package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
)

type provisionerStub struct {
	err   error
	orgID uint64
	calls int
}

func (s *provisionerStub) Provision(_ context.Context, orgID uint64) error {
	s.calls++
	s.orgID = orgID
	return s.err
}

func TestOrganizationProvisioningHandler(t *testing.T) {
	t.Run("provisions organization", func(t *testing.T) {
		stub := &provisionerStub{}
		handler := NewOrganizationProvisioningHandler(stub)
		task, err := tasks.NewOrganizationProvisioningTask(11, "req-1")
		if err != nil {
			t.Fatalf("task error = %v", err)
		}
		if err := handler.Handle(context.Background(), task); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if stub.calls != 1 || stub.orgID != 11 {
			t.Errorf("provision calls = %d org = %d", stub.calls, stub.orgID)
		}
	})

	t.Run("rejects invalid payload", func(t *testing.T) {
		handler := NewOrganizationProvisioningHandler(&provisionerStub{})
		if err := handler.Handle(context.Background(), asynqTask(t, tasks.TypeOrganizationProvisioning, "not-an-object")); err == nil {
			t.Fatal("Handle() expected unmarshal error")
		}
	})

	t.Run("propagates provision error", func(t *testing.T) {
		handler := NewOrganizationProvisioningHandler(&provisionerStub{err: errors.New("seed down")})
		task, err := tasks.NewOrganizationProvisioningTask(11, "")
		if err != nil {
			t.Fatalf("task error = %v", err)
		}
		if err := handler.Handle(context.Background(), task); err == nil {
			t.Fatal("Handle() expected provision error")
		}
	})
}
