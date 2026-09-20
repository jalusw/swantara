package giftcard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestGiftCardService_ForfeitExpired_PostsBreakageAndClearsBalance(t *testing.T) {
	ctx := context.Background()
	expiry := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	due := []*GiftCard{
		{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), Code: "GC-001", Balance: 100, State: GiftCardStateActive, ExpiryDate: &expiry},
		{Base: model.Base{ID: 43}, OrganizationID: helper.Ptr(uint64(10)), Code: "GC-002", Balance: 40, State: GiftCardStateActive, ExpiryDate: &expiry},
	}
	expiredCard := &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), Code: "GC-001", Balance: 100, State: GiftCardStateActive, ExpiryDate: &expiry}
	var postings []accounting.PostRequest
	var createdTransactions []*GiftCardTransaction
	svc := NewGiftCardService(
		GiftCardDAOMock{
			ListDueForExpiryFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*GiftCard, error) { return due, nil },
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, card *GiftCard) (*GiftCard, error) {
				expiredCard = card
				return card, nil
			},
		},
		GiftCardTransactionDAOMock{
			CreateTxFunc: func(_ context.Context, _ *gorm.DB, transaction *GiftCardTransaction) (*GiftCardTransaction, error) {
				createdTransactions = append(createdTransactions, transaction)
				return transaction, nil
			},
		},
		PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			postings = append(postings, request)
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		}},
		TransactionerMock{},
	)

	forfeited, err := svc.ForfeitExpired(ctx, ForfeitExpiredRequest{
		OrganizationID: 10, AsOf: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), JournalID: 3, LiabilityAccountID: 300, IncomeAccountID: 400,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forfeited) != 2 {
		t.Fatalf("forfeited = %d, want 2", len(forfeited))
	}
	if len(postings) != 2 {
		t.Fatalf("postings = %d, want 2", len(postings))
	}
	if postings[0].Lines[0].AccountID != 300 || !postings[0].Lines[0].Debit.Equal(amount.FromFloat64(100)) {
		t.Errorf("liability line = %+v, want Dr 100 on 300", postings[0].Lines[0])
	}
	if postings[0].Lines[1].AccountID != 400 || !postings[0].Lines[1].Credit.Equal(amount.FromFloat64(100)) {
		t.Errorf("income line = %+v, want Cr 100 on 400", postings[0].Lines[1])
	}
	if !postings[1].Lines[0].Debit.Equal(amount.FromFloat64(40)) {
		t.Errorf("second forfeit = %+v, want Dr 40", postings[1].Lines[0])
	}
	if expiredCard.Balance != 0 || expiredCard.State != GiftCardStateExpired {
		t.Errorf("card after forfeit = %+v, want balance 0 / expired", expiredCard)
	}
	if len(createdTransactions) != 2 || createdTransactions[0].Type != TransactionForfeit {
		t.Errorf("transactions = %+v, want 2 forfeit rows", createdTransactions)
	}
}

func TestGiftCardService_ForfeitExpired_SkipsLiveCards(t *testing.T) {
	svc := NewGiftCardService(
		GiftCardDAOMock{ListDueForExpiryFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*GiftCard, error) {
			return []*GiftCard{}, nil
		}},
		GiftCardTransactionDAOMock{},
		PosterMock{},
		TransactionerMock{},
	)

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

func TestGiftCardService_ForfeitExpired_RequiresAccounts(t *testing.T) {
	svc := NewGiftCardService(GiftCardDAOMock{}, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.ForfeitExpired(context.Background(), ForfeitExpiredRequest{
		OrganizationID: 10, AsOf: time.Now(), JournalID: 3,
	})
	if !errors.Is(err, ErrGiftCardForfeitAccount) {
		t.Fatalf("err = %v, want ErrGiftCardForfeitAccount", err)
	}
}

func TestGiftCardService_ForfeitExpired_RequiresOrganization(t *testing.T) {
	svc := NewGiftCardService(GiftCardDAOMock{}, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.ForfeitExpired(context.Background(), ForfeitExpiredRequest{
		AsOf: time.Now(), JournalID: 3, LiabilityAccountID: 300, IncomeAccountID: 400,
	})
	if !errors.Is(err, ErrGiftCardOrganization) {
		t.Fatalf("err = %v, want ErrGiftCardOrganization", err)
	}
}
