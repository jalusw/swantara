package tasks

import (
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
)

func TestTaskConstructors(t *testing.T) {
	tests := []struct {
		name   string
		typ    string
		task   func() (*asynq.Task, error)
		verify func(t *testing.T, task *asynq.Task)
	}{
		{
			name: "email verification",
			typ:  TypeSendEmailVerification,
			task: func() (*asynq.Task, error) {
				return NewSendEmailVerificationTask(7, "tok", "req-1")
			},
			verify: func(t *testing.T, task *asynq.Task) {
				var payload SendEmailPayload
				if err := json.Unmarshal(task.Payload(), &payload); err != nil {
					t.Fatalf("unmarshal error = %v", err)
				}
				if payload.UserID != 7 || payload.Token != "tok" || payload.RequestID != "req-1" {
					t.Errorf("payload = %+v", payload)
				}
			},
		},
		{
			name: "password reset",
			typ:  TypeSendPasswordReset,
			task: func() (*asynq.Task, error) {
				return NewSendPasswordResetTask(8, "tok2", "req-2")
			},
			verify: func(t *testing.T, task *asynq.Task) {
				var payload SendEmailPayload
				if err := json.Unmarshal(task.Payload(), &payload); err != nil {
					t.Fatalf("unmarshal error = %v", err)
				}
				if payload.UserID != 8 || payload.Token != "tok2" || payload.RequestID != "req-2" {
					t.Errorf("payload = %+v", payload)
				}
			},
		},
		{
			name: "deferral recognition",
			typ:  TypeDeferralRecognition,
			task: func() (*asynq.Task, error) {
				return NewDeferralRecognitionTask("req-3")
			},
		},
		{
			name: "subscription billing",
			typ:  TypeSubscriptionBilling,
			task: func() (*asynq.Task, error) {
				return NewSubscriptionBillingTask("req-4")
			},
		},
		{
			name: "integration event dispatch",
			typ:  TypeIntegrationEventDispatch,
			task: func() (*asynq.Task, error) {
				return NewIntegrationEventDispatchTask(42, "req-5")
			},
			verify: func(t *testing.T, task *asynq.Task) {
				var payload IntegrationEventDispatchPayload
				if err := json.Unmarshal(task.Payload(), &payload); err != nil {
					t.Fatalf("unmarshal error = %v", err)
				}
				if payload.EventID != 42 || payload.RequestID != "req-5" {
					t.Errorf("payload = %+v", payload)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := tt.task()
			if err != nil {
				t.Fatalf("task constructor error = %v", err)
			}
			if task.Type() != tt.typ {
				t.Errorf("Type() = %q, want %q", task.Type(), tt.typ)
			}
			if tt.verify != nil {
				tt.verify(t, task)
			}
		})
	}
}

func TestNewTask_MarshalError(t *testing.T) {
	_, err := newTask("bad", make(chan int))
	if err == nil {
		t.Fatal("newTask() expected marshal error for un-serializable payload")
	}
}
