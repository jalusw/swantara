package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"time"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/jobs"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/mail"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
)

func testMailer(t *testing.T) *mail.Mailer {
	t.Helper()
	m, err := mail.New(&config.Config{
		SMTPHost:      "localhost",
		SMTPPort:      1025,
		SMTPFromEmail: "noreply@swantara.com",
		SMTPFromName:  "Swantara",
	})
	if err != nil {
		t.Fatalf("mail.New() error = %v", err)
	}
	return m
}

func testJobService(create, update func(ctx context.Context, run *jobs.JobRun) (*jobs.JobRun, error)) jobs.Service {
	return jobs.NewService(jobs.DAOMock{CreateFunc: create, UpdateFunc: update})
}

func testEventService() xtradata.EventService {
	return xtradata.NewEventService(xtradata.IntegrationEventDAOMock{}, queue.TaskEnqueuerMock{})
}

func testDeferralService() accounting.DeferralService {
	return accounting.NewDeferralService(
		accounting.DeferredScheduleDAOMock{},
		accounting.DeferredScheduleLineDAOMock{},
		subscription.PosterMock{},
		subscription.SubscriptionConfigSourceMock{},
		subscription.TransactionerMock{},
	)
}

func testSubscriptionService(deferrals accounting.DeferralService) subscription.SubscriptionService {
	return subscription.NewSubscriptionService(
		subscription.SubscriptionDAOMock{},
		subscription.SubscriptionLineDAOMock{},
		dao.CRUDMock[reference.SubscriptionPlan]{},
		contacts.ContactDAOMock{},
		products.PriceBookDAOMock{},
		subscription.IncomeAccountResolverMock{},
		subscription.InvoiceEngineMock{},
		deferrals,
		subscription.SubscriptionConfigSourceMock{},
		subscription.TransactionerMock{},
	)
}

func asynqTask(t *testing.T, typ string, payload any) *asynq.Task {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload error = %v", err)
	}
	return asynq.NewTask(typ, raw)
}

func TestSendEmailHandler_HandlesInvalidPayload(t *testing.T) {
	handler := NewSendEmailHandler(testMailer(t), iam.UserDAOMock{}, mail.EmailVerificationTemplate, "http://localhost/verify", "VerifyLink", "Verify")
	err := handler.Handle(context.Background(), asynqTask(t, tasks.TypeSendEmailVerification, "not-an-object"))
	if err == nil {
		t.Fatal("Handle() expected unmarshal error")
	}
}

func TestSendEmailHandler_FindUserError(t *testing.T) {
	handler := NewSendEmailHandler(testMailer(t), iam.UserDAOMock{DAOMock: iam.DAOMock[iam.User]{FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
		return nil, errors.New("db error")
	}}}, mail.EmailVerificationTemplate, "http://localhost/verify", "VerifyLink", "Verify")
	err := handler.Handle(context.Background(), asynqTask(t, tasks.TypeSendEmailVerification, tasks.SendEmailPayload{UserID: 1}))
	if err == nil {
		t.Fatal("Handle() expected error when user lookup fails")
	}
}

func TestSendEmailHandler_SkipsMissingUser(t *testing.T) {
	handler := NewSendEmailHandler(testMailer(t), iam.UserDAOMock{}, mail.EmailVerificationTemplate, "http://localhost/verify", "VerifyLink", "Verify")
	err := handler.Handle(context.Background(), asynqTask(t, tasks.TypeSendEmailVerification, tasks.SendEmailPayload{UserID: 99}))
	if err != nil {
		t.Fatalf("Handle() error = %v, want nil skip", err)
	}
}

func TestSendEmailHandler_InvalidTemplate(t *testing.T) {
	handler := NewSendEmailHandler(testMailer(t), iam.UserDAOMock{DAOMock: iam.DAOMock[iam.User]{FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
		return &iam.User{FirstName: "Ada", Email: "ada@example.com"}, nil
	}}}, "templates/missing.html", "http://localhost/verify", "VerifyLink", "Verify")
	err := handler.Handle(context.Background(), asynqTask(t, tasks.TypeSendEmailVerification, tasks.SendEmailPayload{UserID: 1}))
	if err == nil {
		t.Fatal("Handle() expected template parse error")
	}
}

func TestSendEmailHandler_SendError(t *testing.T) {
	handler := NewSendEmailHandler(testMailer(t), iam.UserDAOMock{DAOMock: iam.DAOMock[iam.User]{FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
		return &iam.User{FirstName: "Ada", Email: "not-an-email"}, nil
	}}}, mail.EmailVerificationTemplate, "http://localhost/verify", "VerifyLink", "Verify")
	err := handler.Handle(context.Background(), asynqTask(t, tasks.TypeSendEmailVerification, tasks.SendEmailPayload{UserID: 1}))
	if err == nil {
		t.Fatal("Handle() expected send error")
	}
}

