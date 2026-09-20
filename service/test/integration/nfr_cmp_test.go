//go:build integration

package integration

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/audit"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
	"github.com/jalusw/swantara/apps/service/test/testutil"
	"gorm.io/gorm"
)

func TestPostedMoveImmutableAndAudited(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := model.ContextWithActor(testutil.SystemContext(), fx.actorID)

	auditSvc := audit.NewAuditLogService(audit.NewLogDAO(testDB))
	logs := audit.NewLogDAO(testDB)

	orgs := dao.NewBase[reference.Organization](testDB)
	org, err := orgs.Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	org.Name = gofakeit.Company()
	if _, err := orgs.Update(ctx, org); err != nil {
		t.Fatalf("update organization failed: %v", err)
	}

	orgLogsPage, err := logs.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "table_name", Operator: query.Equal, Value: "organizations"},
		{Field: "record_id", Operator: query.Equal, Value: org.ID},
	}})
	if err != nil {
		t.Fatalf("list organization audit logs failed: %v", err)
	}
	orgLogs := orgLogsPage.Items
	if len(orgLogs) != 2 {
		t.Fatalf("expected 2 audit logs for the edited organization, got %d", len(orgLogs))
	}
	for _, entry := range orgLogs {
		if entry.ChangedBy != fx.actorID {
			t.Fatalf("expected actor %d on audit log, got %d", fx.actorID, entry.ChangedBy)
		}
		if entry.RecordID != org.ID {
			t.Fatalf("expected record id %d on audit log, got %d", org.ID, entry.RecordID)
		}
		if entry.Action != audit.ActionInsert && entry.Action != audit.ActionUpdate {
			t.Fatalf("expected insert/update action, got %q", entry.Action)
		}
		if len(entry.Diff) == 0 {
			t.Fatal("expected audit log to capture a diff")
		}
	}

	posted, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "NFR-CMP-001",
		Lines: []accounting.PostingLine{
			{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(100)},
			{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(100)},
		},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}

	auditedMoves := audit.NewAudited[accounting.JournalEntry](fx.moveDAO, auditSvc, "journal_entrys")
	posted.Name = helper.Ptr("tampered")
	if _, err := auditedMoves.Update(ctx, posted); !errors.Is(err, audit.ErrImmutable) {
		t.Fatalf("expected ErrImmutable when editing a posted movement, got %v", err)
	}

	moveLogsPage, err := logs.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "table_name", Operator: query.Equal, Value: "journal_entrys"},
		{Field: "record_id", Operator: query.Equal, Value: posted.ID},
	}})
	if err != nil {
		t.Fatalf("list movement audit logs failed: %v", err)
	}
	moveLogs := moveLogsPage.Items
	if len(moveLogs) != 1 {
		t.Fatalf("expected only the post insert audit log for the movement, got %d", len(moveLogs))
	}
	if moveLogs[0].Action != audit.ActionInsert {
		t.Fatalf("expected the rejected movement edit to leave no update log, got %q", moveLogs[0].Action)
	}
}

func TestMoveTraceableToOriginAndActor(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := model.ContextWithActor(testutil.SystemContext(), fx.actorID)

	posted, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Ref:            "SO-1001",
		OriginType:     "sale_order",
		OriginID:       1001,
		Description:    "NFR-CMP-002",
		Lines: []accounting.PostingLine{
			{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(100)},
			{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(100)},
		},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}

	if posted.PostedBy == nil || *posted.PostedBy != fx.actorID {
		t.Fatalf("expected posted_by %d, got %+v", fx.actorID, posted.PostedBy)
	}
	if posted.PostedAt == nil {
		t.Fatal("expected posted_at to be set")
	}
	if posted.OriginType == nil || *posted.OriginType != "sale_order" {
		t.Fatalf("expected origin_type sale_order, got %+v", posted.OriginType)
	}
	if posted.OriginID == nil || *posted.OriginID != 1001 {
		t.Fatalf("expected origin_id 1001, got %+v", posted.OriginID)
	}

	reversal, err := fx.postingSvc.Reverse(ctx, accounting.ReverseRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "NFR-CMP-002 reversal",
		MovementID:     posted.ID,
	})
	if err != nil {
		t.Fatalf("reverse failed: %v", err)
	}
	if reversal.OriginID == nil || *reversal.OriginID != posted.ID {
		t.Fatalf("expected reversal to trace to original movement, got %+v", reversal.OriginID)
	}

	lines, err := fx.moveLineDAO.ListByMovement(ctx, posted.ID)
	if err != nil {
		t.Fatalf("list lines failed: %v", err)
	}
	for _, line := range lines {
		if line.Date.IsZero() {
			t.Fatal("expected each line to carry the movement date")
		}
	}
}

