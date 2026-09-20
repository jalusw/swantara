package giftcard

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func giftCardSvc(cards GiftCardDAOMock) GiftCardService {
	return NewGiftCardService(cards, GiftCardTransactionDAOMock{}, PosterMock{}, TransactionerMock{})
}

func activeGiftCard() *GiftCard {
	expiry := time.Now().AddDate(0, 6, 0)
	return &GiftCard{Base: model.Base{ID: 1}, State: GiftCardStateActive, Balance: 100, ExpiryDate: &expiry}
}

func TestGiftCardService_CardBranches(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("card lookup error and missing", func(t *testing.T) {
		svc := giftCardSvc(GiftCardDAOMock{
			CRUDMock: dao.CRUDMock[GiftCard]{
				FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) { return nil, dbErr },
			},
		})
		_, err := svc.card(ctx, 1)
		helper.AssertError(t, err, true, dbErr)

		svc = giftCardSvc(GiftCardDAOMock{})
		_, err = svc.card(ctx, 1)
		helper.AssertError(t, err, true, ErrGiftCardNotFound)
	})

	t.Run("assert active branches", func(t *testing.T) {
		svc := giftCardSvc(GiftCardDAOMock{})
		now := time.Now()
		if err := svc.assertActive(activeGiftCard(), now); err != nil {
			t.Errorf("active = %v", err)
		}

		expired := activeGiftCard()
		expired.State = GiftCardStateExpired
		helper.AssertError(t, svc.assertActive(expired, now), true, ErrGiftCardExpired)

		inactive := activeGiftCard()
		inactive.State = GiftCardStateUsed
		helper.AssertError(t, svc.assertActive(inactive, now), true, ErrGiftCardInactive)

		past := activeGiftCard()
		past.ExpiryDate = &now
		future := now.Add(time.Hour)
		_ = future
		pastDate := now.Add(-time.Hour)
		past.ExpiryDate = &pastDate
		helper.AssertError(t, svc.assertActive(past, now), true, ErrGiftCardExpired)
	})
}

func TestGiftCardService_RedeemRefund_Branches(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	newSvc := func(card *GiftCard, cardErr error) GiftCardService {
		return giftCardSvc(GiftCardDAOMock{
			CRUDMock: dao.CRUDMock[GiftCard]{
				FindFunc: func(_ context.Context, _ uint64) (*GiftCard, error) { return card, cardErr },
			},
		})
	}
	t.Run("redeem validations", func(t *testing.T) {
		svc := newSvc(activeGiftCard(), nil)
		_, err := svc.Redeem(ctx, RedeemRequest{Amount: 0})
		helper.AssertError(t, err, true, ErrGiftCardAmountInvalid)

		_, err = svc.Redeem(ctx, RedeemRequest{GiftCardID: 1, Amount: 10, RevenueAccountID: 2, Date: now})
		helper.AssertError(t, err, true, ErrGiftCardLiabilityOnly)

		_, err = svc.Redeem(ctx, RedeemRequest{GiftCardID: 1, Amount: 10, LiabilityAccountID: 3, Date: now})
		helper.AssertError(t, err, true, ErrGiftCardRevenue)
	})

	t.Run("refund validations", func(t *testing.T) {
		svc := newSvc(activeGiftCard(), nil)
		_, err := svc.Refund(ctx, RefundRequest{Amount: 0})
		helper.AssertError(t, err, true, ErrGiftCardAmountInvalid)

		_, err = svc.Refund(ctx, RefundRequest{GiftCardID: 1, Amount: 10, RefundAccountID: 4, Date: now})
		helper.AssertError(t, err, true, ErrGiftCardLiabilityOnly)

		_, err = svc.Refund(ctx, RefundRequest{GiftCardID: 1, Amount: 10, LiabilityAccountID: 3, Date: now})
		helper.AssertError(t, err, true, ErrGiftCardRefundAccount)

		used := activeGiftCard()
		used.State = GiftCardStateUsed
		_ = used

		expired := activeGiftCard()
		expired.State = GiftCardStateExpired
		svc = newSvc(expired, nil)
		_, err = svc.Refund(ctx, RefundRequest{GiftCardID: 1, Amount: 10, LiabilityAccountID: 3, RefundAccountID: 4, Date: now})
		helper.AssertError(t, err, true, ErrGiftCardInactive)
	})
}

func TestGiftCardDAO_CreateTx(t *testing.T) {
	ctx := context.Background()

	t.Run("creates in tx", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "gift_cards"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		tx := db.Begin()
		got, err := NewGiftCardDAO(db).CreateTx(ctx, tx, &GiftCard{})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit = %v", err)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "gift_cards"`)).WillReturnError(dbErr)

		tx := db.Begin()
		_, err := NewGiftCardDAO(db).CreateTx(ctx, tx, &GiftCard{})
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestGiftCardMock_Fallbacks(t *testing.T) {
	ctx := context.Background()

	t.Run("dao fallbacks", func(t *testing.T) {
		bare := GiftCardDAOMock{}
		if _, err := bare.CreateTx(ctx, nil, &GiftCard{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		wired := GiftCardDAOMock{
			CreateTxFunc: func(_ context.Context, _ *gorm.DB, c *GiftCard) (*GiftCard, error) { return c, nil },
		}
		if _, err := wired.CreateTx(ctx, nil, &GiftCard{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		trans := GiftCardTransactionDAOMock{}
		if _, err := trans.ListByGiftCard(ctx, 1); err != nil {
			t.Errorf("ListByGiftCard = %v", err)
		}
	})

	t.Run("poster and tx", func(t *testing.T) {
		bare := PosterMock{}
		if _, err := bare.Post(ctx, accounting.PostRequest{}); err != nil {
			t.Errorf("Post = %v", err)
		}
		if _, err := bare.Reverse(ctx, accounting.ReverseRequest{}); err != nil {
			t.Errorf("Reverse = %v", err)
		}
		if _, err := bare.ReverseTx(ctx, nil, accounting.ReverseRequest{}); err != nil {
			t.Errorf("ReverseTx = %v", err)
		}
		wired := PosterMock{
			PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
				return &accounting.JournalEntry{}, nil
			},
		}
		if _, err := wired.Post(ctx, accounting.PostRequest{}); err != nil {
			t.Errorf("Post = %v", err)
		}
		if err := (TransactionerMock{}).Run(ctx, func(_ *gorm.DB) error { return nil }); err != nil {
			t.Errorf("Run = %v", err)
		}
	})
}

func TestGiftCardFixtures(t *testing.T) {
	if GiftCardFixture() == nil {
		t.Error("card = nil")
	}
	if GiftCardTransactionFixture() == nil {
		t.Error("transaction = nil")
	}
	if CouponFixture() == nil {
		t.Error("coupon = nil")
	}
}
