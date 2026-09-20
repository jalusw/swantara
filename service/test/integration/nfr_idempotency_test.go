//go:build integration

package integration

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestDeferralRecognitionResumeNoDuplicate(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}
	bsAccount, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2400", Name: "Deferred Revenue", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create bs account failed: %v", err)
	}
	plAccount, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4000", Name: "Sales Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create pl account failed: %v", err)
	}
	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "Deferral Journal", Code: helper.Ptr("DEF"),
		Type: "general", DefaultAccountID: helper.Ptr(bsAccount.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}
	raw, err := json.Marshal(journal.ID)
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}
	if _, err := dao.NewBase[reference.SystemConfig](testDB).Create(ctx, &reference.SystemConfig{
		OrganizationID: helper.Ptr(org.ID), Key: "subscription.journal_id", Value: raw,
	}); err != nil {
		t.Fatalf("create system config failed: %v", err)
	}

	journalEntryDAO := accounting.NewJournalEntryDAO(testDB)
	postingSvc := accounting.NewPostingService(journalEntryDAO)
	deferralSvc := accounting.NewDeferralService(
		accounting.NewDeferredScheduleDAO(testDB),
		accounting.NewDeferredScheduleLineDAO(testDB),
		postingSvc,
		subscription.NewSubscriptionConfigSource(dao.NewBase[reference.SystemConfig](testDB)),
		db.NewDBTransactioner(testDB),
	)

	dateStart := time.Now().UTC().Truncate(24 * time.Hour)
	schedule, err := deferralSvc.Create(ctx, accounting.CreateScheduleRequest{
		OrganizationID:        org.ID,
		Type:                  accounting.DeferredTypeDeferredRevenue,
		SourceType:            "invoice_line",
		SourceID:              101,
		TotalAmount:           300,
		BalanceSheetAccountID: bsAccount.ID,
		PLAccountID:           plAccount.ID,
		Method:                accounting.DeferredMethodLinear,
		DateStart:             dateStart,
		DateEnd:               dateStart.AddDate(0, 3, 0),
		Periods:               3,
	})
	if err != nil {
		t.Fatalf("create schedule failed: %v", err)
	}

	asOf := dateStart.AddDate(0, 4, 0)
	posted, err := deferralSvc.RecognizeDue(ctx, nil, asOf)
	if err != nil {
		t.Fatalf("first recognize failed: %v", err)
	}
	if posted != 3 {
		t.Fatalf("first recognize = %d, want 3", posted)
	}
	var firstMoves int64
	if err := testDB.WithContext(ctx).Model(&accounting.JournalEntry{}).Count(&firstMoves).Error; err != nil {
		t.Fatalf("count movements failed: %v", err)
	}

	replayed, err := deferralSvc.RecognizeDue(ctx, nil, asOf)
	if err != nil {
		t.Fatalf("resumed recognize failed: %v", err)
	}
	if replayed != 0 {
		t.Errorf("resumed recognize = %d, want 0", replayed)
	}
	var secondMoves int64
	if err := testDB.WithContext(ctx).Model(&accounting.JournalEntry{}).Count(&secondMoves).Error; err != nil {
		t.Fatalf("count movements failed: %v", err)
	}
	if secondMoves != firstMoves {
		t.Errorf("movements after resume = %d, want %d (no duplicates)", secondMoves, firstMoves)
	}

	updated, err := deferralSvc.Get(ctx, org.ID, schedule.ID)
	if err != nil {
		t.Fatalf("get schedule failed: %v", err)
	}
	if updated.State != accounting.DeferredStateDone || updated.RecognizedAmount != 300 {
		t.Errorf("schedule = %+v, want done with 300 recognized", updated)
	}
}