func TestIntegrationEventDispatchHandler(t *testing.T) {
	tests := []struct {
		name    string
		events  xtradata.EventService
		payload any
		wantErr bool
	}{
		{
			name:    "invalid payload",
			payload: "not-an-object",
			wantErr: true,
		},
		{
			name: "dispatch error",
			events: xtradata.NewEventService(xtradata.IntegrationEventDAOMock{FindFunc: func(_ context.Context, _ uint64) (*xtradata.IntegrationEvent, error) {
				return nil, errors.New("db error")
			}}, queue.TaskEnqueuerMock{}),
			payload: tasks.IntegrationEventDispatchPayload{EventID: 1},
			wantErr: true,
		},
		{
			name: "success",
			events: xtradata.NewEventService(xtradata.IntegrationEventDAOMock{FindFunc: func(_ context.Context, _ uint64) (*xtradata.IntegrationEvent, error) {
				return &xtradata.IntegrationEvent{Base: model.Base{ID: 1}, Status: xtradata.IntegrationEventStatusPending}, nil
			}}, queue.TaskEnqueuerMock{}),
			payload: tasks.IntegrationEventDispatchPayload{EventID: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewIntegrationEventDispatchHandler(tt.events)
			err := handler.Handle(context.Background(), asynqTask(t, tasks.TypeIntegrationEventDispatch, tt.payload))
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeferralRecognitionHandler(t *testing.T) {
	invalidPayload := asynqTask(t, tasks.TypeDeferralRecognition, "not-an-object")

	tests := []struct {
		name    string
		handler func() DeferralRecognitionHandler
		task    *asynq.Task
		wantErr bool
	}{
		{
			name: "invalid payload",
			handler: func() DeferralRecognitionHandler {
				return NewDeferralRecognitionHandler(testDeferralService(), testJobService(nil, nil), testEventService())
			},
			task:    invalidPayload,
			wantErr: true,
		},
		{
			name: "job start failure",
			handler: func() DeferralRecognitionHandler {
				return NewDeferralRecognitionHandler(testDeferralService(), testJobService(func(_ context.Context, _ *jobs.JobRun) (*jobs.JobRun, error) {
					return nil, errors.New("db error")
				}, nil), testEventService())
			},
			task:    asynqTask(t, tasks.TypeDeferralRecognition, tasks.DeferralRecognitionPayload{RequestID: "r1"}),
			wantErr: true,
		},
		{
			name: "recognition failure marks job failed",
			handler: func() DeferralRecognitionHandler {
				return NewDeferralRecognitionHandler(accounting.NewDeferralService(
					accounting.DeferredScheduleDAOMock{ListRunningFunc: func(_ context.Context, _ *uint64) ([]*accounting.DeferredSchedule, error) {
						return nil, errors.New("db error")
					}},
					accounting.DeferredScheduleLineDAOMock{},
					subscription.PosterMock{},
					subscription.SubscriptionConfigSourceMock{},
					subscription.TransactionerMock{},
				), testJobService(func(_ context.Context, _ *jobs.JobRun) (*jobs.JobRun, error) {
					return &jobs.JobRun{Base: model.Base{ID: 1}}, nil
				}, nil), testEventService())
			},
			task:    asynqTask(t, tasks.TypeDeferralRecognition, tasks.DeferralRecognitionPayload{}),
			wantErr: true,
		},
		{
			name: "success",
			handler: func() DeferralRecognitionHandler {
				return NewDeferralRecognitionHandler(testDeferralService(), testJobService(func(_ context.Context, _ *jobs.JobRun) (*jobs.JobRun, error) {
					return &jobs.JobRun{Base: model.Base{ID: 1}}, nil
				}, nil), testEventService())
			},
			task: asynqTask(t, tasks.TypeDeferralRecognition, tasks.DeferralRecognitionPayload{}),
		},
		{
			name: "job completion failure",
			handler: func() DeferralRecognitionHandler {
				return NewDeferralRecognitionHandler(testDeferralService(), testJobService(func(_ context.Context, _ *jobs.JobRun) (*jobs.JobRun, error) {
					return &jobs.JobRun{Base: model.Base{ID: 1}}, nil
				}, func(_ context.Context, _ *jobs.JobRun) (*jobs.JobRun, error) {
					return nil, errors.New("db error")
				}), testEventService())
			},
			task:    asynqTask(t, tasks.TypeDeferralRecognition, tasks.DeferralRecognitionPayload{}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.handler().Handle(context.Background(), tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubscriptionBillingHandler(t *testing.T) {
	valid := func() SubscriptionBillingHandler {
		return NewSubscriptionBillingHandler(
			testSubscriptionService(testDeferralService()),
			testDeferralService(),
			testJobService(func(_ context.Context, _ *jobs.JobRun) (*jobs.JobRun, error) {
				return &jobs.JobRun{Base: model.Base{ID: 1}}, nil
			}, nil),
			testEventService(),
		)
	}

	tests := []struct {
		name    string
		handler func() SubscriptionBillingHandler
		task    *asynq.Task
		wantErr bool
	}{
		{
			name:    "invalid payload",
			handler: valid,
			task:    asynqTask(t, tasks.TypeSubscriptionBilling, "not-an-object"),
			wantErr: true,
		},
		{
			name: "job start failure",
			handler: func() SubscriptionBillingHandler {
				return NewSubscriptionBillingHandler(
					testSubscriptionService(testDeferralService()),
					testDeferralService(),
					testJobService(func(_ context.Context, _ *jobs.JobRun) (*jobs.JobRun, error) {
						return nil, errors.New("db error")
					}, nil),
					testEventService(),
				)
			},
			task:    asynqTask(t, tasks.TypeSubscriptionBilling, tasks.SubscriptionBillingPayload{}),
			wantErr: true,
		},
		{
			name:    "success",
			handler: valid,
			task:    asynqTask(t, tasks.TypeSubscriptionBilling, tasks.SubscriptionBillingPayload{}),
		},
		{
			name: "billing failure marks job failed",
			handler: func() SubscriptionBillingHandler {
				subscriptions := subscription.SubscriptionDAOMock{ListDueFunc: func(_ context.Context, _ time.Time) ([]*subscription.Subscription, error) {
					return nil, errors.New("db error")
				}}
				svc := subscription.NewSubscriptionService(
					subscriptions,
					subscription.SubscriptionLineDAOMock{},
					dao.CRUDMock[reference.SubscriptionPlan]{},
					contacts.ContactDAOMock{},
					products.PriceBookDAOMock{},
					subscription.IncomeAccountResolverMock{},
					subscription.InvoiceEngineMock{},
					testDeferralService(),
					subscription.SubscriptionConfigSourceMock{},
					subscription.TransactionerMock{},
				)
				return NewSubscriptionBillingHandler(
					svc,
					testDeferralService(),
					testJobService(func(_ context.Context, _ *jobs.JobRun) (*jobs.JobRun, error) {
						return &jobs.JobRun{Base: model.Base{ID: 1}}, nil
					}, nil),
					testEventService(),
				)
			},
			task:    asynqTask(t, tasks.TypeSubscriptionBilling, tasks.SubscriptionBillingPayload{}),
			wantErr: true,
		},
		{
			name: "deferral recognition failure",
			handler: func() SubscriptionBillingHandler {
				return NewSubscriptionBillingHandler(
					testSubscriptionService(testDeferralService()),
					accounting.NewDeferralService(
						accounting.DeferredScheduleDAOMock{ListRunningFunc: func(_ context.Context, _ *uint64) ([]*accounting.DeferredSchedule, error) {
							return nil, errors.New("db error")
						}},
						accounting.DeferredScheduleLineDAOMock{},
						subscription.PosterMock{},
						subscription.SubscriptionConfigSourceMock{},
						subscription.TransactionerMock{},
					),
					testJobService(func(_ context.Context, _ *jobs.JobRun) (*jobs.JobRun, error) {
						return &jobs.JobRun{Base: model.Base{ID: 1}}, nil
					}, nil),
					testEventService(),
				)
			},
			task:    asynqTask(t, tasks.TypeSubscriptionBilling, tasks.SubscriptionBillingPayload{}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.handler().Handle(context.Background(), tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestPublishRunEvent(t *testing.T) {
	ctx := context.Background()
	events := testEventService()
	publishRunEvent(ctx, events, "deferral.recognition.completed", map[string]any{"recognized": 3})

	failing := xtradata.NewEventService(xtradata.IntegrationEventDAOMock{CreateFunc: func(_ context.Context, _ *xtradata.IntegrationEvent) (*xtradata.IntegrationEvent, error) {
		return nil, errors.New("db error")
	}}, queue.TaskEnqueuerMock{})
	publishRunEvent(ctx, failing, "deferral.recognition.completed", map[string]any{"recognized": 3})
}
