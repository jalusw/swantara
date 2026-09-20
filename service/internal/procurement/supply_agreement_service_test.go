package procurement

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
)

func testAgreementService() (SupplyAgreementService, *SupplyAgreementDAOMock, *SupplyAgreementLineDAOMock) {
	agreements := &SupplyAgreementDAOMock{}
	lines := &SupplyAgreementLineDAOMock{}
	contactDAO := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 10}}, nil
			},
		},
	}
	seqDAO := sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "AGR/00001"}, nil
		},
	}
	svc := NewSupplyAgreementService(agreements, lines, sequence.NewSequenceService(seqDAO), contactDAO)
	return svc, agreements, lines
}

func TestSupplyAgreementService_Create_Success(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := testAgreementService()
	orgID := uint64(1)
	agreement := &SupplyAgreement{
		OrganizationID: &orgID,
		SupplierID:     10,
		StartDate:      helper.Ptr(time.Now()),
		EndDate:        helper.Ptr(time.Now().AddDate(0, 6, 0)),
	}
	lines := []*SupplyAgreementLine{
		{ItemID: helper.Ptr(uint64(100)), Qty: 10, UnitPrice: 1000},
	}
	created, err := svc.Create(ctx, agreement, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Name == nil {
		t.Error("expected name to be set")
	}
	if created.State != SupplyAgreementStateDraft {
		t.Errorf("expected state %q, got %q", SupplyAgreementStateDraft, created.State)
	}
	if created.ID == 0 {
		t.Error("expected agreement ID to be set")
	}
}

func TestSupplyAgreementService_Create_NoLines(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := testAgreementService()
	orgID := uint64(1)
	agreement := &SupplyAgreement{
		OrganizationID: &orgID,
		SupplierID:     10,
	}
	_, err := svc.Create(ctx, agreement, []*SupplyAgreementLine{})
	if err != ErrAgreementNoLines {
		t.Errorf("expected ErrAgreementNoLines, got %v", err)
	}
}

func TestSupplyAgreementService_Create_VendorNotFound(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := testAgreementService()
	svc.contacts = contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, id uint64) (*contacts.Contact, error) {
				if id == 999 {
					return nil, nil
				}
				return &contacts.Contact{Base: model.Base{ID: id}}, nil
			},
		},
	}
	orgID := uint64(1)
	agreement := &SupplyAgreement{
		OrganizationID: &orgID,
		SupplierID:     999,
	}
	lines := []*SupplyAgreementLine{
		{ItemID: helper.Ptr(uint64(100)), Qty: 10, UnitPrice: 1000},
	}
	_, err := svc.Create(ctx, agreement, lines)
	if err != ErrAgreementVendor {
		t.Errorf("expected ErrAgreementVendor, got %v", err)
	}
}

func TestSupplyAgreementService_Create_LineQtyZero(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := testAgreementService()
	orgID := uint64(1)
	agreement := &SupplyAgreement{
		OrganizationID: &orgID,
		SupplierID:     10,
	}
	lines := []*SupplyAgreementLine{
		{ItemID: helper.Ptr(uint64(100)), Qty: 0, UnitPrice: 1000},
	}
	_, err := svc.Create(ctx, agreement, lines)
	if err != ErrAgreementLineQty {
		t.Errorf("expected ErrAgreementLineQty, got %v", err)
	}
}

func TestSupplyAgreementService_Activate_Success(t *testing.T) {
	ctx := context.Background()
	svc, agreements, _ := testAgreementService()
	agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
		return &SupplyAgreement{
			Base:  model.Base{ID: 1},
			State: SupplyAgreementStateDraft,
		}, nil
	}
	agreements.UpdateFunc = func(_ context.Context, agreement *SupplyAgreement) (*SupplyAgreement, error) {
		return agreement, nil
	}
	result, err := svc.Activate(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.State != SupplyAgreementStateActive {
		t.Errorf("expected state %q, got %q", SupplyAgreementStateActive, result.State)
	}
}

