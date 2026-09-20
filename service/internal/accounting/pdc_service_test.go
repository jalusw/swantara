package accounting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func pdcTestService(instruments PdcInstrumentDAO, payments PaymentCreator) PdcService {
	return NewPdcService(instruments, payments, TransactionerMock{})
}

func heldPdc(id uint64) *PdcInstrument {
	due := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	return &PdcInstrument{
		Base:           model.Base{ID: id},
		OrganizationID: helper.Ptr(uint64(10)),
		ContactID:      5,
		Direction:      PaymentTypeInbound,
		Number:         helper.Ptr("CHQ-001"),
		Amount:         amount.FromFloat64(500),
		DueDate:        &due,
		State:          PdcStateHeld,
		InvoiceID:      helper.Ptr(uint64(11)),
		JournalID:      helper.Ptr(uint64(20)),
	}
}

func TestPdcService_Register(t *testing.T) {
	var created *PdcInstrument
	instruments := PdcInstrumentDAOMock{
		CRUDMock: dao.CRUDMock[PdcInstrument]{
			CreateFunc: func(_ context.Context, instrument *PdcInstrument) (*PdcInstrument, error) {
				instrument.ID = 1
				created = instrument
				return instrument, nil
			},
		},
	}
	svc := pdcTestService(instruments, PaymentCreatorMock{})
	due := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	instrument, err := svc.Register(context.Background(), RegisterPdcRequest{
		OrganizationID: 10,
		ContactID:      5,
		Direction:      PaymentTypeInbound,
		Number:         "CHQ-001",
		Amount:         500,
		DueDate:        due,
		JournalID:      20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.State != PdcStateHeld {
		t.Errorf("state = %s, want held", created.State)
	}
	if instrument.Amount.Float64() != 500 {
		t.Errorf("amount = %v, want 500", instrument.Amount)
	}
}

func TestPdcService_Transitions(t *testing.T) {
	tests := []struct {
		name    string
		state   string
		action  string
		want    string
		wantErr error
	}{
		{name: "deposit held", state: PdcStateHeld, action: "deposit", want: PdcStateDeposited},
		{name: "cancel held", state: PdcStateHeld, action: "cancel", want: PdcStateCancelled},
		{name: "bounce deposited", state: PdcStateDeposited, action: "bounce", want: PdcStateBounced},
		{name: "deposit deposited fails", state: PdcStateDeposited, action: "deposit", wantErr: ErrPdcInvalidState},
		{name: "cancel deposited fails", state: PdcStateDeposited, action: "cancel", wantErr: ErrPdcInvalidState},
		{name: "bounce held fails", state: PdcStateHeld, action: "bounce", wantErr: ErrPdcInvalidState},
		{name: "transition cleared fails", state: PdcStateCleared, action: "bounce", wantErr: ErrPdcInvalidState},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored := heldPdc(7)
			stored.State = tt.state
			instruments := PdcInstrumentDAOMock{
				CRUDMock: dao.CRUDMock[PdcInstrument]{
					FindFunc: func(_ context.Context, _ uint64) (*PdcInstrument, error) {
						return stored, nil
					},
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, instrument *PdcInstrument) (*PdcInstrument, error) {
					return instrument, nil
				},
			}
			svc := pdcTestService(instruments, PaymentCreatorMock{})
			var result *PdcInstrument
			var err error
			switch tt.action {
			case "deposit":
				result, err = svc.Deposit(context.Background(), 7, 10)
			case "cancel":
				result, err = svc.Cancel(context.Background(), 7, 10)
			case "bounce":
				result, err = svc.Bounce(context.Background(), 7, 10)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.State != tt.want {
				t.Errorf("state = %s, want %s", result.State, tt.want)
			}
		})
	}
}

func TestPdcService_Clear(t *testing.T) {
	stored := heldPdc(7)
	stored.State = PdcStateDeposited
	instruments := PdcInstrumentDAOMock{
		CRUDMock: dao.CRUDMock[PdcInstrument]{
			FindFunc: func(_ context.Context, _ uint64) (*PdcInstrument, error) {
				return stored, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, instrument *PdcInstrument) (*PdcInstrument, error) {
			return instrument, nil
		},
	}
	var paid *CreatePaymentRequest
	payments := PaymentCreatorMock{
		CreateFunc: func(_ context.Context, request CreatePaymentRequest) (*Payment, error) {
			paid = &request
			return &Payment{Base: model.Base{ID: 44}}, nil
		},
	}
	svc := pdcTestService(instruments, payments)
	cleared, err := svc.Clear(context.Background(), 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleared.State != PdcStateCleared {
		t.Errorf("state = %s, want cleared", cleared.State)
	}
	if cleared.PaymentID == nil || *cleared.PaymentID != 44 {
		t.Errorf("payment = %v, want 44", cleared.PaymentID)
	}
	if paid == nil || paid.Amount != 500 || len(paid.InvoiceIDs) != 1 || paid.InvoiceIDs[0] != 11 {
		t.Errorf("payment request = %+v, want 500 allocated to invoice 11", paid)
	}
}

func TestPdcService_ClearFromHeldFails(t *testing.T) {
	instruments := PdcInstrumentDAOMock{
		CRUDMock: dao.CRUDMock[PdcInstrument]{
			FindFunc: func(_ context.Context, _ uint64) (*PdcInstrument, error) {
				return heldPdc(7), nil
			},
		},
	}
	svc := pdcTestService(instruments, PaymentCreatorMock{})
	if _, err := svc.Clear(context.Background(), 7, 10); !errors.Is(err, ErrPdcInvalidState) {
		t.Fatalf("error = %v, want %v", err, ErrPdcInvalidState)
	}
}
