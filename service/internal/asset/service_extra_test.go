package asset

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestAssetService_Register_Extra(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ctx *assetTestContext)
		wantErr error
		wantMsg string
		assert  func(t *testing.T, result *FixedAsset)
	}{
		{
			name:    "RejectsNegativePurchase",
			setup:   func(ctx *assetTestContext) {},
			wantErr: ErrAssetInvalidValues,
		},
		{
			name:    "RejectsNegativeSalvage",
			setup:   func(ctx *assetTestContext) {},
			wantErr: ErrAssetInvalidValues,
		},
		{
			name:    "RejectsZeroAcquisitionDate",
			setup:   func(ctx *assetTestContext) {},
			wantErr: ErrAssetInvalidDates,
		},
		{
			name:    "RejectsZeroInServiceDate",
			setup:   func(ctx *assetTestContext) {},
			wantErr: ErrAssetInvalidDates,
		},
		{
			name:    "RejectsInServiceBeforeAcquisition",
			setup:   func(ctx *assetTestContext) {},
			wantErr: ErrAssetInvalidDates,
		},
		{
			name: "PropagatesCategoryLookupError",
			setup: func(ctx *assetTestContext) {
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "RejectsMissingCategory",
			setup: func(ctx *assetTestContext) {
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return nil, nil
				}
			},
			wantErr: ErrAssetCategoryNotFound,
		},
		{
			name: "RejectsMissingMethod",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				category.Method = nil
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
			},
			wantErr: ErrAssetInvalidMethod,
		},
		{
			name: "RejectsInvalidMethod",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				method := "sum_of_digits"
				category.Method = &method
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
			},
			wantErr: ErrAssetInvalidMethod,
		},
		{
			name: "RejectsNonPositivePeriods",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				periods := 0
				category.MethodNumber = &periods
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
			},
			wantErr: ErrAssetInvalidPeriods,
		},
		{
			name: "RejectsMissingAssetAccounts",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				category.AssetAccountID = nil
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
			},
			wantErr: ErrAssetCategoryAccounts,
		},
		{
			name: "RejectsMissingGainAccount",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				category.GainAccountID = nil
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
			},
			wantErr: ErrAssetCategoryAccounts,
		},
		{
			name: "PropagatesInvoiceLineLookupError",
			setup: func(ctx *assetTestContext) {
				stubRegisterDependencies(ctx)
				ctx.invoiceLines.FindFunc = func(_ context.Context, _ uint64) (*accounting.InvoiceLine, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "RejectsMissingInvoiceLine",
			setup: func(ctx *assetTestContext) {
				stubRegisterDependencies(ctx)
				ctx.invoiceLines.FindFunc = func(_ context.Context, _ uint64) (*accounting.InvoiceLine, error) {
					return nil, nil
				}
			},
			wantErr: ErrAssetInvoiceLineNotFound,
		},
		{
			name: "PropagatesInvoiceLookupError",
			setup: func(ctx *assetTestContext) {
				stubRegisterDependencies(ctx)
				ctx.invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "RejectsMissingInvoice",
			setup: func(ctx *assetTestContext) {
				stubRegisterDependencies(ctx)
				ctx.invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
					return nil, nil
				}
			},
			wantErr: ErrAssetInvoiceNotSupplierBill,
		},
		{
			name: "RejectsNonSupplierBillType",
			setup: func(ctx *assetTestContext) {
				stubRegisterDependencies(ctx)
				ctx.invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
					return &accounting.Invoice{Base: model.Base{ID: 9}, Type: accounting.InvoiceTypeCustomerInvoice, State: accounting.InvoiceStatePosted}, nil
				}
			},
			wantErr: ErrAssetInvoiceNotSupplierBill,
		},
		{
			name: "PropagatesCreateError",
			setup: func(ctx *assetTestContext) {
				stubRegisterDependencies(ctx)
				ctx.assets.CreateFunc = func(_ context.Context, _ *FixedAsset) (*FixedAsset, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newAssetTestContext()
			tt.setup(ctx)
			request := registerRequest()
			if tt.name == "RejectsNegativePurchase" {
				request.PurchaseValue = -1
			}
			if tt.name == "RejectsNegativeSalvage" {
				request.SalvageValue = -1
			}
			if tt.name == "RejectsZeroAcquisitionDate" {
				request.AcquisitionDate = time.Time{}
			}
			if tt.name == "RejectsZeroInServiceDate" {
				request.InServiceDate = time.Time{}
			}
			if tt.name == "RejectsInServiceBeforeAcquisition" {
				request.InServiceDate = time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
			}
			_, err := ctx.svc.Register(context.Background(), request)
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if tt.wantMsg != "" {
				helper.AssertError(t, err, true, nil)
				if err.Error() != tt.wantMsg {
					t.Fatalf("error = %v, want %v", err, tt.wantMsg)
				}
				return
			}
			helper.AssertError(t, err, false, nil)
		})
	}
}

