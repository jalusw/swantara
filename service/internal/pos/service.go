package pos

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/kernel/state"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

const (
	locationUsageInternal = "internal"
	locationUsageCustomer = "customer"
	paymentMethodCash     = "cash"
	accountTypeReceivable = "receivable"
	originPOSOrder        = "pos_order"
	originPOSRefund       = "pos_refund"
	originPOSInvoiceShift = "pos_order_invoice_shift"
)

type POSService struct {
	configs    dao.CRUD[reference.POSConfig]
	sessions   POSSessionDAO
	orders     POSOrderDAO
	lines      POSOrderLineDAO
	payments   POSPaymentDAO
	productSvc products.ProductService
	taxes      dao.CRUD[reference.Tax]
	journals   dao.CRUD[reference.Journal]
	accounts   AccountLookup
	payMethods PaymentAccountLookup
	locations  inventory.StockLocationDAO
	quants     inventory.StockBalanceDAO
	movements  inventory.StockMovementDAO
	layers     inventory.CostLayerDAO
	shipments  inventory.ShipmentDAO
	valuer     StockValuer
	invoice    InvoiceEngine
	poster     accounting.Poster
	members    iam.MemberDAO
	sequences  sequence.Service
	tx         db.Transactioner
	machine    state.Machine
	now        func() time.Time
}

func NewPOSService(
	configs dao.CRUD[reference.POSConfig],
	sessions POSSessionDAO,
	orders POSOrderDAO,
	lines POSOrderLineDAO,
	payments POSPaymentDAO,
	productSvc products.ProductService,
	taxes dao.CRUD[reference.Tax],
	journals dao.CRUD[reference.Journal],
	accounts AccountLookup,
	payMethods PaymentAccountLookup,
	locations inventory.StockLocationDAO,
	quants inventory.StockBalanceDAO,
	movements inventory.StockMovementDAO,
	layers inventory.CostLayerDAO,
	shipments inventory.ShipmentDAO,
	valuer StockValuer,
	invoice InvoiceEngine,
	poster accounting.Poster,
	members iam.MemberDAO,
	sequences sequence.Service,
	tx db.Transactioner,
) POSService {
	return POSService{
		configs:    configs,
		sessions:   sessions,
		orders:     orders,
		lines:      lines,
		payments:   payments,
		productSvc: productSvc,
		taxes:      taxes,
		journals:   journals,
		accounts:   accounts,
		payMethods: payMethods,
		locations:  locations,
		quants:     quants,
		movements:  movements,
		layers:     layers,
		shipments:  shipments,
		valuer:     valuer,
		invoice:    invoice,
		poster:     poster,
		members:    members,
		sequences:  sequences,
		tx:         tx,
		machine: state.NewMachine(
			state.Transition{From: model.Status(SessionStateOpened), To: model.Status(SessionStateClosing)},
			state.Transition{From: model.Status(SessionStateClosing), To: model.Status(SessionStateClosed)},
		),
		now: time.Now,
	}
}

type POSConfigService struct {
	configs dao.CRUD[reference.POSConfig]
}

func NewPOSConfigService(configs dao.CRUD[reference.POSConfig]) POSConfigService {
	return POSConfigService{configs: configs}
}

func (s POSConfigService) List(ctx context.Context, q *query.Query) (*query.Page[reference.POSConfig], error) {
	return s.configs.List(ctx, q)
}

func (s POSConfigService) Find(ctx context.Context, id uint64) (*reference.POSConfig, error) {
	return s.configs.Find(ctx, id)
}

func (s POSConfigService) Create(ctx context.Context, config *reference.POSConfig) (*reference.POSConfig, error) {
	return s.configs.Create(ctx, config)
}

func (s POSConfigService) Update(ctx context.Context, config *reference.POSConfig) (*reference.POSConfig, error) {
	return s.configs.Update(ctx, config)
}

func (s POSConfigService) Delete(ctx context.Context, id uint64) error {
	return s.configs.Delete(ctx, id)
}

