package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type DeferralScheduleResponse struct {
	ID                    uint64     `json:"id"`
	OrganizationID        *uint64    `json:"organization_id"`
	Type                  string     `json:"type"`
	SourceType            string     `json:"source_type"`
	SourceID              uint64     `json:"source_id"`
	TotalAmount           float64    `json:"total_amount"`
	BalanceSheetAccountID *uint64    `json:"balance_sheet_account_id"`
	PLAccountID           *uint64    `json:"pl_account_id"`
	Method                string     `json:"method"`
	DateStart             *time.Time `json:"date_start"`
	State                 string     `json:"state"`
	RecognizedAmount      float64    `json:"recognized_amount"`
}

func newDeferralScheduleResponse(schedule *accounting.DeferredSchedule) DeferralScheduleResponse {
	return DeferralScheduleResponse{
		ID:                    schedule.ID,
		OrganizationID:        schedule.OrganizationID,
		Type:                  schedule.Type,
		SourceType:            schedule.SourceType,
		SourceID:              schedule.SourceID,
		TotalAmount:           schedule.TotalAmount,
		BalanceSheetAccountID: schedule.BalanceSheetAccountID,
		PLAccountID:           schedule.PLAccountID,
		Method:                schedule.Method,
		DateStart:             schedule.DateStart,
		State:                 schedule.State,
		RecognizedAmount:      schedule.RecognizedAmount,
	}
}

type DeferralLineResponse struct {
	ID              uint64     `json:"id"`
	ScheduleID      uint64     `json:"schedule_id"`
	Sequence        int        `json:"sequence"`
	RecognitionDate *time.Time `json:"recognition_date"`
	Amount          float64    `json:"amount"`
	Posted          bool       `json:"posted"`
	EntryID         *uint64    `json:"entry_id"`
}

type ListDeferralSchedulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListDeferralSchedulesResponse `json:"data"`
}
type ListDeferralSchedulesResponse struct {
	Schedules []DeferralScheduleResponse `json:"schedules"`
}

type GetDeferralScheduleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data DeferralScheduleResponse `json:"data"`
}

type ListDeferralLinesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListDeferralLinesResponse `json:"data"`
}
type ListDeferralLinesResponse struct {
	Lines []DeferralLineResponse `json:"lines"`
}

type CreateDeferralScheduleRequest struct {
	Type                  string                      `json:"type" validate:"required,oneof=deferred_revenue deferred_expense prepaid"`
	SourceType            string                      `json:"source_type" validate:"required"`
	SourceID              uint64                      `json:"source_id" validate:"required"`
	ContactID             *uint64                     `json:"contact_id"`
	ItemID                *uint64                     `json:"item_id"`
	TotalAmount           float64                     `json:"total_amount" validate:"required,gt=0"`
	BalanceSheetAccountID uint64                      `json:"balance_sheet_account_id" validate:"required"`
	PLAccountID           uint64                      `json:"pl_account_id" validate:"required"`
	DimensionID           *uint64                     `json:"dimension_id"`
	Method                string                      `json:"method" validate:"required,oneof=linear manual milestone"`
	DateStart             *time.Time                  `json:"date_start" validate:"required"`
	DateEnd               *time.Time                  `json:"date_end"`
	Periods               int                         `json:"periods" validate:"gte=0"`
	Lines                 []CreateDeferralLineRequest `json:"lines" validate:"dive"`
}

type CreateDeferralLineRequest struct {
	RecognitionDate *time.Time `json:"recognition_date" validate:"required"`
	Amount          float64    `json:"amount" validate:"required,gt=0"`
}

type CreateDeferralScheduleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data DeferralScheduleResponse `json:"data"`
}
