package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
)

func testRequisitionService(requisitions PurchaseRequestDAOMock, lines PurchaseRequestLineDAOMock, contacts contacts.ContactDAOMock) PurchaseRequestService {
	return NewPurchaseRequestService(
		requisitions,
		lines,
		sequence.NewSequenceService(sequence.DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
				return &sequence.Reservation{Value: 1, Number: "REQ/00001"}, nil
			},
		}),
		contacts,
	)
}

func TestPurchaseRequestService_Create_SetsDraftStateAndName(t *testing.T) {
	ctx := context.Background()
	var saved *PurchaseRequest
	requisitions := PurchaseRequestDAOMock{
		CreateWithLinesFunc: func(_ context.Context, requisition *PurchaseRequest, _ []*PurchaseRequestLine) (*PurchaseRequest, error) {
			saved = requisition
			return requisition, nil
		},
	}
	contactDAO := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 5}}, nil
			},
		},
	}
	svc := testRequisitionService(requisitions, PurchaseRequestLineDAOMock{}, contactDAO)

	created, err := svc.Create(ctx, &PurchaseRequest{
		OrganizationID: helper.Ptr(uint64(1)),
		RequesterID:    5,
	}, []*PurchaseRequestLine{{Qty: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.State != RequestStateDraft {
		t.Errorf("state = %s, want draft", created.State)
	}
	if saved == nil || saved.Name == nil || *saved.Name != "REQ/00001" {
		t.Errorf("name = %v, want REQ/00001", saved.Name)
	}
}

func TestPurchaseRequestService_Create_RejectsWithoutLines(t *testing.T) {
	svc := testRequisitionService(PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})

	_, err := svc.Create(context.Background(), &PurchaseRequest{
		OrganizationID: helper.Ptr(uint64(1)),
		RequesterID:    5,
	}, nil)
	if helper.AssertError(t, err, true, ErrRequisitionNoLines) {
		return
	}
}

func TestPurchaseRequestService_Create_RejectsMissingRequester(t *testing.T) {
	contactDAO := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return nil, nil
			},
		},
	}
	svc := testRequisitionService(PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, contactDAO)

	_, err := svc.Create(context.Background(), &PurchaseRequest{
		OrganizationID: helper.Ptr(uint64(1)),
		RequesterID:    5,
	}, []*PurchaseRequestLine{{Qty: 1}})
	if helper.AssertError(t, err, true, ErrRequisitionRequester) {
		return
	}
}

func TestPurchaseRequestService_Create_RejectsZeroQty(t *testing.T) {
	contactDAO := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 5}}, nil
			},
		},
	}
	svc := testRequisitionService(PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, contactDAO)

	_, err := svc.Create(context.Background(), &PurchaseRequest{
		OrganizationID: helper.Ptr(uint64(1)),
		RequesterID:    5,
	}, []*PurchaseRequestLine{{Qty: 0}})
	if helper.AssertError(t, err, true, ErrRequisitionLineQty) {
		return
	}
}

func TestPurchaseRequestService_Confirm_TransitionsDraftToConfirmed(t *testing.T) {
	ctx := context.Background()
	requisition := &PurchaseRequest{Base: model.Base{ID: 1}, State: RequestStateDraft}
	var updated *PurchaseRequest
	requisitions := PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) { return requisition, nil },
			UpdateFunc: func(_ context.Context, req *PurchaseRequest) (*PurchaseRequest, error) {
				updated = req
				return req, nil
			},
		},
	}
	svc := testRequisitionService(requisitions, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})

	confirmed, err := svc.Confirm(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if confirmed.State != RequestStateConfirmed || updated.State != RequestStateConfirmed {
		t.Errorf("state = %s, want confirmed", confirmed.State)
	}
}

func TestPurchaseRequestService_Confirm_RejectsWhenApproved(t *testing.T) {
	ctx := context.Background()
	requisition := &PurchaseRequest{Base: model.Base{ID: 1}, State: RequestStateApproved}
	requisitions := PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) { return requisition, nil },
		},
	}
	svc := testRequisitionService(requisitions, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})

	_, err := svc.Confirm(ctx, 1)
	if helper.AssertError(t, err, true, ErrRequestState) {
		return
	}
}

func TestPurchaseRequestService_Approve_RejectsWhenDraft(t *testing.T) {
	ctx := context.Background()
	requisition := &PurchaseRequest{Base: model.Base{ID: 1}, State: RequestStateDraft}
	requisitions := PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) { return requisition, nil },
		},
	}
	svc := testRequisitionService(requisitions, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})

	_, err := svc.Approve(ctx, 1)
	if helper.AssertError(t, err, true, ErrRequestState) {
		return
	}
}

func TestPurchaseRequestService_Approve_TransitionsConfirmedToApproved(t *testing.T) {
	ctx := context.Background()
	requisition := &PurchaseRequest{Base: model.Base{ID: 1}, State: RequestStateConfirmed}
	var updated *PurchaseRequest
	requisitions := PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) { return requisition, nil },
			UpdateFunc: func(_ context.Context, req *PurchaseRequest) (*PurchaseRequest, error) {
				updated = req
				return req, nil
			},
		},
	}
	svc := testRequisitionService(requisitions, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})

	approved, err := svc.Approve(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved.State != RequestStateApproved || updated.State != RequestStateApproved {
		t.Errorf("state = %s, want approved", approved.State)
	}
}

func TestPurchaseRequestService_Approve_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	requisitions := PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := testRequisitionService(requisitions, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})

	_, err := svc.Approve(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPurchaseRequestService_Cancel_TransitionsDraftToCancelled(t *testing.T) {
	ctx := context.Background()
	requisition := &PurchaseRequest{Base: model.Base{ID: 1}, State: RequestStateDraft}
	var updated *PurchaseRequest
	requisitions := PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) { return requisition, nil },
			UpdateFunc: func(_ context.Context, req *PurchaseRequest) (*PurchaseRequest, error) {
				updated = req
				return req, nil
			},
		},
	}
	svc := testRequisitionService(requisitions, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})

	cancelled, err := svc.Cancel(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancelled.State != RequestStateCancelled || updated.State != RequestStateCancelled {
		t.Errorf("state = %s, want cancelled", cancelled.State)
	}
}

func TestPurchaseRequestService_Cancel_RejectsWhenDone(t *testing.T) {
	ctx := context.Background()
	requisition := &PurchaseRequest{Base: model.Base{ID: 1}, State: RequestStateDone}
	requisitions := PurchaseRequestDAOMock{
		CRUDMock: dao.CRUDMock[PurchaseRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) { return requisition, nil },
		},
	}
	svc := testRequisitionService(requisitions, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})

	_, err := svc.Cancel(ctx, 1)
	if helper.AssertError(t, err, true, ErrRequestState) {
		return
	}
}

func TestPurchaseRequestService_ListLines_ReturnsRequisitionLines(t *testing.T) {
	ctx := context.Background()
	lines := PurchaseRequestLineDAOMock{
		ListByRequestFunc: func(_ context.Context, _ uint64) ([]*PurchaseRequestLine, error) {
			return []*PurchaseRequestLine{{Base: model.Base{ID: 9}, Qty: 4}}, nil
		},
	}
	svc := testRequisitionService(PurchaseRequestDAOMock{}, lines, contacts.ContactDAOMock{})

	got, err := svc.ListLines(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Qty != 4 {
		t.Errorf("lines = %v, want one line with qty 4", got)
	}
}
