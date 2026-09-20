package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
)

func TestSupplyAgreementService_Create_Branches(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	orgID := uint64(10)

	t.Run("rejects empty lines", func(t *testing.T) {
		svc, _, _ := testAgreementService()
		_, err := svc.Create(ctx, &SupplyAgreement{OrganizationID: &orgID, SupplierID: 10}, nil)
		helper.AssertError(t, err, true, ErrAgreementNoLines)
	})

	t.Run("rejects zero line qty", func(t *testing.T) {
		svc, _, _ := testAgreementService()
		_, err := svc.Create(ctx,
			&SupplyAgreement{OrganizationID: &orgID, SupplierID: 10},
			[]*SupplyAgreementLine{{Qty: 0, UnitPrice: 100}})
		helper.AssertError(t, err, true, ErrAgreementLineQty)
	})

	t.Run("propagates sequence error", func(t *testing.T) {
		svc, _, _ := testAgreementService()
		svc.sequences = errSequenceService(dbErr)
		_, err := svc.Create(ctx,
			&SupplyAgreement{OrganizationID: &orgID, SupplierID: 10},
			[]*SupplyAgreementLine{{Qty: 2, UnitPrice: 100}})
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("cancel propagates lookup error", func(t *testing.T) {
		svc, agreements, _ := testAgreementService()
		agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) { return nil, dbErr }
		_, err := svc.Cancel(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("consume propagates lookup error", func(t *testing.T) {
		svc, agreements, _ := testAgreementService()
		agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) { return nil, dbErr }
		err := svc.Consume(ctx, 1, 1, 100)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("find maps missing", func(t *testing.T) {
		svc, agreements, _ := testAgreementService()
		agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) { return nil, nil }
		_, err := svc.Find(ctx, 1)
		helper.AssertError(t, err, true, ErrAgreementNotFound)
	})
}

func TestSupplierScorecardService_LookupErrors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	doneAt := time.Now()

	t.Run("update propagates lookup error", func(t *testing.T) {
		svc, scorecards := testScorecardService()
		scorecards.FindByVendorAndPeriodFunc = func(_ context.Context, _ uint64, _, _ *time.Time) (*SupplierScorecard, error) {
			return nil, dbErr
		}
		err := svc.UpdateScorecard(ctx, nil, 10, doneAt)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("record failure propagates lookup error", func(t *testing.T) {
		svc, scorecards := testScorecardService()
		scorecards.FindByVendorAndPeriodFunc = func(_ context.Context, _ uint64, _, _ *time.Time) (*SupplierScorecard, error) {
			return nil, dbErr
		}
		err := svc.RecordQualityFailure(ctx, nil, 10, doneAt)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("record failure creates new scorecard", func(t *testing.T) {
		svc, scorecards := testScorecardService()
		scorecards.FindByVendorAndPeriodFunc = func(_ context.Context, _ uint64, _, _ *time.Time) (*SupplierScorecard, error) {
			return nil, nil
		}
		scorecards.CreateFunc = func(_ context.Context, sc *SupplierScorecard) (*SupplierScorecard, error) {
			sc.ID = 2
			return sc, nil
		}
		if err := svc.RecordQualityFailure(ctx, helper.Ptr(uint64(10)), 10, doneAt); helper.AssertError(t, err, false, nil) {
			return
		}
	})
}

func TestPurchaseOrderService_GetBatch_LookupError(t *testing.T) {
	dbErr := errors.New("db down")
	svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
	svc.batches = &PaymentBatchDAOMock{
		CRUDMock: dao.CRUDMock[PaymentBatch]{
			FindFunc: func(_ context.Context, _ uint64) (*PaymentBatch, error) { return nil, dbErr },
		},
	}
	_, _, err := svc.GetPaymentBatch(context.Background(), 1)
	helper.AssertError(t, err, true, dbErr)
}
