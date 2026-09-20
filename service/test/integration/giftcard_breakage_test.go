//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestGiftCardBreakage_ForfeitsExpiredCardsWithGLPosting(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	svc := giftcard.NewGiftCardService(
		giftcard.NewGiftCardDAO(testDB),
		giftcard.NewGiftCardTransactionDAO(testDB),
		fx.postingSvc,
		db.NewDBTransactioner(testDB),
	)

	past := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	future := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	if _, err := svc.Issue(ctx, giftcard.IssueRequest{
		OrganizationID: fx.orgID, Code: "BREAK-001", Amount: 100, CurrencyCode: "USD",
		ExpiryDate: &past, JournalID: fx.journalID, CashAccountID: fx.bankID, LiabilityAccountID: fx.liabilityID, Date: asOf,
	}); err != nil {
		t.Fatalf("issue expired card failed: %v", err)
	}
	if _, err := svc.Issue(ctx, giftcard.IssueRequest{
		OrganizationID: fx.orgID, Code: "LIVE-001", Amount: 100, CurrencyCode: "USD",
		ExpiryDate: &future, JournalID: fx.journalID, CashAccountID: fx.bankID, LiabilityAccountID: fx.liabilityID, Date: asOf,
	}); err != nil {
		t.Fatalf("issue live card failed: %v", err)
	}
	if balance := glBalance(t, fx.orgID, fx.liabilityID); balance != -200 {
		t.Fatalf("liability GL = %v, want -200 (credit) after issue", balance)
	}

	forfeited, err := svc.ForfeitExpired(ctx, giftcard.ForfeitExpiredRequest{
		OrganizationID: fx.orgID, AsOf: asOf, Date: asOf, JournalID: fx.journalID, LiabilityAccountID: fx.liabilityID, IncomeAccountID: fx.otherIncomeID,
	})
	if err != nil {
		t.Fatalf("forfeit expired failed: %v", err)
	}
	if len(forfeited) != 1 {
		t.Fatalf("forfeited = %d, want 1", len(forfeited))
	}
	if forfeited[0].Balance != 0 || forfeited[0].State != giftcard.GiftCardStateExpired {
		t.Errorf("forfeited card = %+v, want balance 0 / expired", forfeited[0])
	}

	cardDAO := giftcard.NewGiftCardDAO(testDB)
	live, err := cardDAO.Search(ctx, "code", "LIVE-001")
	if err != nil {
		t.Fatalf("find live card failed: %v", err)
	}
	if live.State != giftcard.GiftCardStateActive || live.Balance != 100 {
		t.Errorf("live card = %+v, want active with balance 100", live)
	}

	if balance := glBalance(t, fx.orgID, fx.liabilityID); balance != -100 {
		t.Fatalf("liability GL = %v, want -100 (credit) after forfeit", balance)
	}
	if balance := glBalance(t, fx.orgID, fx.otherIncomeID); balance != -100 {
		t.Fatalf("other income GL = %v, want -100 (credit) after forfeit", balance)
	}

	expired, err := cardDAO.Search(ctx, "code", "BREAK-001")
	if err != nil {
		t.Fatalf("find expired card failed: %v", err)
	}
	transactions, err := giftcard.NewGiftCardTransactionDAO(testDB).ListByGiftCard(ctx, expired.ID)
	if err != nil {
		t.Fatalf("list transactions failed: %v", err)
	}
	if len(transactions) != 2 || transactions[1].Type != giftcard.TransactionForfeit {
		t.Fatalf("transactions = %+v, want issue + forfeit", transactions)
	}
	if transactions[1].MovementID == nil {
		t.Errorf("forfeit transaction missing account movement reference")
	}

	again, err := svc.ForfeitExpired(ctx, giftcard.ForfeitExpiredRequest{
		OrganizationID: fx.orgID, AsOf: asOf, Date: asOf, JournalID: fx.journalID, LiabilityAccountID: fx.liabilityID, IncomeAccountID: fx.otherIncomeID,
	})
	if err != nil {
		t.Fatalf("second forfeit run failed: %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("second forfeit = %d, want 0 (idempotent)", len(again))
	}
}
