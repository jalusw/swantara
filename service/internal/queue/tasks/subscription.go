package tasks

import "github.com/hibiken/asynq"

const TypeSubscriptionBilling = "subscription-billing-daily"

type SubscriptionBillingPayload struct {
	RequestID string `json:"request_id,omitempty"`
}

func NewSubscriptionBillingTask(requestID string) (*asynq.Task, error) {
	return newTask(TypeSubscriptionBilling, SubscriptionBillingPayload{RequestID: requestID})
}
