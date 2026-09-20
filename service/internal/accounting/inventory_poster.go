package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"gorm.io/gorm"
)

type StockValuation struct {
	ItemID             uint64
	Quantity           float64
	UnitCost           amount.Amount
	AccountID          uint64
	ValuationAccountID uint64
	CogsAccountID      *uint64
}

type InventoryPoster struct {
	poster Poster
}

func NewInventoryPoster(poster Poster) InventoryPoster {
	return InventoryPoster{poster: poster}
}

func (p InventoryPoster) PostStockMovement(ctx context.Context, tx *gorm.DB, req StockMovementPostRequest) (*JournalEntry, error) {
	if len(req.Lines) == 0 {
		return nil, ErrNoLines
	}
	postReq := PostRequest{
		OrganizationID: req.OrganizationID,
		JournalID:      req.JournalID,
		Date:           req.Date,
		Ref:            req.Ref,
		OriginType:     req.OriginType,
		OriginID:       req.OriginID,
		Description:    req.Description,
		Lines:          req.Lines,
	}
	if tx != nil {
		return p.poster.PostTx(ctx, tx, postReq)
	}
	return p.poster.Post(ctx, postReq)
}

func (p InventoryPoster) PostGoodsReceipt(ctx context.Context, organizationID, journalID uint64, date time.Time, ref, description string, stockAccount, inputAccount uint64, cost amount.Amount) (*JournalEntry, error) {
	return p.poster.Post(ctx, PostRequest{
		OrganizationID: organizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            ref,
		OriginType:     OriginTypeStockMovement,
		Description:    description,
		Lines: []PostingLine{
			{AccountID: stockAccount, Name: "Inventory", Debit: cost},
			{AccountID: inputAccount, Name: "Stock Input", Credit: cost},
		},
	})
}

func (p InventoryPoster) PostGoodsIssue(ctx context.Context, organizationID, journalID uint64, date time.Time, ref, description string, cogsAccount, stockAccount uint64, cost amount.Amount) (*JournalEntry, error) {
	return p.poster.Post(ctx, PostRequest{
		OrganizationID: organizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            ref,
		OriginType:     OriginTypeStockMovement,
		Description:    description,
		Lines: []PostingLine{
			{AccountID: cogsAccount, Name: "COGS", Debit: cost},
			{AccountID: stockAccount, Name: "Inventory", Credit: cost},
		},
	})
}

func (p InventoryPoster) PostInboundCost(ctx context.Context, organizationID, journalID uint64, date time.Time, ref string, stockAccount, varianceAccount uint64, cost amount.Amount) (*JournalEntry, error) {
	return p.poster.Post(ctx, PostRequest{
		OrganizationID: organizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            ref,
		OriginType:     OriginTypeInboundCost,
		Description:    "Landed cost allocation",
		Lines: []PostingLine{
			{AccountID: stockAccount, Name: "Inventory - Landed Cost", Debit: cost},
			{AccountID: varianceAccount, Name: "Landed Cost Variance", Credit: cost},
		},
	})
}

func (p InventoryPoster) PostScrap(ctx context.Context, organizationID, journalID uint64, date time.Time, ref string, lossAccount, stockAccount uint64, cost amount.Amount) (*JournalEntry, error) {
	return p.poster.Post(ctx, PostRequest{
		OrganizationID: organizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            ref,
		OriginType:     OriginTypeQualityScrap,
		Description:    "Quality scrap",
		Lines: []PostingLine{
			{AccountID: lossAccount, Name: "Scrap Loss", Debit: cost},
			{AccountID: stockAccount, Name: "Inventory", Credit: cost},
		},
	})
}

func (p InventoryPoster) PostStockCount(ctx context.Context, organizationID, journalID uint64, date time.Time, ref string, stockAccount, gainLossAccount uint64, diff amount.Amount) (*JournalEntry, error) {
	if diff.IsPositive() {
		return p.poster.Post(ctx, PostRequest{
			OrganizationID: organizationID,
			JournalID:      journalID,
			Date:           date,
			Ref:            ref,
			OriginType:     OriginTypeStockCount,
			Description:    "Inventory count surplus",
			Lines: []PostingLine{
				{AccountID: stockAccount, Name: "Inventory", Debit: diff},
				{AccountID: gainLossAccount, Name: "Inventory Gain", Credit: diff},
			},
		})
	}
	return p.poster.Post(ctx, PostRequest{
		OrganizationID: organizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            ref,
		OriginType:     OriginTypeStockCount,
		Description:    "Inventory count shortage",
		Lines: []PostingLine{
			{AccountID: gainLossAccount, Name: "Inventory Loss", Debit: diff.Abs()},
			{AccountID: stockAccount, Name: "Inventory", Credit: diff.Abs()},
		},
	})
}
