package asset

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/state"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AssetCategoryService struct {
	categories dao.CRUD[reference.AssetCategory]
}

func NewAssetCategoryService(categories dao.CRUD[reference.AssetCategory]) AssetCategoryService {
	return AssetCategoryService{categories: categories}
}

func (s AssetCategoryService) List(ctx context.Context, q *query.Query) (*query.Page[reference.AssetCategory], error) {
	return s.categories.List(ctx, q)
}

func (s AssetCategoryService) Find(ctx context.Context, id uint64) (*reference.AssetCategory, error) {
	return s.categories.Find(ctx, id)
}

func (s AssetCategoryService) Create(ctx context.Context, category *reference.AssetCategory) (*reference.AssetCategory, error) {
	return s.categories.Create(ctx, category)
}

func (s AssetCategoryService) Update(ctx context.Context, category *reference.AssetCategory) (*reference.AssetCategory, error) {
	return s.categories.Update(ctx, category)
}

func (s AssetCategoryService) Delete(ctx context.Context, id uint64) error {
	return s.categories.Delete(ctx, id)
}

type AssetService struct {
	assets       FixedAssetDAO
	lines        AssetDepreciationLineDAO
	categories   AssetCategoryLookup
	invoiceLines InvoiceLineLookup
	invoices     InvoiceLookup
	poster       accounting.Poster
	tx           db.Transactioner
	machine      state.Machine
}

func NewAssetService(
	assets FixedAssetDAO,
	lines AssetDepreciationLineDAO,
	categories AssetCategoryLookup,
	invoiceLines InvoiceLineLookup,
	invoices InvoiceLookup,
	poster accounting.Poster,
	tx db.Transactioner,
) AssetService {
	return AssetService{
		assets:       assets,
		lines:        lines,
		categories:   categories,
		invoiceLines: invoiceLines,
		invoices:     invoices,
		poster:       poster,
		tx:           tx,
		machine: state.NewMachine(
			state.Transition{From: model.Status(AssetStateRunning), To: model.Status(AssetStateDisposed)},
			state.Transition{From: model.Status(AssetStateRunning), To: model.Status(AssetStateSold)},
		),
	}
}

func (s AssetService) List(ctx context.Context, q *query.Query) (*query.Page[FixedAsset], error) {
	return s.assets.List(ctx, q)
}

func (s AssetService) Find(ctx context.Context, id uint64) (*FixedAsset, error) {
	return s.assets.Find(ctx, id)
}

func (s AssetService) category(ctx context.Context, id uint64) (*reference.AssetCategory, error) {
	category, err := s.categories.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrAssetCategoryNotFound
	}
	if category.Method == nil || category.MethodNumber == nil || category.MethodPeriod == nil {
		return nil, ErrAssetInvalidMethod
	}
	switch *category.Method {
	case MethodLinear, MethodDeclining, MethodDecliningThenLinear:
	default:
		return nil, ErrAssetInvalidMethod
	}
	if *category.MethodNumber <= 0 {
		return nil, ErrAssetInvalidPeriods
	}
	if category.AssetAccountID == nil || category.DepreciationAccountID == nil || category.ExpenseAccountID == nil {
		return nil, ErrAssetCategoryAccounts
	}
	if category.GainAccountID == nil || category.LossAccountID == nil {
		return nil, ErrAssetCategoryAccounts
	}
	return category, nil
}

type RegisterAssetRequest struct {
	OrganizationID  uint64
	Name            string
	CategoryID      uint64
	PurchaseValue   float64
	SalvageValue    float64
	AcquisitionDate time.Time
	InServiceDate   time.Time
	InvoiceLineID   uint64
}