func (s POSService) OpenSession(ctx context.Context, configID, cashierID uint64, openingBalance float64) (*POSSession, error) {
	if cashierID == 0 {
		return nil, ErrNoCashier
	}
	config, err := s.configs.Find(ctx, configID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, ErrConfigNotFound
	}
	if config.OrganizationID == nil {
		return nil, ErrConfigNotFound
	}
	member, err := s.members.FindByUserAndOrganization(ctx, cashierID, *config.OrganizationID)
	if err != nil {
		return nil, err
	}
	if member == nil || member.ID == 0 {
		return nil, ErrNoCashier
	}
	page, err := s.sessions.List(ctx, &query.Query{Filters: []query.Filter{{Field: "config_id", Operator: query.Equal, Value: configID}}})
	if err != nil {
		return nil, err
	}
	for _, existing := range page.Items {
		if existing.State == SessionStateOpened {
			return nil, ErrSessionOpen
		}
	}

	now := s.now().UTC()
	session := &POSSession{
		ConfigID:       configID,
		CashierID:      cashierID,
		OpenedAt:       &now,
		OpeningBalance: openingBalance,
		State:          SessionStateOpened,
	}
	created, err := s.sessions.Create(ctx, session)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrSessionOpen
		}
		return nil, err
	}
	return created, nil
}

func (s POSService) ListSessions(ctx context.Context, organizationID uint64, q *query.Query) (*query.Page[POSSession], error) {
	if organizationID == 0 {
		return s.sessions.List(ctx, q)
	}
	return s.sessions.ListInOrganization(ctx, q, organizationID)
}