func TestTaxIsDataDrivenNotCountryHardcoded(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	outputTax, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: fx.orgID, Code: "2200", Name: "Output Tax", Type: "tax", Active: true,
	})
	if err != nil {
		t.Fatalf("create output tax account failed: %v", err)
	}

	whole, err := dao.NewBase[reference.Tax](testDB).Create(ctx, &reference.Tax{
		OrganizationID: helper.Ptr(fx.orgID),
		Name:           "VAT",
		Amount:         helper.Ptr(10.0),
		Type:           reference.TaxTypePercent,
		Scope:          reference.TaxScopeSale,
		TaxAccountID:   helper.Ptr(outputTax.ID),
		Active:         true,
	})
	if err != nil {
		t.Fatalf("create whole-percent tax failed: %v", err)
	}
	fractional, err := dao.NewBase[reference.Tax](testDB).Create(ctx, &reference.Tax{
		OrganizationID: helper.Ptr(fx.orgID),
		Name:           "Digital Levy",
		Amount:         helper.Ptr(12.5),
		Type:           reference.TaxTypePercent,
		Scope:          reference.TaxScopeSale,
		TaxAccountID:   helper.Ptr(outputTax.ID),
		Active:         true,
	})
	if err != nil {
		t.Fatalf("create fractional tax failed: %v", err)
	}

	loadedWhole, err := dao.NewBase[reference.Tax](testDB).Find(ctx, whole.ID)
	if err != nil {
		t.Fatalf("load whole tax failed: %v", err)
	}
	loadedFractional, err := dao.NewBase[reference.Tax](testDB).Find(ctx, fractional.ID)
	if err != nil {
		t.Fatalf("load fractional tax failed: %v", err)
	}
	if loadedWhole.Amount == nil || *loadedWhole.Amount != 10 {
		t.Fatalf("expected whole tax rate 10, got %+v", loadedWhole.Amount)
	}
	if loadedFractional.Amount == nil || *loadedFractional.Amount != 12.5 {
		t.Fatalf("expected fractional tax rate 12.5, got %+v", loadedFractional.Amount)
	}
	if loadedWhole.Type != reference.TaxTypePercent || loadedWhole.Scope != reference.TaxScopeSale {
		t.Fatalf("expected tax type/scope preserved, got %+v / %+v", loadedWhole.Type, loadedWhole.Scope)
	}
}

func TestNewEventTypePlugsIntoPostingEngine(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	first, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		OriginType:     "verification_event",
		OriginID:       4242,
		Description:    "NFR-MNT-002 event A",
		Lines: []accounting.PostingLine{
			{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(100)},
			{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(100)},
		},
	})
	if err != nil {
		t.Fatalf("post first event failed: %v", err)
	}
	second, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		OriginType:     "verification_event",
		OriginID:       4243,
		Description:    "NFR-MNT-002 event B",
		Lines: []accounting.PostingLine{
			{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(50)},
			{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(50)},
		},
	})
	if err != nil {
		t.Fatalf("post second event failed: %v", err)
	}

	for _, movement := range []*accounting.JournalEntry{first, second} {
		if movement.OriginType == nil || *movement.OriginType != "verification_event" {
			t.Fatalf("expected verification_event origin on movement %d, got %+v", movement.ID, movement.OriginType)
		}
		lines, err := fx.moveLineDAO.ListByMovement(ctx, movement.ID)
		if err != nil {
			t.Fatalf("list lines for movement %d failed: %v", movement.ID, err)
		}
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines on movement %d, got %d", movement.ID, len(lines))
		}
	}
}

func TestGlobalAuditPluginTracksMutationsAndSkipsTransientTables(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := model.ContextWithActor(testutil.SystemContext(), fx.actorID)

	logs := audit.NewLogDAO(testDB)
	accounts := dao.NewBase[reference.Account](testDB)

	account, err := accounts.Create(ctx, &reference.Account{
		OrganizationID: fx.orgID, Code: "2100", Name: "Supplier Payable", Type: "payable", Active: true,
	})
	if err != nil {
		t.Fatalf("create account failed: %v", err)
	}

	account.Name = "Trade Payable"
	if _, err := accounts.Update(ctx, account); err != nil {
		t.Fatalf("update account failed: %v", err)
	}

	if err := accounts.Delete(ctx, account.ID); err != nil {
		t.Fatalf("delete account failed: %v", err)
	}

	page, err := logs.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "table_name", Operator: query.Equal, Value: "accounts"},
		{Field: "record_id", Operator: query.Equal, Value: account.ID},
	}})
	if err != nil {
		t.Fatalf("list account audit logs failed: %v", err)
	}
	entries := page.Items
	if len(entries) != 3 {
		t.Fatalf("expected insert, update and delete audit logs, got %d", len(entries))
	}
	actions := map[audit.Action]bool{}
	for _, entry := range entries {
		if entry.ChangedBy != fx.actorID {
			t.Fatalf("expected actor %d on audit log, got %d", fx.actorID, entry.ChangedBy)
		}
		if entry.RecordID != account.ID {
			t.Fatalf("expected record id %d on audit log, got %d", account.ID, entry.RecordID)
		}
		if len(entry.Diff) == 0 {
			t.Fatal("expected audit log to capture a diff")
		}
		actions[entry.Action] = true
	}
	for _, action := range []audit.Action{audit.ActionInsert, audit.ActionUpdate, audit.ActionDelete} {
		if !actions[action] {
			t.Fatalf("expected %q audit log to exist", action)
		}
	}

	if _, err := dao.NewBase[xtradata.IdempotencyKey](testDB).Create(ctx, &xtradata.IdempotencyKey{
		OrganizationID: helper.Ptr(fx.orgID),
		Key:            "audit-skip-check",
		Resource:       "test",
		StatusCode:     200,
	}); err != nil {
		t.Fatalf("create idempotency key failed: %v", err)
	}
	skipped, err := logs.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "table_name", Operator: query.Equal, Value: "idempotency_keys"},
	}})
	if err != nil {
		t.Fatalf("list idempotency audit logs failed: %v", err)
	}
	if len(skipped.Items) != 0 {
		t.Fatalf("expected transient tables to be skipped by the audit plugin, got %d logs", len(skipped.Items))
	}
}

