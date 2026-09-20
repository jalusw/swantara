package xtradata

import (
	"context"
	"encoding/json"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ConfigService struct {
	configs SystemConfigDAO
}

func NewConfigService(configs SystemConfigDAO) ConfigService {
	return ConfigService{configs: configs}
}

func (s ConfigService) Create(ctx context.Context, config *reference.SystemConfig) (*reference.SystemConfig, error) {
	existing, err := s.configs.Search(ctx, "key", config.Key)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrConfigKeyExists
	}
	return s.configs.Create(ctx, config)
}

func (s ConfigService) Update(ctx context.Context, config *reference.SystemConfig) (*reference.SystemConfig, error) {
	return s.configs.Update(ctx, config)
}

func (s ConfigService) List(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error) {
	return s.configs.List(ctx, q)
}

func (s ConfigService) Find(ctx context.Context, id uint64) (*reference.SystemConfig, error) {
	return s.configs.Find(ctx, id)
}

func (s ConfigService) Delete(ctx context.Context, id uint64) error {
	return s.configs.Delete(ctx, id)
}

type EventService struct {
	events   IntegrationEventDAO
	enqueuer queue.TaskEnqueuer
	deliver  func(ctx context.Context, event *IntegrationEvent) error
}

func NewEventService(events IntegrationEventDAO, enqueuer queue.TaskEnqueuer) EventService {
	return EventService{
		events:   events,
		enqueuer: enqueuer,
		deliver:  func(context.Context, *IntegrationEvent) error { return nil },
	}
}

func (s EventService) List(ctx context.Context, q *query.Query) (*query.Page[IntegrationEvent], error) {
	return s.events.List(ctx, q)
}

func (s EventService) Find(ctx context.Context, id uint64) (*IntegrationEvent, error) {
	return s.events.Find(ctx, id)
}

func (s EventService) Enqueue(ctx context.Context, organizationID *uint64, topic string, payload any) (*IntegrationEvent, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	event, err := s.events.Create(ctx, &IntegrationEvent{
		OrganizationID: organizationID,
		Topic:          topic,
		Payload:        raw,
		Status:         IntegrationEventStatusPending,
	})
	if err != nil {
		return nil, err
	}

	task, err := tasks.NewIntegrationEventDispatchTask(event.ID, helper.RequestID(ctx))
	if err != nil {
		return event, err
	}
	if _, err := s.enqueuer.Enqueue(task); err != nil {
		return event, err
	}
	return event, nil
}

func (s EventService) Dispatch(ctx context.Context, id uint64) error {
	event, err := s.events.Find(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return ErrEventNotFound
	}
	if event.Status == IntegrationEventStatusSent {
		return ErrEventAlreadySent
	}

	if err := s.deliver(ctx, event); err != nil {
		return s.fail(ctx, event, err)
	}

	event.Status = IntegrationEventStatusSent
	_, err = s.events.Update(ctx, event)
	return err
}

func (s EventService) fail(ctx context.Context, event *IntegrationEvent, cause error) error {
	event.Retries++
	event.Status = IntegrationEventStatusFailed
	if _, err := s.events.Update(ctx, event); err != nil {
		return err
	}
	return cause
}

func (s EventService) WithDeliver(deliver func(ctx context.Context, event *IntegrationEvent) error) EventService {
	s.deliver = deliver
	return s
}

type IdempotencyService struct {
	keys IdempotencyKeyDAO
}

func NewIdempotencyService(keys IdempotencyKeyDAO) IdempotencyService {
	return IdempotencyService{keys: keys}
}

func (s IdempotencyService) Claim(ctx context.Context, organizationID *uint64, key, resource string) (*httpx.IdempotencyEntry, error) {
	claimed, existing, err := s.keys.Claim(ctx, organizationID, key, resource)
	if err != nil {
		return nil, err
	}
	if claimed {
		return nil, nil
	}
	if existing.Resource != resource {
		return nil, httpx.ErrIdempotencyKeyReused
	}
	if existing.Response == nil {
		return nil, httpx.ErrIdempotencyKeyInFlight
	}
	return &httpx.IdempotencyEntry{
		StatusCode: existing.StatusCode,
		Response:   existing.Response,
	}, nil
}

func (s IdempotencyService) Complete(ctx context.Context, organizationID *uint64, key string, statusCode int, response []byte) error {
	return s.keys.Complete(ctx, organizationID, key, statusCode, response)
}

func (s IdempotencyService) Release(ctx context.Context, organizationID *uint64, key string) error {
	return s.keys.Release(ctx, organizationID, key)
}
