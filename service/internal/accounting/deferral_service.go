package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type JournalResolver interface {
	JournalID(ctx context.Context, organizationID uint64) (uint64, error)
}

type ScheduleLineRequest struct {
	RecognitionDate time.Time
	Amount          float64
}

type CreateScheduleRequest struct {
	OrganizationID        uint64
	Type                  string
	SourceType            string
	SourceID              uint64
	ContactID             *uint64
	ItemID                *uint64
	TotalAmount           float64
	BalanceSheetAccountID uint64
	PLAccountID           uint64
	DimensionID           *uint64
	Method                string
	DateStart             time.Time
	DateEnd               time.Time
	Periods               int
	Lines                 []ScheduleLineRequest
}

type DeferralService struct {
	schedules     DeferredScheduleDAO
	scheduleLines DeferredScheduleLineDAO
	poster        Poster
	journals      JournalResolver
	tx            db.Transactioner
}

func NewDeferralService(
	schedules DeferredScheduleDAO,
	scheduleLines DeferredScheduleLineDAO,
	poster Poster,
	journals JournalResolver,
	tx db.Transactioner,
) DeferralService {
	return DeferralService{
		schedules:     schedules,
		scheduleLines: scheduleLines,
		poster:        poster,
		journals:      journals,
		tx:            tx,
	}
}

