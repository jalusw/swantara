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

type PurchaseRequestService struct {
	requisitions PurchaseRequestDAO
	lines        PurchaseRequestLineDAO
	sequences    sequence.Service
	contacts     contacts.ContactDAO
	machine      state.Machine
}

func NewPurchaseRequestService(
	requisitions PurchaseRequestDAO,
	lines PurchaseRequestLineDAO,
	sequences sequence.Service,
	contacts contacts.ContactDAO,
) PurchaseRequestService {
	return PurchaseRequestService{
		requisitions: requisitions,
		lines:        lines,
		sequences:    sequences,
		contacts:     contacts,
		machine: state.NewMachine(
			state.Transition{From: model.Status(RequestStateDraft), To: model.Status(RequestStateConfirmed)},
			state.Transition{From: model.Status(RequestStateDraft), To: model.Status(RequestStateCancelled)},
			state.Transition{From: model.Status(RequestStateConfirmed), To: model.Status(RequestStateApproved)},
			state.Transition{From: model.Status(RequestStateConfirmed), To: model.Status(RequestStateCancelled)},
			state.Transition{From: model.Status(RequestStateApproved), To: model.Status(RequestStateDone)},
			state.Transition{From: model.Status(RequestStateApproved), To: model.Status(RequestStateCancelled)},
		),
	}
}

func (s PurchaseRequestService) List(ctx context.Context, q *query.Query) (*query.Page[PurchaseRequest], error) {
	return s.requisitions.List(ctx, q)
}

func (s PurchaseRequestService) Find(ctx context.Context, id uint64) (*PurchaseRequest, error) {
	return s.requisitions.Find(ctx, id)
}

func (s PurchaseRequestService) Create(ctx context.Context, requisition *PurchaseRequest, lines []*PurchaseRequestLine) (*PurchaseRequest, error) {
	if requisition.OrganizationID == nil {
		return nil, ErrRequisitionNotFound
	}
	if len(lines) == 0 {
		return nil, ErrRequisitionNoLines
	}
	requester, err := s.contacts.Find(ctx, requisition.RequesterID)
	if err != nil {
		return nil, err
	}
	if requester == nil {
		return nil, ErrRequisitionRequester
	}
	for _, line := range lines {
		if !amount.FromFloat64(line.Qty).GreaterThan(amount.Zero()) {
			return nil, ErrRequisitionLineQty
		}
	}

	name, err := s.sequences.Next(ctx, *requisition.OrganizationID, SequencePurchaseRequestCode)
	if err != nil {
		return nil, err
	}
	requisition.Name = helper.Ptr(name)
	requisition.State = RequestStateDraft
	return s.requisitions.CreateWithLines(ctx, requisition, lines)
}

func (s PurchaseRequestService) Confirm(ctx context.Context, requestID uint64) (*PurchaseRequest, error) {
	requisition, err := s.findRequisition(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(requisition.State), model.Status(RequestStateConfirmed)); err != nil {
		return nil, ErrRequestState
	}
	requisition.State = RequestStateConfirmed
	return s.requisitions.Update(ctx, requisition)
}

func (s PurchaseRequestService) Approve(ctx context.Context, requestID uint64) (*PurchaseRequest, error) {
	requisition, err := s.findRequisition(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(requisition.State), model.Status(RequestStateApproved)); err != nil {
		return nil, ErrRequestState
	}
	requisition.State = RequestStateApproved
	return s.requisitions.Update(ctx, requisition)
}

func (s PurchaseRequestService) Cancel(ctx context.Context, requestID uint64) (*PurchaseRequest, error) {
	requisition, err := s.findRequisition(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(requisition.State), model.Status(RequestStateCancelled)); err != nil {
		return nil, ErrRequestState
	}
	requisition.State = RequestStateCancelled
	return s.requisitions.Update(ctx, requisition)
}

func (s PurchaseRequestService) ListLines(ctx context.Context, requestID uint64) ([]*PurchaseRequestLine, error) {
	return s.lines.ListByRequest(ctx, requestID)
}

func (s PurchaseRequestService) findRequisition(ctx context.Context, requestID uint64) (*PurchaseRequest, error) {
	requisition, err := s.requisitions.Find(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if requisition == nil {
		return nil, ErrRequisitionNotFound
	}
	return requisition, nil
}
