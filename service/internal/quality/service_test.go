package quality

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func testQualityService(points QualityPointDAOMock, checks QualityCheckDAOMock, alerts QualityAlertDAOMock) QualityCheckService {
	return NewQualityCheckService(points, checks, alerts)
}

type scrapRouterMock struct {
	routeFunc func(ctx context.Context, organizationID, shipmentID uint64, itemID uint64, journalID, expenseAccountID uint64, date time.Time) error
}

func (m scrapRouterMock) RouteToScrap(ctx context.Context, organizationID, shipmentID uint64, itemID uint64, journalID, expenseAccountID uint64, date time.Time) error {
	if m.routeFunc != nil {
		return m.routeFunc(ctx, organizationID, shipmentID, itemID, journalID, expenseAccountID, date)
	}
	return nil
}

func measurePoint(normMin, normMax *float64) *reference.QualityPoint {
	return &reference.QualityPoint{
		Base:     model.Base{ID: 1},
		ItemID:   helper.Ptr(uint64(200)),
		TestType: TestTypeMeasure,
		NormMin:  normMin,
		NormMax:  normMax,
	}
}

func TestQualityCheckService_TriggerChecks(t *testing.T) {
	points := QualityPointDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*reference.QualityPoint, error) {
			return []*reference.QualityPoint{
				{Base: model.Base{ID: 1}, TestType: TestTypePassFail},
				{Base: model.Base{ID: 2}, TestType: TestTypeMeasure},
			}, nil
		},
	}
	created := 0
	checks := QualityCheckDAOMock{
		CRUDMock: dao.CRUDMock[QualityCheck]{
			CreateFunc: func(_ context.Context, check *QualityCheck) (*QualityCheck, error) {
				created++
				return check, nil
			},
		},
	}
	svc := testQualityService(points, checks, QualityAlertDAOMock{})

	count, err := svc.TriggerChecks(context.Background(), 1, 50, []uint64{200})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 || created != 2 {
		t.Errorf("count = %d created = %d, want 2", count, created)
	}
}

func TestQualityCheckService_RecordResult(t *testing.T) {
	tests := []struct {
		name       string
		points     QualityPointDAOMock
		check      *QualityCheck
		pass       bool
		measured   *float64
		wantResult string
		wantAlert  bool
		wantErr    error
	}{
		{
			name: "pass fail records pass",
			points: QualityPointDAOMock{
				CRUDMock: dao.CRUDMock[reference.QualityPoint]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.QualityPoint, error) {
						return &reference.QualityPoint{Base: model.Base{ID: 1}, TestType: TestTypePassFail}, nil
					},
				},
			},
			check:      &QualityCheck{Base: model.Base{ID: 1}, PointID: helper.Ptr(uint64(1)), ItemID: helper.Ptr(uint64(200)), Result: CheckResultPending},
			pass:       true,
			wantResult: CheckResultPass,
		},
		{
			name: "measure compares against norm",
			points: QualityPointDAOMock{
				CRUDMock: dao.CRUDMock[reference.QualityPoint]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.QualityPoint, error) {
						return measurePoint(helper.Ptr(10.0), helper.Ptr(20.0)), nil
					},
				},
			},
			check:      &QualityCheck{Base: model.Base{ID: 1}, PointID: helper.Ptr(uint64(1)), Result: CheckResultPending},
			pass:       false,
			measured:   helper.Ptr(15.0),
			wantResult: CheckResultPass,
		},
		{
			name: "out of range fails and creates alert",
			points: QualityPointDAOMock{
				CRUDMock: dao.CRUDMock[reference.QualityPoint]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.QualityPoint, error) {
						return measurePoint(helper.Ptr(10.0), helper.Ptr(20.0)), nil
					},
				},
			},
			check:      &QualityCheck{Base: model.Base{ID: 1}, PointID: helper.Ptr(uint64(1)), ItemID: helper.Ptr(uint64(200)), Result: CheckResultPending},
			pass:       true,
			measured:   helper.Ptr(25.0),
			wantResult: CheckResultFail,
			wantAlert:  true,
		},
		{
			name: "rejects already resolved",
			points: QualityPointDAOMock{
				CRUDMock: dao.CRUDMock[reference.QualityPoint]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.QualityPoint, error) {
						return &reference.QualityPoint{Base: model.Base{ID: 1}, TestType: TestTypePassFail}, nil
					},
				},
			},
			check:   &QualityCheck{Base: model.Base{ID: 1}, PointID: helper.Ptr(uint64(1)), Result: CheckResultPass},
			pass:    true,
			wantErr: ErrQualityCheckDone,
		},
		{
			name: "rejects measure without value",
			points: QualityPointDAOMock{
				CRUDMock: dao.CRUDMock[reference.QualityPoint]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.QualityPoint, error) {
						return measurePoint(nil, nil), nil
					},
				},
			},
			check:   &QualityCheck{Base: model.Base{ID: 1}, PointID: helper.Ptr(uint64(1)), Result: CheckResultPending},
			pass:    true,
			wantErr: ErrQualityCheckNoValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			check := tt.check
			checks := QualityCheckDAOMock{
				CRUDMock: dao.CRUDMock[QualityCheck]{
					FindFunc: func(_ context.Context, _ uint64) (*QualityCheck, error) { return check, nil },
					UpdateFunc: func(_ context.Context, c *QualityCheck) (*QualityCheck, error) {
						check = c
						return c, nil
					},
				},
			}
			alertCreated := false
			alerts := QualityAlertDAOMock{
				CRUDMock: dao.CRUDMock[QualityAlert]{
					CreateFunc: func(_ context.Context, alert *QualityAlert) (*QualityAlert, error) {
						alertCreated = true
						return alert, nil
					},
				},
			}
			svc := testQualityService(tt.points, checks, alerts)

			updated, err := svc.RecordResult(ctx, 1, tt.pass, tt.measured, 9)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if tt.wantResult != "" && updated.Result != tt.wantResult {
				t.Errorf("result = %s, want %s", updated.Result, tt.wantResult)
			}
			if tt.wantAlert != alertCreated {
				t.Errorf("alert created = %v, want %v", alertCreated, tt.wantAlert)
			}
		})
	}
}