func TestAssetService_GenerateSchedule_Extra(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ctx *assetTestContext)
		wantErr error
		wantMsg string
		assert  func(t *testing.T, schedule []AssetDepreciationLine)
	}{
		{
			name: "PropagatesAssetLookupError",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "RejectsMissingAsset",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return nil, nil
				}
			},
			wantErr: ErrAssetNotFound,
		},
		{
			name: "RejectsNonRunningAsset",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return &FixedAsset{Base: model.Base{ID: 1}, State: AssetStateDisposed}, nil
				}
			},
			wantErr: ErrAssetNotRunning,
		},
		{
			name: "PropagatesCategoryError",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "PropagatesListError",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "PropagatesCreateError",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
				ctx.lines.CreateTxFunc = func(_ context.Context, _ *gorm.DB, _ *AssetDepreciationLine) (*AssetDepreciationLine, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "DecliningThenLinearEndsAtSalvage",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				method := MethodDecliningThenLinear
				category.Method = &method
				category.MethodNumber = intPtr(5)
				asset := &FixedAsset{
					Base:           model.Base{ID: 1},
					OrganizationID: 1,
					Name:           "Server",
					CategoryID:     1,
					PurchaseValue:  100,
					SalvageValue:   0,
					State:          AssetStateRunning,
					InServiceDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
				}
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return asset, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
			},
			assert: func(t *testing.T, schedule []AssetDepreciationLine) {
				if len(schedule) != 5 {
					t.Fatalf("schedule lines = %d, want 5", len(schedule))
				}
				if last := schedule[len(schedule)-1]; last.Accumulated != 100 {
					t.Fatalf("final accumulated = %v, want 100", last.Accumulated)
				}
			},
		},
		{
			name: "DecliningWithoutSalvage",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				method := MethodDeclining
				category.Method = &method
				asset := &FixedAsset{
					Base:           model.Base{ID: 1},
					OrganizationID: 1,
					Name:           "Server",
					CategoryID:     1,
					PurchaseValue:  100,
					SalvageValue:   0,
					State:          AssetStateRunning,
					InServiceDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
				}
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return asset, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
			},
			assert: func(t *testing.T, schedule []AssetDepreciationLine) {
				if last := schedule[len(schedule)-1]; last.Accumulated != 100 {
					t.Fatalf("final accumulated = %v, want 100", last.Accumulated)
				}
			},
		},
		{
			name: "DecliningWithZeroPurchase",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				method := MethodDeclining
				category.Method = &method
				asset := &FixedAsset{
					Base:           model.Base{ID: 1},
					OrganizationID: 1,
					Name:           "Server",
					CategoryID:     1,
					PurchaseValue:  0,
					SalvageValue:   100,
					State:          AssetStateRunning,
					InServiceDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
				}
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return asset, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
			},
			assert: func(t *testing.T, schedule []AssetDepreciationLine) {
				if len(schedule) != 5 {
					t.Fatalf("schedule lines = %d, want 5", len(schedule))
				}
			},
		},
		{
			name: "YearlyPeriod",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				period := "year"
				category.MethodPeriod = &period
				category.MethodNumber = intPtr(2)
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
			},
			assert: func(t *testing.T, schedule []AssetDepreciationLine) {
				if len(schedule) != 2 {
					t.Fatalf("schedule lines = %d, want 2", len(schedule))
				}
				wantSecond := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
				if !schedule[1].DepreciationDate.Equal(wantSecond) {
					t.Fatalf("second depreciation date = %v, want %v", schedule[1].DepreciationDate, wantSecond)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newAssetTestContext()
			tt.setup(ctx)
			schedule, err := ctx.svc.GenerateSchedule(context.Background(), GenerateScheduleRequest{AssetID: 1})
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if tt.wantMsg != "" {
				helper.AssertError(t, err, true, nil)
				if err.Error() != tt.wantMsg {
					t.Fatalf("error = %v, want %v", err, tt.wantMsg)
				}
				return
			}
			helper.AssertError(t, err, false, nil)
			if tt.assert != nil {
				tt.assert(t, schedule)
			}
		})
	}
}

