package giftcard

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestGiftCardService_Issue_RejectsNonPositiveAmount(t *testing.T) {
	svc := NewGiftCardService(GiftCardDAOMock{}, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Issue(context.Background(), IssueRequest{Amount: 0})
	if helper.AssertError(t, err, true, ErrGiftCardAmountInvalid) {
		return
	}
}

func TestGiftCardService_Issue_RequiresAccounts(t *testing.T) {
	svc := NewGiftCardService(GiftCardDAOMock{}, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Issue(context.Background(), IssueRequest{Amount: 100, CashAccountID: 200})
	if helper.AssertError(t, err, true, ErrGiftCardAccounts) {
		return
	}
}

func TestGiftCardService_Issue_GeneratesCodeAndDefaultsDate(t *testing.T) {
	ctx := context.Background()
	var createdCode string
	svc := NewGiftCardService(
		GiftCardDAOMock{
			CRUDMock: dao.CRUDMock[GiftCard]{
				SearchFunc: func(_ context.Context, _ string, _ any) (*GiftCard, error) { return nil, nil },
			},
			CreateTxFunc: func(_ context.Context, _ *gorm.DB, card *GiftCard) (*GiftCard, error) {
				createdCode = card.Code
				card.ID = 42
				return card, nil
			},
		},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		}},
		TransactionerMock{},
	)

	card, err := svc.Issue(ctx, IssueRequest{OrganizationID: 10, Amount: 100, JournalID: 3, CashAccountID: 200, LiabilityAccountID: 300})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(createdCode) != 12 || card.Code != createdCode {
		t.Errorf("code = %q, want 12-char generated code", card.Code)
	}
	if card.State != GiftCardStateActive {
		t.Errorf("card = %+v, want active", card)
	}
}

func TestGiftCardService_Issue_RejectsDuplicateGeneratedCode(t *testing.T) {
	existing := &GiftCard{Base: model.Base{ID: 9}}
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*GiftCard, error) { return existing, nil },
		}},
		GiftCardTransactionDAOMock{},
		PosterMock{},
		TransactionerMock{},
	)

	_, err := svc.Issue(context.Background(), IssueRequest{OrganizationID: 10, Amount: 100, JournalID: 3, CashAccountID: 200, LiabilityAccountID: 300})
	if helper.AssertError(t, err, true, ErrGiftCardAlreadyExists) {
		return
	}
}

func TestGiftCardService_Issue_RejectsProvidedDuplicateCode(t *testing.T) {
	existing := &GiftCard{Base: model.Base{ID: 9}, Code: "GC-X"}
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*GiftCard, error) { return existing, nil },
		}},
		GiftCardTransactionDAOMock{},
		PosterMock{},
		TransactionerMock{},
	)

	_, err := svc.Issue(context.Background(), IssueRequest{OrganizationID: 10, Code: "GC-X", Amount: 100, JournalID: 3, CashAccountID: 200, LiabilityAccountID: 300})
	if helper.AssertError(t, err, true, ErrGiftCardAlreadyExists) {
		return
	}
}