func (s POSService) GetSession(ctx context.Context, organizationID, sessionID uint64) (*POSSession, error) {
	session, err := s.findSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if organizationID == 0 {
		return session, nil
	}
	config, err := s.configs.Find(ctx, session.ConfigID)
	if err != nil {
		return nil, err
	}
	if config == nil || config.OrganizationID == nil || *config.OrganizationID != organizationID {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s POSService) GetOrder(ctx context.Context, organizationID, orderID uint64) (*POSOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if organizationID == 0 {
		return order, nil
	}
	session, err := s.findSession(ctx, order.SessionID)
	if err != nil {
		return nil, err
	}
	config, err := s.configs.Find(ctx, session.ConfigID)
	if err != nil {
		return nil, err
	}
	if config == nil || config.OrganizationID == nil || *config.OrganizationID != organizationID {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s POSService) ListOrderLines(ctx context.Context, orderID uint64) ([]*POSOrderLine, error) {
	return s.lines.ListByOrder(ctx, orderID)
}

func (s POSService) ListOrderPayments(ctx context.Context, orderID uint64) ([]*POSPayment, error) {
	return s.payments.ListByOrder(ctx, orderID)
}

func (s POSService) ListOrders(ctx context.Context, organizationID uint64, q *query.Query) (*query.Page[POSOrder], error) {
	if organizationID == 0 {
		return s.orders.List(ctx, q)
	}
	return s.orders.ListInOrganization(ctx, q, organizationID)
}

func (s POSService) StartClosing(ctx context.Context, sessionID uint64) (*POSSession, error) {
	session, err := s.findSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(session.State), model.Status(SessionStateClosing)); err != nil {
		return nil, ErrSessionState
	}
	session.State = SessionStateClosing
	return s.sessions.Update(ctx, session)
}

func (s POSService) CloseSession(ctx context.Context, sessionID uint64, closingBalance float64) (*POSSession, error) {
	session, err := s.findSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(session.State), model.Status(SessionStateClosed)); err != nil {
		return nil, ErrSessionState
	}

	expected, err := s.sessionCashTotal(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	expected = expected.Add(amount.FromFloat64(session.OpeningBalance)).Round(4)
	if !amount.FromFloat64(closingBalance).Round(4).Equal(expected) {
		return nil, ErrSessionReconciliation
	}

	now := s.now().UTC()
	session.State = SessionStateClosed
	session.ClosedAt = &now
	session.ClosingBalance = helper.Ptr(expected.Float64())
	return s.sessions.Update(ctx, session)
}

type SellLineRequest struct {
	ItemID      uint64
	Qty         float64
	DiscountPct float64
	TaxIDs      helper.Int64Array
}

type SellPaymentRequest struct {
	Method string
	Amount float64
}

type SellRequest struct {
	SessionID uint64
	ContactID *uint64
	Lines     []SellLineRequest
	Payments  []SellPaymentRequest
}

func (s POSService) Sell(ctx context.Context, request SellRequest) (*POSOrder, error) {
	if len(request.Lines) == 0 {
		return nil, ErrOrderNoLines
	}
	if len(request.Payments) == 0 {
		return nil, ErrPaymentMissing
	}
	session, err := s.findSession(ctx, request.SessionID)
	if err != nil {
		return nil, err
	}
	if session.State != SessionStateOpened {
		return nil, ErrSessionState
	}
	config, err := s.configs.Find(ctx, session.ConfigID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, ErrConfigNotFound
	}
	if config.WarehouseID == nil {
		return nil, ErrConfigWarehouse
	}
	if config.JournalID == nil {
		return nil, ErrConfigJournal
	}
	if config.PriceBookID == nil {
		return nil, ErrConfigPriceBook
	}
	if config.OrganizationID == nil {
		return nil, ErrConfigNotFound
	}

	stockLocationID, err := s.warehouseStockLocation(ctx, *config.OrganizationID, *config.WarehouseID)
	if err != nil {
		return nil, err
	}
	customerLocationID, err := s.customerLocation(ctx, *config.OrganizationID)
	if err != nil {
		return nil, err
	}

	date := s.now().UTC()
	taxes, err := s.findTaxes(ctx, collectTaxIDs(request.Lines))
	if err != nil {
		return nil, err
	}
	orderLines := make([]*POSOrderLine, 0, len(request.Lines))
	untaxed := amount.Zero()
	tax := amount.Zero()
	for _, lineRequest := range request.Lines {
		line, err := s.buildLine(ctx, config, lineRequest, taxes, date)
		if err != nil {
			return nil, err
		}
		orderLines = append(orderLines, line)
		untaxed = untaxed.Add(line.PriceSubtotal)
		tax = tax.Add(line.PriceTax)
	}
	total := untaxed.Add(tax).Round(4)

	paymentSum := amount.Zero()
	payments := make([]*POSPayment, 0, len(request.Payments))
	for _, paymentRequest := range request.Payments {
		if paymentRequest.Method == "" || !amount.FromFloat64(paymentRequest.Amount).GreaterThan(amount.Zero()) {
			return nil, ErrPaymentTotal
		}
		paymentSum = paymentSum.Add(amount.FromFloat64(paymentRequest.Amount))
		payments = append(payments, &POSPayment{Method: paymentRequest.Method, Amount: paymentRequest.Amount})
	}
	if !paymentSum.Round(4).Equal(total) {
		return nil, ErrPaymentTotal
	}

	order := &POSOrder{
		SessionID:   request.SessionID,
		ContactID:   request.ContactID,
		AmountTotal: total,
		AmountTax:   tax,
		State:       OrderStateDone,
		OrderTime:   &date,
	}

	var created *POSOrder
	if err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		for _, line := range orderLines {
			if line.ItemID == nil {
				continue
			}
			if err := s.ensureStockAvailableTx(ctx, tx, *line.ItemID, stockLocationID, line.Qty); err != nil {
				return err
			}
		}

		name, err := s.sequences.NextTx(ctx, tx, *config.OrganizationID, SequencePOSOrderCode)
		if err != nil {
			return err
		}
		order.Name = helper.Ptr(name)

		created, err = s.orders.CreateWithLinesAndPaymentsTx(ctx, tx, order, orderLines, payments)
		if err != nil {
			return err
		}

		if err := s.postStockMovementsTx(ctx, tx, created, orderLines, config, stockLocationID, customerLocationID, date); err != nil {
			return err
		}

		if err := s.postRevenueMoveTx(ctx, tx, created, config, orderLines, payments, date); err != nil {
			return err
		}

		if created.ContactID != nil {
			invoice, err := s.invoiceForOrderTx(ctx, tx, created, orderLines, config, *config.JournalID, date)
			if err != nil {
				return err
			}
			created.InvoiceID = &invoice.ID
			if _, err := s.orders.UpdateTx(ctx, tx, created); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return created, nil
}

func (s POSService) CreateInvoice(ctx context.Context, orderID, journalID uint64, date time.Time) (*accounting.Invoice, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.ContactID == nil {
		return nil, ErrOrderNotFound
	}
	if order.InvoiceID != nil {
		return nil, ErrOrderInvoiced
	}
	session, err := s.findSession(ctx, order.SessionID)
	if err != nil {
		return nil, err
	}
	config, err := s.configs.Find(ctx, session.ConfigID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, ErrConfigNotFound
	}

	lines, err := s.lines.ListByOrder(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	var invoice *accounting.Invoice
	if err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		invoice, err = s.invoiceForOrderTx(ctx, tx, order, lines, config, journalID, date)
		if err != nil {
			return err
		}
		order.InvoiceID = &invoice.ID
		_, err = s.orders.UpdateTx(ctx, tx, order)
		return err
	}); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s POSService) SessionPaymentBreakdown(ctx context.Context, sessionID uint64) (map[string]float64, error) {
	totals, err := s.payments.SumBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	breakdown := make(map[string]float64, len(totals))
	for _, total := range totals {
		breakdown[total.Method] = total.Total
	}
	return breakdown, nil
}

func (s POSService) Refund(ctx context.Context, orderID, journalID uint64, date time.Time) (*POSOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.State != OrderStateDone {
		return nil, ErrOrderState
	}
	session, err := s.findSession(ctx, order.SessionID)
	if err != nil {
		return nil, err
	}
	config, err := s.configs.Find(ctx, session.ConfigID)
	if err != nil {
		return nil, err
	}
	if config == nil || config.OrganizationID == nil {
		return nil, ErrConfigNotFound
	}

	lines, err := s.lines.ListByOrder(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrOrderNoLines
	}

	if err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		if err := s.restockOrderTx(ctx, tx, order, lines, config, journalID, date); err != nil {
			return err
		}
		if err := s.postRefundMoveTx(ctx, tx, order, config, lines, journalID, date); err != nil {
			return err
		}
		if order.InvoiceID != nil {
			if _, err := s.invoice.CreateCreditNoteTx(ctx, tx, accounting.CreateCreditNoteRequest{
				OrganizationID:    *config.OrganizationID,
				OriginalInvoiceID: *order.InvoiceID,
				JournalID:         journalID,
				Date:              date,
				Reference:         helper.Deref(order.Name, ""),
			}); err != nil {
				return err
			}
		}

		order.State = OrderStateRefunded
		_, err := s.orders.UpdateTx(ctx, tx, order)
		return err
	}); err != nil {
		return nil, err
	}
	return order, nil
}