func TestAssetService_PostDepreciation_Extra(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ctx *assetTestContext)
		wantErr error
		wantMsg string
		assert  func(t *testing.T, posted *AssetDepreciationLine)
	}{
		{
			name: "PropagatesAssetLookupError",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "RejectsMissingAsset",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return nil, nil
				}
			},
			wantErr: ErrAssetNotFound,
		},
		{
			name: "RejectsNonRunningAsset",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return &FixedAsset{Base: model.Base{ID: 1}, State: AssetStateSold}, nil
				}
			},
			wantErr: ErrAssetNotRunning,
		},
		{
			name: "PropagatesCategoryError",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "PropagatesListError",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "SkipsPostedAndFutureLines",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				postedLine := &AssetDepreciationLine{Base: model.Base{ID: 1}, Sequence: 1, DepreciationDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Amount: 100, Posted: true}
				futureLine := &AssetDepreciationLine{Base: model.Base{ID: 2}, Sequence: 2, DepreciationDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), Amount: 200}
				dueLine := &AssetDepreciationLine{Base: model.Base{ID: 3}, Sequence: 3, DepreciationDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Amount: 300}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{postedLine, futureLine, dueLine}, nil
				}
			},
			assert: func(t *testing.T, posted *AssetDepreciationLine) {
				if posted.ID != 3 {
					t.Fatalf("posted line id = %d, want 3", posted.ID)
				}
			},
		},
		{
			name: "PropagatesPosterError",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{{Base: model.Base{ID: 1}, Sequence: 1, DepreciationDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Amount: 100}}, nil
				}
				ctx.poster.PostTxFunc = func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "PropagatesUpdateError",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{{Base: model.Base{ID: 1}, Sequence: 1, DepreciationDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Amount: 100}}, nil
				}
				ctx.lines.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, _ *AssetDepreciationLine) (*AssetDepreciationLine, error) {
					return nil, errors.New("db down")
				}
			},
			wantMsg: "db down",
		},
		{
			name: "PostsWithMockFallbacks",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				unposted := &AssetDepreciationLine{Base: model.Base{ID: 1}, Sequence: 1, DepreciationDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Amount: 100}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{unposted}, nil
				}
			},
			assert: func(t *testing.T, posted *AssetDepreciationLine) {
				if !posted.Posted {
					t.Fatalf("expected line to be posted")
				}
			},
		},
		{
			name: "UsesMockListFallback",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
			},
			wantErr: ErrAssetNothingToPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newAssetTestContext()
			tt.setup(ctx)
			posted, err := ctx.svc.PostDepreciation(context.Background(), PostDepreciationRequest{AssetID: 1, JournalID: 9, Date: time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)})
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if tt.wantMsg != "" {
				helper.AssertError(t, err, true, nil)
				if err.Error() != tt.wantMsg {
					t.Fatalf("error = %v, want %v", err, tt.wantMsg)
				}
				return
			}
			helper.AssertError(t, err, false, nil)
			if tt.assert != nil {
				tt.assert(t, posted)
			}
		})
	}
}