func (s AssetService) Register(ctx context.Context, request RegisterAssetRequest) (*FixedAsset, error) {
	if request.Name == "" {
		return nil, ErrAssetNameRequired
	}
	if request.PurchaseValue < 0 {
		return nil, ErrAssetInvalidValues
	}
	if request.SalvageValue < 0 || request.SalvageValue >= request.PurchaseValue {
		return nil, ErrAssetInvalidValues
	}
	if request.AcquisitionDate.IsZero() || request.InServiceDate.IsZero() {
		return nil, ErrAssetInvalidDates
	}
	if request.InServiceDate.Before(request.AcquisitionDate) {
		return nil, ErrAssetInvalidDates
	}

	if _, err := s.category(ctx, request.CategoryID); err != nil {
		return nil, err
	}

	line, err := s.invoiceLines.Find(ctx, request.InvoiceLineID)
	if err != nil {
		return nil, err
	}
	if line == nil {
		return nil, ErrAssetInvoiceLineNotFound
	}
	invoice, err := s.invoices.Find(ctx, line.InvoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil || invoice.Type != accounting.InvoiceTypeSupplierBill || invoice.State != accounting.InvoiceStatePosted {
		return nil, ErrAssetInvoiceNotSupplierBill
	}

	acquisition := request.AcquisitionDate
	inService := request.InServiceDate
	return s.assets.Create(ctx, &FixedAsset{
		OrganizationID:  request.OrganizationID,
		Name:            request.Name,
		CategoryID:      request.CategoryID,
		PurchaseValue:   request.PurchaseValue,
		SalvageValue:    request.SalvageValue,
		AcquisitionDate: &acquisition,
		InServiceDate:   &inService,
		OriginalEntryID: invoice.EntryID,
		InvoiceLineID:   &request.InvoiceLineID,
		State:           AssetStateRunning,
	})
}

func (s AssetService) generateSchedule(category *reference.AssetCategory, asset *FixedAsset) ([]AssetDepreciationLine, error) {
	purchase := amount.FromFloat64(asset.PurchaseValue)
	salvage := amount.FromFloat64(asset.SalvageValue)
	depreciable := purchase.Sub(salvage)
	periods := *category.MethodNumber
	periodUnit := *category.MethodPeriod

	var lines []AssetDepreciationLine
	accumulated := amount.Zero()
	remaining := depreciable

	for i := 0; i < periods; i++ {
		date := addPeriod(*asset.InServiceDate, i, periodUnit)
		var lineAmount amount.Amount
		switch *category.Method {
		case MethodLinear:
			var divErr error
			lineAmount, divErr = depreciable.Div(amount.FromInt64(int64(periods)))
			if divErr != nil {
				return nil, divErr
			}
		case MethodDeclining:
			lineAmount = decliningAmount(purchase, salvage, remaining, periods, i)
		case MethodDecliningThenLinear:
			lineAmount = decliningThenLinearAmount(purchase, salvage, remaining, periods, i)
		}
		lineAmount = lineAmount.Round(4)
		if i == periods-1 {
			lineAmount = remaining.Round(4)
		}
		accumulated = accumulated.Add(lineAmount)
		remaining = remaining.Sub(lineAmount)
		if remaining.IsNegative() {
			remaining = amount.Zero()
		}
		lines = append(lines, AssetDepreciationLine{
			Sequence:         i + 1,
			DepreciationDate: date,
			Amount:           lineAmount.Float64(),
			Accumulated:      accumulated.Float64(),
			RemainingValue:   remaining.Float64(),
		})
	}
	return lines, nil
}

func decliningAmount(purchase, salvage, remaining amount.Amount, periods, index int) amount.Amount {
	rate := decliningRate(purchase, salvage, periods)
	return remaining.Mul(rate).Round(4)
}

func decliningThenLinearAmount(purchase, salvage, remaining amount.Amount, periods, index int) amount.Amount {
	declining := decliningAmount(purchase, salvage, remaining, periods, index)
	remainingPeriods := amount.FromInt64(int64(periods - index))
	linear, err := remaining.Div(remainingPeriods)
	if err != nil {
		return amount.Zero()
	}
	if linear.GreaterThan(declining) {
		return linear
	}
	return declining
}

func decliningRate(purchase, salvage amount.Amount, periods int) amount.Amount {
	exponent := 1 / float64(periods)
	if salvage.IsPositive() {
		ratio, err := salvage.Div(purchase)
		if err != nil {
			return amount.Zero()
		}
		ratioDecimal := decimal.Decimal(ratio)
		rate := ratioDecimal.Pow(decimal.NewFromFloat(exponent))
		return amount.FromFloat64(1 - rate.InexactFloat64())
	}
	return amount.FromFloat64(2 / float64(periods))
}

func addPeriod(date time.Time, count int, unit string) time.Time {
	switch unit {
	case "year":
		return date.AddDate(count, 0, 0)
	default:
		return date.AddDate(0, count, 0)
	}
}

type GenerateScheduleRequest struct {
	AssetID   uint64
	JournalID uint64
	Date      time.Time
}

func (s AssetService) GenerateSchedule(ctx context.Context, request GenerateScheduleRequest) ([]AssetDepreciationLine, error) {
	asset, err := s.assets.Find(ctx, request.AssetID)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}
	if asset.State != AssetStateRunning {
		return nil, ErrAssetNotRunning
	}
	category, err := s.category(ctx, asset.CategoryID)
	if err != nil {
		return nil, err
	}

	schedule, err := s.generateSchedule(category, asset)
	if err != nil {
		return nil, err
	}

	var created []AssetDepreciationLine
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		existing, err := s.lines.ListByAsset(ctx, asset.ID)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return ErrAssetScheduleExists
		}
		for i := range schedule {
			line := schedule[i]
			line.AssetID = asset.ID
			createdLine, err := s.lines.CreateTx(ctx, tx, &line)
			if err != nil {
				return err
			}
			created = append(created, *createdLine)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s AssetService) nextUnpostedLine(ctx context.Context, assetID uint64, asOf time.Time) (*AssetDepreciationLine, error) {
	lines, err := s.lines.ListByAsset(ctx, assetID)
	if err != nil {
		return nil, err
	}
	var candidate *AssetDepreciationLine
	for _, line := range lines {
		if line.Posted {
			continue
		}
		if !asOf.IsZero() && line.DepreciationDate.After(asOf) {
			continue
		}
		if candidate == nil || line.Sequence < candidate.Sequence {
			candidate = line
		}
	}
	return candidate, nil
}

