package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/state"
	"gorm.io/gorm"
)

type ServiceService struct {
	equipments EquipmentDAO
	contracts  ServiceContractDAO
	orders     ServiceOrderDAO
	lines      ServiceOrderLineDAO
	poster     accounting.Poster
	invoices   InvoiceBuilder
	tx         db.Transactioner
	orderState state.Machine
}

func NewServiceService(
	equipments EquipmentDAO,
	contracts ServiceContractDAO,
	orders ServiceOrderDAO,
	lines ServiceOrderLineDAO,
	poster accounting.Poster,
	invoices InvoiceBuilder,
	tx db.Transactioner,
) ServiceService {
	return ServiceService{
		equipments: equipments,
		contracts:  contracts,
		orders:     orders,
		lines:      lines,
		poster:     poster,
		invoices:   invoices,
		tx:         tx,
		orderState: state.NewMachine(
			state.Transition{From: "new", To: OrderStateScheduled},
			state.Transition{From: "new", To: OrderStateCancelled},
			state.Transition{From: OrderStateScheduled, To: OrderStateInProgress},
			state.Transition{From: OrderStateScheduled, To: OrderStateCancelled},
			state.Transition{From: OrderStateInProgress, To: OrderStateDone},
			state.Transition{From: OrderStateInProgress, To: OrderStateCancelled},
			state.Transition{From: OrderStateDone, To: OrderStateInvoiced},
			state.Transition{From: OrderStateDone, To: OrderStateCancelled},
		),
	}
}

func validateLineType(lineType string) error {
	switch lineType {
	case LineTypePart, LineTypeLabor, LineTypeExpense:
		return nil
	}
	return ErrOrderLineInvalidType
}

func (s ServiceService) ListEquipments(ctx context.Context, q *query.Query) (*query.Page[Equipment], error) {
	return s.equipments.List(ctx, q)
}

func (s ServiceService) FindEquipment(ctx context.Context, id uint64) (*Equipment, error) {
	return s.equipments.Find(ctx, id)
}

func (s ServiceService) CreateEquipment(ctx context.Context, equipment *Equipment) (*Equipment, error) {
	return s.equipments.Create(ctx, equipment)
}

func (s ServiceService) ListContracts(ctx context.Context, q *query.Query) (*query.Page[ServiceContract], error) {
	return s.contracts.List(ctx, q)
}

func (s ServiceService) FindContract(ctx context.Context, id uint64) (*ServiceContract, error) {
	return s.contracts.Find(ctx, id)
}

func (s ServiceService) ListOrders(ctx context.Context, q *query.Query) (*query.Page[ServiceOrder], error) {
	return s.orders.List(ctx, q)
}

func (s ServiceService) FindOrder(ctx context.Context, id uint64) (*ServiceOrder, error) {
	return s.orders.Find(ctx, id)
}

func (s ServiceService) ListOrderLines(ctx context.Context, orderID uint64) ([]*ServiceOrderLine, error) {
	return s.lines.ListByOrder(ctx, orderID)
}

type CreateOrderRequest struct {
	OrganizationID uint64
	Name           string
	ContactID      *uint64
	EquipmentID    *uint64
	ContractID     *uint64
	Type           string
	Priority       int
	ScheduledDate  *time.Time
	TechnicianID   *uint64
	DimensionID    *uint64
	ReportedIssue  string
	Lines          []LineRequest
}

type LineRequest struct {
	Type              string
	ItemID            *uint64
	Description       string
	Qty               float64
	UnitID            *uint64
	UnitCost          float64
	UnitPrice         float64
	Billable          bool
	CoveredByWarranty bool
}

