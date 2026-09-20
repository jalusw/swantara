package subscription

import (
	"context"
	"math"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type SubscriptionService struct {
	subscriptions SubscriptionDAO
	lines         SubscriptionLineDAO
	plans         dao.CRUD[reference.SubscriptionPlan]
	contacts      contacts.ContactDAO
	price_books   products.PriceBookDAO
	productSvc    IncomeAccountResolver
	invoices      InvoiceEngine
	deferrals     accounting.DeferralService
	configs       SubscriptionConfigSource
	tx            db.Transactioner
	now           func() time.Time
}

func NewSubscriptionService(
	subscriptions SubscriptionDAO,
	lines SubscriptionLineDAO,
	plans dao.CRUD[reference.SubscriptionPlan],
	contacts contacts.ContactDAO,
	price_books products.PriceBookDAO,
	productSvc IncomeAccountResolver,
	invoices InvoiceEngine,
	deferrals accounting.DeferralService,
	configs SubscriptionConfigSource,
	tx db.Transactioner,
) SubscriptionService {
	return SubscriptionService{
		subscriptions: subscriptions,
		lines:         lines,
		plans:         plans,
		contacts:      contacts,
		price_books:   price_books,
		productSvc:    productSvc,
		invoices:      invoices,
		deferrals:     deferrals,
		configs:       configs,
		tx:            tx,
		now:           time.Now,
	}
}

type CreateSubscriptionLineRequest struct {
	ItemID      uint64
	Qty         float64
	UnitPrice   float64
	DiscountPct float64
}

type CreateSubscriptionRequest struct {
	OrganizationID uint64
	Name           string
	ContactID      uint64
	PlanID         uint64
	PriceBookID    uint64
	CurrencyCode   string
	Lines          []CreateSubscriptionLineRequest
}

func (s SubscriptionService) Create(ctx context.Context, request CreateSubscriptionRequest) (*Subscription, error) {
	if request.ContactID == 0 {
		return nil, ErrSubscriptionContact
	}
	if request.PlanID == 0 {
		return nil, ErrSubscriptionNoPlan
	}
	if request.PriceBookID == 0 {
		return nil, ErrSubscriptionPriceBook
	}
	if request.CurrencyCode == "" {
		return nil, ErrSubscriptionCurrency
	}
	if len(request.Lines) == 0 {
		return nil, ErrSubscriptionNoLines
	}

	plan, err := s.plans.Find(ctx, request.PlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrSubscriptionPlanNotFound
	}
	if plan.OrganizationID != nil && *plan.OrganizationID != request.OrganizationID {
		return nil, ErrSubscriptionPlanOrganization
	}

	if err := s.validateContact(ctx, request.OrganizationID, request.ContactID); err != nil {
		return nil, err
	}
	if err := s.validatePriceBook(ctx, request.OrganizationID, request.PriceBookID); err != nil {
		return nil, err
	}

	total := amount.Zero()
	entities := make([]*SubscriptionLine, 0, len(request.Lines))
	for _, line := range request.Lines {
		if line.ItemID == 0 {
			return nil, ErrSubscriptionLineProduct
		}
		variantOrg, err := s.productSvc.ResolveVariantOrganization(ctx, line.ItemID)
		if err != nil {
			return nil, err
		}
		if variantOrg != nil && *variantOrg != request.OrganizationID {
			return nil, ErrSubscriptionLineProductOrganization
		}
		if !amount.FromFloat64(line.Qty).GreaterThan(amount.Zero()) {
			return nil, ErrSubscriptionLineQty
		}
		if amount.FromFloat64(line.UnitPrice).IsNegative() {
			return nil, ErrSubscriptionLinePrice
		}
		if line.DiscountPct < 0 || line.DiscountPct > 100 {
			return nil, ErrSubscriptionLineDiscount
		}
		subtotal := amount.FromFloat64(line.Qty).
			Mul(amount.FromFloat64(line.UnitPrice)).
			Mul(amount.FromFloat64(1 - line.DiscountPct/100)).
			Round(4)
		total = total.Add(subtotal)
		entities = append(entities, &SubscriptionLine{
			ItemID:      helper.Ptr(line.ItemID),
			Qty:         line.Qty,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
		})
	}

	periods := monthsPerPeriod(plan.RecurringInterval, plan.RecurringCount)
	mrr := total
	if periods > 0 {
		divided, err := total.Div(amount.FromFloat64(periods))
		if err != nil {
			return nil, err
		}
		mrr = divided
	}

	subscription := &Subscription{
		OrganizationID: helper.Ptr(request.OrganizationID),
		Name:           request.Name,
		ContactID:      helper.Ptr(request.ContactID),
		PlanID:         helper.Ptr(request.PlanID),
		PriceBookID:    helper.Ptr(request.PriceBookID),
		CurrencyCode:   helper.Ptr(request.CurrencyCode),
		State:          SubscriptionStateDraft,
		MRR:            mrr.Round(4).Float64(),
	}

	created, err := s.subscriptions.Create(ctx, subscription)
	if err != nil {
		return nil, err
	}
	for _, line := range entities {
		line.SubscriptionID = created.ID
		if _, err := s.lines.Create(ctx, line); err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (s SubscriptionService) Get(ctx context.Context, organizationID, subscriptionID uint64) (*Subscription, error) {
	subscription, err := s.subscriptions.Search(ctx, "id", subscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription == nil || subscription.OrganizationID == nil || *subscription.OrganizationID != organizationID {
		return nil, ErrSubscriptionNotFound
	}
	return subscription, nil
}

func (s SubscriptionService) List(ctx context.Context, organizationID uint64) ([]*Subscription, error) {
	page, err := s.subscriptions.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s SubscriptionService) Activate(ctx context.Context, organizationID, subscriptionID uint64) (*Subscription, error) {
	subscription, err := s.Get(ctx, organizationID, subscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription.State != SubscriptionStateDraft {
		return nil, ErrSubscriptionState
	}
	now := s.now().UTC()
	if subscription.DateStart == nil {
		subscription.DateStart = &now
	}
	if subscription.NextInvoiceDate == nil {
		next := now
		subscription.NextInvoiceDate = &next
	}
	subscription.State = SubscriptionStateActive
	return s.subscriptions.Update(ctx, subscription)
}

func (s SubscriptionService) Pause(ctx context.Context, organizationID, subscriptionID uint64) (*Subscription, error) {
	subscription, err := s.Get(ctx, organizationID, subscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription.State != SubscriptionStateActive {
		return nil, ErrSubscriptionState
	}
	subscription.State = SubscriptionStatePaused
	return s.subscriptions.Update(ctx, subscription)
}

func (s SubscriptionService) Resume(ctx context.Context, organizationID, subscriptionID uint64) (*Subscription, error) {
	subscription, err := s.Get(ctx, organizationID, subscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription.State != SubscriptionStatePaused {
		return nil, ErrSubscriptionState
	}
	subscription.State = SubscriptionStateActive
	return s.subscriptions.Update(ctx, subscription)
}

func (s SubscriptionService) Churn(ctx context.Context, organizationID, subscriptionID uint64) (*Subscription, error) {
	subscription, err := s.Get(ctx, organizationID, subscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription.State != SubscriptionStateActive && subscription.State != SubscriptionStatePaused {
		return nil, ErrSubscriptionState
	}
	now := s.now().UTC()
	subscription.State = SubscriptionStateChurned
	subscription.DateEnd = &now
	return s.subscriptions.Update(ctx, subscription)
}

func (s SubscriptionService) Close(ctx context.Context, organizationID, subscriptionID uint64) (*Subscription, error) {
	subscription, err := s.Get(ctx, organizationID, subscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription.State != SubscriptionStateActive && subscription.State != SubscriptionStatePaused {
		return nil, ErrSubscriptionState
	}
	now := s.now().UTC()
	subscription.State = SubscriptionStateClosed
	subscription.DateEnd = &now
	return s.subscriptions.Update(ctx, subscription)
}

type Metrics struct {
	MRR       float64
	ARR       float64
	Churned   int
	ChurnRate float64
	LTV       float64
}

func (s SubscriptionService) Metrics(ctx context.Context, organizationID uint64) (Metrics, error) {
	subscriptions, err := s.List(ctx, organizationID)
	if err != nil {
		return Metrics{}, err
	}
	metrics := Metrics{}
	recurring := 0
	for _, subscription := range subscriptions {
		switch subscription.State {
		case SubscriptionStateActive, SubscriptionStatePaused:
			metrics.MRR += subscription.MRR
			recurring++
		case SubscriptionStateChurned:
			metrics.Churned++
		}
	}
	metrics.ARR = metrics.MRR * 12
	base := recurring + metrics.Churned
	if base > 0 {
		metrics.ChurnRate = float64(metrics.Churned) / float64(base)
	}
	if metrics.ChurnRate > 0 {
		metrics.LTV = metrics.ARR / metrics.ChurnRate
	}
	return metrics, nil
}

func (s SubscriptionService) BillDue(ctx context.Context, asOf time.Time) (int, error) {
	due, err := s.subscriptions.ListDue(ctx, asOf)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, subscription := range due {
		if _, err := s.GenerateNextInvoice(ctx, subscription.ID, asOf); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s SubscriptionService) GenerateNextInvoice(ctx context.Context, subscriptionID uint64, asOf time.Time) (*accounting.Invoice, error) {
	subscription, err := s.subscriptions.Find(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription == nil || subscription.OrganizationID == nil {
		return nil, ErrSubscriptionNotFound
	}
	if subscription.State != SubscriptionStateActive {
		return nil, ErrSubscriptionState
	}
	if subscription.NextInvoiceDate == nil || subscription.NextInvoiceDate.After(asOf) {
		return nil, ErrSubscriptionNotDue
	}
	if subscription.ContactID == nil || subscription.PlanID == nil || subscription.PriceBookID == nil {
		return nil, ErrSubscriptionNotFound
	}

	plan, err := s.plans.Find(ctx, *subscription.PlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrSubscriptionPlanNotFound
	}
	if plan.OrganizationID != nil && *plan.OrganizationID != *subscription.OrganizationID {
		return nil, ErrSubscriptionPlanOrganization
	}

	if err := s.validateContact(ctx, *subscription.OrganizationID, *subscription.ContactID); err != nil {
		return nil, err
	}
	if err := s.validatePriceBook(ctx, *subscription.OrganizationID, *subscription.PriceBookID); err != nil {
		return nil, err
	}

	lines, err := s.lines.ListBySubscription(ctx, subscription.ID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrSubscriptionNoLines
	}

	organizationID := *subscription.OrganizationID
	journalID, err := s.configs.JournalID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if journalID == 0 {
		return nil, ErrSubscriptionConfig
	}
	deferredAccountID, err := s.configs.DeferredRevenueAccountID(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	invoiceLines := make([]accounting.InvoiceLineRequest, 0, len(lines))
	accounts := make(map[uint64]uint64, len(lines))
	for _, line := range lines {
		if line.ItemID == nil {
			return nil, ErrSubscriptionLineProduct
		}
		accountID, err := s.productSvc.ResolveIncomeAccount(ctx, *line.ItemID)
		if err != nil {
			return nil, err
		}
		accounts[line.ID] = accountID
		invoiceLines = append(invoiceLines, accounting.InvoiceLineRequest{
			ItemID:      line.ItemID,
			Qty:         line.Qty,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
			AccountID:   accountID,
		})
	}

	var deferredAccount *uint64
	if deferredAccountID != 0 {
		deferredAccount = helper.Ptr(deferredAccountID)
	}

	var created *accounting.Invoice
	invoiceDate := *subscription.NextInvoiceDate
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		invoice, err := s.invoices.CreateTx(ctx, tx, accounting.CreateInvoiceRequest{
			OrganizationID:  organizationID,
			JournalID:       journalID,
			ContactID:       *subscription.ContactID,
			Date:            invoiceDate,
			Reference:       subscription.Name,
			DeferredAccount: deferredAccount,
			Lines:           invoiceLines,
		})
		if err != nil {
			return err
		}
		created = invoice
		if deferredAccountID != 0 {
			if err := s.createDeferralSchedules(ctx, tx, subscription, lines, accounts, invoiceDate, deferredAccountID, plan); err != nil {
				return err
			}
		}
		next := advanceDate(invoiceDate, plan.RecurringInterval, plan.RecurringCount)
		subscription.NextInvoiceDate = &next
		if _, err := s.subscriptions.UpdateTx(ctx, tx, subscription); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s SubscriptionService) createDeferralSchedules(ctx context.Context, tx *gorm.DB, subscription *Subscription, lines []*SubscriptionLine, accounts map[uint64]uint64, invoiceDate time.Time, deferredAccountID uint64, plan *reference.SubscriptionPlan) error {
	periods := schedulePeriods(plan.RecurringInterval, plan.RecurringCount)
	dateEnd := invoiceDate.AddDate(0, periods, 0)
	for _, line := range lines {
		if line.ItemID == nil {
			return ErrSubscriptionLineProduct
		}
		subtotal := amount.FromFloat64(line.Qty).
			Mul(amount.FromFloat64(line.UnitPrice)).
			Mul(amount.FromFloat64(1 - line.DiscountPct/100)).
			Round(4)
		_, err := s.deferrals.CreateTx(ctx, tx, accounting.CreateScheduleRequest{
			OrganizationID:        *subscription.OrganizationID,
			Type:                  accounting.DeferredTypeDeferredRevenue,
			SourceType:            "subscription_line",
			SourceID:              line.ID,
			ContactID:             subscription.ContactID,
			ItemID:                line.ItemID,
			TotalAmount:           subtotal.Float64(),
			BalanceSheetAccountID: deferredAccountID,
			PLAccountID:           accounts[line.ID],
			Method:                accounting.DeferredMethodLinear,
			DateStart:             invoiceDate,
			DateEnd:               dateEnd,
			Periods:               periods,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s SubscriptionService) validateContact(ctx context.Context, organizationID, contactID uint64) error {
	contact, err := s.contacts.Find(ctx, contactID)
	if err != nil {
		return err
	}
	if contact == nil {
		return ErrSubscriptionContact
	}
	if contact.OrganizationID != nil && *contact.OrganizationID != organizationID {
		return ErrSubscriptionContactOrganization
	}
	return nil
}

func (s SubscriptionService) validatePriceBook(ctx context.Context, organizationID, price_bookID uint64) error {
	price_book, err := s.price_books.Find(ctx, price_bookID)
	if err != nil {
		return err
	}
	if price_book == nil {
		return ErrSubscriptionPriceBook
	}
	if price_book.OrganizationID != nil && *price_book.OrganizationID != organizationID {
		return ErrSubscriptionPriceBookOrganization
	}
	return nil
}

func monthsPerPeriod(interval string, count int) float64 {
	switch interval {
	case "day":
		return float64(count) / 30
	case "week":
		return float64(count) / 4
	case "year":
		return float64(count) * 12
	default:
		return float64(count)
	}
}

func schedulePeriods(interval string, count int) int {
	periods := int(math.Round(monthsPerPeriod(interval, count)))
	if periods < 1 {
		return 1
	}
	return periods
}

func advanceDate(date time.Time, interval string, count int) time.Time {
	switch interval {
	case "day":
		return date.AddDate(0, 0, count)
	case "week":
		return date.AddDate(0, 0, count*7)
	case "year":
		return date.AddDate(count, 0, 0)
	default:
		return date.AddDate(0, count, 0)
	}
}

type SubscriptionPlanService struct {
	plans dao.CRUD[reference.SubscriptionPlan]
}

func NewSubscriptionPlanService(plans dao.CRUD[reference.SubscriptionPlan]) SubscriptionPlanService {
	return SubscriptionPlanService{plans: plans}
}

func (s SubscriptionPlanService) List(ctx context.Context, q *query.Query) (*query.Page[reference.SubscriptionPlan], error) {
	return s.plans.List(ctx, q)
}

func (s SubscriptionPlanService) Find(ctx context.Context, id uint64) (*reference.SubscriptionPlan, error) {
	return s.plans.Find(ctx, id)
}

func (s SubscriptionPlanService) Create(ctx context.Context, plan *reference.SubscriptionPlan) (*reference.SubscriptionPlan, error) {
	return s.plans.Create(ctx, plan)
}

func (s SubscriptionPlanService) Update(ctx context.Context, plan *reference.SubscriptionPlan) (*reference.SubscriptionPlan, error) {
	return s.plans.Update(ctx, plan)
}

func (s SubscriptionPlanService) Delete(ctx context.Context, id uint64) error {
	return s.plans.Delete(ctx, id)
}
