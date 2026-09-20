package tasks

import "github.com/hibiken/asynq"

const TypeIntegrationEventDispatch = "integration-event-dispatch"

type IntegrationEventDispatchPayload struct {
	EventID   uint64 `json:"event_id"`
	RequestID string `json:"request_id,omitempty"`
}

func NewIntegrationEventDispatchTask(eventID uint64, requestID string) (*asynq.Task, error) {
	return newTask(TypeIntegrationEventDispatch, IntegrationEventDispatchPayload{
		EventID:   eventID,
		RequestID: requestID,
	})
}
