package asset

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type assetTestContext struct {
	assets       FixedAssetDAOMock
	lines        AssetDepreciationLineDAOMock
	categories   assetCategoryLookupMock
	invoiceLines invoiceLineLookupMock
	invoices     invoiceLookupMock
	poster       inventory.PosterMock
	svc          AssetService
}

type assetCategoryLookupMock struct {
	FindFunc func(ctx context.Context, id uint64) (*reference.AssetCategory, error)
}

func (m assetCategoryLookupMock) Find(ctx context.Context, id uint64) (*reference.AssetCategory, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

type invoiceLineLookupMock struct {
	FindFunc func(ctx context.Context, id uint64) (*accounting.InvoiceLine, error)
}

func (m invoiceLineLookupMock) Find(ctx context.Context, id uint64) (*accounting.InvoiceLine, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

type invoiceLookupMock struct {
	FindFunc func(ctx context.Context, id uint64) (*accounting.Invoice, error)
}

func (m invoiceLookupMock) Find(ctx context.Context, id uint64) (*accounting.Invoice, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func assetCategoryFixture() *reference.AssetCategory {
	assetAccount := uint64(1500)
	depreciationAccount := uint64(1510)
	expenseAccount := uint64(6600)
	gainAccount := uint64(4500)
	lossAccount := uint64(6000)
	method := "linear"
	period := "month"
	periods := 5
	return &reference.AssetCategory{
		Base:                  model.Base{ID: 1},
		AssetAccountID:        &assetAccount,
		DepreciationAccountID: &depreciationAccount,
		ExpenseAccountID:      &expenseAccount,
		GainAccountID:         &gainAccount,
		LossAccountID:         &lossAccount,
		Method:                &method,
		MethodNumber:          &periods,
		MethodPeriod:          &period,
	}
}

func postedSupplierBillLine() (*accounting.InvoiceLine, *accounting.Invoice) {
	line := &accounting.InvoiceLine{Base: model.Base{ID: 1}, InvoiceID: 9}
	moveID := uint64(30)
	invoice := &accounting.Invoice{
		Base:    model.Base{ID: 9},
		Type:    accounting.InvoiceTypeSupplierBill,
		State:   accounting.InvoiceStatePosted,
		EntryID: &moveID,
	}
	return line, invoice
}

func newAssetTestContext() *assetTestContext {
	ctx := &assetTestContext{}
	ctx.svc = NewAssetService(
		&ctx.assets,
		&ctx.lines,
		&ctx.categories,
		&ctx.invoiceLines,
		&ctx.invoices,
		&ctx.poster,
		inventory.TransactionerMock{},
	)
	return ctx
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func TestAssetService_Register(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ctx *assetTestContext)
		request   RegisterAssetRequest
		wantErr   bool
		wantErrFn func(error) bool
		assert    func(t *testing.T, created *FixedAsset)
	}{
		{
			name: "RegistersFromPostedSupplierBill",
			setup: func(ctx *assetTestContext) {
				line, invoice := postedSupplierBillLine()
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.invoiceLines.FindFunc = func(_ context.Context, _ uint64) (*accounting.InvoiceLine, error) {
					return line, nil
				}
				ctx.invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
					return invoice, nil
				}
				ctx.assets.CreateFunc = func(_ context.Context, a *FixedAsset) (*FixedAsset, error) {
					return a, nil
				}
			},
			request: RegisterAssetRequest{
				OrganizationID:  1,
				Name:            "Laptop",
				CategoryID:      1,
				PurchaseValue:   10000000,
				SalvageValue:    1000000,
				AcquisitionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				InServiceDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				InvoiceLineID:   1,
			},
			assert: func(t *testing.T, created *FixedAsset) {
				if created.State != AssetStateRunning {
					t.Fatalf("state = %s, want %s", created.State, AssetStateRunning)
				}
				if created.OriginalEntryID == nil || *created.OriginalEntryID != 30 {
					t.Fatalf("original movement id = %v, want 30", created.OriginalEntryID)
				}
			},
		},
		{
			name: "RejectsUnpostedInvoice",
			setup: func(ctx *assetTestContext) {
				line, _ := postedSupplierBillLine()
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.invoiceLines.FindFunc = func(_ context.Context, _ uint64) (*accounting.InvoiceLine, error) {
					return line, nil
				}
				ctx.invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
					return &accounting.Invoice{Base: model.Base{ID: 9}, Type: accounting.InvoiceTypeSupplierBill, State: accounting.InvoiceStateDraft}, nil
				}
			},
			request: RegisterAssetRequest{
				OrganizationID:  1,
				Name:            "Laptop",
				CategoryID:      1,
				PurchaseValue:   10000000,
				SalvageValue:    1000000,
				AcquisitionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				InServiceDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				InvoiceLineID:   1,
			},
			wantErr:   true,
			wantErrFn: func(err error) bool { return err == ErrAssetInvoiceNotSupplierBill },
		},
		{
			name:  "RejectsSalvageAtLeastPurchase",
			setup: func(ctx *assetTestContext) {},
			request: RegisterAssetRequest{
				OrganizationID:  1,
				Name:            "Laptop",
				CategoryID:      1,
				PurchaseValue:   10000000,
				SalvageValue:    10000000,
				AcquisitionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				InServiceDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				InvoiceLineID:   1,
			},
			wantErr:   true,
			wantErrFn: func(err error) bool { return err == ErrAssetInvalidValues },
		},
		{
			name:  "RejectsMissingName",
			setup: func(ctx *assetTestContext) {},
			request: RegisterAssetRequest{
				OrganizationID:  1,
				CategoryID:      1,
				PurchaseValue:   10000000,
				SalvageValue:    1000000,
				AcquisitionDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				InServiceDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				InvoiceLineID:   1,
			},
			wantErr:   true,
			wantErrFn: func(err error) bool { return err == ErrAssetNameRequired },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newAssetTestContext()
			tt.setup(ctx)
			created, err := ctx.svc.Register(context.Background(), tt.request)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				if tt.wantErrFn != nil && !tt.wantErrFn(err) {
					t.Fatalf("error = %v, does not match expected error", err)
				}
				return
			}
			if tt.assert != nil {
				tt.assert(t, created)
			}
		})
	}
}

