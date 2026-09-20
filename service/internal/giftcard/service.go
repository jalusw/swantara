package giftcard

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type GiftCardService struct {
	cards    GiftCardDAO
	trans    GiftCardTransactionDAO
	poster   accounting.Poster
	tx       db.Transactioner
	now      func() time.Time
	generate func() string
}

func NewGiftCardService(
	cards GiftCardDAO,
	trans GiftCardTransactionDAO,
	poster accounting.Poster,
	tx db.Transactioner,
) GiftCardService {
	return GiftCardService{
		cards:    cards,
		trans:    trans,
		poster:   poster,
		tx:       tx,
		now:      time.Now,
		generate: generateGiftCardCode,
	}
}

type IssueRequest struct {
	OrganizationID     uint64
	Code               string
	ContactID          *uint64
	Amount             float64
	CurrencyCode       string
	ExpiryDate         *time.Time
	IssuedFromOrderID  *uint64
	JournalID          uint64
	CashAccountID      uint64
	LiabilityAccountID uint64
	Date               time.Time
}

func (s GiftCardService) Issue(ctx context.Context, request IssueRequest) (*GiftCard, error) {
	if request.Amount <= 0 {
		return nil, ErrGiftCardAmountInvalid
	}
	if request.CashAccountID == 0 || request.LiabilityAccountID == 0 {
		return nil, ErrGiftCardAccounts
	}
	code := request.Code
	if code == "" {
		for i := 0; i < 10; i++ {
			code = s.generate()
			existing, err := s.cards.Search(ctx, "code", code)
			if err != nil {
				return nil, err
			}
			if existing == nil {
				break
			}
			code = ""
		}
		if code == "" {
			return nil, ErrGiftCardAlreadyExists
		}
	} else {
		existing, err := s.cards.Search(ctx, "code", code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrGiftCardAlreadyExists
		}
	}

	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	amountValue := amount.FromFloat64(request.Amount)

	card := &GiftCard{
		OrganizationID:    &request.OrganizationID,
		Code:              code,
		ContactID:         request.ContactID,
		InitialAmount:     request.Amount,
		Balance:           request.Amount,
		CurrencyCode:      request.CurrencyCode,
		ExpiryDate:        request.ExpiryDate,
		State:             GiftCardStateActive,
		IssuedFromOrderID: request.IssuedFromOrderID,
	}

	var created *GiftCard
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		journal, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
			OrganizationID: request.OrganizationID,
			JournalID:      request.JournalID,
			Date:           date,
			Ref:            fmt.Sprintf("GC/%s", code),
			OriginType:     TransactionIssue,
			Description:    "Gift card issue",
			Lines: []accounting.PostingLine{
				{AccountID: request.CashAccountID, Name: "Cash", Debit: amountValue},
				{AccountID: request.LiabilityAccountID, Name: "Gift Card Liability", Credit: amountValue},
			},
		})
		if err != nil {
			return err
		}
		created, err = s.cards.CreateTx(ctx, tx, card)
		if err != nil {
			return err
		}
		_, err = s.trans.CreateTx(ctx, tx, &GiftCardTransaction{
			GiftCardID: created.ID,
			Type:       TransactionIssue,
			Amount:     request.Amount,
			EntryID:    &journal.ID,
		})
		return err
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrGiftCardAlreadyExists
		}
		return nil, err
	}
	return created, nil
}

func (s GiftCardService) card(ctx context.Context, id uint64) (*GiftCard, error) {
	card, err := s.cards.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if card == nil {
		return nil, ErrGiftCardNotFound
	}
	return card, nil
}

func (s GiftCardService) assertActive(card *GiftCard, date time.Time) error {
	switch card.State {
	case GiftCardStateActive:
	case GiftCardStateExpired:
		return ErrGiftCardExpired
	default:
		return ErrGiftCardInactive
	}
	if card.ExpiryDate != nil && !date.IsZero() && date.After(*card.ExpiryDate) {
		return ErrGiftCardExpired
	}
	return nil
}

type RedeemRequest struct {
	GiftCardID         uint64
	Amount             float64
	OrderType          string
	OrderID            uint64
	JournalID          uint64
	RevenueAccountID   uint64
	LiabilityAccountID uint64
	Date               time.Time
}

