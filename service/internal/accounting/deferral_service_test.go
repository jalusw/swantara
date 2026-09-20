package accounting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestDeferralService_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "linear splits evenly and absorbs rounding",
			run: func(t *testing.T) {
				var createdLines []*DeferredScheduleLine
				schedules := DeferredScheduleDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error) {
						schedule.ID = 7
						return schedule, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error) {
						createdLines = append(createdLines, line)
						return line, nil
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				schedule, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredRevenue,
					SourceType:            "invoice_line",
					SourceID:              3,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodLinear,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Periods:               3,
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if schedule.State != DeferredStateRunning || schedule.ID != 7 {
					t.Errorf("schedule = %+v, want running state and id 7", schedule)
				}
				if len(createdLines) != 3 {
					t.Fatalf("lines = %d, want 3", len(createdLines))
				}
				if createdLines[0].Amount != 33.3333 || createdLines[1].Amount != 33.3333 || createdLines[2].Amount != 33.3334 {
					t.Errorf("line amounts = %v/%v/%v, want 33.3333/33.3333/33.3334", createdLines[0].Amount, createdLines[1].Amount, createdLines[2].Amount)
				}
				if !createdLines[0].RecognitionDate.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
					t.Errorf("first recognition date = %v, want 2026-01-01", createdLines[0].RecognitionDate)
				}
			},
		},
		{
			name: "manual requires explicit lines",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredExpense,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodManual,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				})
				if helper.AssertError(t, err, true, ErrScheduleNoLines) {
					return
				}
			},
		},
		{
			name: "rejects invalid type",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  "wrong",
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodLinear,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Periods:               2,
				})
				if helper.AssertError(t, err, true, ErrScheduleType) {
					return
				}
			},
		},
		{
			name: "rejects invalid method",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredRevenue,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                "wrong",
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Periods:               2,
				})
				if helper.AssertError(t, err, true, ErrScheduleMethod) {
					return
				}
			},
		},
		{
			name: "rejects non-positive amount",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredRevenue,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodLinear,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Periods:               2,
				})
				if helper.AssertError(t, err, true, ErrScheduleAmount) {
					return
				}
			},
		},
		{
			name: "rejects without accounts",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID: 10,
					Type:           DeferredTypeDeferredRevenue,
					TotalAmount:    100,
					Method:         DeferredMethodLinear,
					DateStart:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Periods:        2,
				})
				if helper.AssertError(t, err, true, ErrScheduleNoAccount) {
					return
				}
			},
		},
		{
			name: "manual builds explicit lines",
			run: func(t *testing.T) {
				var createdLines []*DeferredScheduleLine
				schedules := DeferredScheduleDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error) {
						schedule.ID = 7
						return schedule, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error) {
						createdLines = append(createdLines, line)
						return line, nil
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				schedule, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredExpense,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodManual,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Lines: []ScheduleLineRequest{
						{RecognitionDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Amount: 40},
						{RecognitionDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Amount: 60},
					},
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if schedule.ID != 7 {
					t.Errorf("schedule = %+v, want id 7", schedule)
				}
				if len(createdLines) != 2 {
					t.Fatalf("lines = %d, want 2", len(createdLines))
				}
				if createdLines[0].Sequence != 1 || createdLines[0].Amount != 40 || createdLines[0].ScheduleID != 7 {
					t.Errorf("line 0 = %+v, want sequence 1 amount 40 on schedule 7", createdLines[0])
				}
				if createdLines[1].Sequence != 2 || createdLines[1].Amount != 60 {
					t.Errorf("line 1 = %+v, want sequence 2 amount 60", createdLines[1])
				}
			},
		},
		{
			name: "rejects invalid date",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredRevenue,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodLinear,
					Periods:               2,
				})
				if helper.AssertError(t, err, true, ErrScheduleInvalidDate) {
					return
				}
			},
		},
		{
			name: "rejects linear without periods",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredRevenue,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodLinear,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				})
				if helper.AssertError(t, err, true, ErrScheduleNoLines) {
					return
				}
			},
		},
		{
			name: "rejects invalid line amount",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredRevenue,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodManual,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Lines: []ScheduleLineRequest{
						{RecognitionDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Amount: 0},
					},
				})
				if helper.AssertError(t, err, true, ErrScheduleLineAmount) {
					return
				}
			},
		},
		{
			name: "rejects invalid line date",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredRevenue,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodManual,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Lines: []ScheduleLineRequest{
						{Amount: 100},
					},
				})
				if helper.AssertError(t, err, true, ErrScheduleInvalidDate) {
					return
				}
			},
		},
		{
			name: "propagates schedule create error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *DeferredSchedule) (*DeferredSchedule, error) {
						return nil, errors.New("insert failed")
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredRevenue,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodLinear,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Periods:               2,
				})
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates line create error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error) {
						schedule.ID = 7
						return schedule, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *DeferredScheduleLine) (*DeferredScheduleLine, error) {
						return nil, errors.New("insert failed")
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateScheduleRequest{
					OrganizationID:        10,
					Type:                  DeferredTypeDeferredRevenue,
					TotalAmount:           100,
					BalanceSheetAccountID: 300,
					PLAccountID:           200,
					Method:                DeferredMethodLinear,
					DateStart:             time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Periods:               2,
				})
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestDeferralService_RecognizeDue(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "revenue debits deferred credits revenue",
			run: func(t *testing.T) {
				var posted PostRequest
				var updatedLine *DeferredScheduleLine
				var updatedSchedule *DeferredSchedule
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							Type:                  DeferredTypeDeferredRevenue,
							SourceID:              3,
							TotalAmount:           240,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
							State:                 DeferredStateRunning,
						}}, nil
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error) {
						updatedSchedule = schedule
						return schedule, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*DeferredScheduleLine, error) {
						return []*DeferredScheduleLine{{
							Base:            model.Base{ID: 11},
							Sequence:        1,
							Amount:          240,
							RecognitionDate: helper.Ptr(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)),
						}}, nil
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error) {
						updatedLine = line
						return line, nil
					},
				}
				poster := PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
						posted = request
						return &JournalEntry{Base: model.Base{ID: 99}}, nil
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, poster, JournalResolverMock{}, TransactionerMock{})

				postedCount, err := svc.RecognizeDue(ctx, nil, asOf)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if postedCount != 1 {
					t.Errorf("posted = %d, want 1", postedCount)
				}
				if len(posted.Lines) != 2 {
					t.Fatalf("posting lines = %d, want 2", len(posted.Lines))
				}
				if posted.Lines[0].AccountID != 300 || !posted.Lines[0].Debit.Equal(amount.FromFloat64(240)) || posted.Lines[1].AccountID != 200 || !posted.Lines[1].Credit.Equal(amount.FromFloat64(240)) {
					t.Errorf("posting lines = %+v, want Dr 300 / Cr 200 for 240", posted.Lines)
				}
				if !updatedLine.Posted || updatedLine.EntryID == nil || *updatedLine.EntryID != 99 {
					t.Errorf("line not marked posted = %+v", updatedLine)
				}
				if updatedSchedule.State != DeferredStateDone || updatedSchedule.RecognizedAmount != 240 {
					t.Errorf("schedule = %+v, want done with recognized 240", updatedSchedule)
				}
			},
		},
		{
			name: "prepaid debits expense credits prepaid",
			run: func(t *testing.T) {
				var posted PostRequest
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							Type:                  DeferredTypePrepaid,
							TotalAmount:           120,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
							State:                 DeferredStateRunning,
						}}, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*DeferredScheduleLine, error) {
						return []*DeferredScheduleLine{{
							Base:            model.Base{ID: 11},
							Amount:          120,
							RecognitionDate: helper.Ptr(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)),
						}}, nil
					},
				}
				poster := PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
						posted = request
						return &JournalEntry{Base: model.Base{ID: 99}}, nil
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, poster, JournalResolverMock{}, TransactionerMock{})

				if _, err := svc.RecognizeDue(ctx, nil, asOf); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if posted.OriginType != DeferredTypePrepaid {
					t.Errorf("origin type = %s, want %s", posted.OriginType, DeferredTypePrepaid)
				}
				if posted.Lines[0].AccountID != 200 || posted.Lines[1].AccountID != 300 {
					t.Errorf("posting lines = %+v, want Dr PL 200 / Cr BS 300", posted.Lines)
				}
			},
		},
		{
			name: "requires configured journal",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							TotalAmount:           240,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
							State:                 DeferredStateRunning,
						}}, nil
					},
				}
				journals := JournalResolverMock{
					JournalIDFunc: func(_ context.Context, _ uint64) (uint64, error) { return 0, nil },
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, journals, TransactionerMock{})

				if helper.AssertError(t, mustRecognize(ctx, svc), true, ErrScheduleNoJournal) {
					return
				}
			},
		},
		{
			name: "stays running until fully recognized",
			run: func(t *testing.T) {
				var updatedSchedule *DeferredSchedule
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							TotalAmount:           240,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
							State:                 DeferredStateRunning,
						}}, nil
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error) {
						updatedSchedule = schedule
						return schedule, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*DeferredScheduleLine, error) {
						return []*DeferredScheduleLine{{
							Base:            model.Base{ID: 11},
							Amount:          120,
							RecognitionDate: helper.Ptr(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)),
						}}, nil
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				if _, err := svc.RecognizeDue(ctx, nil, asOf); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updatedSchedule.State != DeferredStateRunning || updatedSchedule.RecognizedAmount != 120 {
					t.Errorf("schedule = %+v, want running with recognized 120", updatedSchedule)
				}
			},
		},
		{
			name: "scopes to organization",
			run: func(t *testing.T) {
				var scopedTo *uint64
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, organizationID *uint64) ([]*DeferredSchedule, error) {
						scopedTo = organizationID
						return []*DeferredSchedule{}, nil
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				if _, err := svc.RecognizeDue(ctx, helper.Ptr(uint64(42)), asOf); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if scopedTo == nil || *scopedTo != 42 {
					t.Errorf("scoped to = %v, want 42", scopedTo)
				}
			},
		},
		{
			name: "propagates list running error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.RecognizeDue(ctx, nil, asOf)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "skips schedule without organization",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{Base: model.Base{ID: 7}, Type: DeferredTypeDeferredRevenue}}, nil
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				posted, err := svc.RecognizeDue(ctx, nil, asOf)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if posted != 0 {
					t.Errorf("posted = %d, want 0", posted)
				}
			},
		},
		{
			name: "rejects missing accounts",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), Type: DeferredTypeDeferredRevenue}}, nil
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.RecognizeDue(ctx, nil, asOf)
				if helper.AssertError(t, err, true, ErrScheduleNoAccount) {
					return
				}
			},
		},
		{
			name: "propagates journal error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							Type:                  DeferredTypeDeferredRevenue,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
						}}, nil
					},
				}
				journals := JournalResolverMock{
					JournalIDFunc: func(_ context.Context, _ uint64) (uint64, error) { return 0, errors.New("db down") },
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, journals, TransactionerMock{})

				_, err := svc.RecognizeDue(ctx, nil, asOf)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "skips schedule without due lines",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							Type:                  DeferredTypeDeferredRevenue,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
						}}, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*DeferredScheduleLine, error) {
						return []*DeferredScheduleLine{}, nil
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				posted, err := svc.RecognizeDue(ctx, nil, asOf)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if posted != 0 {
					t.Errorf("posted = %d, want 0", posted)
				}
			},
		},
		{
			name: "propagates list due error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							Type:                  DeferredTypeDeferredRevenue,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
						}}, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*DeferredScheduleLine, error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.RecognizeDue(ctx, nil, asOf)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates poster error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							Type:                  DeferredTypeDeferredRevenue,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
						}}, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*DeferredScheduleLine, error) {
						return []*DeferredScheduleLine{{Base: model.Base{ID: 11}, Amount: 120, RecognitionDate: helper.Ptr(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))}}, nil
					},
				}
				poster := PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, _ PostRequest) (*JournalEntry, error) {
						return nil, errors.New("post failed")
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, poster, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.RecognizeDue(ctx, nil, asOf)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates line update error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							Type:                  DeferredTypeDeferredRevenue,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
						}}, nil
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*DeferredScheduleLine, error) {
						return []*DeferredScheduleLine{{Base: model.Base{ID: 11}, Amount: 120, RecognitionDate: helper.Ptr(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))}}, nil
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *DeferredScheduleLine) (*DeferredScheduleLine, error) {
						return nil, errors.New("update failed")
					},
				}
				poster := PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, _ PostRequest) (*JournalEntry, error) {
						return &JournalEntry{Base: model.Base{ID: 99}}, nil
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, poster, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.RecognizeDue(ctx, nil, asOf)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates schedule update error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					ListRunningFunc: func(_ context.Context, _ *uint64) ([]*DeferredSchedule, error) {
						return []*DeferredSchedule{{
							Base:                  model.Base{ID: 7},
							OrganizationID:        helper.Ptr(uint64(10)),
							Type:                  DeferredTypeDeferredRevenue,
							BalanceSheetAccountID: helper.Ptr(uint64(300)),
							PLAccountID:           helper.Ptr(uint64(200)),
						}}, nil
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *DeferredSchedule) (*DeferredSchedule, error) {
						return nil, errors.New("update failed")
					},
				}
				scheduleLines := DeferredScheduleLineDAOMock{
					ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*DeferredScheduleLine, error) {
						return []*DeferredScheduleLine{{Base: model.Base{ID: 11}, Amount: 120, RecognitionDate: helper.Ptr(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))}}, nil
					},
				}
				svc := NewDeferralService(schedules, scheduleLines, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.RecognizeDue(ctx, nil, asOf)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestDeferralService_Get(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "scopes to organization",
			run: func(t *testing.T) {
				svc := NewDeferralService(DeferredScheduleDAOMock{
					CRUDMock: dao.CRUDMock[DeferredSchedule]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[DeferredSchedule], error) {
							return &query.Page[DeferredSchedule]{Items: []*DeferredSchedule{{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99))}}}, nil
						},
					},
				}, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				if helper.AssertError(t, mustGet(ctx, svc), true, ErrScheduleNotFound) {
					return
				}
			},
		},
		{
			name: "returns schedule",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					CRUDMock: dao.CRUDMock[DeferredSchedule]{
						SearchFunc: func(_ context.Context, field string, value any) (*DeferredSchedule, error) {
							if field != "id" || value != uint64(7) {
								t.Errorf("search = %s/%v, want id/7", field, value)
							}
							return &DeferredSchedule{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10))}, nil
						},
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				schedule, err := svc.Get(ctx, 10, 7)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if schedule.ID != 7 {
					t.Errorf("schedule = %+v, want id 7", schedule)
				}
			},
		},
		{
			name: "propagates search error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					CRUDMock: dao.CRUDMock[DeferredSchedule]{
						SearchFunc: func(_ context.Context, _ string, _ any) (*DeferredSchedule, error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.Get(ctx, 10, 7)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "rejects schedule without organization",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					CRUDMock: dao.CRUDMock[DeferredSchedule]{
						SearchFunc: func(_ context.Context, _ string, _ any) (*DeferredSchedule, error) {
							return &DeferredSchedule{Base: model.Base{ID: 7}}, nil
						},
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				if helper.AssertError(t, mustGet(ctx, svc), true, ErrScheduleNotFound) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestDeferralService_List(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "returns organization schedules",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					CRUDMock: dao.CRUDMock[DeferredSchedule]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[DeferredSchedule], error) {
							if len(q.Filters) != 1 || q.Filters[0].Field != "organization_id" || q.Filters[0].Value != uint64(10) {
								t.Errorf("filter = %+v, want organization_id=10", q.Filters)
							}
							return &query.Page[DeferredSchedule]{Items: []*DeferredSchedule{
								{OrganizationID: helper.Ptr(uint64(10)), Type: DeferredTypeDeferredRevenue},
								{OrganizationID: helper.Ptr(uint64(10)), Type: DeferredTypePrepaid},
							}}, nil
						},
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				items, err := svc.List(ctx, 10)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(items) != 2 || items[0].Type != DeferredTypeDeferredRevenue || items[1].Type != DeferredTypePrepaid {
					t.Errorf("items = %+v, want two schedules", items)
				}
			},
		},
		{
			name: "propagates error",
			run: func(t *testing.T) {
				schedules := DeferredScheduleDAOMock{
					CRUDMock: dao.CRUDMock[DeferredSchedule]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[DeferredSchedule], error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := NewDeferralService(schedules, DeferredScheduleLineDAOMock{}, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				_, err := svc.List(ctx, 10)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestDeferralService_ListLines(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "returns schedule lines",
			run: func(t *testing.T) {
				lines := DeferredScheduleLineDAOMock{
					ListByScheduleFunc: func(_ context.Context, scheduleID uint64) ([]*DeferredScheduleLine, error) {
						if scheduleID != 7 {
							t.Errorf("schedule id = %d, want 7", scheduleID)
						}
						return []*DeferredScheduleLine{{Sequence: 1, Amount: 100}, {Sequence: 2, Amount: 100}}, nil
					},
				}
				svc := NewDeferralService(DeferredScheduleDAOMock{}, lines, PosterMock{}, JournalResolverMock{}, TransactionerMock{})

				items, err := svc.ListLines(ctx, 7)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(items) != 2 || items[1].Sequence != 2 {
					t.Errorf("items = %+v, want two lines", items)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func mustRecognize(ctx context.Context, svc DeferralService) (err error) {
	_, err = svc.RecognizeDue(ctx, nil, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	return err
}

func mustGet(ctx context.Context, svc DeferralService) (err error) {
	_, err = svc.Get(ctx, 10, 7)
	return err
}
