package crosscutting

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestCrosscuttingService_Create(t *testing.T) {
	tests := []struct {
		name    string
		orgID   uint64
		ownerID uint64
		steps   []uint64
		setup   func(requests *ApprovalRequestDAOMock)
		wantID  uint64
		wantErr error
	}{
		{
			name:    "creates approval request and steps with correct organization",
			orgID:   7,
			ownerID: 10,
			steps:   []uint64{3, 4},
			wantID:  7,
		},
		{
			name:    "rejects zero organization",
			orgID:   0,
			ownerID: 10,
			steps:   []uint64{3},
			wantErr: ErrApprovalOrganization,
		},
		{
			name:    "rejects nil approvers",
			orgID:   7,
			ownerID: 10,
			steps:   nil,
			wantErr: ErrApprovalNoApprovers,
		},
		{
			name:    "rejects zero owner",
			orgID:   7,
			ownerID: 0,
			steps:   []uint64{3},
			wantErr: ErrApprovalOwner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, requests, _ := testApprovalService()
			var createdSteps []*ApprovalStep
			requests.CreateWithStepsFunc = func(_ context.Context, request *ApprovalRequest, steps []*ApprovalStep) (*ApprovalRequest, error) {
				request.ID = 7
				createdSteps = steps
				return request, nil
			}

			if tt.setup != nil {
				tt.setup(requests)
			}

			request, err := svc.Create(context.Background(), tt.orgID, "purchase_order", tt.ownerID, 2, tt.steps)

			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			if request.ID != tt.wantID {
				t.Fatalf("expected created approval request id %d, got %d", tt.wantID, request.ID)
			}
			if request.OrganizationID != tt.orgID {
				t.Errorf("request organization = %d, want %d", request.OrganizationID, tt.orgID)
			}
			if len(createdSteps) != len(tt.steps) {
				t.Fatalf("created steps = %d, want %d", len(createdSteps), len(tt.steps))
			}
			for _, step := range createdSteps {
				if step.OrganizationID != tt.orgID {
					t.Errorf("step organization = %d, want %d", step.OrganizationID, tt.orgID)
				}
			}
		})
	}
}

func TestCrosscuttingService_Decide(t *testing.T) {
	tests := []struct {
		name       string
		orgID      uint64
		requestID  uint64
		sequence   uint64
		approver   uint64
		approve    bool
		reason     string
		setup      func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock)
		wantState  string
		wantErr    bool
		wantErrVal error
	}{
		{
			name:      "approves all steps when final approver decides",
			orgID:     7,
			requestID: 1,
			sequence:  2,
			approver:  4,
			approve:   true,
			reason:    "approved",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{
						{Base: model.Base{ID: 1}, RequestID: 1, ApproverID: 3, Sequence: 10, Decision: ApprovalStepApproved},
						{Base: model.Base{ID: 2}, RequestID: 1, ApproverID: 4, Sequence: 20, Decision: ApprovalStepPending},
					}, nil
				}
				requests.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, request *ApprovalRequest) (*ApprovalRequest, error) {
					return request, nil
				}
			},
			wantState: ApprovalStateApproved,
		},
		{
			name:      "refuses when approver rejects",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   false,
			reason:    "rejected price",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{
						{Base: model.Base{ID: 1}, RequestID: 1, ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending},
					}, nil
				}
				requests.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, request *ApprovalRequest) (*ApprovalRequest, error) {
					return request, nil
				}
			},
			wantState: ApprovalStateRefused,
		},
		{
			name:      "rejects out of order decision",
			orgID:     7,
			requestID: 1,
			sequence:  2,
			approver:  4,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{
						{Base: model.Base{ID: 1}, RequestID: 1, ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending},
						{Base: model.Base{ID: 2}, RequestID: 1, ApproverID: 4, Sequence: 20, Decision: ApprovalStepPending},
					}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrApprovalOrder,
		},
		{
			name:      "rejects wrong approver",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  99,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{
						{Base: model.Base{ID: 1}, RequestID: 1, ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending},
					}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrApprovalNotApprover,
		},
		{
			name:      "rejects decision on already decided request",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return &ApprovalRequest{Base: model.Base{ID: 1}, OrganizationID: 7, OwnerType: "purchase_order", OwnerID: 10, RequestedBy: 2, State: ApprovalStateApproved}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrApprovalState,
		},
		{
			name:      "rejects foreign organization",
			orgID:     8,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrApprovalNotFound,
		},
		{
			name:      "fails when request lookup fails",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: true,
		},
		{
			name:      "rejects missing request",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrApprovalNotFound,
		},
		{
			name:      "fails when steps lookup fails",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: true,
		},
		{
			name:      "rejects missing steps",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrApprovalNotFound,
		},
		{
			name:      "rejects missing steps by default",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrApprovalNotFound,
		},
		{
			name:      "rejects missing step",
			orgID:     7,
			requestID: 1,
			sequence:  99,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{{Base: model.Base{ID: 1}, ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending}}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrApprovalStepNotFound,
		},
		{
			name:      "rejects decided step",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{{Base: model.Base{ID: 1}, ApproverID: 3, Sequence: 10, Decision: ApprovalStepApproved}}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrApprovalState,
		},
		{
			name:      "keeps request pending when some steps remain",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{
						{Base: model.Base{ID: 1}, RequestID: 1, ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending},
						{Base: model.Base{ID: 2}, RequestID: 1, ApproverID: 4, Sequence: 20, Decision: ApprovalStepPending},
					}, nil
				}
			},
			wantState: ApprovalStatePending,
		},
		{
			name:      "fails when step update fails",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   false,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{{Base: model.Base{ID: 1}, ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending}}, nil
				}
				steps.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, _ *ApprovalStep) (*ApprovalStep, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: true,
		},
		{
			name:      "fails when request update fails",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   false,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{{Base: model.Base{ID: 1}, ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending}}, nil
				}
				requests.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, _ *ApprovalRequest) (*ApprovalRequest, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: true,
		},
		{
			name:      "propagates transaction error",
			orgID:     7,
			requestID: 1,
			sequence:  1,
			approver:  3,
			approve:   true,
			reason:    "",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindFunc = func(_ context.Context, id uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{{Base: model.Base{ID: 1}, ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending}}, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := &ApprovalRequestDAOMock{}
			steps := &ApprovalStepDAOMock{}
			var svc ApprovalService
			if tt.name == "propagates transaction error" {
				svc = NewApprovalService(requests, steps, TransactionerMock{RunFunc: func(_ context.Context, _ func(*gorm.DB) error) error {
					return errors.New("tx failed")
				}})
			} else {
				svc = NewApprovalService(requests, steps, TransactionerMock{})
			}
			if tt.setup != nil {
				tt.setup(requests, steps)
			}

			request, err := svc.Decide(context.Background(), tt.orgID, tt.requestID, tt.sequence, tt.approver, tt.approve, tt.reason)

			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantErr {
				return
			}
			if request.State != tt.wantState {
				t.Fatalf("expected request state %q, got %q", tt.wantState, request.State)
			}
		})
	}
}