func (s POSService) restockOrderTx(ctx context.Context, tx *gorm.DB, order *POSOrder, lines []*POSOrderLine, config *reference.POSConfig, journalID uint64, date time.Time) error {
	originalMoves, err := s.movements.ListByOrigin(ctx, originPOSOrder, order.ID)
	if err != nil {
		return err
	}
	customerLocationID, err := s.customerLocation(ctx, *config.OrganizationID)
	if err != nil {
		return err
	}
	for _, line := range lines {
		if line.ItemID == nil {
			continue
		}
		original := findDoneMove(originalMoves, *line.ItemID)
		if original == nil {
			return ErrOrderNoStock
		}
		unitCost, err := s.originalUnitCost(ctx, original)
		if err != nil {
			return err
		}
		movement, err := s.movements.CreateTx(ctx, tx, &inventory.StockMovement{
			OrganizationID: config.OrganizationID,
			ItemID:         *line.ItemID,
			Qty:            line.Qty,
			SrcLocationID:  customerLocationID,
			DstLocationID:  original.SrcLocationID,
			State:          inventory.MovementStateConfirmed,
			OriginType:     helper.Ptr(originPOSRefund),
			OriginID:       helper.Ptr(order.ID),
			ScheduledDate:  &date,
		})
		if err != nil {
			return err
		}
		if _, err := s.valuer.RestockTx(ctx, tx, movement.ID, unitCost, journalID, date); err != nil {
			return err
		}
	}
	return nil
}

