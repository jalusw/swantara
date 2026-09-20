package procurement

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/kernel/state"
)

type SupplyAgreementService struct {
	agreements SupplyAgreementDAO
	lines      SupplyAgreementLineDAO
	sequences  sequence.Service
	contacts   contacts.ContactDAO
	machine    state.Machine
}

func NewSupplyAgreementService(
	agreements SupplyAgreementDAO,
	lines SupplyAgreementLineDAO,
	sequences sequence.Service,
	contacts contacts.ContactDAO,
) SupplyAgreementService {
	return SupplyAgreementService{
		agreements: agreements,
		lines:      lines,
		sequences:  sequences,
		contacts:   contacts,
		machine: state.NewMachine(
			state.Transition{From: model.Status(SupplyAgreementStateDraft), To: model.Status(SupplyAgreementStateActive)},
			state.Transition{From: model.Status(SupplyAgreementStateDraft), To: model.Status(SupplyAgreementStateCancelled)},
			state.Transition{From: model.Status(SupplyAgreementStateActive), To: model.Status(SupplyAgreementStateExpired)},
			state.Transition{From: model.Status(SupplyAgreementStateActive), To: model.Status(SupplyAgreementStateCancelled)},
		),
	}
}

func (s SupplyAgreementService) Find(ctx context.Context, id uint64) (*SupplyAgreement, error) {
	agreement, err := s.agreements.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if agreement == nil {
		return nil, ErrAgreementNotFound
	}
	return agreement, nil
}

func (s SupplyAgreementService) List(ctx context.Context, q *query.Query) (*query.Page[SupplyAgreement], error) {
	return s.agreements.List(ctx, q)
}

func (s SupplyAgreementService) Create(ctx context.Context, agreement *SupplyAgreement, lines []*SupplyAgreementLine) (*SupplyAgreement, error) {
	if agreement.OrganizationID == nil {
		return nil, ErrAgreementNotFound
	}
	if len(lines) == 0 {
		return nil, ErrAgreementNoLines
	}
	supplier, err := s.contacts.Find(ctx, agreement.SupplierID)
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, ErrAgreementVendor
	}
	for _, line := range lines {
		if !amount.FromFloat64(line.Qty).GreaterThan(amount.Zero()) {
			return nil, ErrAgreementLineQty
		}
	}

	name, err := s.sequences.Next(ctx, *agreement.OrganizationID, SequenceSupplyAgreementCode)
	if err != nil {
		return nil, err
	}
	agreement.Name = helper.Ptr(name)
	agreement.State = SupplyAgreementStateDraft
	if agreement.CurrencyCode == nil {
		agreement.CurrencyCode = helper.Ptr("IDR")
	}
	return s.agreements.CreateWithLines(ctx, agreement, lines)
}

func (s SupplyAgreementService) Activate(ctx context.Context, agreementID uint64) (*SupplyAgreement, error) {
	agreement, err := s.findAgreement(ctx, agreementID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(agreement.State), model.Status(SupplyAgreementStateActive)); err != nil {
		return nil, ErrSupplyAgreementState
	}
	agreement.State = SupplyAgreementStateActive
	return s.agreements.Update(ctx, agreement)
}

func (s SupplyAgreementService) Cancel(ctx context.Context, agreementID uint64) (*SupplyAgreement, error) {
	agreement, err := s.findAgreement(ctx, agreementID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(agreement.State), model.Status(SupplyAgreementStateCancelled)); err != nil {
		return nil, ErrSupplyAgreementState
	}
	agreement.State = SupplyAgreementStateCancelled
	return s.agreements.Update(ctx, agreement)
}

func (s SupplyAgreementService) ListLines(ctx context.Context, agreementID uint64) ([]*SupplyAgreementLine, error) {
	return s.lines.ListByAgreement(ctx, agreementID)
}

func (s SupplyAgreementService) Consume(ctx context.Context, agreementID uint64, qtyDelta, amountDelta float64) error {
	agreement, err := s.findAgreement(ctx, agreementID)
	if err != nil {
		return err
	}
	if agreement.State != SupplyAgreementStateActive {
		return ErrAgreementNotActive
	}
	if agreement.QtyLimit > 0 {
		newQty := amount.FromFloat64(agreement.ConsumedQty).Add(amount.FromFloat64(qtyDelta))
		if newQty.GreaterThan(amount.FromFloat64(agreement.QtyLimit)) {
			return ErrAgreementQtyExceeded
		}
	}
	if agreement.AmountLimit > 0 {
		newAmount := amount.FromFloat64(agreement.ConsumedAmount).Add(amount.FromFloat64(amountDelta))
		if newAmount.GreaterThan(amount.FromFloat64(agreement.AmountLimit)) {
			return ErrAgreementAmountExceeded
		}
	}
	agreement.ConsumedQty = amount.FromFloat64(agreement.ConsumedQty).Add(amount.FromFloat64(qtyDelta)).Float64()
	agreement.ConsumedAmount = amount.FromFloat64(agreement.ConsumedAmount).Add(amount.FromFloat64(amountDelta)).Float64()
	_, err = s.agreements.Update(ctx, agreement)
	return err
}

func (s SupplyAgreementService) findAgreement(ctx context.Context, agreementID uint64) (*SupplyAgreement, error) {
	agreement, err := s.agreements.Find(ctx, agreementID)
	if err != nil {
		return nil, err
	}
	if agreement == nil {
		return nil, ErrAgreementNotFound
	}
	return agreement, nil
}
