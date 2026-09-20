package accounting

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestReminderService_Generate_SendsLevelForOverdueInvoice(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := now.AddDate(0, 0, -40)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{
				{Base: model.Base{ID: 1}, DaysOverdue: 30},
				{Base: model.Base{ID: 2}, DaysOverdue: 60},
			}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10)), ContactID: 5, DueDate: &dueDate}}, nil
		},
	}
	var created []*ReminderAction
	actions := ReminderActionDAOMock{
		CRUDMock: dao.CRUDMock[ReminderAction]{
			CreateFunc: func(_ context.Context, action *ReminderAction) (*ReminderAction, error) {
				action.ID = 90
				created = append(created, action)
				return action, nil
			},
		},
	}
	svc := NewReminderService(invoices, levels, actions, TransactionerMock{})
	svc.now = func() time.Time { return now }

	result, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].InvoiceID != 11 || result[0].LevelID != 1 {
		t.Errorf("result = %+v, want invoice 11 at level 1", result)
	}
	if len(created) != 1 || created[0].ContactID != 5 {
		t.Errorf("created = %+v, want single action for contact 5", created)
	}
}

func TestReminderService_Generate_SkipsAlreadySent(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := now.AddDate(0, 0, -40)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10)), ContactID: 5, DueDate: &dueDate}}, nil
		},
	}
	actions := ReminderActionDAOMock{
		FindByInvoiceLevelFunc: func(_ context.Context, _, _ uint64) (*ReminderAction, error) {
			return &ReminderAction{Base: model.Base{ID: 90}}, nil
		},
	}
	svc := NewReminderService(invoices, levels, actions, TransactionerMock{})
	svc.now = func() time.Time { return now }

	result, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("result = %d, want 0", len(result))
	}
}

func TestReminderService_Generate_RejectsWithoutLevels(t *testing.T) {
	ctx := context.Background()
	svc := NewReminderService(InvoiceDAOMock{}, ReminderLevelDAOMock{}, ReminderActionDAOMock{}, TransactionerMock{})

	_, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if helper.AssertError(t, err, true, ErrReminderLevelNotFound) {
		return
	}
}

func TestReminderService_Generate_IgnoresInvoiceNotOverdue(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := now.AddDate(0, 0, 10)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10)), DueDate: &dueDate}}, nil
		},
	}
	svc := NewReminderService(invoices, levels, ReminderActionDAOMock{}, TransactionerMock{})
	svc.now = func() time.Time { return now }

	result, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("result = %d, want 0", len(result))
	}
}

func TestReminderLevelDAO_ListSorted_ReturnsLevelsBySequence(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reminder_levels" ORDER BY sequence ASC`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "days_overdue"}).AddRow(1, 30).AddRow(2, 60))

	levels := NewReminderLevelDAO(db)

	items, err := levels.ListSorted(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].DaysOverdue != 30 || items[1].DaysOverdue != 60 {
		t.Errorf("items = %+v, want 30/60", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReminderLevelDAO_ListSorted_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reminder_levels" ORDER BY sequence ASC`)).
		WillReturnError(errors.New("db down"))

	levels := NewReminderLevelDAO(db)

	_, err := levels.ListSorted(ctx)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReminderService_Generate_PropagatesLevelListError(t *testing.T) {
	ctx := context.Background()
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewReminderService(InvoiceDAOMock{}, levels, ReminderActionDAOMock{}, TransactionerMock{})

	_, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestReminderService_Generate_PropagatesOverdueListError(t *testing.T) {
	ctx := context.Background()
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewReminderService(invoices, levels, ReminderActionDAOMock{}, TransactionerMock{})

	_, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestReminderService_Generate_PropagatesFindByInvoiceLevelError(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := now.AddDate(0, 0, -40)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10)), DueDate: &dueDate}}, nil
		},
	}
	actions := ReminderActionDAOMock{
		FindByInvoiceLevelFunc: func(_ context.Context, _, _ uint64) (*ReminderAction, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewReminderService(invoices, levels, actions, TransactionerMock{})
	svc.now = func() time.Time { return now }

	_, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestReminderService_Generate_PropagatesCreateError(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := now.AddDate(0, 0, -40)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10)), DueDate: &dueDate}}, nil
		},
	}
	actions := ReminderActionDAOMock{
		CRUDMock: dao.CRUDMock[ReminderAction]{
			CreateFunc: func(_ context.Context, _ *ReminderAction) (*ReminderAction, error) {
				return nil, errors.New("insert failed")
			},
		},
	}
	svc := NewReminderService(invoices, levels, actions, TransactionerMock{})
	svc.now = func() time.Time { return now }

	_, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestReminderService_Generate_UsesAsOfFromRequest(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := asOf.AddDate(0, 0, -10)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 5}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, got *time.Time) ([]*Invoice, error) {
			if got == nil || !got.Equal(asOf) {
				t.Errorf("asOf = %v, want %v", got, asOf)
			}
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10)), DueDate: &dueDate}}, nil
		},
	}
	var created *ReminderAction
	actions := ReminderActionDAOMock{
		CRUDMock: dao.CRUDMock[ReminderAction]{
			CreateFunc: func(_ context.Context, action *ReminderAction) (*ReminderAction, error) {
				action.ID = 90
				created = action
				return action, nil
			},
		},
	}
	svc := NewReminderService(invoices, levels, actions, TransactionerMock{})
	svc.now = func() time.Time { return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) }

	result, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10, AsOf: &asOf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("result = %d, want 1", len(result))
	}
	if created.SentAt != nil {
		t.Errorf("sent at = %v, want nil pending worker send", created.SentAt)
	}
}