func (s POSService) postRefundMoveTx(ctx context.Context, tx *gorm.DB, order *POSOrder, config *reference.POSConfig, lines []*POSOrderLine, journalID uint64, date time.Time) error {
	cashAccount, err := s.journalDefaultAccount(ctx, journalID)
	if err != nil {
		return err
	}
	orderPayments, err := s.payments.ListByOrder(ctx, order.ID)
	if err != nil {
		return err
	}
	creditByAccount, err := s.paymentsByAccount(ctx, *config.OrganizationID, orderPayments, cashAccount)
	if err != nil {
		return err
	}

	revenueByAccount, taxByAccount, err := s.postingAggregates(ctx, lines)
	if err != nil {
		return err
	}

	postLines := make([]accounting.PostingLine, 0, len(creditByAccount)+len(revenueByAccount)+len(taxByAccount))
	for accountID, credit := range creditByAccount {
		postLines = append(postLines, accounting.PostingLine{AccountID: accountID, Name: "Refund", Credit: credit})
	}
	for accountID, revenue := range revenueByAccount {
		postLines = append(postLines, accounting.PostingLine{AccountID: accountID, Name: "Revenue", Debit: revenue})
	}
	for accountID, taxAmount := range taxByAccount {
		postLines = append(postLines, accounting.PostingLine{AccountID: accountID, Name: "Output Tax", Debit: taxAmount})
	}

	_, err = s.poster.PostTx(ctx, tx, accounting.PostRequest{
		OrganizationID: *config.OrganizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            helper.Deref(order.Name, "") + "-refund",
		OriginType:     originPOSRefund,
		OriginID:       order.ID,
		Description:    "POS refund",
		Lines:          postLines,
	})
	return err
}

func (s POSService) originalUnitCost(ctx context.Context, original *inventory.StockMovement) (amount.Amount, error) {
	layers, err := s.layers.ListByMovement(ctx, original.ID)
	if err != nil {
		return amount.Amount{}, err
	}
	for _, layer := range layers {
		if layer.UnitCost != nil {
			return amount.FromFloat64(*layer.UnitCost), nil
		}
	}
	return amount.Amount{}, ErrOrderCost
}

func findDoneMove(movements []*inventory.StockMovement, itemID uint64) *inventory.StockMovement {
	for _, movement := range movements {
		if movement.ItemID == itemID && movement.State == inventory.MovementStateDone {
			return movement
		}
	}
	return nil
}