func (s GiftCardService) Redeem(ctx context.Context, request RedeemRequest) (*GiftCard, error) {
	if request.Amount <= 0 {
		return nil, ErrGiftCardAmountInvalid
	}
	if request.LiabilityAccountID == 0 {
		return nil, ErrGiftCardLiabilityOnly
	}
	if request.RevenueAccountID == 0 {
		return nil, ErrGiftCardRevenue
	}
	card, err := s.card(ctx, request.GiftCardID)
	if err != nil {
		return nil, err
	}
	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	if err := s.assertActive(card, date); err != nil {
		return nil, err
	}
	balance := amount.FromFloat64(card.Balance)
	redeem := amount.FromFloat64(request.Amount)
	if redeem.GreaterThan(balance) {
		return nil, ErrGiftCardInsufficient
	}

	var updated *GiftCard
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		journal, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
			OrganizationID: *card.OrganizationID,
			JournalID:      request.JournalID,
			Date:           date,
			Ref:            fmt.Sprintf("GC-REDEEM/%d", card.ID),
			OriginType:     TransactionRedeem,
			Description:    "Gift card redemption",
			Lines: []accounting.PostingLine{
				{AccountID: request.LiabilityAccountID, Name: "Gift Card Liability", Debit: redeem},
				{AccountID: request.RevenueAccountID, Name: "Revenue", Credit: redeem},
			},
		})
		if err != nil {
			return err
		}
		card.Balance = balance.Sub(redeem).Round(4).Float64()
		if card.Balance == 0 {
			card.State = GiftCardStateUsed
		}
		updated, err = s.cards.UpdateTx(ctx, tx, card)
		if err != nil {
			return err
		}
		_, err = s.trans.CreateTx(ctx, tx, &GiftCardTransaction{
			GiftCardID: card.ID,
			Type:       TransactionRedeem,
			Amount:     request.Amount,
			OrderType:  request.OrderType,
			OrderID:    request.OrderID,
			EntryID:    &journal.ID,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

type RefundRequest struct {
	GiftCardID         uint64
	Amount             float64
	OrderType          string
	OrderID            uint64
	JournalID          uint64
	RefundAccountID    uint64
	LiabilityAccountID uint64
	Date               time.Time
}

func (s GiftCardService) Refund(ctx context.Context, request RefundRequest) (*GiftCard, error) {
	if request.Amount <= 0 {
		return nil, ErrGiftCardAmountInvalid
	}
	if request.LiabilityAccountID == 0 {
		return nil, ErrGiftCardLiabilityOnly
	}
	if request.RefundAccountID == 0 {
		return nil, ErrGiftCardRefundAccount
	}
	card, err := s.card(ctx, request.GiftCardID)
	if err != nil {
		return nil, err
	}
	if card.State != GiftCardStateActive && card.State != GiftCardStateUsed {
		return nil, ErrGiftCardInactive
	}
	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	refund := amount.FromFloat64(request.Amount)

	var updated *GiftCard
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		journal, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
			OrganizationID: *card.OrganizationID,
			JournalID:      request.JournalID,
			Date:           date,
			Ref:            fmt.Sprintf("GC-REFUND/%d", card.ID),
			OriginType:     TransactionRefund,
			Description:    "Gift card refund",
			Lines: []accounting.PostingLine{
				{AccountID: request.RefundAccountID, Name: "Sales Refund", Debit: refund},
				{AccountID: request.LiabilityAccountID, Name: "Gift Card Liability", Credit: refund},
			},
		})
		if err != nil {
			return err
		}
		card.Balance = amount.FromFloat64(card.Balance).Add(refund).Round(4).Float64()
		card.State = GiftCardStateActive
		updated, err = s.cards.UpdateTx(ctx, tx, card)
		if err != nil {
			return err
		}
		_, err = s.trans.CreateTx(ctx, tx, &GiftCardTransaction{
			GiftCardID: card.ID,
			Type:       TransactionRefund,
			Amount:     request.Amount,
			OrderType:  request.OrderType,
			OrderID:    request.OrderID,
			EntryID:    &journal.ID,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

type AdjustRequest struct {
	GiftCardID         uint64
	Delta              float64
	JournalID          uint64
	LiabilityAccountID uint64
	OffsetAccountID    uint64
	Date               time.Time
}

func (s GiftCardService) Adjust(ctx context.Context, request AdjustRequest) (*GiftCard, error) {
	if request.Delta == 0 {
		return nil, ErrGiftCardAmountInvalid
	}
	if request.LiabilityAccountID == 0 || request.OffsetAccountID == 0 {
		return nil, ErrGiftCardLiabilityOnly
	}
	card, err := s.card(ctx, request.GiftCardID)
	if err != nil {
		return nil, err
	}
	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	delta := amount.FromFloat64(request.Delta)
	newBalance := amount.FromFloat64(card.Balance).Add(delta)
	if newBalance.IsNegative() {
		return nil, ErrGiftCardInsufficient
	}

	var updated *GiftCard
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		var lines []accounting.PostingLine
		if delta.IsPositive() {
			lines = []accounting.PostingLine{
				{AccountID: request.OffsetAccountID, Name: "Gift Card Adjustment", Debit: delta},
				{AccountID: request.LiabilityAccountID, Name: "Gift Card Liability", Credit: delta},
			}
		} else {
			lines = []accounting.PostingLine{
				{AccountID: request.LiabilityAccountID, Name: "Gift Card Liability", Debit: delta.Neg()},
				{AccountID: request.OffsetAccountID, Name: "Gift Card Adjustment", Credit: delta.Neg()},
			}
		}
		journal, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
			OrganizationID: *card.OrganizationID,
			JournalID:      request.JournalID,
			Date:           date,
			Ref:            fmt.Sprintf("GC-ADJ/%d", card.ID),
			OriginType:     TransactionAdjust,
			Description:    "Gift card balance adjustment",
			Lines:          lines,
		})
		if err != nil {
			return err
		}
		card.Balance = newBalance.Round(4).Float64()
		if card.Balance == 0 {
			card.State = GiftCardStateUsed
		}
		updated, err = s.cards.UpdateTx(ctx, tx, card)
		if err != nil {
			return err
		}
		_, err = s.trans.CreateTx(ctx, tx, &GiftCardTransaction{
			GiftCardID: card.ID,
			Type:       TransactionAdjust,
			Amount:     request.Delta,
			EntryID:    &journal.ID,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

type ForfeitExpiredRequest struct {
	OrganizationID     uint64
	AsOf               time.Time
	Date               time.Time
	JournalID          uint64
	LiabilityAccountID uint64
	IncomeAccountID    uint64
}

func (s GiftCardService) ForfeitExpired(ctx context.Context, request ForfeitExpiredRequest) ([]*GiftCard, error) {
	if request.OrganizationID == 0 {
		return nil, ErrGiftCardOrganization
	}
	if request.LiabilityAccountID == 0 || request.IncomeAccountID == 0 {
		return nil, ErrGiftCardForfeitAccount
	}
	asOf := request.AsOf
	if asOf.IsZero() {
		asOf = s.now().UTC()
	}
	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	due, err := s.cards.ListDueForExpiry(ctx, request.OrganizationID, asOf)
	if err != nil {
		return nil, err
	}
	var forfeited []*GiftCard
	for _, card := range due {
		if card.OrganizationID == nil {
			continue
		}
		balance := amount.FromFloat64(card.Balance).Round(4)
		if !balance.IsPositive() {
			continue
		}
		var updated *GiftCard
		err := s.tx.Run(ctx, func(tx *gorm.DB) error {
			journal, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
				OrganizationID: *card.OrganizationID,
				JournalID:      request.JournalID,
				Date:           date,
				Ref:            fmt.Sprintf("GC-FORFEIT/%d", card.ID),
				OriginType:     TransactionForfeit,
				Description:    "Gift card expiry breakage",
				Lines: []accounting.PostingLine{
					{AccountID: request.LiabilityAccountID, Name: "Gift Card Liability", Debit: balance},
					{AccountID: request.IncomeAccountID, Name: "Other Income", Credit: balance},
				},
			})
			if err != nil {
				return err
			}
			card.Balance = 0
			card.State = GiftCardStateExpired
			updated, err = s.cards.UpdateTx(ctx, tx, card)
			if err != nil {
				return err
			}
			_, err = s.trans.CreateTx(ctx, tx, &GiftCardTransaction{
				GiftCardID: card.ID,
				Type:       TransactionForfeit,
				Amount:     balance.Float64(),
				EntryID:    &journal.ID,
			})
			return err
		})
		if err != nil {
			return nil, err
		}
		forfeited = append(forfeited, updated)
	}
	return forfeited, nil
}

func (s GiftCardService) List(ctx context.Context, q *query.Query) (*query.Page[GiftCard], error) {
	return s.cards.List(ctx, q)
}

func (s GiftCardService) Find(ctx context.Context, id uint64) (*GiftCard, error) {
	return s.cards.Find(ctx, id)
}

func (s GiftCardService) Delete(ctx context.Context, id uint64) error {
	return s.cards.Delete(ctx, id)
}

func (s GiftCardService) ListTransactions(ctx context.Context, giftCardID uint64) ([]*GiftCardTransaction, error) {
	if _, err := s.card(ctx, giftCardID); err != nil {
		return nil, err
	}
	return s.trans.ListByGiftCard(ctx, giftCardID)
}

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateGiftCardCode() string {
	var b []byte
	max := big.NewInt(int64(len(codeAlphabet)))
	for i := 0; i < 12; i++ {
		n, _ := rand.Int(rand.Reader, max)
		b = append(b, codeAlphabet[n.Int64()])
	}
	return string(b)
}