type PostDepreciationRequest struct {
	AssetID   uint64
	JournalID uint64
	Date      time.Time
}

func (s AssetService) PostDepreciation(ctx context.Context, request PostDepreciationRequest) (*AssetDepreciationLine, error) {
	asset, err := s.assets.Find(ctx, request.AssetID)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}
	if asset.State != AssetStateRunning {
		return nil, ErrAssetNotRunning
	}
	category, err := s.category(ctx, asset.CategoryID)
	if err != nil {
		return nil, err
	}

	var posted *AssetDepreciationLine
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		line, err := s.nextUnpostedLine(ctx, asset.ID, request.Date)
		if err != nil {
			return err
		}
		if line == nil {
			return ErrAssetNothingToPost
		}
		journal, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
			OrganizationID: asset.OrganizationID,
			JournalID:      request.JournalID,
			Date:           line.DepreciationDate,
			Ref:            fmt.Sprintf("DEP/%d", asset.ID),
			OriginType:     OriginDepreciation,
			OriginID:       line.ID,
			Description:    fmt.Sprintf("Depreciation %s", asset.Name),
			Lines: []accounting.PostingLine{
				{AccountID: *category.ExpenseAccountID, Name: "Depreciation Expense", Debit: amount.FromFloat64(line.Amount)},
				{AccountID: *category.DepreciationAccountID, Name: "Accumulated Depreciation", Credit: amount.FromFloat64(line.Amount)},
			},
		})
		if err != nil {
			return err
		}
		line.Posted = true
		line.EntryID = &journal.ID
		updated, err := s.lines.UpdateTx(ctx, tx, line)
		if err != nil {
			return err
		}
		posted = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return posted, nil
}

