package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type InboundCostService struct {
	costs           InboundCostDAO
	lines           InboundCostLineDAO
	adjustments     InboundCostAdjustmentDAO
	movements       StockMovementDAO
	layers          CostLayerDAO
	resolver        ItemResolver
	poster          Poster
	configs         InboundCostConfigSource
	vendorBillLines SupplierBillLineLookup
	tx              db.Transactioner
	now             func() time.Time
}

type SupplierBillLineLookup interface {
	Find(ctx context.Context, id uint64) (*accounting.InvoiceLine, error)
}

func NewInboundCostService(
	costs InboundCostDAO,
	lines InboundCostLineDAO,
	adjustments InboundCostAdjustmentDAO,
	movements StockMovementDAO,
	layers CostLayerDAO,
	resolver ItemResolver,
	poster Poster,
	configs InboundCostConfigSource,
	vendorBillLines SupplierBillLineLookup,
	tx db.Transactioner,
) InboundCostService {
	return InboundCostService{
		costs:           costs,
		lines:           lines,
		adjustments:     adjustments,
		movements:       movements,
		layers:          layers,
		resolver:        resolver,
		poster:          poster,
		configs:         configs,
		vendorBillLines: vendorBillLines,
		tx:              tx,
		now:             time.Now,
	}
}

type CreateInboundCostLineRequest struct {
	ItemID             uint64
	Description        string
	Amount             float64
	SupplierBillLineID *uint64
	SplitMethod        string
	AccountID          *uint64
}

type CreateInboundCostRequest struct {
	OrganizationID    uint64
	Name              string
	Date              *time.Time
	TargetShipmentIDs helper.Int64Array
	Lines             []CreateInboundCostLineRequest
}

func (s InboundCostService) Create(ctx context.Context, request CreateInboundCostRequest) (*InboundCost, error) {
	if len(request.Lines) == 0 {
		return nil, ErrInboundCostNoLines
	}
	if len(request.TargetShipmentIDs) == 0 {
		return nil, ErrInboundCostNoShipments
	}
	cost := &InboundCost{
		OrganizationID:    &request.OrganizationID,
		Name:              request.Name,
		Date:              request.Date,
		State:             InboundCostStateDraft,
		TargetShipmentIDs: request.TargetShipmentIDs,
	}
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		created, err := s.costs.CreateTx(ctx, tx, cost)
		if err != nil {
			return err
		}
		cost = created
		for _, lineRequest := range request.Lines {
			line := &InboundCostLine{
				InboundCostID:      cost.ID,
				ItemID:             lineRequest.ItemID,
				Description:        &lineRequest.Description,
				Amount:             lineRequest.Amount,
				SupplierBillLineID: lineRequest.SupplierBillLineID,
				SplitMethod:        lineRequest.SplitMethod,
				AccountID:          lineRequest.AccountID,
			}
			if _, err := s.lines.CreateTx(ctx, tx, line); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cost, nil
}

func (s InboundCostService) Get(ctx context.Context, organizationID, costID uint64) (*InboundCost, error) {
	return s.find(ctx, organizationID, costID)
}

func (s InboundCostService) List(ctx context.Context, organizationID uint64) ([]*InboundCost, error) {
	page, err := s.costs.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s InboundCostService) ListLines(ctx context.Context, organizationID, costID uint64) ([]*InboundCostLine, error) {
	if _, err := s.find(ctx, organizationID, costID); err != nil {
		return nil, err
	}
	return s.lines.ListByInboundCost(ctx, costID)
}

func (s InboundCostService) ListAdjustments(ctx context.Context, organizationID, costID uint64) ([]*InboundCostAdjustment, error) {
	if _, err := s.find(ctx, organizationID, costID); err != nil {
		return nil, err
	}
	return s.adjustments.ListByInboundCost(ctx, costID)
}

func (s InboundCostService) Post(ctx context.Context, organizationID, costID uint64) (*InboundCost, error) {
	cost, err := s.find(ctx, organizationID, costID)
	if err != nil {
		return nil, err
	}
	if cost.State != InboundCostStateDraft {
		return nil, ErrInboundCostState
	}
	lines, err := s.lines.ListByInboundCost(ctx, cost.ID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrInboundCostNoLines
	}
	if len(cost.TargetShipmentIDs) == 0 {
		return nil, ErrInboundCostNoShipments
	}
	journalID, err := s.configs.JournalID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if journalID == 0 {
		return nil, ErrInboundCostNoJournal
	}
	date := s.now().UTC()
	if cost.Date != nil {
		date = *cost.Date
	}

	productMoves := map[uint64][]*StockMovement{}
	for _, shipmentID := range cost.TargetShipmentIDs {
		movements, err := s.movements.ListByShipment(ctx, uint64(shipmentID))
		if err != nil {
			return nil, err
		}
		for _, movement := range movements {
			if movement.ItemID != 0 {
				productMoves[movement.ItemID] = append(productMoves[movement.ItemID], movement)
			}
		}
	}

	postLines := []accounting.PostingLine{}
	drByAccount := map[uint64]amount.Amount{}
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		for _, line := range lines {
			movements := productMoves[line.ItemID]
			if len(movements) == 0 {
				return ErrInboundCostNoMovements
			}
			if line.AccountID == nil && line.SupplierBillLineID != nil && s.vendorBillLines != nil {
				billLine, err := s.vendorBillLines.Find(ctx, *line.SupplierBillLineID)
				if err != nil {
					return err
				}
				if billLine != nil && billLine.AccountID != nil {
					line.AccountID = billLine.AccountID
				}
			}
			if line.AccountID == nil {
				return ErrInboundCostNoAccount
			}
			resolved, err := s.resolver.Resolve(ctx, line.ItemID)
			if err != nil {
				return err
			}
			if resolved.StockValuationAccountID == 0 {
				return ErrValuationAccount
			}

			layers := make([]*CostLayer, len(movements))
			weights := make([]amount.Amount, len(movements))
			totalBasis := amount.Zero()
			for i, movement := range movements {
				layer, err := s.openLayerForMove(ctx, movement)
				if err != nil {
					return err
				}
				layers[i] = layer
				basis, err := s.moveBasis(line.SplitMethod, movement, layer, resolved)
				if err != nil {
					return err
				}
				weights[i] = basis
				totalBasis = totalBasis.Add(basis)
			}
			if totalBasis.IsZero() {
				return ErrInboundCostSplit
			}

			lineAmount := amount.FromFloat64(line.Amount)
			assigned := amount.Zero()
			for i, movement := range movements {
				var additional amount.Amount
				if i == len(movements)-1 {
					additional = lineAmount.Sub(assigned)
				} else {
					ratio, err := weights[i].Div(totalBasis)
					if err != nil {
						return err
					}
					additional = lineAmount.Mul(ratio).Round(4)
					assigned = assigned.Add(additional)
				}
				if _, err := s.adjustments.CreateTx(ctx, tx, &InboundCostAdjustment{
					InboundCostID:   cost.ID,
					StockMovementID: movement.ID,
					ItemID:          line.ItemID,
					AdditionalCost:  additional.Float64(),
					CostLayerID:     layers[i].ID,
				}); err != nil {
					return err
				}
				layers[i].RemainingValue += additional.Float64()
				if _, err := s.layers.UpdateTx(ctx, tx, layers[i]); err != nil {
					return err
				}
				drByAccount[resolved.StockValuationAccountID] = drByAccount[resolved.StockValuationAccountID].Add(additional)
			}
			postLines = append(postLines, accounting.PostingLine{
				AccountID: *line.AccountID,
				Name:      inboundCostLineName(line),
				Credit:    lineAmount,
			})
		}
		for accountID, debit := range drByAccount {
			if debit.IsZero() {
				continue
			}
			postLines = append(postLines, accounting.PostingLine{AccountID: accountID, Name: "Landed cost", Debit: debit.Round(4)})
		}

		post, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
			OrganizationID: organizationID,
			JournalID:      journalID,
			Date:           date,
			Ref:            fmt.Sprintf("LND/%d", cost.ID),
			OriginType:     accounting.OriginTypeInboundCost,
			OriginID:       cost.ID,
			Description:    "Landed cost " + cost.Name,
			Lines:          postLines,
		})
		if err != nil {
			return err
		}
		cost.MovementID = &post.ID
		cost.State = InboundCostStatePosted
		updated, err := s.costs.UpdateTx(ctx, tx, cost)
		if err != nil {
			return err
		}
		cost = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cost, nil
}