func TestQualityCheckService_HasFailedChecks(t *testing.T) {
	tests := []struct {
		name     string
		checks   QualityCheckDAOMock
		wantFail bool
	}{
		{
			name: "detects failures",
			checks: QualityCheckDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*QualityCheck, error) {
					return []*QualityCheck{{Base: model.Base{ID: 1}, Result: CheckResultFail}}, nil
				},
			},
			wantFail: true,
		},
		{
			name: "no failures when all pass",
			checks: QualityCheckDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*QualityCheck, error) {
					return []*QualityCheck{{Base: model.Base{ID: 1}, Result: CheckResultPass}}, nil
				},
			},
			wantFail: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testQualityService(QualityPointDAOMock{}, tt.checks, QualityAlertDAOMock{})
			failed, err := svc.HasFailedChecks(context.Background(), 50)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if failed != tt.wantFail {
				t.Errorf("failed = %v, want %v", failed, tt.wantFail)
			}
		})
	}
}

func TestQualityCheckService_RouteFailedToScrap(t *testing.T) {
	tests := []struct {
		name      string
		checks    QualityCheckDAOMock
		router    *scrapRouterMock
		wantScrap int
		wantErr   error
	}{
		{
			name: "routes only failed products",
			checks: QualityCheckDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*QualityCheck, error) {
					return []*QualityCheck{
						{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(200)), Result: CheckResultFail},
						{Base: model.Base{ID: 2}, ItemID: helper.Ptr(uint64(201)), Result: CheckResultPass},
						{Base: model.Base{ID: 3}, ItemID: helper.Ptr(uint64(202)), Result: CheckResultFail},
					}, nil
				},
			},
			router: &scrapRouterMock{
				routeFunc: func(_ context.Context, _ uint64, _ uint64, _ uint64, _ uint64, _ uint64, _ time.Time) error {
					return nil
				},
			},
			wantScrap: 2,
		},
		{
			name:    "rejects without router",
			checks:  QualityCheckDAOMock{},
			wantErr: ErrQualityScrapRouter,
		},
		{
			name: "skips when no failures",
			checks: QualityCheckDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*QualityCheck, error) {
					return []*QualityCheck{{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(200)), Result: CheckResultPass}}, nil
				},
			},
			router: &scrapRouterMock{
				routeFunc: func(_ context.Context, _ uint64, _ uint64, _ uint64, _ uint64, _ uint64, _ time.Time) error {
					return nil
				},
			},
			wantScrap: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testQualityService(QualityPointDAOMock{}, tt.checks, QualityAlertDAOMock{})
			if tt.router != nil {
				svc.SetScrapRouter(*tt.router)
			}

			scrapped, err := svc.RouteFailedToScrap(context.Background(), 1, 50, 90, 6000, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if len(scrapped) != tt.wantScrap {
				t.Errorf("scrapped = %v, want %d products", len(scrapped), tt.wantScrap)
			}
		})
	}
}

func TestQualityCheckService_SetAlertState(t *testing.T) {
	tests := []struct {
		name    string
		alert   *QualityAlert
		target  string
		wantErr error
	}{
		{
			name:   "transitions open to in progress",
			alert:  &QualityAlert{Base: model.Base{ID: 1}, State: AlertStateOpen},
			target: AlertStateInProgress,
		},
		{
			name:    "rejects illegal transition",
			alert:   &QualityAlert{Base: model.Base{ID: 1}, State: AlertStateSolved},
			target:  AlertStateOpen,
			wantErr: ErrQualityAlertState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			alert := tt.alert
			alerts := QualityAlertDAOMock{
				CRUDMock: dao.CRUDMock[QualityAlert]{
					FindFunc: func(_ context.Context, _ uint64) (*QualityAlert, error) { return alert, nil },
					UpdateFunc: func(_ context.Context, a *QualityAlert) (*QualityAlert, error) {
						alert = a
						return a, nil
					},
				},
			}
			svc := testQualityService(QualityPointDAOMock{}, QualityCheckDAOMock{}, alerts)

			updated, err := svc.SetAlertState(ctx, 1, tt.target)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if updated.State != tt.target {
				t.Errorf("state = %s, want %s", updated.State, tt.target)
			}
		})
	}
}
