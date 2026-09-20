package procurement

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
)

func errContactDAO(dbErr error) contacts.ContactDAOMock {
	return contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return nil, dbErr
			},
		},
	}
}

func errSequenceService(dbErr error) sequence.Service {
	return sequence.NewSequenceService(sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return nil, dbErr
		},
	})
}

func TestPurchaseRequestService_Create_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	orgID := uint64(10)

	newReq := func() *PurchaseRequest {
		return &PurchaseRequest{OrganizationID: &orgID, RequesterID: 5}
	}
	lines := []*PurchaseRequestLine{{Qty: 2}}

	t.Run("propagates contact lookup error", func(t *testing.T) {
		svc := NewPurchaseRequestService(PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{},
			sequence.NewSequenceService(sequence.DAOMock{}), errContactDAO(dbErr))
		_, err := svc.Create(ctx, newReq(), lines)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("propagates sequence error", func(t *testing.T) {
		svc := NewPurchaseRequestService(PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{},
			errSequenceService(dbErr), contacts.ContactDAOMock{
				CRUDMock: dao.CRUDMock[contacts.Contact]{
					FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
						return &contacts.Contact{Base: model.Base{ID: 5}}, nil
					},
				},
			})
		_, err := svc.Create(ctx, newReq(), lines)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("confirm propagates lookup error", func(t *testing.T) {
		svc := testRequisitionService(PurchaseRequestDAOMock{
			CRUDMock: dao.CRUDMock[PurchaseRequest]{
				FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) { return nil, dbErr },
			},
		}, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})
		_, err := svc.Confirm(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("cancel propagates lookup error", func(t *testing.T) {
		svc := testRequisitionService(PurchaseRequestDAOMock{
			CRUDMock: dao.CRUDMock[PurchaseRequest]{
				FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) { return nil, dbErr },
			},
		}, PurchaseRequestLineDAOMock{}, contacts.ContactDAOMock{})
		_, err := svc.Cancel(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestSupplyAgreementService_Transitions(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("activate propagates lookup error", func(t *testing.T) {
		svc, agreements, _ := testAgreementService()
		agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) { return nil, dbErr }
		_, err := svc.Activate(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("cancel rejects invalid transition", func(t *testing.T) {
		svc, agreements, _ := testAgreementService()
		agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
			return &SupplyAgreement{Base: model.Base{ID: 1}, State: SupplyAgreementStateCancelled}, nil
		}
		_, err := svc.Cancel(ctx, 1)
		helper.AssertError(t, err, true, ErrSupplyAgreementState)
	})

	t.Run("cancel propagates update error", func(t *testing.T) {
		svc, agreements, _ := testAgreementService()
		agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
			return &SupplyAgreement{Base: model.Base{ID: 1}, State: SupplyAgreementStateDraft}, nil
		}
		agreements.UpdateFunc = func(_ context.Context, _ *SupplyAgreement) (*SupplyAgreement, error) { return nil, dbErr }
		_, err := svc.Cancel(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestReplaceLines_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("requisition delete error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "purchase_request_lines"`)).WillReturnError(dbErr)
		mock.ExpectRollback()

		err := NewPurchaseRequestLineDAO(db).ReplaceLines(ctx, 7, []*PurchaseRequestLine{{Qty: 1}})
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("requisition create error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "purchase_request_lines"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_request_lines"`)).WillReturnError(dbErr)
		mock.ExpectRollback()

		err := NewPurchaseRequestLineDAO(db).ReplaceLines(ctx, 7, []*PurchaseRequestLine{{Qty: 1}})
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("order delete error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "purchase_order_lines"`)).WillReturnError(dbErr)
		mock.ExpectRollback()

		err := NewPurchaseOrderLineDAO(db).ReplaceLines(ctx, 7, []*PurchaseOrderLine{{QtyOrdered: 1}})
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("order create error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "purchase_order_lines"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_order_lines"`)).WillReturnError(dbErr)
		mock.ExpectRollback()

		err := NewPurchaseOrderLineDAO(db).ReplaceLines(ctx, 7, []*PurchaseOrderLine{{QtyOrdered: 1}})
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestPurchaseOrderService_Batch_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("confirm propagates lines error", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.batches = &PaymentBatchDAOMock{
			CRUDMock: dao.CRUDMock[PaymentBatch]{
				FindFunc: func(_ context.Context, _ uint64) (*PaymentBatch, error) {
					return &PaymentBatch{Base: model.Base{ID: 1}, State: PaymentBatchStateDraft}, nil
				},
			},
		}
		svc.batchLines = &PaymentBatchLineDAOMock{
			ListByBatchFunc: func(_ context.Context, _ uint64) ([]*PaymentBatchLine, error) { return nil, dbErr },
		}
		_, err := svc.ConfirmBatch(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("confirm propagates order error", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.batches = &PaymentBatchDAOMock{
			CRUDMock: dao.CRUDMock[PaymentBatch]{
				FindFunc: func(_ context.Context, _ uint64) (*PaymentBatch, error) {
					return &PaymentBatch{Base: model.Base{ID: 1}, State: PaymentBatchStateDraft}, nil
				},
			},
		}
		svc.batchLines = &PaymentBatchLineDAOMock{
			ListByBatchFunc: func(_ context.Context, _ uint64) ([]*PaymentBatchLine, error) {
				return []*PaymentBatchLine{{Base: model.Base{ID: 1}, OrderID: 9}}, nil
			},
		}
		svc.orders = &PurchaseOrderDAOMock{
			CRUDMock: dao.CRUDMock[PurchaseOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return nil, dbErr },
			},
		}
		_, err := svc.ConfirmBatch(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("get propagates lines error", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.batches = &PaymentBatchDAOMock{
			CRUDMock: dao.CRUDMock[PaymentBatch]{
				FindFunc: func(_ context.Context, _ uint64) (*PaymentBatch, error) {
					return &PaymentBatch{Base: model.Base{ID: 1}, State: PaymentBatchStateDraft}, nil
				},
			},
		}
		svc.batchLines = &PaymentBatchLineDAOMock{
			ListByBatchFunc: func(_ context.Context, _ uint64) ([]*PaymentBatchLine, error) { return nil, dbErr },
		}
		_, _, err := svc.GetPaymentBatch(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("debit memo propagates persist error", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.orders = confirmedPOMock(PurchaseOrderStateConfirmed)
		svc.debitMemos = &PurchaseDebitMemoDAOMock{
			CRUDMock: dao.CRUDMock[PurchaseDebitMemo]{
				CreateFunc: func(_ context.Context, _ *PurchaseDebitMemo) (*PurchaseDebitMemo, error) { return nil, dbErr },
			},
		}
		_, err := svc.CreateVendorDebitMemo(ctx, 1, 5, time.Now(), 250, "shortage")
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("credit memo propagates persist error", func(t *testing.T) {
		svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
		svc.orders = confirmedPOMock(PurchaseOrderStateConfirmed)
		svc.creditMemos = &PurchaseCreditMemoDAOMock{
			CRUDMock: dao.CRUDMock[PurchaseCreditMemo]{
				CreateFunc: func(_ context.Context, _ *PurchaseCreditMemo) (*PurchaseCreditMemo, error) { return nil, dbErr },
			},
		}
		_, err := svc.CreateVendorCreditMemo(ctx, 1, 5, time.Now(), 500, "return")
		helper.AssertError(t, err, true, dbErr)
	})
}