func TestSupplyAgreementService_Activate_InvalidTransition(t *testing.T) {
	ctx := context.Background()
	svc, agreements, _ := testAgreementService()
	agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
		return &SupplyAgreement{
			Base:  model.Base{ID: 1},
			State: SupplyAgreementStateActive,
		}, nil
	}
	_, err := svc.Activate(ctx, 1)
	if err != ErrSupplyAgreementState {
		t.Errorf("expected ErrSupplyAgreementState, got %v", err)
	}
}

func TestSupplyAgreementService_Cancel_Success(t *testing.T) {
	ctx := context.Background()
	svc, agreements, _ := testAgreementService()
	agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
		return &SupplyAgreement{
			Base:  model.Base{ID: 1},
			State: SupplyAgreementStateDraft,
		}, nil
	}
	agreements.UpdateFunc = func(_ context.Context, agreement *SupplyAgreement) (*SupplyAgreement, error) {
		return agreement, nil
	}
	result, err := svc.Cancel(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.State != SupplyAgreementStateCancelled {
		t.Errorf("expected state %q, got %q", SupplyAgreementStateCancelled, result.State)
	}
}

func TestSupplyAgreementService_Consume_Success(t *testing.T) {
	ctx := context.Background()
	svc, agreements, _ := testAgreementService()
	agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
		return &SupplyAgreement{
			Base:           model.Base{ID: 1},
			State:          SupplyAgreementStateActive,
			QtyLimit:       100,
			AmountLimit:    1000000,
			ConsumedQty:    10,
			ConsumedAmount: 100000,
		}, nil
	}
	agreements.UpdateFunc = func(_ context.Context, agreement *SupplyAgreement) (*SupplyAgreement, error) {
		return agreement, nil
	}
	err := svc.Consume(ctx, 1, 5, 50000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSupplyAgreementService_Consume_QtyExceeded(t *testing.T) {
	ctx := context.Background()
	svc, agreements, _ := testAgreementService()
	agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
		return &SupplyAgreement{
			Base:           model.Base{ID: 1},
			State:          SupplyAgreementStateActive,
			QtyLimit:       100,
			ConsumedQty:    95,
			ConsumedAmount: 100000,
		}, nil
	}
	err := svc.Consume(ctx, 1, 10, 0)
	if err != ErrAgreementQtyExceeded {
		t.Errorf("expected ErrAgreementQtyExceeded, got %v", err)
	}
}

func TestSupplyAgreementService_Consume_AmountExceeded(t *testing.T) {
	ctx := context.Background()
	svc, agreements, _ := testAgreementService()
	agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
		return &SupplyAgreement{
			Base:           model.Base{ID: 1},
			State:          SupplyAgreementStateActive,
			AmountLimit:    1000000,
			ConsumedAmount: 950000,
		}, nil
	}
	err := svc.Consume(ctx, 1, 0, 100000)
	if err != ErrAgreementAmountExceeded {
		t.Errorf("expected ErrAgreementAmountExceeded, got %v", err)
	}
}

func TestSupplyAgreementService_Consume_NotActive(t *testing.T) {
	ctx := context.Background()
	svc, agreements, _ := testAgreementService()
	agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
		return &SupplyAgreement{
			Base:  model.Base{ID: 1},
			State: SupplyAgreementStateDraft,
		}, nil
	}
	err := svc.Consume(ctx, 1, 10, 10000)
	if err != ErrAgreementNotActive {
		t.Errorf("expected ErrAgreementNotActive, got %v", err)
	}
}

func TestSupplyAgreementService_ListLines_Success(t *testing.T) {
	ctx := context.Background()
	svc, _, lines := testAgreementService()
	lines.ListByAgreementFunc = func(_ context.Context, _ uint64) ([]*SupplyAgreementLine, error) {
		return []*SupplyAgreementLine{
			{Base: model.Base{ID: 1}, AgreementID: 1, ItemID: helper.Ptr(uint64(100)), Qty: 10, UnitPrice: 1000},
			{Base: model.Base{ID: 2}, AgreementID: 1, ItemID: helper.Ptr(uint64(200)), Qty: 5, UnitPrice: 2000},
		}, nil
	}
	result, err := svc.ListLines(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 lines, got %d", len(result))
	}
}
