package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestService_Start(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		create  func(ctx context.Context, run *JobRun) (*JobRun, error)
		wantErr bool
	}{
		{
			name: "creates running run",
			create: func(_ context.Context, run *JobRun) (*JobRun, error) {
				if run.Status != StatusRunning {
					t.Errorf("status = %q, want %q", run.Status, StatusRunning)
				}
				return run, nil
			},
		},
		{
			name: "dao error propagates",
			create: func(_ context.Context, _ *JobRun) (*JobRun, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{runs: DAOMock{CreateFunc: tt.create}, now: func() time.Time { return now }}
			run, err := svc.Start(ctx, TypeSubscriptionBilling)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if run == nil {
				t.Fatal("Start() returned nil run")
			}
			if !run.StartedAt.Equal(now) {
				t.Errorf("StartedAt = %v, want %v", run.StartedAt, now)
			}
		})
	}
}

func TestService_Complete(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		update  func(ctx context.Context, run *JobRun) (*JobRun, error)
		wantErr bool
	}{
		{
			name: "marks run completed",
			update: func(_ context.Context, run *JobRun) (*JobRun, error) {
				if run.Status != StatusCompleted {
					t.Errorf("status = %q, want %q", run.Status, StatusCompleted)
				}
				if run.Processed != 7 {
					t.Errorf("processed = %d, want 7", run.Processed)
				}
				if run.FinishedAt == nil || !run.FinishedAt.Equal(now) {
					t.Errorf("FinishedAt = %v, want %v", run.FinishedAt, now)
				}
				return run, nil
			},
		},
		{
			name: "dao error propagates",
			update: func(_ context.Context, _ *JobRun) (*JobRun, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{runs: DAOMock{UpdateFunc: tt.update}, now: func() time.Time { return now }}
			run := &JobRun{JobType: TypeDeferralRecognition}
			err := svc.Complete(ctx, run, 7)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if run.Status != StatusCompleted {
				t.Errorf("status = %q, want %q", run.Status, StatusCompleted)
			}
		})
	}
}

func TestService_Fail(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)

	svc := Service{runs: DAOMock{UpdateFunc: func(_ context.Context, run *JobRun) (*JobRun, error) {
		if run.Status != StatusFailed {
			t.Errorf("status = %q, want %q", run.Status, StatusFailed)
		}
		if run.ErrorMessage == nil || *run.ErrorMessage != "boom" {
			t.Errorf("ErrorMessage = %v, want boom", run.ErrorMessage)
		}
		return run, nil
	}}, now: func() time.Time { return now }}

	run := &JobRun{JobType: TypeIntegrationEventSweep}
	if err := svc.Fail(ctx, run, "boom"); err != nil {
		t.Fatalf("Fail() error = %v", err)
	}
	if run.Status != StatusFailed {
		t.Errorf("status = %q, want %q", run.Status, StatusFailed)
	}
}

func TestJobRun_TableName(t *testing.T) {
	if got := (JobRun{}).TableName(); got != "job_runs" {
		t.Errorf("TableName() = %q, want %q", got, "job_runs")
	}
}

func TestNewService(t *testing.T) {
	svc := NewService(DAOMock{})
	if svc.now == nil {
		t.Error("NewService() now function is nil")
	}
}

func TestNewJobRunDAO(t *testing.T) {
	db, mock := query.NewMockDB(t)
	dao := NewJobRunDAO(db)
	if dao == nil {
		t.Fatal("NewJobRunDAO() returned nil")
	}
	query.AssertDBMockDone(t, mock)
}

func TestDAOMockDefaults(t *testing.T) {
	ctx := context.Background()
	mock := DAOMock{}

	page, err := mock.List(ctx, nil)
	if err != nil || page.Count != 0 || len(page.Items) != 0 {
		t.Errorf("List() = (%v, %v), want empty page", page, err)
	}
	if got, err := mock.Search(ctx, "job_type", "x"); err != nil || got != nil {
		t.Errorf("Search() = (%v, %v), want (nil, nil)", got, err)
	}
	if got, err := mock.Find(ctx, 1); err != nil || got != nil {
		t.Errorf("Find() = (%v, %v), want (nil, nil)", got, err)
	}
	run := &JobRun{}
	if got, err := mock.Create(ctx, run); err != nil || got != run {
		t.Errorf("Create() = (%v, %v), want original entity", got, err)
	}
	if got, err := mock.Update(ctx, run); err != nil || got != run {
		t.Errorf("Update() = (%v, %v), want original entity", got, err)
	}
	if err := mock.Delete(ctx, 1); err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}
	if err := mock.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete() error = %v, want nil", err)
	}
}