func (s POSService) findSession(ctx context.Context, sessionID uint64) (*POSSession, error) {
	session, err := s.sessions.Find(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s POSService) findOrder(ctx context.Context, orderID uint64) (*POSOrder, error) {
	order, err := s.orders.Find(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s POSService) sessionCashTotal(ctx context.Context, sessionID uint64) (amount.Amount, error) {
	totals, err := s.payments.SumBySession(ctx, sessionID)
	if err != nil {
		return amount.Zero(), err
	}
	total := amount.Zero()
	for _, methodTotal := range totals {
		if methodTotal.Method != paymentMethodCash {
			continue
		}
		total = total.Add(amount.FromFloat64(methodTotal.Total))
	}
	return total.Round(4), nil
}

func collectTaxIDs(lines []SellLineRequest) []int64 {
	var ids []int64
	for _, line := range lines {
		ids = append(ids, line.TaxIDs...)
	}
	return ids
}

func (s POSService) findTaxes(ctx context.Context, ids []int64) (map[int64]*reference.Tax, error) {
	uniqueIDs := make([]any, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	taxesByID := make(map[int64]*reference.Tax, len(uniqueIDs))
	if len(uniqueIDs) == 0 {
		return taxesByID, nil
	}
	page, err := s.taxes.List(ctx, &query.Query{Filters: []query.Filter{{Field: "id", Operator: query.In, Value: uniqueIDs}}})
	if err != nil {
		return nil, err
	}
	for _, tax := range page.Items {
		taxesByID[int64(tax.ID)] = tax
	}
	return taxesByID, nil
}

func (s POSService) ensureStockAvailableTx(ctx context.Context, tx *gorm.DB, itemID, locationID uint64, qty float64) error {
	quant, err := s.quants.FindByKeyTx(ctx, tx, itemID, locationID, nil, true)
	if err != nil {
		return err
	}
	available := amount.Zero()
	if quant != nil {
		available = amount.FromFloat64(quant.Quantity).Sub(amount.FromFloat64(quant.ReservedQty))
	}
	if amount.FromFloat64(qty).GreaterThan(available) {
		return ErrOrderStockUnavailable
	}
	return nil
}

func (s POSService) buildLine(ctx context.Context, config *reference.POSConfig, request SellLineRequest, taxes map[int64]*reference.Tax, date time.Time) (*POSOrderLine, error) {
	if request.ItemID == 0 {
		return nil, ErrOrderProduct
	}
	if !amount.FromFloat64(request.Qty).GreaterThan(amount.Zero()) {
		return nil, ErrOrderQty
	}
	if request.DiscountPct < 0 || request.DiscountPct > 100 {
		return nil, ErrOrderDiscount
	}

	resolved, err := s.productSvc.ResolvePrice(ctx, *config.PriceBookID, request.ItemID, amount.FromFloat64(request.Qty), date)
	if err != nil {
		return nil, err
	}

	qty := amount.FromFloat64(request.Qty)
	discount := amount.FromFloat64(1 - request.DiscountPct/100)
	subtotal := qty.Mul(resolved.Price).Mul(discount).Round(4)
	lineTax := amount.Zero()
	for _, taxID := range request.TaxIDs {
		tax, ok := taxes[taxID]
		if !ok || tax == nil {
			return nil, ErrOrderTax
		}
		if tax.Scope != reference.TaxScopeSale && tax.Scope != reference.TaxScopeNone {
			return nil, ErrOrderTax
		}
		taxAmount, err := reference.TaxLineAmount(subtotal, qty, tax)
		if err != nil {
			return nil, err
		}
		lineTax = lineTax.Add(taxAmount)
	}
	lineTax = lineTax.Round(4)

	return &POSOrderLine{
		ItemID:        helper.Ptr(request.ItemID),
		Qty:           request.Qty,
		UnitPrice:     resolved.Price,
		DiscountPct:   request.DiscountPct,
		TaxIDs:        request.TaxIDs,
		PriceSubtotal: subtotal,
		PriceTax:      lineTax,
		PriceTotal:    subtotal.Add(lineTax).Round(4),
	}, nil
}

func (s POSService) postStockMovementsTx(ctx context.Context, tx *gorm.DB, order *POSOrder, lines []*POSOrderLine, config *reference.POSConfig, stockLocationID, customerLocationID uint64, date time.Time) error {
	movements := make([]*inventory.StockMovement, 0, len(lines))
	for _, line := range lines {
		if line.ItemID == nil {
			continue
		}
		movements = append(movements, &inventory.StockMovement{
			OrganizationID: config.OrganizationID,
			ItemID:         *line.ItemID,
			Qty:            line.Qty,
			SrcLocationID:  stockLocationID,
			DstLocationID:  customerLocationID,
			State:          inventory.MovementStateConfirmed,
			OriginType:     helper.Ptr(originPOSOrder),
			OriginID:       helper.Ptr(order.ID),
			ScheduledDate:  &date,
		})
	}
	if len(movements) == 0 {
		return ErrOrderNoStock
	}

	shipment := &inventory.Shipment{
		OrganizationID: config.OrganizationID,
		Name:           order.Name,
		Type:           inventory.ShipmentTypeOutgoing,
		SrcLocationID:  helper.Ptr(stockLocationID),
		DstLocationID:  helper.Ptr(customerLocationID),
		State:          inventory.ShipmentStateConfirmed,
		Origin:         order.Name,
		ScheduledDate:  &date,
	}

	created, err := s.shipments.CreateWithMovementsTx(ctx, tx, shipment, movements)
	if err != nil {
		return err
	}
	for _, movement := range movements {
		if _, err := s.valuer.ShipTx(ctx, tx, movement.ID, *config.JournalID, date); err != nil {
			return err
		}
	}
	created.State = inventory.ShipmentStateDone
	_, err = s.shipments.UpdateTx(ctx, tx, created)
	return err
}

func (s POSService) postRevenueMoveTx(ctx context.Context, tx *gorm.DB, order *POSOrder, config *reference.POSConfig, lines []*POSOrderLine, payments []*POSPayment, date time.Time) error {
	cashAccount, err := s.journalDefaultAccount(ctx, *config.JournalID)
	if err != nil {
		return err
	}
	debitByAccount, err := s.paymentsByAccount(ctx, *config.OrganizationID, payments, cashAccount)
	if err != nil {
		return err
	}

	revenueByAccount, taxByAccount, err := s.postingAggregates(ctx, lines)
	if err != nil {
		return err
	}

	postLines := make([]accounting.PostingLine, 0, len(debitByAccount)+len(revenueByAccount)+len(taxByAccount))
	for accountID, debit := range debitByAccount {
		postLines = append(postLines, accounting.PostingLine{AccountID: accountID, Name: "Payment", Debit: debit})
	}
	for accountID, revenue := range revenueByAccount {
		postLines = append(postLines, accounting.PostingLine{AccountID: accountID, Name: "Revenue", Credit: revenue})
	}
	for accountID, taxAmount := range taxByAccount {
		postLines = append(postLines, accounting.PostingLine{AccountID: accountID, Name: "Output Tax", Credit: taxAmount})
	}

	_, err = s.poster.PostTx(ctx, tx, accounting.PostRequest{
		OrganizationID: *config.OrganizationID,
		JournalID:      *config.JournalID,
		Date:           date,
		Ref:            helper.Deref(order.Name, ""),
		OriginType:     originPOSOrder,
		OriginID:       order.ID,
		Description:    "POS sale",
		Lines:          postLines,
	})
	return err
}

func (s POSService) postingAggregates(ctx context.Context, lines []*POSOrderLine) (map[uint64]amount.Amount, map[uint64]amount.Amount, error) {
	var taxIDs []int64
	for _, line := range lines {
		taxIDs = append(taxIDs, line.TaxIDs...)
	}
	taxesByID, err := s.findTaxes(ctx, taxIDs)
	if err != nil {
		return nil, nil, err
	}

	revenueByAccount := map[uint64]amount.Amount{}
	taxByAccount := map[uint64]amount.Amount{}
	for _, line := range lines {
		if line.ItemID == nil {
			continue
		}
		incomeAccount, err := s.productSvc.ResolveIncomeAccount(ctx, *line.ItemID)
		if err != nil {
			return nil, nil, err
		}
		revenueByAccount[incomeAccount] = revenueByAccount[incomeAccount].Add(line.PriceSubtotal)
		for _, taxID := range line.TaxIDs {
			tax := taxesByID[taxID]
			if tax == nil || tax.TaxAccountID == nil {
				continue
			}
			taxByAccount[*tax.TaxAccountID] = taxByAccount[*tax.TaxAccountID].Add(line.PriceTax)
		}
	}
	return revenueByAccount, taxByAccount, nil
}

func (s POSService) invoiceForOrderTx(ctx context.Context, tx *gorm.DB, order *POSOrder, lines []*POSOrderLine, config *reference.POSConfig, journalID uint64, date time.Time) (*accounting.Invoice, error) {
	requests := make([]accounting.InvoiceLineRequest, 0, len(lines))
	revenueByAccount := map[uint64]amount.Amount{}
	for _, line := range lines {
		if line.ItemID == nil {
			continue
		}
		incomeAccount, err := s.productSvc.ResolveIncomeAccount(ctx, *line.ItemID)
		if err != nil {
			return nil, err
		}
		requests = append(requests, accounting.InvoiceLineRequest{
			ItemID:      line.ItemID,
			Description: helper.Deref(order.Name, ""),
			Qty:         line.Qty,
			UnitPrice:   line.UnitPrice.Float64(),
			DiscountPct: line.DiscountPct,
			TaxIDs:      line.TaxIDs,
			AccountID:   incomeAccount,
		})
		revenueByAccount[incomeAccount] = revenueByAccount[incomeAccount].Add(line.PriceSubtotal)
	}

	invoice, err := s.invoice.CreateTx(ctx, tx, accounting.CreateInvoiceRequest{
		OrganizationID: *config.OrganizationID,
		JournalID:      journalID,
		ContactID:      *order.ContactID,
		Date:           date,
		Reference:      helper.Deref(order.Name, ""),
		Lines:          requests,
	})
	if err != nil {
		return nil, err
	}

	receivable, err := s.receivableAccount(ctx, *config.OrganizationID)
	if err != nil {
		return nil, err
	}
	var taxIDs []int64
	for _, line := range lines {
		taxIDs = append(taxIDs, line.TaxIDs...)
	}
	taxesByID, err := s.findTaxes(ctx, taxIDs)
	if err != nil {
		return nil, err
	}
	taxByAccount := map[uint64]amount.Amount{}
	for _, line := range lines {
		for _, taxID := range line.TaxIDs {
			tax := taxesByID[taxID]
			if tax == nil || tax.TaxAccountID == nil {
				continue
			}
			taxByAccount[*tax.TaxAccountID] = taxByAccount[*tax.TaxAccountID].Add(line.PriceTax)
		}
	}

	shiftLines := []accounting.PostingLine{
		{AccountID: receivable, Name: "Accounts Receivable", Credit: order.AmountTotal},
	}
	for accountID, revenue := range revenueByAccount {
		shiftLines = append(shiftLines, accounting.PostingLine{AccountID: accountID, Name: "Revenue", Debit: revenue})
	}
	for accountID, taxAmount := range taxByAccount {
		shiftLines = append(shiftLines, accounting.PostingLine{AccountID: accountID, Name: "Output Tax", Debit: taxAmount})
	}
	if _, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
		OrganizationID: *config.OrganizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            helper.Deref(order.Name, "") + "-shift",
		OriginType:     originPOSInvoiceShift,
		OriginID:       order.ID,
		Description:    "POS invoice revenue shift",
		Lines:          shiftLines,
	}); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s POSService) journalDefaultAccount(ctx context.Context, journalID uint64) (uint64, error) {
	journal, err := s.journals.Find(ctx, journalID)
	if err != nil {
		return 0, err
	}
	if journal == nil || journal.DefaultAccountID == nil {
		return 0, ErrConfigJournal
	}
	return *journal.DefaultAccountID, nil
}

func (s POSService) paymentsByAccount(ctx context.Context, organizationID uint64, payments []*POSPayment, fallback uint64) (map[uint64]amount.Amount, error) {
	mapping := map[string]uint64{}
	page, err := s.payMethods.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	for _, item := range page.Items {
		mapping[item.Method] = item.AccountID
	}
	byAccount := map[uint64]amount.Amount{}
	for _, payment := range payments {
		accountID := fallback
		if mapped, ok := mapping[payment.Method]; ok {
			accountID = mapped
		}
		byAccount[accountID] = byAccount[accountID].Add(amount.FromFloat64(payment.Amount))
	}
	return byAccount, nil
}

func (s POSService) receivableAccount(ctx context.Context, organizationID uint64) (uint64, error) {
	accounts, err := s.accounts.List(ctx, &query.Query{Filters: []query.Filter{{Field: "type", Operator: query.Equal, Value: accountTypeReceivable}}})
	if err != nil {
		return 0, err
	}
	for _, account := range accounts.Items {
		if account.OrganizationID == organizationID && account.Active {
			return account.ID, nil
		}
	}
	return 0, ErrConfigJournal
}

func (s POSService) warehouseStockLocation(ctx context.Context, organizationID, warehouseID uint64) (uint64, error) {
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "warehouse_id", Operator: query.Equal, Value: warehouseID}}})
	if err != nil {
		return 0, err
	}
	for _, location := range locations.Items {
		if location.Usage == locationUsageInternal && location.OrganizationID != nil && *location.OrganizationID == organizationID {
			return location.ID, nil
		}
	}
	return 0, ErrOrderNoStock
}

func (s POSService) customerLocation(ctx context.Context, organizationID uint64) (uint64, error) {
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "usage", Operator: query.Equal, Value: locationUsageCustomer}}})
	if err != nil {
		return 0, err
	}
	for _, location := range locations.Items {
		if location.OrganizationID != nil && *location.OrganizationID == organizationID {
			return location.ID, nil
		}
	}
	return 0, ErrOrderNoStock
}
