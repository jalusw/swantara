package expense

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestExpenseConfigSource_JournalID_ReadsGlobalConfig(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configExpenseJournalID, Value: json.RawMessage("5")},
			}}, nil
		},
	}
	source := NewExpenseConfigSource(configs)

	journal, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journal != 5 {
		t.Errorf("journal = %d, want 5", journal)
	}
}

func TestExpenseConfigSource_JournalID_ReadsOrganizationScopedConfig(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configExpenseJournalID, OrganizationID: &org, Value: json.RawMessage("5")},
			}}, nil
		},
	}
	source := NewExpenseConfigSource(configs)

	journal, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journal != 5 {
		t.Errorf("journal = %d, want 5", journal)
	}
}

func TestExpenseConfigSource_JournalID_IgnoresOtherOrganization(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configExpenseJournalID, OrganizationID: &org, Value: json.RawMessage("5")},
			}}, nil
		},
	}
	source := NewExpenseConfigSource(configs)

	journal, err := source.JournalID(ctx, 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journal != 0 {
		t.Errorf("journal = %d, want 0", journal)
	}
}

func TestExpenseConfigSource_JournalID_EmptyValueReturnsZero(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configExpenseJournalID, Value: json.RawMessage{}},
			}}, nil
		},
	}
	source := NewExpenseConfigSource(configs)

	journal, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journal != 0 {
		t.Errorf("journal = %d, want 0", journal)
	}
}

func TestExpenseConfigSource_JournalID_PropagatesError(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("db down")
		},
	}
	source := NewExpenseConfigSource(configs)

	if helper.AssertError(t, mustJournalID(ctx, source), true, nil) {
		return
	}
}

func TestExpenseConfigSource_EmployeePayableAccountID_ReadsConfig(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configExpenseEmployeePayable, Value: json.RawMessage("300")},
			}}, nil
		},
	}
	source := NewExpenseConfigSource(configs)

	account, err := source.EmployeePayableAccountID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account != 300 {
		t.Errorf("account = %d, want 300", account)
	}
}

func TestExpenseConfigSource_EmployeePayableAccountID_PropagatesError(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("db down")
		},
	}
	source := NewExpenseConfigSource(configs)

	_, err := source.EmployeePayableAccountID(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestExpenseConfigSource_CardClearingAccountID_ReadsConfig(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configExpenseCardClearing, Value: json.RawMessage("301")},
			}}, nil
		},
	}
	source := NewExpenseConfigSource(configs)

	account, err := source.CardClearingAccountID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account != 301 {
		t.Errorf("account = %d, want 301", account)
	}
}

func TestExpenseConfigSource_CardClearingAccountID_PropagatesError(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("db down")
		},
	}
	source := NewExpenseConfigSource(configs)

	_, err := source.CardClearingAccountID(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestExpenseConfigSource_ReimbursementBankAccountID_ReadsConfig(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configExpenseReimbursementBank, Value: json.RawMessage("400")},
			}}, nil
		},
	}
	source := NewExpenseConfigSource(configs)

	account, err := source.ReimbursementBankAccountID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account != 400 {
		t.Errorf("account = %d, want 400", account)
	}
}

func TestExpenseConfigSource_ReimbursementBankAccountID_PropagatesError(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.SystemConfig]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("db down")
		},
	}
	source := NewExpenseConfigSource(configs)

	_, err := source.ReimbursementBankAccountID(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func mustJournalID(ctx context.Context, source ExpenseConfigSource) (err error) {
	_, err = source.JournalID(ctx, 10)
	return err
}
