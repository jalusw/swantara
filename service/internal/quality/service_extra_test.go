package quality

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestQualityCheckService_TriggerChecks_SkipsForeignOrgAndPropagatesErrors(t *testing.T) {
	ctx := context.Background()

	var created int
	points := QualityPointDAOMock{
		ListByItemFunc: func(_ context.Context, itemID uint64) ([]*reference.QualityPoint, error) {
			if itemID == 1 {
				return []*reference.QualityPoint{
					{Base: model.Base{ID: 1}, TestType: TestTypePassFail},
					{Base: model.Base{ID: 2}, OrganizationID: helper.Ptr(uint64(99)), TestType: TestTypePassFail},
				}, nil
			}
			return []*reference.QualityPoint{}, nil
		},
	}
	checks := QualityCheckDAOMock{
		CRUDMock: dao.CRUDMock[QualityCheck]{
			CreateFunc: func(_ context.Context, check *QualityCheck) (*QualityCheck, error) {
				created++
				return check, nil
			},
		},
	}
	svc := testQualityService(points, checks, QualityAlertDAOMock{})

	total, err := svc.TriggerChecks(ctx, 10, 3, []uint64{1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || created != 1 {
		t.Errorf("total = %d, created = %d, want 1/1 (foreign org point skipped)", total, created)
	}

	pointErr := QualityPointDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*reference.QualityPoint, error) {
			return nil, errors.New("point list failed")
		},
	}
	svc = testQualityService(pointErr, QualityCheckDAOMock{}, QualityAlertDAOMock{})
	if _, err := svc.TriggerChecks(ctx, 10, 3, []uint64{1}); err == nil {
		t.Error("expected ListByItem error propagation")
	}

	checkErr := QualityCheckDAOMock{
		CRUDMock: dao.CRUDMock[QualityCheck]{
			CreateFunc: func(_ context.Context, _ *QualityCheck) (*QualityCheck, error) {
				return nil, errors.New("create failed")
			},
		},
	}
	pointsWithOne := QualityPointDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*reference.QualityPoint, error) {
			return []*reference.QualityPoint{{Base: model.Base{ID: 1}, TestType: TestTypePassFail}}, nil
		},
	}
	svc = testQualityService(pointsWithOne, checkErr, QualityAlertDAOMock{})
	if _, err := svc.TriggerChecks(ctx, 10, 3, []uint64{9}); err == nil {
		t.Error("expected Create error propagation")
	}
}

func TestQualityCheckService_RecordResult_ErrorBranches(t *testing.T) {
	ctx := context.Background()

	svc := testQualityService(
		QualityPointDAOMock{},
		QualityCheckDAOMock{CRUDMock: dao.CRUDMock[QualityCheck]{FindFunc: func(_ context.Context, _ uint64) (*QualityCheck, error) {
			return nil, nil
		}}},
		QualityAlertDAOMock{},
	)
	if _, err := svc.RecordResult(ctx, 1, true, nil, 1); !errors.Is(err, ErrQualityCheckNotFound) {
		t.Errorf("nil check err = %v, want ErrQualityCheckNotFound", err)
	}

	svc = testQualityService(
		QualityPointDAOMock{},
		QualityCheckDAOMock{CRUDMock: dao.CRUDMock[QualityCheck]{FindFunc: func(_ context.Context, _ uint64) (*QualityCheck, error) {
			return &QualityCheck{Base: model.Base{ID: 1}, Result: CheckResultPending, PointID: nil}, nil
		}}},
		QualityAlertDAOMock{},
	)
	if _, err := svc.RecordResult(ctx, 1, true, nil, 1); !errors.Is(err, ErrQualityCheckNoProduct) {
		t.Errorf("nil point-id err = %v, want ErrQualityCheckNoProduct", err)
	}

	svc = testQualityService(
		QualityPointDAOMock{CRUDMock: dao.CRUDMock[reference.QualityPoint]{FindFunc: func(_ context.Context, _ uint64) (*reference.QualityPoint, error) {
			return nil, nil
		}}},
		QualityCheckDAOMock{CRUDMock: dao.CRUDMock[QualityCheck]{FindFunc: func(_ context.Context, _ uint64) (*QualityCheck, error) {
			return &QualityCheck{Base: model.Base{ID: 1}, Result: CheckResultPending, PointID: helper.Ptr(uint64(1))}, nil
		}}},
		QualityAlertDAOMock{},
	)
	if _, err := svc.RecordResult(ctx, 1, true, nil, 1); !errors.Is(err, ErrQualityPointNotFound) {
		t.Errorf("nil point err = %v, want ErrQualityPointNotFound", err)
	}

	svc = testQualityService(
		QualityPointDAOMock{CRUDMock: dao.CRUDMock[reference.QualityPoint]{FindFunc: func(_ context.Context, _ uint64) (*reference.QualityPoint, error) {
			return &reference.QualityPoint{Base: model.Base{ID: 1}, TestType: TestTypePassFail}, nil
		}}},
		QualityCheckDAOMock{CRUDMock: dao.CRUDMock[QualityCheck]{FindFunc: func(_ context.Context, _ uint64) (*QualityCheck, error) {
			return &QualityCheck{Base: model.Base{ID: 1}, Result: CheckResultPass}, nil
		}}},
		QualityAlertDAOMock{},
	)
	if _, err := svc.RecordResult(ctx, 1, true, nil, 1); !errors.Is(err, ErrQualityCheckDone) {
		t.Errorf("done check err = %v, want ErrQualityCheckDone", err)
	}
}