func (s InboundCostService) openLayerForMove(ctx context.Context, movement *StockMovement) (*CostLayer, error) {
	layers, err := s.layers.ListByMovement(ctx, movement.ID)
	if err != nil {
		return nil, err
	}
	for _, layer := range layers {
		if layer.RemainingQty != 0 {
			return layer, nil
		}
	}
	return nil, ErrInboundCostValue
}

func (s InboundCostService) moveBasis(method string, movement *StockMovement, layer *CostLayer, resolved ResolvedItem) (amount.Amount, error) {
	switch method {
	case SplitMethodQuantity:
		return amount.FromFloat64(movement.Qty), nil
	case SplitMethodWeight:
		return amount.FromFloat64(movement.Qty).Mul(amount.FromFloat64(resolved.Weight)), nil
	case SplitMethodVolume:
		return amount.FromFloat64(movement.Qty).Mul(amount.FromFloat64(resolved.Volume)), nil
	case SplitMethodValue:
		if layer == nil {
			return amount.Zero(), ErrInboundCostValue
		}
		return amount.FromFloat64(layer.Value), nil
	case SplitMethodEqual:
		return amount.FromInt64(1), nil
	default:
		return amount.Zero(), ErrInboundCostSplit
	}
}

func inboundCostLineName(line *InboundCostLine) string {
	if line.Description == nil {
		return "Landed cost"
	}
	return *line.Description
}

func (s InboundCostService) find(ctx context.Context, organizationID, costID uint64) (*InboundCost, error) {
	cost, err := s.costs.Search(ctx, "id", costID)
	if err != nil {
		return nil, err
	}
	if cost == nil || cost.OrganizationID == nil || *cost.OrganizationID != organizationID {
		return nil, ErrInboundCostNotFound
	}
	return cost, nil
}
