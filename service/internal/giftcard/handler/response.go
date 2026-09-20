package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GiftCardResponse struct {
	ID                uint64     `json:"id"`
	OrganizationID    *uint64    `json:"organization_id"`
	Code              string     `json:"code"`
	ContactID         *uint64    `json:"contact_id"`
	InitialAmount     float64    `json:"initial_amount"`
	Balance           float64    `json:"balance"`
	CurrencyCode      string     `json:"currency_code"`
	ExpiryDate        *time.Time `json:"expiry_date"`
	State             string     `json:"state"`
	IssuedFromOrderID *uint64    `json:"issued_from_order_id"`
}

func newGiftCardResponse(card *giftcard.GiftCard) GiftCardResponse {
	return GiftCardResponse{
		ID: card.ID, OrganizationID: card.OrganizationID, Code: card.Code, ContactID: card.ContactID,
		InitialAmount: card.InitialAmount, Balance: card.Balance, CurrencyCode: card.CurrencyCode,
		ExpiryDate: card.ExpiryDate, State: card.State, IssuedFromOrderID: card.IssuedFromOrderID,
	}
}

var giftCardQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"code":            {},
	"contact_id":      {},
	"state":           {},
	"currency_code":   {},
	"created_at":      {},
	"updated_at":      {},
}

type ListGiftCardsResponse struct {
	GiftCards []GiftCardResponse `json:"gift_cards"`
}

type ListGiftCardsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListGiftCardsResponse `json:"data"`
}

type GetGiftCardResponse struct {
	GiftCard GiftCardResponse `json:"gift_card"`
}

type GetGiftCardResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetGiftCardResponse `json:"data"`
}

type CreateGiftCardResponse struct {
	GiftCard GiftCardResponse `json:"gift_card"`
}

type CreateGiftCardResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateGiftCardResponse `json:"data"`
}

type GiftCardTransactionResponse struct {
	ID         uint64  `json:"id"`
	GiftCardID uint64  `json:"gift_card_id"`
	Type       string  `json:"type"`
	Amount     float64 `json:"amount"`
	OrderType  string  `json:"order_type"`
	OrderID    uint64  `json:"order_id"`
	EntryID    *uint64 `json:"entry_id"`
}

func newGiftCardTransactionResponse(t *giftcard.GiftCardTransaction) GiftCardTransactionResponse {
	return GiftCardTransactionResponse{
		ID: t.ID, GiftCardID: t.GiftCardID, Type: t.Type, Amount: t.Amount,
		OrderType: t.OrderType, OrderID: t.OrderID, EntryID: t.EntryID,
	}
}

type ListGiftCardTransactionsResponse struct {
	GiftCardTransactions []GiftCardTransactionResponse `json:"gift_card_transactions"`
}

type ListGiftCardTransactionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListGiftCardTransactionsResponse `json:"data"`
}

type ForfeitExpiredGiftCardsResponse struct {
	GiftCards []GiftCardResponse `json:"gift_cards"`
}

type ForfeitExpiredGiftCardsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ForfeitExpiredGiftCardsResponse `json:"data"`
}