func TestCrosscuttingService_State(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock)
		wantErr   bool
		wantNil   bool
		wantCount int
	}{
		{
			name: "returns request with steps",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindByOwnerFunc = func(_ context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return []*ApprovalStep{{Base: model.Base{ID: 1}, RequestID: 1, ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending}}, nil
				}
			},
			wantCount: 1,
		},
		{
			name: "fails when request lookup fails",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindByOwnerFunc = func(_ context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: true,
		},
		{
			name: "returns nil when no request",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindByOwnerFunc = func(_ context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error) {
					return nil, nil
				}
			},
			wantNil: true,
		},
		{
			name: "fails when steps lookup fails",
			setup: func(requests *ApprovalRequestDAOMock, steps *ApprovalStepDAOMock) {
				requests.FindByOwnerFunc = func(_ context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error) {
					return pendingRequest(), nil
				}
				steps.ListByRequestFunc = func(_ context.Context, requestID uint64) ([]*ApprovalStep, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, requests, steps := testApprovalService()
			tt.setup(requests, steps)

			request, stepList, err := svc.State(context.Background(), "purchase_order", 10)

			helper.AssertError(t, err, tt.wantErr, nil)
			if tt.wantNil && request != nil {
				t.Fatalf("request = %+v, want nil", request)
			}
			if tt.wantCount > 0 && (request == nil || len(stepList) != tt.wantCount) {
				t.Fatalf("expected request with %d steps", tt.wantCount)
			}
		})
	}
}

func TestCrosscuttingService_IsApproved(t *testing.T) {
	tests := []struct {
		name       string
		wantOK     bool
		setup      func(requests *ApprovalRequestDAOMock)
		wantErr    bool
		wantErrVal error
	}{
		{
			name:   "returns true when approved",
			wantOK: true,
			setup: func(requests *ApprovalRequestDAOMock) {
				requests.FindByOwnerFunc = func(_ context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error) {
					return &ApprovalRequest{Base: model.Base{ID: 1}, OwnerType: ownerType, OwnerID: ownerID, RequestedBy: 2, State: ApprovalStateApproved}, nil
				}
			},
		},
		{
			name:   "returns false when missing",
			wantOK: false,
			setup: func(requests *ApprovalRequestDAOMock) {
				requests.FindByOwnerFunc = func(_ context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error) {
					return nil, nil
				}
			},
		},
		{
			name: "fails on lookup error",
			setup: func(requests *ApprovalRequestDAOMock) {
				requests.FindByOwnerFunc = func(_ context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, requests, _ := testApprovalService()
			if tt.setup != nil {
				tt.setup(requests)
			}

			approved, err := svc.IsApproved(context.Background(), "purchase_order", 10)

			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if approved != tt.wantOK {
				t.Fatalf("expected approval to be %v", tt.wantOK)
			}
		})
	}
}

func testApprovalService() (ApprovalService, *ApprovalRequestDAOMock, *ApprovalStepDAOMock) {
	requests := &ApprovalRequestDAOMock{}
	steps := &ApprovalStepDAOMock{}
	svc := NewApprovalService(requests, steps, TransactionerMock{})
	return svc, requests, steps
}

func pendingRequest() *ApprovalRequest {
	return &ApprovalRequest{Base: model.Base{ID: 1}, OrganizationID: 7, OwnerType: "purchase_order", OwnerID: 10, RequestedBy: 2, State: ApprovalStatePending}
}