func TestReminderService_Generate_EnqueuesSendTask(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := now.AddDate(0, 0, -40)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10)), ContactID: 5, DueDate: &dueDate}}, nil
		},
	}
	actions := ReminderActionDAOMock{
		CRUDMock: dao.CRUDMock[ReminderAction]{
			CreateFunc: func(_ context.Context, action *ReminderAction) (*ReminderAction, error) {
				action.ID = 90
				return action, nil
			},
		},
	}
	var enqueued int
	svc := NewReminderService(invoices, levels, actions, TransactionerMock{}).WithSender(queue.TaskEnqueuerMock{
		EnqueueFunc: func(_ *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
			enqueued++
			return &asynq.TaskInfo{}, nil
		},
	})
	svc.now = func() time.Time { return now }

	result, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("result = %d, want 1", len(result))
	}
	if result[0].SentAt != nil {
		t.Errorf("sent at = %v, want nil pending worker send", result[0].SentAt)
	}
	if enqueued != 1 {
		t.Errorf("enqueued = %d, want 1 send task", enqueued)
	}
}

func TestReminderService_Generate_EnqueueFailureRollsBack(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := now.AddDate(0, 0, -40)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10)), ContactID: 5, DueDate: &dueDate}}, nil
		},
	}
	var deleted []uint64
	actions := ReminderActionDAOMock{
		CRUDMock: dao.CRUDMock[ReminderAction]{
			CreateFunc: func(_ context.Context, action *ReminderAction) (*ReminderAction, error) {
				action.ID = 90
				return action, nil
			},
			DeleteFunc: func(_ context.Context, id uint64) error {
				deleted = append(deleted, id)
				return nil
			},
		},
	}
	svc := NewReminderService(invoices, levels, actions, TransactionerMock{}).WithSender(queue.TaskEnqueuerMock{
		EnqueueFunc: func(_ *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
			return nil, errors.New("queue down")
		},
	})
	svc.now = func() time.Time { return now }

	if _, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10}); err == nil {
		t.Fatalf("expected enqueue error, got nil")
	}
	if len(deleted) != 1 || deleted[0] != 90 {
		t.Errorf("deleted = %v, want [90] so retry can resend", deleted)
	}
}

func TestReminderService_Generate_IgnoresOtherOrganization(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := now.AddDate(0, 0, -40)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(99)), DueDate: &dueDate}}, nil
		},
	}
	svc := NewReminderService(invoices, levels, ReminderActionDAOMock{}, TransactionerMock{})
	svc.now = func() time.Time { return now }

	result, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("result = %d, want 0", len(result))
	}
}

func TestReminderService_Generate_IgnoresNoDueDate(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10))}}, nil
		},
	}
	svc := NewReminderService(invoices, levels, ReminderActionDAOMock{}, TransactionerMock{})
	svc.now = func() time.Time { return now }

	result, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("result = %d, want 0", len(result))
	}
}

func TestReminderService_Generate_IgnoresBelowLowestLevel(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	dueDate := now.AddDate(0, 0, -10)
	levels := ReminderLevelDAOMock{
		ListSortedFunc: func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 30}}, nil
		},
	}
	invoices := InvoiceDAOMock{
		ListOverdueFunc: func(_ context.Context, _ *time.Time) ([]*Invoice, error) {
			return []*Invoice{{Base: model.Base{ID: 11}, OrganizationID: helper.Ptr(uint64(10)), DueDate: &dueDate}}, nil
		},
	}
	svc := NewReminderService(invoices, levels, ReminderActionDAOMock{}, TransactionerMock{})
	svc.now = func() time.Time { return now }

	result, err := svc.Generate(ctx, GenerateReminderRequest{OrganizationID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("result = %d, want 0", len(result))
	}
}
