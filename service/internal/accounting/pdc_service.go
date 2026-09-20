package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type PaymentCreator interface {
	Create(ctx context.Context, request CreatePaymentRequest) (*Payment, error)
	CreateOutbound(ctx context.Context, request CreatePaymentRequest) (*Payment, error)
}

type RegisterPdcRequest struct {
	OrganizationID uint64
	ContactID      uint64
	Direction      string
	Number         string
	BankName       string
	Amount         float64
	CurrencyCode   *string
	DueDate        time.Time
	InvoiceID      *uint64
	JournalID      uint64
}

type PdcService struct {
	instruments PdcInstrumentDAO
	payments    PaymentCreator
	tx          db.Transactioner
	now         func() time.Time
}

func NewPdcService(instruments PdcInstrumentDAO, payments PaymentCreator, tx db.Transactioner) PdcService {
	return PdcService{instruments: instruments, payments: payments, tx: tx, now: time.Now}
}

func (s PdcService) Find(ctx context.Context, id uint64) (*PdcInstrument, error) {
	return s.instruments.Find(ctx, id)
}

func (s PdcService) List(ctx context.Context, q *query.Query) (*query.Page[PdcInstrument], error) {
	return s.instruments.List(ctx, q)
}

func (s PdcService) ListDue(ctx context.Context, asOf *time.Time) ([]*PdcInstrument, error) {
	return s.instruments.ListDue(ctx, asOf)
}

func (s PdcService) Register(ctx context.Context, request RegisterPdcRequest) (*PdcInstrument, error) {
	if request.Direction != PaymentTypeInbound && request.Direction != PaymentTypeOutbound {
		return nil, ErrPdcInvalidState
	}
	settled := amount.FromFloat64(request.Amount).Round(4)
	if !settled.IsPositive() {
		return nil, ErrPaymentAmount
	}
	if request.DueDate.IsZero() {
		return nil, ErrPdcInvalidState
	}
	return s.instruments.Create(ctx, &PdcInstrument{
		OrganizationID: helper.Ptr(request.OrganizationID),
		ContactID:      request.ContactID,
		Direction:      request.Direction,
		Number:         helper.Ptr(request.Number),
		BankName:       helper.Ptr(request.BankName),
		Amount:         settled,
		CurrencyCode:   request.CurrencyCode,
		DueDate:        helper.Ptr(request.DueDate),
		State:          PdcStateHeld,
		InvoiceID:      request.InvoiceID,
		JournalID:      helper.Ptr(request.JournalID),
	})
}

func (s PdcService) Deposit(ctx context.Context, id, organizationID uint64) (*PdcInstrument, error) {
	return s.transition(ctx, id, organizationID, PdcStateHeld, PdcStateDeposited, nil)
}

func (s PdcService) Cancel(ctx context.Context, id, organizationID uint64) (*PdcInstrument, error) {
	return s.transition(ctx, id, organizationID, PdcStateHeld, PdcStateCancelled, nil)
}

func (s PdcService) Bounce(ctx context.Context, id, organizationID uint64) (*PdcInstrument, error) {
	return s.transition(ctx, id, organizationID, PdcStateDeposited, PdcStateBounced, nil)
}

func (s PdcService) Clear(ctx context.Context, id, organizationID uint64) (*PdcInstrument, error) {
	instrument, err := s.load(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}
	if instrument.State != PdcStateDeposited {
		return nil, ErrPdcInvalidState
	}
	if instrument.JournalID == nil {
		return nil, ErrNoBankAccount
	}
	if s.payments == nil {
		return nil, ErrPdcInvalidState
	}
	paymentRequest := CreatePaymentRequest{
		OrganizationID: organizationID,
		ContactID:      instrument.ContactID,
		JournalID:      *instrument.JournalID,
		Amount:         instrument.Amount.Float64(),
		CurrencyCode:   instrument.CurrencyCode,
		Date:           s.now().UTC(),
		Reference:      helper.Deref(instrument.Number, ""),
	}
	if instrument.InvoiceID != nil {
		paymentRequest.InvoiceIDs = []uint64{*instrument.InvoiceID}
	}
	var cleared *PdcInstrument
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		var payment *Payment
		var err error
		if instrument.Direction == PaymentTypeOutbound {
			payment, err = s.payments.CreateOutbound(ctx, paymentRequest)
		} else {
			payment, err = s.payments.Create(ctx, paymentRequest)
		}
		if err != nil {
			return err
		}
		instrument.PaymentID = helper.Ptr(payment.ID)
		return s.applyTransition(ctx, tx, instrument, PdcStateCleared, &cleared)
	})
	if err != nil {
		return nil, err
	}
	return cleared, nil
}

func (s PdcService) transition(ctx context.Context, id, organizationID uint64, from, to string, out **PdcInstrument) (*PdcInstrument, error) {
	instrument, err := s.load(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}
	if instrument.State != from {
		return nil, ErrPdcInvalidState
	}
	var updated *PdcInstrument
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		return s.applyTransition(ctx, tx, instrument, to, &updated)
	})
	if err != nil {
		return nil, err
	}
	if out != nil {
		*out = updated
	}
	return updated, nil
}

func (s PdcService) applyTransition(ctx context.Context, tx *gorm.DB, instrument *PdcInstrument, to string, out **PdcInstrument) error {
	instrument.State = to
	updated, err := s.instruments.UpdateTx(ctx, tx, instrument)
	if err != nil {
		return err
	}
	if out != nil {
		*out = updated
	}
	return nil
}

func (s PdcService) load(ctx context.Context, id, organizationID uint64) (*PdcInstrument, error) {
	instrument, err := s.instruments.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if instrument == nil || instrument.OrganizationID == nil || *instrument.OrganizationID != organizationID {
		return nil, ErrPdcNotFound
	}
	return instrument, nil
}