func TestSubscriptionBillingResumeNoDuplicate(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}
	receivable, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1200", Name: "Receivable", Type: "receivable", Active: true,
	})
	if err != nil {
		t.Fatalf("create receivable account failed: %v", err)
	}
	income, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4000", Name: "Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create income account failed: %v", err)
	}
	category, err := dao.NewBase[reference.ItemCategory](testDB).Create(ctx, &reference.ItemCategory{
		Name:            "Service",
		IncomeAccountID: helper.Ptr(income.ID),
		CostMethod:      helper.Ptr("average"),
	})
	if err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	template, err := products.NewItemDAO(testDB).Create(ctx, &products.Item{
		OrganizationID: helper.Ptr(org.ID), Name: "Plan Seat", CategoryID: helper.Ptr(category.ID),
		Type: "service", IsPurchasable: true, IsSellable: true, Tracking: "none",
	})
	if err != nil {
		t.Fatalf("create template failed: %v", err)
	}
	variant, err := products.NewItemVariantDAO(testDB).Create(ctx, &products.ItemVariant{
		ItemID: template.ID, Active: true,
	})
	if err != nil {
		t.Fatalf("create variant failed: %v", err)
	}
	contact, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(org.ID), Name: gofakeit.Company(),
	})
	if err != nil {
		t.Fatalf("create contact failed: %v", err)
	}
	price_book, err := products.NewPriceBookDAO(testDB).Create(ctx, &products.PriceBook{
		Name: "Main", CurrencyCode: helper.Ptr("IDR"), OrganizationID: helper.Ptr(org.ID), Active: true,
	})
	if err != nil {
		t.Fatalf("create price_book failed: %v", err)
	}
	plan, err := dao.NewBase[reference.SubscriptionPlan](testDB).Create(ctx, &reference.SubscriptionPlan{
		Name: "Monthly", RecurringInterval: "month", RecurringCount: 1,
	})
	if err != nil {
		t.Fatalf("create plan failed: %v", err)
	}
	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "Sales Journal", Code: helper.Ptr("SAL"),
		Type: "sale", DefaultAccountID: helper.Ptr(receivable.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}
	raw, err := json.Marshal(journal.ID)
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}
	if _, err := dao.NewBase[reference.SystemConfig](testDB).Create(ctx, &reference.SystemConfig{
		OrganizationID: helper.Ptr(org.ID), Key: "subscription.journal_id", Value: raw,
	}); err != nil {
		t.Fatalf("create system config failed: %v", err)
	}
	if err := testDB.WithContext(ctx).Create(&sequence.DocumentSequence{
		OrganizationID: org.ID, Code: accounting.SequenceInvoiceCode, NextNumber: 1, Padding: 5,
	}).Error; err != nil {
		t.Fatalf("create invoice sequence failed: %v", err)
	}

	postingSvc := accounting.NewPostingService(accounting.NewJournalEntryDAO(testDB))
	invoiceSvc := accounting.NewInvoiceService(
		accounting.NewInvoiceDAO(testDB),
		accounting.NewInvoiceLineDAO(testDB),
		accounting.NewInvoiceTaxDAO(testDB),
		postingSvc,
		dao.NewBase[reference.Account](testDB),
		dao.NewBase[reference.Tax](testDB),
		sequence.NewSequenceService(sequence.NewDAO(testDB)),
		db.NewDBTransactioner(testDB),
	)
	configSource := subscription.NewSubscriptionConfigSource(dao.NewBase[reference.SystemConfig](testDB))
	deferralSvc := accounting.NewDeferralService(
		accounting.NewDeferredScheduleDAO(testDB),
		accounting.NewDeferredScheduleLineDAO(testDB),
		postingSvc,
		configSource,
		db.NewDBTransactioner(testDB),
	)
	subscriptionSvc := subscription.NewSubscriptionService(
		subscription.NewSubscriptionDAO(testDB),
		subscription.NewSubscriptionLineDAO(testDB),
		dao.NewBase[reference.SubscriptionPlan](testDB),
		contacts.NewContactDAO(testDB),
		products.NewPriceBookDAO(testDB),
		products.NewProductService(
			products.NewItemDAO(testDB),
			products.NewItemVariantDAO(testDB),
			dao.NewBase[reference.ItemCategory](testDB),
			products.NewPriceBookDAO(testDB),
			products.NewPriceRuleDAO(testDB),
		),
		invoiceSvc,
		deferralSvc,
		configSource,
		db.NewDBTransactioner(testDB),
	)

	dueDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sub, err := subscription.NewSubscriptionDAO(testDB).Create(ctx, &subscription.Subscription{
		OrganizationID:  helper.Ptr(org.ID),
		Name:            "Resume Test",
		ContactID:       helper.Ptr(contact.ID),
		PlanID:          helper.Ptr(plan.ID),
		PriceBookID:     helper.Ptr(price_book.ID),
		CurrencyCode:    helper.Ptr("IDR"),
		DateStart:       helper.Ptr(dueDate),
		NextInvoiceDate: helper.Ptr(dueDate),
		State:           subscription.SubscriptionStateActive,
	})
	if err != nil {
		t.Fatalf("create subscription failed: %v", err)
	}
	if _, err := subscription.NewSubscriptionLineDAO(testDB).Create(ctx, &subscription.SubscriptionLine{
		SubscriptionID: sub.ID, ItemID: helper.Ptr(variant.ID), Qty: 1, UnitPrice: 120,
	}); err != nil {
		t.Fatalf("create subscription line failed: %v", err)
	}

	asOf := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	invoice, err := subscriptionSvc.GenerateNextInvoice(ctx, sub.ID, asOf)
	if err != nil {
		t.Fatalf("first billing failed: %v", err)
	}
	if invoice.ID == 0 {
		t.Errorf("first billing produced no invoice")
	}

	replayed, err := subscriptionSvc.GenerateNextInvoice(ctx, sub.ID, asOf)
	if !errors.Is(err, subscription.ErrSubscriptionNotDue) {
		t.Errorf("resume error = %v, want ErrSubscriptionNotDue (no duplicate invoice)", err)
	}
	if replayed != nil {
		t.Errorf("resume produced a second invoice, want none")
	}

	invoices := []*accounting.Invoice{}
	if err := testDB.WithContext(ctx).Find(&invoices).Error; err != nil {
		t.Fatalf("list invoices failed: %v", err)
	}
	if len(invoices) != 1 {
		t.Errorf("invoices after resume = %d, want 1 (no duplicates)", len(invoices))
	}

	after, err := subscription.NewSubscriptionDAO(testDB).Find(ctx, sub.ID)
	if err != nil {
		t.Fatalf("find subscription failed: %v", err)
	}
	if after.NextInvoiceDate == nil || !after.NextInvoiceDate.Equal(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("next_invoice_date = %v, want 2026-02-01", after.NextInvoiceDate)
	}
}