func TestQualityCheckService_RecordResult_AlertCreateError(t *testing.T) {
	ctx := context.Background()
	svc := testQualityService(
		QualityPointDAOMock{CRUDMock: dao.CRUDMock[reference.QualityPoint]{FindFunc: func(_ context.Context, _ uint64) (*reference.QualityPoint, error) {
			return &reference.QualityPoint{Base: model.Base{ID: 1}, TestType: TestTypePassFail}, nil
		}}},
		QualityCheckDAOMock{CRUDMock: dao.CRUDMock[QualityCheck]{
			FindFunc: func(_ context.Context, _ uint64) (*QualityCheck, error) {
				return &QualityCheck{Base: model.Base{ID: 1}, PointID: helper.Ptr(uint64(1)), ItemID: helper.Ptr(uint64(2)), Result: CheckResultPending}, nil
			},
			UpdateFunc: func(_ context.Context, check *QualityCheck) (*QualityCheck, error) {
				return check, nil
			},
		}},
		QualityAlertDAOMock{CRUDMock: dao.CRUDMock[QualityAlert]{CreateFunc: func(_ context.Context, _ *QualityAlert) (*QualityAlert, error) {
			return nil, errors.New("alert create failed")
		}}},
	)
	if _, err := svc.RecordResult(ctx, 1, false, nil, 1); err == nil {
		t.Fatal("expected alert create error propagation")
	}
}

func TestQualityCheckService_HasFailedChecks_ListError(t *testing.T) {
	ctx := context.Background()
	svc := testQualityService(
		QualityPointDAOMock{},
		QualityCheckDAOMock{ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*QualityCheck, error) {
			return nil, errors.New("list failed")
		}},
		QualityAlertDAOMock{},
	)
	if _, err := svc.HasFailedChecks(ctx, 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestQualityCheckService_RouteFailedToScrap_RouterError(t *testing.T) {
	ctx := context.Background()
	svc := testQualityService(
		QualityPointDAOMock{},
		QualityCheckDAOMock{ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*QualityCheck, error) {
			return []*QualityCheck{
				{Base: model.Base{ID: 1}, Result: CheckResultFail, ItemID: helper.Ptr(uint64(2))},
			}, nil
		}},
		QualityAlertDAOMock{},
	)
	svc.SetScrapRouter(scrapRouterMock{routeFunc: func(_ context.Context, _, _, _ uint64, _, _ uint64, _ time.Time) error {
		return errors.New("scrap failed")
	}})
	if _, err := svc.RouteFailedToScrap(ctx, 1, 2, 3, 4, time.Now()); err == nil {
		t.Fatal("expected scrap router error propagation")
	}
}

func TestQualityCheckService_SetAlertState_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := testQualityService(
		QualityPointDAOMock{},
		QualityCheckDAOMock{},
		QualityAlertDAOMock{CRUDMock: dao.CRUDMock[QualityAlert]{FindFunc: func(_ context.Context, _ uint64) (*QualityAlert, error) {
			return nil, nil
		}}},
	)
	if _, err := svc.SetAlertState(ctx, 1, AlertStateInProgress); !errors.Is(err, ErrQualityAlertNotFound) {
		t.Fatalf("err = %v, want ErrQualityAlertNotFound", err)
	}
}

func TestQualityCheckService_InRangeBoundaries(t *testing.T) {
	svc := testQualityService(QualityPointDAOMock{}, QualityCheckDAOMock{}, QualityAlertDAOMock{})

	normMin := 10.0
	normMax := 20.0
	point := measurePoint(&normMin, &normMax)

	if !svc.inRange(point, 10) || !svc.inRange(point, 20) {
		t.Error("boundary values should be in range")
	}
	if svc.inRange(point, 9.99) || svc.inRange(point, 20.01) {
		t.Error("out-of-bound values should be out of range")
	}
	if !svc.inRange(measurePoint(nil, nil), 999) {
		t.Error("no norm should always pass")
	}
	if svc.inRange(measurePoint(&normMin, nil), 5) {
		t.Error("below min should fail")
	}
	if svc.inRange(measurePoint(nil, &normMax), 25) {
		t.Error("above max should fail")
	}
}