func TestGiftCardService_Issue_PropagatesSearchError(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*GiftCard, error) { return nil, errors.New("db down") },
		}},
		GiftCardTransactionDAOMock{},
		PosterMock{},
		TransactionerMock{},
	)

	_, err := svc.Issue(context.Background(), IssueRequest{OrganizationID: 10, Code: "GC-X", Amount: 100, JournalID: 3, CashAccountID: 200, LiabilityAccountID: 300})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestGiftCardService_Issue_PropagatesPostingError(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*GiftCard, error) { return nil, nil },
		}},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return nil, errors.New("db down")
		}},
		TransactionerMock{},
	)

	_, err := svc.Issue(context.Background(), IssueRequest{OrganizationID: 10, Amount: 100, JournalID: 3, CashAccountID: 200, LiabilityAccountID: 300})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestGiftCardService_Issue_PropagatesCreateError(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{
			CRUDMock: dao.CRUDMock[GiftCard]{
				SearchFunc: func(_ context.Context, _ string, _ any) (*GiftCard, error) { return nil, nil },
			},
			CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *GiftCard) (*GiftCard, error) { return nil, errors.New("db down") },
		},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		}},
		TransactionerMock{},
	)

	_, err := svc.Issue(context.Background(), IssueRequest{OrganizationID: 10, Amount: 100, JournalID: 3, CashAccountID: 200, LiabilityAccountID: 300})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestGiftCardService_Issue_MapsDuplicateCodeRace(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{
			CRUDMock: dao.CRUDMock[GiftCard]{
				SearchFunc: func(_ context.Context, _ string, _ any) (*GiftCard, error) { return nil, nil },
			},
			CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *GiftCard) (*GiftCard, error) {
				return nil, &pgconn.PgError{Code: "23505"}
			},
		},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		}},
		TransactionerMock{},
	)

	_, err := svc.Issue(context.Background(), IssueRequest{OrganizationID: 10, Code: "GIFT-001", Amount: 100, JournalID: 3, CashAccountID: 200, LiabilityAccountID: 300})
	if helper.AssertError(t, err, true, ErrGiftCardAlreadyExists) {
		return
	}
}

func TestGiftCardService_Adjust_RejectsZeroDelta(t *testing.T) {
	svc := NewGiftCardService(GiftCardDAOMock{}, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Adjust(context.Background(), AdjustRequest{GiftCardID: 42, Delta: 0, JournalID: 3, LiabilityAccountID: 300, OffsetAccountID: 500})
	if helper.AssertError(t, err, true, ErrGiftCardAmountInvalid) {
		return
	}
}

func TestGiftCardService_Adjust_RequiresAccounts(t *testing.T) {
	svc := NewGiftCardService(GiftCardDAOMock{}, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Adjust(context.Background(), AdjustRequest{GiftCardID: 42, Delta: 10, JournalID: 3, LiabilityAccountID: 300})
	if helper.AssertError(t, err, true, ErrGiftCardLiabilityOnly) {
		return
	}
}

func TestGiftCardService_Adjust_CardNotFound(t *testing.T) {
	svc := NewGiftCardService(GiftCardDAOMock{}, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Adjust(context.Background(), AdjustRequest{GiftCardID: 42, Delta: 10, JournalID: 3, LiabilityAccountID: 300, OffsetAccountID: 500})
	if helper.AssertError(t, err, true, ErrGiftCardNotFound) {
		return
	}
}

func TestGiftCardService_Adjust_RejectsNegativeBalance(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
			return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 10}, nil
		}}},
		GiftCardTransactionDAOMock{},
		PosterMock{},
		TransactionerMock{},
	)

	_, err := svc.Adjust(context.Background(), AdjustRequest{GiftCardID: 42, Delta: -30, JournalID: 3, LiabilityAccountID: 300, OffsetAccountID: 500})
	if helper.AssertError(t, err, true, ErrGiftCardInsufficient) {
		return
	}
}

