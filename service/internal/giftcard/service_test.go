package giftcard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestGiftCardService_Issue_PostsCashAndLiability(t *testing.T) {
	var posted accounting.PostRequest
	var createdID uint64
	svc := NewGiftCardService(
		GiftCardDAOMock{
			CRUDMock: dao.CRUDMock[GiftCard]{
				SearchFunc: func(_ context.Context, _ string, _ any) (*GiftCard, error) { return nil, nil },
			},
			CreateTxFunc: func(_ context.Context, _ *gorm.DB, card *GiftCard) (*GiftCard, error) {
				card.ID = 42
				createdID = card.ID
				return card, nil
			},
		},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = request
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		}},
		TransactionerMock{},
	)

	card, err := svc.Issue(context.Background(), IssueRequest{
		OrganizationID: 10, Code: "GIFT-001", Amount: 100,
		JournalID: 3, CashAccountID: 200, LiabilityAccountID: 300,
		Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.State != GiftCardStateActive || card.Balance != 100 {
		t.Errorf("card = %+v, want active with balance 100", card)
	}
	if len(posted.Lines) != 2 {
		t.Fatalf("posted = %+v, want 2 lines", posted.Lines)
	}
	if !posted.Lines[0].Debit.Equal(amount.FromFloat64(100)) || posted.Lines[0].AccountID != 200 {
		t.Errorf("cash line = %+v, want Dr 100 on 200", posted.Lines[0])
	}
	if !posted.Lines[1].Credit.Equal(amount.FromFloat64(100)) || posted.Lines[1].AccountID != 300 {
		t.Errorf("liability line = %+v, want Cr 100 on 300", posted.Lines[1])
	}
	if createdID == 0 {
		t.Errorf("card was not persisted")
	}
}

func TestGiftCardService_Redeem_ReducesBalanceAndPostsLiability(t *testing.T) {
	var posted accounting.PostRequest
	issued := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := NewGiftCardService(
		GiftCardDAOMock{
			CRUDMock: dao.CRUDMock[GiftCard]{
				FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
					return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 100, ExpiryDate: nil}, nil
				},
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, card *GiftCard) (*GiftCard, error) { return card, nil },
		},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = request
			return &accounting.JournalEntry{Base: model.Base{ID: 901}}, nil
		}},
		TransactionerMock{},
	)

	card, err := svc.Redeem(context.Background(), RedeemRequest{
		GiftCardID: 42, Amount: 30, JournalID: 3, RevenueAccountID: 400, LiabilityAccountID: 300,
		Date: issued,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Balance != 70 {
		t.Errorf("balance = %v, want 70", card.Balance)
	}
	if !posted.Lines[0].Debit.Equal(amount.FromFloat64(30)) || posted.Lines[0].AccountID != 300 {
		t.Errorf("liability line = %+v, want Dr 30 on 300", posted.Lines[0])
	}
	if !posted.Lines[1].Credit.Equal(amount.FromFloat64(30)) || posted.Lines[1].AccountID != 400 {
		t.Errorf("revenue line = %+v, want Cr 30 on 400", posted.Lines[1])
	}
}

func TestGiftCardService_Redeem_RejectsInsufficientBalance(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
			return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 20}, nil
		}}},
		GiftCardTransactionDAOMock{},
		PosterMock{},
		TransactionerMock{},
	)

	_, err := svc.Redeem(context.Background(), RedeemRequest{GiftCardID: 42, Amount: 30, JournalID: 3, RevenueAccountID: 400, LiabilityAccountID: 300})
	if !errors.Is(err, ErrGiftCardInsufficient) {
		t.Fatalf("err = %v, want ErrGiftCardInsufficient", err)
	}
}

func TestGiftCardService_Redeem_RejectsExpired(t *testing.T) {
	expired := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := NewGiftCardService(
		GiftCardDAOMock{CRUDMock: dao.CRUDMock[GiftCard]{FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
			return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateActive, Balance: 50, ExpiryDate: &expired}, nil
		}}},
		GiftCardTransactionDAOMock{},
		PosterMock{},
		TransactionerMock{},
	)

	_, err := svc.Redeem(context.Background(), RedeemRequest{
		GiftCardID: 42, Amount: 10, JournalID: 3, RevenueAccountID: 400, LiabilityAccountID: 300,
		Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ErrGiftCardExpired) {
		t.Fatalf("err = %v, want ErrGiftCardExpired", err)
	}
}

func TestGiftCardService_Refund_RestoresBalanceAndLiability(t *testing.T) {
	var posted accounting.PostRequest
	svc := NewGiftCardService(
		GiftCardDAOMock{
			CRUDMock: dao.CRUDMock[GiftCard]{
				FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) {
					return &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: GiftCardStateUsed, Balance: 0}, nil
				},
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, card *GiftCard) (*GiftCard, error) { return card, nil },
		},
		GiftCardTransactionDAOMock{},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = request
			return &accounting.JournalEntry{Base: model.Base{ID: 902}}, nil
		}},
		TransactionerMock{},
	)

	card, err := svc.Refund(context.Background(), RefundRequest{
		GiftCardID: 42, Amount: 10, JournalID: 3, RefundAccountID: 500, LiabilityAccountID: 300,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.State != GiftCardStateActive || card.Balance != 10 {
		t.Errorf("card = %+v, want active with balance 10", card)
	}
	if !posted.Lines[1].Credit.Equal(amount.FromFloat64(10)) || posted.Lines[1].AccountID != 300 {
		t.Errorf("liability line = %+v, want Cr 10 on 300", posted.Lines[1])
	}
}