func TestAuditLogsJoinBusinessTransactionAndRedactSecrets(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := model.ContextWithActor(testutil.SystemContext(), fx.actorID)

	logs := audit.NewLogDAO(testDB)

	rolledBackOrgID := uint64(0)
	err := testDB.Transaction(func(tx *gorm.DB) error {
		org, err := dao.NewBase[reference.Organization](tx).Create(ctx, &reference.Organization{
			Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
		})
		if err != nil {
			return err
		}
		rolledBackOrgID = org.ID
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("expected forced transaction rollback")
	}

	page, err := logs.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "table_name", Operator: query.Equal, Value: "organizations"},
		{Field: "record_id", Operator: query.Equal, Value: rolledBackOrgID},
	}})
	if err != nil {
		t.Fatalf("list rolled-back audit logs failed: %v", err)
	}
	if len(page.Items) != 0 {
		t.Fatalf("expected no audit row after rollback, got %d", len(page.Items))
	}

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create committed organization failed: %v", err)
	}
	page, err = logs.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "table_name", Operator: query.Equal, Value: "organizations"},
		{Field: "record_id", Operator: query.Equal, Value: org.ID},
	}})
	if err != nil {
		t.Fatalf("list committed audit logs failed: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected one audit row after commit, got %d", len(page.Items))
	}

	card, err := giftcard.NewGiftCardDAO(testDB).Create(ctx, &giftcard.GiftCard{
		OrganizationID: helper.Ptr(fx.orgID),
		Code:           "REDACT-ME-001",
		InitialAmount:  100,
		Balance:        100,
		CurrencyCode:   "USD",
		State:          giftcard.GiftCardStateActive,
	})
	if err != nil {
		t.Fatalf("create gift card failed: %v", err)
	}
	page, err = logs.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "table_name", Operator: query.Equal, Value: "gift_cards"},
		{Field: "record_id", Operator: query.Equal, Value: card.ID},
	}})
	if err != nil {
		t.Fatalf("list gift card audit logs failed: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected one gift card audit row, got %d", len(page.Items))
	}
	var diff map[string]any
	if err := json.Unmarshal(page.Items[0].Diff, &diff); err != nil {
		t.Fatalf("unmarshal gift card diff failed: %v", err)
	}
	if diff["code"] != audit.RedactedValue {
		t.Fatalf("expected redacted gift card code, got %#v", diff["code"])
	}
	if diff["balance"] != 100.0 {
		t.Fatalf("expected balance captured, got %#v", diff["balance"])
	}

	contact, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(fx.orgID),
		Name:           "Sensitive Supplier",
		Email:          helper.Ptr("supplier@example.com"),
		Phone:          helper.Ptr("+1-555-0100"),
		TaxID:          helper.Ptr("TAX-987654321"),
	})
	if err != nil {
		t.Fatalf("create contact failed: %v", err)
	}
	page, err = logs.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "table_name", Operator: query.Equal, Value: "contacts"},
		{Field: "record_id", Operator: query.Equal, Value: contact.ID},
	}})
	if err != nil {
		t.Fatalf("list contact audit logs failed: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected one contact audit row, got %d", len(page.Items))
	}
	diff = map[string]any{}
	if err := json.Unmarshal(page.Items[0].Diff, &diff); err != nil {
		t.Fatalf("unmarshal contact diff failed: %v", err)
	}
	for _, key := range []string{"email", "phone", "tax_id"} {
		if diff[key] != audit.RedactedValue {
			t.Errorf("expected redacted contact %s, got %#v", key, diff[key])
		}
	}
	if diff["name"] != "Sensitive Supplier" {
		t.Errorf("expected contact name captured, got %#v", diff["name"])
	}
}
