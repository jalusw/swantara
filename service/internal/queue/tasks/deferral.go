package tasks

import "github.com/hibiken/asynq"

const TypeDeferralRecognition = "deferral-recognition-daily"

type DeferralRecognitionPayload struct {
	RequestID string `json:"request_id,omitempty"`
}

func NewDeferralRecognitionTask(requestID string) (*asynq.Task, error) {
	return newTask(TypeDeferralRecognition, DeferralRecognitionPayload{RequestID: requestID})
}