func TestAssetService_GenerateSchedule(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ctx *assetTestContext)
		request   GenerateScheduleRequest
		wantErr   bool
		wantErrFn func(error) bool
		assert    func(t *testing.T, schedule []AssetDepreciationLine)
	}{
		{
			name: "LinearDepreciatesToSalvage",
			setup: func(ctx *assetTestContext) {
				assetRecord := &FixedAsset{
					Base:           model.Base{ID: 1},
					OrganizationID: 1,
					Name:           "Server",
					CategoryID:     1,
					PurchaseValue:  10000000,
					SalvageValue:   1000000,
					State:          AssetStateRunning,
					InServiceDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
				}
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return assetRecord, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
			},
			request: GenerateScheduleRequest{AssetID: 1},
			assert: func(t *testing.T, schedule []AssetDepreciationLine) {
				if len(schedule) != 5 {
					t.Fatalf("schedule lines = %d, want 5", len(schedule))
				}
				total := amount.Zero()
				for _, line := range schedule {
					total = total.Add(amount.FromFloat64(line.Amount))
				}
				want := amount.FromFloat64(9000000)
				if !total.Equal(want) {
					t.Fatalf("total depreciation = %v, want %v", total, want)
				}
				if last := schedule[len(schedule)-1]; last.Accumulated != 9000000 {
					t.Fatalf("final accumulated = %v, want 9000000", last.Accumulated)
				}
			},
		},
		{
			name: "RejectsScheduleExists",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return &FixedAsset{
						Base:           model.Base{ID: 1},
						OrganizationID: 1,
						Name:           "Server",
						CategoryID:     1,
						PurchaseValue:  10000000,
						SalvageValue:   1000000,
						State:          AssetStateRunning,
						InServiceDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
					}, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{{Base: model.Base{ID: 1}}}, nil
				}
			},
			request:   GenerateScheduleRequest{AssetID: 1},
			wantErr:   true,
			wantErrFn: func(err error) bool { return err == ErrAssetScheduleExists },
		},
		{
			name: "DecliningMethodEndsAtSalvage",
			setup: func(ctx *assetTestContext) {
				category := assetCategoryFixture()
				method := "declining"
				category.Method = &method
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return &FixedAsset{
						Base:           model.Base{ID: 1},
						OrganizationID: 1,
						Name:           "Server",
						CategoryID:     1,
						PurchaseValue:  10000000,
						SalvageValue:   1000000,
						State:          AssetStateRunning,
						InServiceDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
					}, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return category, nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
			},
			request: GenerateScheduleRequest{AssetID: 1},
			assert: func(t *testing.T, schedule []AssetDepreciationLine) {
				if len(schedule) != 5 {
					t.Fatalf("schedule lines = %d, want 5", len(schedule))
				}
				total := amount.Zero()
				for _, line := range schedule {
					total = total.Add(amount.FromFloat64(line.Amount))
				}
				if !total.Equal(amount.FromFloat64(9000000)) {
					t.Fatalf("total depreciation = %v, want 9000000", total)
				}
				if last := schedule[len(schedule)-1]; last.Accumulated != 9000000 {
					t.Fatalf("final accumulated = %v, want 9000000", last.Accumulated)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newAssetTestContext()
			tt.setup(ctx)
			schedule, err := ctx.svc.GenerateSchedule(context.Background(), tt.request)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				if tt.wantErrFn != nil && !tt.wantErrFn(err) {
					t.Fatalf("error = %v, does not match expected error", err)
				}
				return
			}
			if tt.assert != nil {
				tt.assert(t, schedule)
			}
		})
	}
}