func TestAssetService_Dispose_Extra(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ctx *assetTestContext, posted *accounting.PostRequest)
		request DisposalRequest
		wantErr error
		wantMsg string
		assert  func(t *testing.T, disposed *FixedAsset, posted *accounting.PostRequest)
	}{
		{
			name: "PropagatesAssetLookupError",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return nil, errors.New("db down")
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateDisposed},
			wantMsg: "db down",
		},
		{
			name: "RejectsMissingAsset",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return nil, nil
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateDisposed},
			wantErr: ErrAssetNotFound,
		},
		{
			name: "RejectsZeroDate",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateDisposed},
			wantErr: ErrAssetInvalidDates,
		},
		{
			name: "RejectsSoldWithoutProceedsAccount",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateSold, Date: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)},
			wantErr: ErrAssetDisposalProceeds,
		},
		{
			name: "RejectsNegativeProceeds",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
			},
			request: func() DisposalRequest {
				account := uint64(1110)
				return DisposalRequest{AssetID: 1, State: AssetStateSold, Date: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), ProceedsAmount: -1, ProceedsAccountID: &account}
			}(),
			wantErr: ErrAssetDisposalProceeds,
		},
		{
			name: "PropagatesCategoryError",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return nil, errors.New("db down")
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateDisposed, Date: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)},
			wantMsg: "db down",
		},
		{
			name: "PropagatesAccumulatedError",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return nil, errors.New("db down")
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateDisposed, Date: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)},
			wantMsg: "db down",
		},
		{
			name: "PropagatesPosterError",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
				ctx.poster.PostTxFunc = func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return nil, errors.New("db down")
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateDisposed, Date: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)},
			wantMsg: "db down",
		},
		{
			name: "PropagatesUpdateError",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.assets.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, _ *FixedAsset) (*FixedAsset, error) {
					return nil, errors.New("db down")
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateDisposed, Date: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)},
			wantMsg: "db down",
		},
		{
			name: "ClampsNegativeNetBookValue",
			setup: func(ctx *assetTestContext, posted *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.assets.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, updated *FixedAsset) (*FixedAsset, error) {
					return updated, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{{Base: model.Base{ID: 1}, Amount: 20000000, Posted: true}}, nil
				}
				ctx.poster.PostTxFunc = func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					*posted = request
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateDisposed, Date: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)},
			assert: func(t *testing.T, disposed *FixedAsset, posted *accounting.PostRequest) {
				if disposed.State != AssetStateDisposed {
					t.Fatalf("state = %s, want %s", disposed.State, AssetStateDisposed)
				}
				if len(posted.Lines) != 2 {
					t.Fatalf("posting lines = %d, want 2", len(posted.Lines))
				}
			},
		},
		{
			name: "UsesMockUpdateFallback",
			setup: func(ctx *assetTestContext, _ *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return runningAsset(), nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
			},
			request: DisposalRequest{AssetID: 1, State: AssetStateDisposed, Date: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)},
			assert: func(t *testing.T, disposed *FixedAsset, _ *accounting.PostRequest) {
				if disposed.State != AssetStateDisposed {
					t.Fatalf("state = %s, want %s", disposed.State, AssetStateDisposed)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newAssetTestContext()
			var posted accounting.PostRequest
			tt.setup(ctx, &posted)
			disposed, err := ctx.svc.Dispose(context.Background(), tt.request)
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if tt.wantMsg != "" {
				helper.AssertError(t, err, true, nil)
				if err.Error() != tt.wantMsg {
					t.Fatalf("error = %v, want %v", err, tt.wantMsg)
				}
				return
			}
			helper.AssertError(t, err, false, nil)
			if tt.assert != nil {
				tt.assert(t, disposed, &posted)
			}
		})
	}
}

func registerRequest() RegisterAssetRequest {
	return RegisterAssetRequest{
		OrganizationID:  1,
		Name:            "Laptop",
		CategoryID:      1,
		PurchaseValue:   10000000,
		SalvageValue:    1000000,
		AcquisitionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InServiceDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InvoiceLineID:   1,
	}
}

func stubRegisterDependencies(ctx *assetTestContext) {
	ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
		return assetCategoryFixture(), nil
	}
	ctx.invoiceLines.FindFunc = func(_ context.Context, _ uint64) (*accounting.InvoiceLine, error) {
		return &accounting.InvoiceLine{Base: model.Base{ID: 1}, InvoiceID: 9}, nil
	}
	ctx.invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
		moveID := uint64(30)
		return &accounting.Invoice{Base: model.Base{ID: 9}, Type: accounting.InvoiceTypeSupplierBill, State: accounting.InvoiceStatePosted, EntryID: &moveID}, nil
	}
}

func runningAsset() *FixedAsset {
	return &FixedAsset{
		Base:           model.Base{ID: 1},
		OrganizationID: 1,
		Name:           "Server",
		CategoryID:     1,
		PurchaseValue:  10000000,
		SalvageValue:   1000000,
		State:          AssetStateRunning,
		InServiceDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
	}
}

func intPtr(value int) *int {
	return &value
}
