package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

const (
	CountStateDraft  = "draft"
	CountStatePosted = "posted"
)

type StockCountLineRequest struct {
	ItemID     uint64  `json:"item_id"`
	BatchID    *uint64 `json:"batch_id"`
	CountedQty float64 `json:"counted_qty"`
}

type StockCountService struct {
	counts   StockCountDAO
	lines    StockCountLineDAO
	quants   StockBalanceDAO
	layers   CostLayerDAO
	ledger   LedgerService
	resolver ItemResolver
	poster   Poster
	tx       db.Transactioner
}

func NewStockCountService(
	counts StockCountDAO,
	lines StockCountLineDAO,
	quants StockBalanceDAO,
	layers CostLayerDAO,
	ledger LedgerService,
	resolver ItemResolver,
	poster Poster,
	tx db.Transactioner,
) StockCountService {
	return StockCountService{
		counts:   counts,
		lines:    lines,
		quants:   quants,
		layers:   layers,
		ledger:   ledger,
		resolver: resolver,
		poster:   poster,
		tx:       tx,
	}
}

func (s StockCountService) List(ctx context.Context, q *query.Query) (*query.Page[StockCount], error) {
	return s.counts.List(ctx, q)
}

func (s StockCountService) Find(ctx context.Context, id uint64) (*StockCount, error) {
	return s.counts.Find(ctx, id)
}

func (s StockCountService) Delete(ctx context.Context, id uint64) error {
	return s.counts.Delete(ctx, id)
}

func (s StockCountService) ListLines(ctx context.Context, stockCountID uint64) ([]*StockCountLine, error) {
	return s.lines.ListByCount(ctx, stockCountID)
}

func (s StockCountService) Create(ctx context.Context, count *StockCount, requests []StockCountLineRequest) (*StockCount, error) {
	if count.LocationID == nil {
		return nil, ErrLocationRequired
	}
	if count.State == "" {
		count.State = CountStateDraft
	}

	lineModels := make([]*StockCountLine, len(requests))
	for i, request := range requests {
		theoretical, err := s.ledger.OnHand(ctx, count.OrganizationID, request.ItemID, *count.LocationID)
		if err != nil {
			return nil, err
		}
		lineModels[i] = &StockCountLine{
			ItemID:         request.ItemID,
			BatchID:        request.BatchID,
			TheoreticalQty: theoretical,
			CountedQty:     request.CountedQty,
			DiffQty:        request.CountedQty - theoretical,
		}
	}
	return s.counts.CreateWithLines(ctx, count, lineModels)
}

func (s StockCountService) Post(ctx context.Context, stockCountID uint64, journalID, gainLossAccountID uint64, date time.Time) (*StockCount, error) {
	count, err := s.counts.Find(ctx, stockCountID)
	if err != nil {
		return nil, err
	}
	if count == nil {
		return nil, ErrStockCountNotFound
	}
	if count.State != CountStateDraft {
		return nil, ErrStockCountState
	}
	if gainLossAccountID == 0 {
		return nil, ErrGainLossAccount
	}
	if count.OrganizationID == nil {
		return nil, ErrOrganizationMissing
	}

	lines, err := s.lines.ListByCount(ctx, stockCountID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrStockCountNoLines
	}

	if err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		for _, line := range lines {
			diff := amount.FromFloat64(line.DiffQty)
			if diff.IsZero() {
				continue
			}
			unitCost, err := s.currentUnitCost(ctx, line.ItemID)
			if err != nil {
				return err
			}
			if _, err := s.quants.UpsertTx(ctx, tx, count.OrganizationID, line.ItemID, *count.LocationID, line.BatchID, diff.Float64()); err != nil {
				return err
			}

			value := diff.Mul(unitCost).Abs().Round(4)
			if value.IsZero() {
				continue
			}

			resolved, err := s.resolver.Resolve(ctx, line.ItemID)
			if err != nil {
				return err
			}
			if resolved.StockValuationAccountID == 0 {
				return ErrValuationAccount
			}

			var lines []accounting.PostingLine
			if diff.IsPositive() {
				lines = []accounting.PostingLine{
					{AccountID: resolved.StockValuationAccountID, Name: "Inventory", Debit: value},
					{AccountID: gainLossAccountID, Name: "Inventory Gain", Credit: value},
				}
			} else {
				lines = []accounting.PostingLine{
					{AccountID: gainLossAccountID, Name: "Inventory Loss", Debit: value},
					{AccountID: resolved.StockValuationAccountID, Name: "Inventory", Credit: value},
				}
			}
			if _, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
				OrganizationID: *count.OrganizationID,
				JournalID:      journalID,
				Date:           date,
				Ref:            fmt.Sprintf("COUNT/%d", count.ID),
				OriginType:     accounting.OriginTypeStockCount,
				OriginID:       count.ID,
				Description:    "Inventory count adjustment",
				Lines:          lines,
			}); err != nil {
				return err
			}
		}

		count.State = CountStatePosted
		_, err := s.counts.UpdateTx(ctx, tx, count)
		return err
	}); err != nil {
		return nil, err
	}
	return count, nil
}

func (s StockCountService) currentUnitCost(ctx context.Context, itemID uint64) (amount.Amount, error) {
	openLayers, err := s.layers.ListOpenByItem(ctx, itemID)
	if err != nil {
		return amount.Amount{}, err
	}
	return currentUnitCost(openLayers), nil
}