func TestAssetService_PostDepreciation(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ctx *assetTestContext)
		request   PostDepreciationRequest
		wantErr   bool
		wantErrFn func(error) bool
		assert    func(t *testing.T, posted *AssetDepreciationLine)
	}{
		{
			name: "PostsNextUnpostedLine",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return &FixedAsset{
						Base:           model.Base{ID: 1},
						OrganizationID: 1,
						Name:           "Server",
						CategoryID:     1,
						PurchaseValue:  10000000,
						SalvageValue:   1000000,
						State:          AssetStateRunning,
					}, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				unposted := &AssetDepreciationLine{
					Base:             model.Base{ID: 1},
					AssetID:          1,
					Sequence:         1,
					DepreciationDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
					Amount:           1800000,
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{unposted}, nil
				}
				ctx.lines.UpdateTxFunc = func(_ context.Context, tx *gorm.DB, _ *AssetDepreciationLine) (*AssetDepreciationLine, error) {
					return unposted, nil
				}
			},
			request: PostDepreciationRequest{
				AssetID:   1,
				JournalID: 9,
				Date:      time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
			},
			assert: func(t *testing.T, posted *AssetDepreciationLine) {
				if !posted.Posted {
					t.Fatalf("expected line to be posted")
				}
				if posted.EntryID == nil || *posted.EntryID != 1 {
					t.Fatalf("movement id = %v, want 1", posted.EntryID)
				}
			},
		},
		{
			name: "NothingToPost",
			setup: func(ctx *assetTestContext) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return &FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: AssetStateRunning}, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{}, nil
				}
			},
			request: PostDepreciationRequest{
				AssetID:   1,
				JournalID: 9,
				Date:      time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
			},
			wantErr:   true,
			wantErrFn: func(err error) bool { return err == ErrAssetNothingToPost },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newAssetTestContext()
			tt.setup(ctx)
			posted, err := ctx.svc.PostDepreciation(context.Background(), tt.request)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				if tt.wantErrFn != nil && !tt.wantErrFn(err) {
					t.Fatalf("error = %v, does not match expected error", err)
				}
				return
			}
			if tt.assert != nil {
				tt.assert(t, posted)
			}
		})
	}
}