func (s AssetService) accumulatedDepreciation(ctx context.Context, assetID uint64) (amount.Amount, error) {
	lines, err := s.lines.ListByAsset(ctx, assetID)
	if err != nil {
		return amount.Zero(), err
	}
	total := amount.Zero()
	for _, line := range lines {
		if line.Posted {
			total = total.Add(amount.FromFloat64(line.Amount))
		}
	}
	return total, nil
}

type DisposalRequest struct {
	AssetID           uint64
	JournalID         uint64
	Date              time.Time
	State             string
	ProceedsAmount    float64
	ProceedsAccountID *uint64
}

func (s AssetService) Dispose(ctx context.Context, request DisposalRequest) (*FixedAsset, error) {
	asset, err := s.assets.Find(ctx, request.AssetID)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}
	if err := s.machine.TryTransition(model.Status(asset.State), model.Status(request.State)); err != nil {
		return nil, ErrAssetInvalidState
	}
	if request.Date.IsZero() {
		return nil, ErrAssetInvalidDates
	}
	if request.State == AssetStateSold && (request.ProceedsAccountID == nil || request.ProceedsAmount < 0) {
		return nil, ErrAssetDisposalProceeds
	}
	category, err := s.category(ctx, asset.CategoryID)
	if err != nil {
		return nil, err
	}

	proceeds := amount.Zero()
	if request.State == AssetStateSold {
		proceeds = amount.FromFloat64(request.ProceedsAmount)
	}
	accumulated, err := s.accumulatedDepreciation(ctx, asset.ID)
	if err != nil {
		return nil, err
	}
	nbv := amount.FromFloat64(asset.PurchaseValue).Sub(accumulated)
	if nbv.IsNegative() {
		nbv = amount.Zero()
	}

	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		_, err := s.poster.PostTx(ctx, tx, s.disposalPosting(asset, category, proceeds, nbv, request))
		if err != nil {
			return err
		}
		asset.State = request.State
		disposalDate := request.Date
		asset.DisposalDate = &disposalDate
		_, err = s.assets.UpdateTx(ctx, tx, asset)
		return err
	})
	if err != nil {
		return nil, err
	}
	return asset, nil
}

func (s AssetService) disposalPosting(asset *FixedAsset, category *reference.AssetCategory, proceeds, nbv amount.Amount, request DisposalRequest) accounting.PostRequest {
	accumulated := amount.FromFloat64(asset.PurchaseValue).Sub(nbv)
	lines := []accounting.PostingLine{
		{AccountID: *category.AssetAccountID, Name: "Fixed Asset Removal", Credit: amount.FromFloat64(asset.PurchaseValue)},
		{AccountID: *category.DepreciationAccountID, Name: "Accumulated Depreciation", Debit: accumulated},
	}
	if request.State == AssetStateSold && proceeds.IsPositive() {
		lines = append(lines, accounting.PostingLine{AccountID: *request.ProceedsAccountID, Name: "Asset Proceeds", Debit: proceeds})
	}
	if nbv.GreaterThan(proceeds) {
		loss := nbv.Sub(proceeds)
		lines = append(lines, accounting.PostingLine{AccountID: *category.LossAccountID, Name: "Loss on Disposal", Debit: loss})
	} else if proceeds.GreaterThan(nbv) {
		gain := proceeds.Sub(nbv)
		lines = append(lines, accounting.PostingLine{AccountID: *category.GainAccountID, Name: "Gain on Disposal", Credit: gain})
	}
	return accounting.PostRequest{
		OrganizationID: asset.OrganizationID,
		JournalID:      request.JournalID,
		Date:           request.Date,
		Ref:            fmt.Sprintf("DISP/%d", asset.ID),
		OriginType:     OriginDisposal,
		OriginID:       asset.ID,
		Description:    fmt.Sprintf("Disposal %s", asset.Name),
		Lines:          lines,
	}
}