func (s DeferralService) Create(ctx context.Context, request CreateScheduleRequest) (*DeferredSchedule, error) {
	var schedule *DeferredSchedule
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		schedule, err = s.CreateTx(ctx, tx, request)
		return err
	})
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s DeferralService) CreateTx(ctx context.Context, tx *gorm.DB, request CreateScheduleRequest) (*DeferredSchedule, error) {
	if err := validateScheduleRequest(request); err != nil {
		return nil, err
	}
	lines, err := buildScheduleLines(request)
	if err != nil {
		return nil, err
	}

	schedule := &DeferredSchedule{
		OrganizationID:        helper.Ptr(request.OrganizationID),
		Type:                  request.Type,
		SourceType:            request.SourceType,
		SourceID:              request.SourceID,
		ContactID:             request.ContactID,
		ItemID:                request.ItemID,
		TotalAmount:           request.TotalAmount,
		BalanceSheetAccountID: helper.Ptr(request.BalanceSheetAccountID),
		PLAccountID:           helper.Ptr(request.PLAccountID),
		DimensionID:           request.DimensionID,
		Method:                request.Method,
		DateStart:             helper.Ptr(request.DateStart),
		DateEnd:               helper.Ptr(request.DateEnd),
		Periods:               request.Periods,
		State:                 DeferredStateRunning,
	}

	created, err := s.schedules.CreateTx(ctx, tx, schedule)
	if err != nil {
		return nil, err
	}
	for _, line := range lines {
		line.ScheduleID = created.ID
		if _, err := s.scheduleLines.CreateTx(ctx, tx, line); err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (s DeferralService) Get(ctx context.Context, organizationID, scheduleID uint64) (*DeferredSchedule, error) {
	schedule, err := s.schedules.Search(ctx, "id", scheduleID)
	if err != nil {
		return nil, err
	}
	if schedule == nil || schedule.OrganizationID == nil || *schedule.OrganizationID != organizationID {
		return nil, ErrScheduleNotFound
	}
	return schedule, nil
}

func (s DeferralService) List(ctx context.Context, organizationID uint64) ([]*DeferredSchedule, error) {
	page, err := s.schedules.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s DeferralService) ListLines(ctx context.Context, scheduleID uint64) ([]*DeferredScheduleLine, error) {
	return s.scheduleLines.ListBySchedule(ctx, scheduleID)
}

func (s DeferralService) RecognizeDue(ctx context.Context, organizationID *uint64, asOf time.Time) (int, error) {
	schedules, err := s.schedules.ListRunning(ctx, organizationID)
	if err != nil {
		return 0, err
	}
	posted := 0
	for _, schedule := range schedules {
		if schedule.OrganizationID == nil {
			continue
		}
		if schedule.BalanceSheetAccountID == nil || schedule.PLAccountID == nil {
			return posted, ErrScheduleNoAccount
		}
		journalID, err := s.journals.JournalID(ctx, *schedule.OrganizationID)
		if err != nil {
			return posted, err
		}
		if journalID == 0 {
			return posted, ErrScheduleNoJournal
		}
		due, err := s.scheduleLines.ListDue(ctx, schedule.ID, asOf)
		if err != nil {
			return posted, err
		}
		if len(due) == 0 {
			continue
		}
		err = s.tx.Run(ctx, func(tx *gorm.DB) error {
			for _, line := range due {
				entry, err := s.poster.PostTx(ctx, tx, s.recognitionRequest(schedule, line, journalID))
				if err != nil {
					return err
				}
				line.EntryID = &entry.ID
				line.Posted = true
				if _, err := s.scheduleLines.UpdateTx(ctx, tx, line); err != nil {
					return err
				}
				schedule.RecognizedAmount += line.Amount
			}
			if amount.FromFloat64(schedule.RecognizedAmount).Round(4).Equal(amount.FromFloat64(schedule.TotalAmount).Round(4)) {
				schedule.State = DeferredStateDone
			}
			if _, err := s.schedules.UpdateTx(ctx, tx, schedule); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return posted, err
		}
		posted += len(due)
	}
	return posted, nil
}

func (s DeferralService) recognitionRequest(schedule *DeferredSchedule, line *DeferredScheduleLine, journalID uint64) PostRequest {
	amount := amount.FromFloat64(line.Amount)
	debitName := "Deferred Revenue"
	creditName := "Revenue"
	debitAccountID := *schedule.BalanceSheetAccountID
	creditAccountID := *schedule.PLAccountID
	if schedule.Type == DeferredTypeDeferredExpense || schedule.Type == DeferredTypePrepaid {
		debitName = "Expense"
		creditName = "Prepaid Expense"
		debitAccountID, creditAccountID = creditAccountID, debitAccountID
	}
	return PostRequest{
		OrganizationID: *schedule.OrganizationID,
		JournalID:      journalID,
		Date:           *line.RecognitionDate,
		Ref:            fmt.Sprintf("DEF/%d/%d", schedule.SourceID, line.Sequence),
		OriginType:     schedule.Type,
		OriginID:       line.ID,
		Lines: []PostingLine{
			{AccountID: debitAccountID, Name: debitName, Debit: amount},
			{AccountID: creditAccountID, Name: creditName, Credit: amount},
		},
	}
}

func validateScheduleRequest(request CreateScheduleRequest) error {
	switch request.Type {
	case DeferredTypeDeferredRevenue, DeferredTypeDeferredExpense, DeferredTypePrepaid:
	default:
		return ErrScheduleType
	}
	switch request.Method {
	case DeferredMethodLinear, DeferredMethodManual, DeferredMethodMilestone:
	default:
		return ErrScheduleMethod
	}
	if !amount.FromFloat64(request.TotalAmount).IsPositive() {
		return ErrScheduleAmount
	}
	if request.BalanceSheetAccountID == 0 || request.PLAccountID == 0 {
		return ErrScheduleNoAccount
	}
	if request.DateStart.IsZero() {
		return ErrScheduleInvalidDate
	}
	if request.Method == DeferredMethodLinear {
		if request.Periods < 1 {
			return ErrScheduleNoLines
		}
		return nil
	}
	if len(request.Lines) == 0 {
		return ErrScheduleNoLines
	}
	return nil
}

func buildScheduleLines(request CreateScheduleRequest) ([]*DeferredScheduleLine, error) {
	if request.Method != DeferredMethodLinear {
		lines := make([]*DeferredScheduleLine, 0, len(request.Lines))
		for i, line := range request.Lines {
			if !amount.FromFloat64(line.Amount).IsPositive() {
				return nil, ErrScheduleLineAmount
			}
			if line.RecognitionDate.IsZero() {
				return nil, ErrScheduleInvalidDate
			}
			lines = append(lines, &DeferredScheduleLine{
				Sequence:        i + 1,
				RecognitionDate: helper.Ptr(line.RecognitionDate),
				Amount:          line.Amount,
			})
		}
		return lines, nil
	}

	total := amount.FromFloat64(request.TotalAmount).Round(4)
	per, err := total.Div(amount.FromInt64(int64(request.Periods)))
	if err != nil {
		return nil, err
	}
	per = per.Round(4)
	accrued := amount.Zero()
	lines := make([]*DeferredScheduleLine, 0, request.Periods)
	for i := 1; i <= request.Periods; i++ {
		lineAmount := per
		if i == request.Periods {
			lineAmount = total.Sub(accrued).Round(4)
		}
		accrued = accrued.Add(lineAmount)
		lines = append(lines, &DeferredScheduleLine{
			Sequence:        i,
			RecognitionDate: helper.Ptr(request.DateStart.AddDate(0, i-1, 0)),
			Amount:          lineAmount.Float64(),
		})
	}
	return lines, nil
}