func TestAssetService_Dispose(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ctx *assetTestContext, posted *accounting.PostRequest)
		request   DisposalRequest
		wantErr   bool
		wantErrFn func(error) bool
		assert    func(t *testing.T, disposed *FixedAsset, posted *accounting.PostRequest)
	}{
		{
			name: "PostsRemovalAndLoss",
			setup: func(ctx *assetTestContext, posted *accounting.PostRequest) {
				assetRecord := &FixedAsset{
					Base:           model.Base{ID: 1},
					OrganizationID: 1,
					Name:           "Server",
					CategoryID:     1,
					PurchaseValue:  10000000,
					SalvageValue:   1000000,
					State:          AssetStateRunning,
				}
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return assetRecord, nil
				}
				ctx.assets.UpdateTxFunc = func(_ context.Context, tx *gorm.DB, updated *FixedAsset) (*FixedAsset, error) {
					*assetRecord = *updated
					return updated, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{
						{Base: model.Base{ID: 1}, Amount: 1800000, Posted: true},
						{Base: model.Base{ID: 2}, Amount: 1800000, Posted: true},
					}, nil
				}
				ctx.poster.PostTxFunc = func(_ context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					*posted = request
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}
			},
			request: DisposalRequest{
				AssetID:   1,
				JournalID: 9,
				Date:      time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
				State:     AssetStateDisposed,
			},
			assert: func(t *testing.T, disposed *FixedAsset, posted *accounting.PostRequest) {
				if disposed.State != AssetStateDisposed {
					t.Fatalf("state = %s, want %s", disposed.State, AssetStateDisposed)
				}
				debits := amount.Zero()
				credits := amount.Zero()
				for _, line := range posted.Lines {
					debits = debits.Add(line.Debit)
					credits = credits.Add(line.Credit)
				}
				if !amount.IsBalanced(debits, credits, 4) {
					t.Fatalf("expected balanced entry, debits=%v credits=%v", debits, credits)
				}
			},
		},
		{
			name: "SellPostsGain",
			setup: func(ctx *assetTestContext, posted *accounting.PostRequest) {
				assetRecord := &FixedAsset{
					Base:           model.Base{ID: 1},
					OrganizationID: 1,
					Name:           "Server",
					CategoryID:     1,
					PurchaseValue:  10000000,
					SalvageValue:   1000000,
					State:          AssetStateRunning,
				}
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return assetRecord, nil
				}
				ctx.assets.UpdateTxFunc = func(_ context.Context, tx *gorm.DB, updated *FixedAsset) (*FixedAsset, error) {
					*assetRecord = *updated
					return updated, nil
				}
				ctx.categories.FindFunc = func(_ context.Context, _ uint64) (*reference.AssetCategory, error) {
					return assetCategoryFixture(), nil
				}
				ctx.lines.ListByAssetFunc = func(_ context.Context, _ uint64) ([]*AssetDepreciationLine, error) {
					return []*AssetDepreciationLine{
						{Base: model.Base{ID: 1}, Amount: 3600000, Posted: true},
					}, nil
				}
				ctx.poster.PostTxFunc = func(_ context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					*posted = request
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}
			},
			request: func() DisposalRequest {
				proceedsAccount := uint64(1110)
				return DisposalRequest{
					AssetID:           1,
					JournalID:         9,
					Date:              time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
					State:             AssetStateSold,
					ProceedsAmount:    9000000,
					ProceedsAccountID: &proceedsAccount,
				}
			}(),
			assert: func(t *testing.T, disposed *FixedAsset, posted *accounting.PostRequest) {
				if disposed.State != AssetStateSold {
					t.Fatalf("state = %s, want %s", disposed.State, AssetStateSold)
				}
				debits := amount.Zero()
				credits := amount.Zero()
				for _, line := range posted.Lines {
					debits = debits.Add(line.Debit)
					credits = credits.Add(line.Credit)
				}
				if !amount.IsBalanced(debits, credits, 4) {
					t.Fatalf("expected balanced entry, debits=%v credits=%v", debits, credits)
				}
			},
		},
		{
			name: "RejectsWrongState",
			setup: func(ctx *assetTestContext, posted *accounting.PostRequest) {
				ctx.assets.FindFunc = func(_ context.Context, _ uint64) (*FixedAsset, error) {
					return &FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Server", CategoryID: 1, PurchaseValue: 10000000, SalvageValue: 1000000, State: AssetStateDraft}, nil
				}
			},
			request: DisposalRequest{
				AssetID:   1,
				JournalID: 9,
				Date:      time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
				State:     AssetStateDisposed,
			},
			wantErr:   true,
			wantErrFn: func(err error) bool { return err == ErrAssetInvalidState },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newAssetTestContext()
			var posted accounting.PostRequest
			tt.setup(ctx, &posted)
			disposed, err := ctx.svc.Dispose(context.Background(), tt.request)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				if tt.wantErrFn != nil && !tt.wantErrFn(err) {
					t.Fatalf("error = %v, does not match expected error", err)
				}
				return
			}
			if tt.assert != nil {
				tt.assert(t, disposed, &posted)
			}
		})
	}
}
