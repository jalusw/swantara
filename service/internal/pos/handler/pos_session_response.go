package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/pos"
)

type POSSessionResponse struct {
	ID             uint64     `json:"id"`
	ConfigID       uint64     `json:"config_id"`
	CashierID      uint64     `json:"cashier_id"`
	OpenedAt       *time.Time `json:"opened_at"`
	ClosedAt       *time.Time `json:"closed_at"`
	OpeningBalance float64    `json:"opening_balance"`
	ClosingBalance *float64   `json:"closing_balance"`
	State          string     `json:"state"`
}

func newPOSSessionResponse(session *pos.POSSession) POSSessionResponse {
	return POSSessionResponse{
		ID:             session.ID,
		ConfigID:       session.ConfigID,
		CashierID:      session.CashierID,
		OpenedAt:       session.OpenedAt,
		ClosedAt:       session.ClosedAt,
		OpeningBalance: session.OpeningBalance,
		ClosingBalance: session.ClosingBalance,
		State:          session.State,
	}
}

type POSPaymentMethodResponse struct {
	Method string  `json:"method"`
	Amount float64 `json:"amount"`
}

type OpenPOSSessionRequest struct {
	ConfigID       uint64  `json:"config_id" validate:"required,gt=0"`
	CashierID      uint64  `json:"cashier_id" validate:"required,gt=0"`
	OpeningBalance float64 `json:"opening_balance" validate:"gte=0"`
}

type ClosePOSSessionRequest struct {
	ClosingBalance float64 `json:"closing_balance" validate:"gte=0"`
}