func TestGiftCardService_Adjust_PositiveDeltaPostsLiability(t *testing.T) {
	ctx := context.Background()
	var posted accounting.PostRequest
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
			return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 100}, nil
		}}},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = request
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		}},
		TransactionerMock{},
	)

	card, err := svc.Adjust(ctx, AdjustRequest{GiftCardID: 42, Delta: 25, JournalID: 3, LiabilityAccountID: 300, OffsetAccountID: 500, Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Balance != 125 {
		t.Errorf("balance = %v, want 125", card.Balance)
	}
	if posted.Lines[0].AccountID != 500 || !posted.Lines[0].Debit.Equal(amount.FromFloat64(25)) {
		t.Errorf("offset line = %+v, want Dr 25 on 500", posted.Lines[0])
	}
	if posted.Lines[1].AccountID != 300 || !posted.Lines[1].Credit.Equal(amount.FromFloat64(25)) {
		t.Errorf("liability line = %+v, want Cr 25 on 300", posted.Lines[1])
	}
}

func TestGiftCardService_Adjust_NegativeDeltaPostsLiability(t *testing.T) {
	ctx := context.Background()
	var posted accounting.PostRequest
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
			return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 100}, nil
		}}},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = request
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		}},
		TransactionerMock{},
	)

	card, err := svc.Adjust(ctx, AdjustRequest{GiftCardID: 42, Delta: -25, JournalID: 3, LiabilityAccountID: 300, OffsetAccountID: 500, Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Balance != 75 {
		t.Errorf("balance = %v, want 75", card.Balance)
	}
	if posted.Lines[0].AccountID != 300 || !posted.Lines[0].Debit.Equal(amount.FromFloat64(25)) {
		t.Errorf("liability line = %+v, want Dr 25 on 300", posted.Lines[0])
	}
	if posted.Lines[1].AccountID != 500 || !posted.Lines[1].Credit.Equal(amount.FromFloat64(25)) {
		t.Errorf("offset line = %+v, want Cr 25 on 500", posted.Lines[1])
	}
}

func TestGiftCardService_Adjust_ZeroesBalanceMarksUsed(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
			return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 25}, nil
		}}},
		GiftCardTransactionDAOMock{},
		PosterMock{},
		TransactionerMock{},
	)

	card, err := svc.Adjust(context.Background(), AdjustRequest{GiftCardID: 42, Delta: -25, JournalID: 3, LiabilityAccountID: 300, OffsetAccountID: 500})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Balance != 0 || card.State != GiftCardStateUsed {
		t.Errorf("card = %+v, want used with zero balance", card)
	}
}

func TestGiftCardService_Adjust_PropagatesPostingError(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
			return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 100}, nil
		}}},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return nil, errors.New("db down")
		}},
		TransactionerMock{},
	)

	_, err := svc.Adjust(context.Background(), AdjustRequest{GiftCardID: 42, Delta: 25, JournalID: 3, LiabilityAccountID: 300, OffsetAccountID: 500})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestGiftCardService_Adjust_PropagatesUpdateError(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{
			CRUDMock: dao.CRUDMock[GiftCard]{FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
				return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 100}, nil
			}},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *GiftCard) (*GiftCard, error) { return nil, errors.New("db down") },
		},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		}},
		TransactionerMock{},
	)

	_, err := svc.Adjust(context.Background(), AdjustRequest{GiftCardID: 42, Delta: 25, JournalID: 3, LiabilityAccountID: 300, OffsetAccountID: 500})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestGiftCardService_ListTransactions_ReturnsTransactions(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
			return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 50}, nil
		}}},
		GiftCardTransactionDAOMock{},
		PosterMock{},
		TransactionerMock{},
	)

	items, err := svc.ListTransactions(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("items = %+v, want empty", items)
	}
}

func TestGiftCardService_ListTransactions_RejectsUnknownCard(t *testing.T) {
	svc := NewGiftCardService(GiftCardDAOMock{}, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.ListTransactions(context.Background(), 42)
	if helper.AssertError(t, err, true, ErrGiftCardNotFound) {
		return
	}
}

func TestGiftCardService_ForfeitExpired_DefaultsToNoDueCards(t *testing.T) {
	svc := NewGiftCardService(GiftCardDAOMock{}, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})

	forfeited, err := svc.ForfeitExpired(context.Background(), ForfeitExpiredRequest{
		OrganizationID: 10, AsOf: time.Now(), JournalID: 3, LiabilityAccountID: 300, IncomeAccountID: 400,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forfeited) != 0 {
		t.Errorf("forfeited = %d, want 0", len(forfeited))
	}
}

func TestGenerateGiftCardCode_ProducesValidCodes(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := generateGiftCardCode()
		if len(code) != 12 {
			t.Fatalf("code = %q, want 12 chars", code)
		}
		for _, r := range code {
			if !strings.ContainsRune(codeAlphabet, r) {
				t.Fatalf("code = %q, contains invalid char %q", code, r)
			}
		}
	}
}