func (s ServiceService) CreateOrder(ctx context.Context, request CreateOrderRequest) (*ServiceOrder, error) {
	if request.Name == "" {
		return nil, ErrOrderNameRequired
	}
	switch request.Type {
	case OrderTypeRepair, OrderTypeMaintenance, OrderTypeInstallation, OrderTypeInspection:
	default:
		return nil, ErrOrderInvalidType
	}
	if len(request.Lines) == 0 {
		return nil, ErrOrderEmptyLines
	}
	for _, line := range request.Lines {
		if err := validateLineType(line.Type); err != nil {
			return nil, err
		}
		if line.Qty <= 0 {
			return nil, ErrOrderLineInvalidQty
		}
		if line.UnitCost < 0 || line.UnitPrice < 0 {
			return nil, ErrOrderLineInvalidPrice
		}
	}

	order := &ServiceOrder{
		OrganizationID: &request.OrganizationID,
		Name:           request.Name,
		ContactID:      request.ContactID,
		EquipmentID:    request.EquipmentID,
		ContractID:     request.ContractID,
		Type:           request.Type,
		Priority:       request.Priority,
		State:          OrderStateNew,
		ScheduledDate:  request.ScheduledDate,
		TechnicianID:   request.TechnicianID,
		DimensionID:    request.DimensionID,
		ReportedIssue:  request.ReportedIssue,
	}

	var created *ServiceOrder
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		createdOrder, err := s.orders.CreateTx(ctx, tx, order)
		if err != nil {
			return err
		}
		for i := range request.Lines {
			line := request.Lines[i]
			_, err := s.lines.CreateTx(ctx, tx, &ServiceOrderLine{
				ServiceOrderID:    createdOrder.ID,
				Type:              line.Type,
				ItemID:            line.ItemID,
				Description:       line.Description,
				Qty:               line.Qty,
				UnitID:            line.UnitID,
				UnitCost:          line.UnitCost,
				UnitPrice:         line.UnitPrice,
				Billable:          line.Billable,
				CoveredByWarranty: line.CoveredByWarranty,
			})
			if err != nil {
				return err
			}
		}
		created = createdOrder
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

type AddLineRequest struct {
	OrderID uint64
	Line    LineRequest
}

func (s ServiceService) AddLine(ctx context.Context, request AddLineRequest) (*ServiceOrderLine, error) {
	order, err := s.order(ctx, request.OrderID)
	if err != nil {
		return nil, err
	}
	if order.State != OrderStateNew && order.State != OrderStateScheduled && order.State != OrderStateInProgress {
		return nil, ErrOrderInvalidState
	}
	if err := validateLineType(request.Line.Type); err != nil {
		return nil, err
	}
	if request.Line.Qty <= 0 {
		return nil, ErrOrderLineInvalidQty
	}
	if request.Line.UnitCost < 0 || request.Line.UnitPrice < 0 {
		return nil, ErrOrderLineInvalidPrice
	}
	var created *ServiceOrderLine
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		line, err := s.lines.CreateTx(ctx, tx, &ServiceOrderLine{
			ServiceOrderID:    order.ID,
			Type:              request.Line.Type,
			ItemID:            request.Line.ItemID,
			Description:       request.Line.Description,
			Qty:               request.Line.Qty,
			UnitID:            request.Line.UnitID,
			UnitCost:          request.Line.UnitCost,
			UnitPrice:         request.Line.UnitPrice,
			Billable:          request.Line.Billable,
			CoveredByWarranty: request.Line.CoveredByWarranty,
		})
		if err != nil {
			return err
		}
		created = line
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s ServiceService) order(ctx context.Context, id uint64) (*ServiceOrder, error) {
	order, err := s.orders.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s ServiceService) transition(ctx context.Context, order *ServiceOrder, to string) (*ServiceOrder, error) {
	if err := s.orderState.TryTransition(model.Status(order.State), model.Status(to)); err != nil {
		return nil, ErrOrderInvalidState
	}
	order.State = to
	var updated *ServiceOrder
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		u, err := s.orders.UpdateTx(ctx, tx, order)
		if err != nil {
			return err
		}
		updated = u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

type ScheduleRequest struct {
	OrderID       uint64
	ScheduledDate time.Time
	TechnicianID  *uint64
}

func (s ServiceService) Schedule(ctx context.Context, request ScheduleRequest) (*ServiceOrder, error) {
	order, err := s.order(ctx, request.OrderID)
	if err != nil {
		return nil, err
	}
	updated, err := s.transition(ctx, order, OrderStateScheduled)
	if err != nil {
		return nil, err
	}
	date := request.ScheduledDate
	updated.ScheduledDate = &date
	updated.TechnicianID = request.TechnicianID
	return s.orders.Update(ctx, updated)
}

func (s ServiceService) Start(ctx context.Context, orderID uint64) (*ServiceOrder, error) {
	order, err := s.order(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return s.transition(ctx, order, OrderStateInProgress)
}

type CompleteRequest struct {
	OrderID                 uint64
	JournalID               uint64
	Date                    time.Time
	COGSAccountID           uint64
	StockValuationAccountID uint64
	Resolution              string
}

func (s ServiceService) Complete(ctx context.Context, request CompleteRequest) (*ServiceOrder, error) {
	if request.JournalID == 0 {
		return nil, ErrOrderRequiresJournal
	}
	if request.COGSAccountID == 0 || request.StockValuationAccountID == 0 {
		return nil, ErrOrderRequiresAccounts
	}
	order, err := s.order(ctx, request.OrderID)
	if err != nil {
		return nil, err
	}
	if err := s.orderState.TryTransition(model.Status(order.State), model.Status(OrderStateDone)); err != nil {
		return nil, ErrOrderInvalidState
	}
	lines, err := s.lines.ListByOrder(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	var updated *ServiceOrder
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		for _, line := range lines {
			if line.Type != LineTypePart {
				continue
			}
			cost := amount.FromFloat64(line.Qty).Mul(amount.FromFloat64(line.UnitCost)).Round(4)
			if cost.IsZero() {
				continue
			}
			_, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
				OrganizationID: *order.OrganizationID,
				JournalID:      request.JournalID,
				Date:           request.Date,
				Ref:            fmt.Sprintf("SVC-%d", order.ID),
				OriginType:     accounting.OriginTypeServicePart,
				OriginID:       line.ID,
				Description:    fmt.Sprintf("Service part %s", order.Name),
				Lines: []accounting.PostingLine{
					{AccountID: request.COGSAccountID, Name: "COGS", Debit: cost},
					{AccountID: request.StockValuationAccountID, Name: "Stock Valuation", Credit: cost},
				},
			})
			if err != nil {
				return err
			}
		}
		order.State = OrderStateDone
		if request.Resolution != "" {
			order.Resolution = request.Resolution
		}
		u, err := s.orders.UpdateTx(ctx, tx, order)
		if err != nil {
			return err
		}
		updated = u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

type BillRequest struct {
	OrderID          uint64
	JournalID        uint64
	Date             time.Time
	RevenueAccountID uint64
	DeferredAccount  *uint64
}

func (s ServiceService) Bill(ctx context.Context, request BillRequest) (*ServiceOrder, error) {
	if request.JournalID == 0 {
		return nil, ErrOrderRequiresJournal
	}
	if request.RevenueAccountID == 0 {
		return nil, ErrOrderRequiresAccounts
	}
	order, err := s.order(ctx, request.OrderID)
	if err != nil {
		return nil, err
	}
	if order.State != OrderStateDone {
		return nil, ErrOrderNotDone
	}
	if order.ContactID == nil {
		return nil, ErrOrderRequiresContact
	}
	lines, err := s.lines.ListByOrder(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	invoiceLines := make([]accounting.InvoiceLineRequest, 0, len(lines))
	for _, line := range lines {
		if !line.Billable || line.UnitPrice <= 0 {
			continue
		}
		invoiceLines = append(invoiceLines, accounting.InvoiceLineRequest{
			ItemID:      line.ItemID,
			Description: line.Description,
			Qty:         line.Qty,
			UnitID:      line.UnitID,
			UnitPrice:   line.UnitPrice,
			AccountID:   request.RevenueAccountID,
		})
	}
	if len(invoiceLines) == 0 {
		return nil, ErrOrderNoLines
	}

	date := request.Date
	if date.IsZero() {
		date = time.Now().UTC()
	}
	invoice, err := s.invoices.Create(ctx, accounting.CreateInvoiceRequest{
		OrganizationID:  *order.OrganizationID,
		JournalID:       request.JournalID,
		ContactID:       *order.ContactID,
		Date:            date,
		Reference:       order.Name,
		DeferredAccount: request.DeferredAccount,
		Lines:           invoiceLines,
	})
	if err != nil {
		return nil, err
	}

	updated := order
	updated.State = OrderStateInvoiced
	updated.InvoiceID = &invoice.ID
	return s.orders.Update(ctx, updated)
}

func (s ServiceService) CancelOrder(ctx context.Context, orderID uint64) (*ServiceOrder, error) {
	order, err := s.order(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.State == OrderStateDone || order.State == OrderStateInvoiced {
		return nil, ErrOrderInvalidState
	}
	return s.transition(ctx, order, OrderStateCancelled)
}

type CreateContractRequest struct {
	OrganizationID   uint64
	Name             string
	ContactID        *uint64
	EquipmentID      *uint64
	SubscriptionID   *uint64
	Coverage         string
	SLAResponseHours *int
	DateStart        *time.Time
	DateEnd          *time.Time
}

func (s ServiceService) CreateContract(ctx context.Context, request CreateContractRequest) (*ServiceContract, error) {
	if request.Name == "" {
		return nil, ErrContractNameRequired
	}
	stateValue := ContractStateActive
	if request.DateStart == nil || request.DateEnd == nil {
		stateValue = ContractStateDraft
	}
	return s.contracts.Create(ctx, &ServiceContract{
		OrganizationID:   &request.OrganizationID,
		Name:             request.Name,
		ContactID:        request.ContactID,
		EquipmentID:      request.EquipmentID,
		SubscriptionID:   request.SubscriptionID,
		Coverage:         request.Coverage,
		SLAResponseHours: request.SLAResponseHours,
		DateStart:        request.DateStart,
		DateEnd:          request.DateEnd,
		State:            stateValue,
	})
}

func (s ServiceService) CancelContract(ctx context.Context, contractID uint64) (*ServiceContract, error) {
	contract, err := s.contracts.Find(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if contract == nil {
		return nil, ErrContractNotFound
	}
	if contract.State != ContractStateActive {
		return nil, ErrContractInvalidState
	}
	contract.State = ContractStateCancelled
	return s.contracts.Update(ctx, contract)
}

func (s ServiceService) ActivateContract(ctx context.Context, contractID uint64) (*ServiceContract, error) {
	contract, err := s.contracts.Find(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if contract == nil {
		return nil, ErrContractNotFound
	}
	if contract.State != ContractStateDraft {
		return nil, ErrContractInvalidState
	}
	if contract.DateStart == nil || contract.DateEnd == nil {
		return nil, ErrContractInvalidState
	}
	contract.State = ContractStateActive
	return s.contracts.Update(ctx, contract)
}
