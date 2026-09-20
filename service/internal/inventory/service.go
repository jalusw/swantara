package inventory

import (
	"context"
	"strconv"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type WarehouseService struct {
	warehouses WarehouseDAO
	locations  StockLocationDAO
}

var validLocationUsage = map[string]bool{
	"internal": true, "customer": true, "supplier": true, "production": true,
	"inventory": true, "transit": true, "scrap": true, "view": true,
}

func NewWarehouseService(warehouses WarehouseDAO, locations StockLocationDAO) WarehouseService {
	return WarehouseService{warehouses: warehouses, locations: locations}
}

func (s WarehouseService) ListWarehouses(ctx context.Context, q *query.Query) (*query.Page[reference.Warehouse], error) {
	return s.warehouses.List(ctx, q)
}

func (s WarehouseService) FindWarehouse(ctx context.Context, id uint64) (*reference.Warehouse, error) {
	return s.warehouses.Find(ctx, id)
}

func (s WarehouseService) DeleteWarehouse(ctx context.Context, id uint64) error {
	return s.warehouses.Delete(ctx, id)
}

func (s WarehouseService) ListLocations(ctx context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
	return s.locations.List(ctx, q)
}

func (s WarehouseService) FindLocation(ctx context.Context, id uint64) (*reference.StockLocation, error) {
	return s.locations.Find(ctx, id)
}

func (s WarehouseService) DeleteLocation(ctx context.Context, id uint64) error {
	return s.locations.Delete(ctx, id)
}

func (s WarehouseService) CreateWarehouse(ctx context.Context, warehouse *reference.Warehouse) (*reference.Warehouse, error) {
	if err := s.validateWarehouse(ctx, warehouse, 0); err != nil {
		return nil, err
	}
	return s.warehouses.Create(ctx, warehouse)
}

func (s WarehouseService) UpdateWarehouse(ctx context.Context, warehouse *reference.Warehouse) (*reference.Warehouse, error) {
	if err := s.validateWarehouse(ctx, warehouse, warehouse.ID); err != nil {
		return nil, err
	}
	return s.warehouses.Update(ctx, warehouse)
}

func (s WarehouseService) validateWarehouse(ctx context.Context, warehouse *reference.Warehouse, excludeID uint64) error {
	if warehouse.Name == "" {
		return ErrWarehouseNameRequired
	}
	if warehouse.Code != nil && *warehouse.Code != "" {
		existing, err := s.warehouses.Search(ctx, "code", *warehouse.Code)
		if err != nil {
			return err
		}
		if existing != nil && existing.ID != excludeID {
			return ErrWarehouseCodeTaken
		}
	}
	return nil
}

func (s WarehouseService) CreateLocation(ctx context.Context, location *reference.StockLocation) (*reference.StockLocation, error) {
	if err := s.validateLocation(ctx, location, 0); err != nil {
		return nil, err
	}
	return s.locations.Create(ctx, location)
}

func (s WarehouseService) UpdateLocation(ctx context.Context, location *reference.StockLocation) (*reference.StockLocation, error) {
	if err := s.validateLocation(ctx, location, location.ID); err != nil {
		return nil, err
	}
	return s.locations.Update(ctx, location)
}

func (s WarehouseService) validateLocation(ctx context.Context, location *reference.StockLocation, excludeID uint64) error {
	if location.Name == "" {
		return ErrLocationNameRequired
	}
	if !validLocationUsage[location.Usage] {
		return ErrInvalidLocationType
	}
	if location.WarehouseID != nil {
		warehouse, err := s.warehouses.Find(ctx, *location.WarehouseID)
		if err != nil {
			return err
		}
		if warehouse == nil {
			return ErrWarehouseNotFound
		}
		if location.OrganizationID != nil && warehouse.OrganizationID != nil && *location.OrganizationID != *warehouse.OrganizationID {
			return ErrWarehouseMismatch
		}
	}
	if location.ParentID != nil {
		parent, err := s.locations.Find(ctx, *location.ParentID)
		if err != nil {
			return err
		}
		if parent == nil {
			return ErrLocationParent
		}
		if location.OrganizationID != nil && parent.OrganizationID != nil && *location.OrganizationID != *parent.OrganizationID {
			return ErrLocationParent
		}
	}
	return nil
}

type LedgerService struct {
	movements StockMovementDAO
	quants    StockBalanceDAO
	locations StockLocationDAO
	tx        db.Transactioner
}

func NewLedgerService(movements StockMovementDAO, quants StockBalanceDAO, locations StockLocationDAO, tx db.Transactioner) LedgerService {
	return LedgerService{movements: movements, quants: quants, locations: locations, tx: tx}
}

func (s LedgerService) CreateMovement(ctx context.Context, movement *StockMovement) (*StockMovement, error) {
	if err := s.validateMove(ctx, movement); err != nil {
		return nil, err
	}
	movement.State = MovementStateDraft
	return s.movements.Create(ctx, movement)
}

func (s LedgerService) validateMove(ctx context.Context, movement *StockMovement) error {
	if movement.ItemID == 0 {
		return ErrMovementItem
	}
	if movement.Qty <= 0 {
		return ErrMovementQty
	}
	src, err := s.locations.Find(ctx, movement.SrcLocationID)
	if err != nil {
		return err
	}
	if src == nil {
		return ErrLocationNotFound
	}
	dst, err := s.locations.Find(ctx, movement.DstLocationID)
	if err != nil {
		return err
	}
	if dst == nil {
		return ErrLocationNotFound
	}
	return nil
}

func (s LedgerService) OnHand(ctx context.Context, organizationID *uint64, itemID, locationID uint64) (float64, error) {
	location, err := s.locations.Find(ctx, locationID)
	if err != nil {
		return 0, err
	}
	if location == nil || !organizationMatches(location.OrganizationID, organizationID) {
		return 0, ErrLocationNotFound
	}
	quant, err := s.quants.FindByKey(ctx, itemID, locationID, nil)
	if err != nil {
		return 0, err
	}
	if quant == nil {
		return 0, nil
	}
	return quant.Quantity, nil
}

func organizationMatches(record, caller *uint64) bool {
	if caller == nil {
		return true
	}
	return record != nil && *record == *caller
}

func (s LedgerService) OnHandByProduct(ctx context.Context, itemID uint64) (float64, error) {
	quants, err := s.quants.ListByItem(ctx, itemID)
	if err != nil {
		return 0, err
	}
	total := 0.0
	for _, quant := range quants {
		total += quant.Quantity
	}
	return total, nil
}

func (s LedgerService) OnHandByWarehouse(ctx context.Context, itemID uint64) (map[uint64]float64, error) {
	quants, err := s.quants.ListByItem(ctx, itemID)
	if err != nil {
		return nil, err
	}
	byWarehouse := map[uint64]float64{}
	for _, quant := range quants {
		location, err := s.locations.Find(ctx, quant.LocationID)
		if err != nil {
			return nil, err
		}
		if location == nil || location.WarehouseID == nil {
			continue
		}
		byWarehouse[*location.WarehouseID] += quant.Quantity
	}
	return byWarehouse, nil
}

func (s LedgerService) AvailableToPromise(ctx context.Context, organizationID *uint64, itemID, locationID uint64) (float64, error) {
	location, err := s.locations.Find(ctx, locationID)
	if err != nil {
		return 0, err
	}
	if location == nil || !organizationMatches(location.OrganizationID, organizationID) {
		return 0, ErrLocationNotFound
	}
	quant, err := s.quants.FindByKey(ctx, itemID, locationID, nil)
	if err != nil {
		return 0, err
	}
	if quant == nil {
		return 0, nil
	}
	return quant.Quantity - quant.ReservedQty, nil
}

func (s LedgerService) RebuildBalances(ctx context.Context, organizationID *uint64) error {
	totals, err := s.movements.LedgerTotals(ctx, organizationID)
	if err != nil {
		return err
	}
	quants, err := s.quants.ListAll(ctx, organizationID)
	if err != nil {
		return err
	}

	seen := map[string]struct{}{}
	for _, total := range totals {
		key := quantKey(total.ItemID, total.LocationID, total.BatchID)
		seen[key] = struct{}{}

		existing, err := s.quants.FindByKey(ctx, total.ItemID, total.LocationID, total.BatchID)
		if err != nil {
			return err
		}
		if existing != nil {
			existing.Quantity = total.Total
			if _, err := s.quants.Update(ctx, existing); err != nil {
				return err
			}
			continue
		}
		if _, err := s.quants.Create(ctx, &StockBalance{
			OrganizationID: organizationID,
			ItemID:         total.ItemID,
			LocationID:     total.LocationID,
			BatchID:        total.BatchID,
			Quantity:       total.Total,
		}); err != nil {
			return err
		}
	}

	for _, quant := range quants {
		key := quantKey(quant.ItemID, quant.LocationID, quant.BatchID)
		if _, ok := seen[key]; ok {
			continue
		}
		if quant.Quantity != 0 {
			quant.Quantity = 0
			if _, err := s.quants.Update(ctx, quant); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s LedgerService) InternalTransfer(ctx context.Context, movement *StockMovement, doneAt time.Time) (*StockMovement, error) {
	created, err := s.CreateMovement(ctx, movement)
	if err != nil {
		return nil, err
	}
	var applied *StockMovement
	if err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		locked, err := FindApplicableMovementsTx(ctx, tx, s.movements, []uint64{created.ID})
		if err != nil {
			return err
		}
		locked[0].State = MovementStateDone
		locked[0].DateDone = &doneAt
		applied, err = s.movements.ApplyTx(ctx, tx, locked[0])
		return err
	}); err != nil {
		return nil, err
	}
	return applied, nil
}

func quantKey(itemID, locationID uint64, batchID *uint64) string {
	if batchID == nil {
		return strconv.FormatUint(itemID, 10) + ":" + strconv.FormatUint(locationID, 10) + ":0"
	}
	return strconv.FormatUint(itemID, 10) + ":" + strconv.FormatUint(locationID, 10) + ":" + strconv.FormatUint(*batchID, 10)
}
