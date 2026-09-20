package organization

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
)

func TestOrganizationService_QuickCreate_EnqueuesProvisioning(t *testing.T) {
	ctx := context.Background()

	t.Run("enqueues instead of seeding synchronously", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		seeder := &seederStub{}
		svc := NewOrganizationServiceWithPaymentTerms(orgs, currencies, members, seeder)
		var enqueued *asynq.Task
		svc.SetProvisionEnqueuer(queue.TaskEnqueuerMock{
			EnqueueFunc: func(task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
				enqueued = task
				return nil, nil
			},
		})

		created, err := svc.QuickCreate(ctx, "Toko Maju", "ID", 5)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if enqueued == nil {
			t.Fatal("expected provisioning task to be enqueued")
		}
		if enqueued.Type() != tasks.TypeOrganizationProvisioning {
			t.Errorf("task type = %q, want %q", enqueued.Type(), tasks.TypeOrganizationProvisioning)
		}
		if len(seeder.seeded) != 0 {
			t.Errorf("seeder called synchronously = %v, want no sync seeding", seeder.seeded)
		}
		if created.ID != 11 {
			t.Errorf("org id = %d, want 11", created.ID)
		}
	})

	t.Run("propagates enqueue error", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		svc := NewOrganizationService(orgs, currencies, members)
		enqueueErr := errors.New("redis down")
		svc.SetProvisionEnqueuer(queue.TaskEnqueuerMock{
			EnqueueFunc: func(_ *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
				return nil, enqueueErr
			},
		})
		_, err := svc.QuickCreate(ctx, "Toko Maju", "ID", 5)
		helper.AssertError(t, err, true, enqueueErr)
	})
}

func TestOrganizationService_Provision(t *testing.T) {
	ctx := context.Background()

	t.Run("seeds payment terms and modules", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		seeder := &seederStub{}
		svc := NewOrganizationServiceWithPaymentTerms(orgs, currencies, members, seeder)
		svc.SetModuleStore(moduleStoreWith())

		if err := svc.Provision(ctx, 11); helper.AssertError(t, err, false, nil) {
			return
		}
		if len(seeder.seeded) != 1 || seeder.seeded[0] != 11 {
			t.Errorf("seeded = %v, want [11]", seeder.seeded)
		}
	})

	t.Run("propagates seeder error", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		seedErr := errors.New("seed down")
		svc := NewOrganizationServiceWithPaymentTerms(orgs, currencies, members, &seederStub{err: seedErr})
		helper.AssertError(t, svc.Provision(ctx, 11), true, seedErr)
	})
}
