package tasks

import "github.com/hibiken/asynq"

const (
	TypeSendEmailVerification = "send-email-verification"
	TypeSendPasswordReset     = "send-password-reset"
)

type SendEmailPayload struct {
	UserID    uint64 `json:"user_id"`
	Token     string `json:"token"`
	RequestID string `json:"request_id,omitempty"`
}

func NewSendEmailVerificationTask(userID uint64, token, requestID string) (*asynq.Task, error) {
	return newTask(TypeSendEmailVerification, SendEmailPayload{UserID: userID, Token: token, RequestID: requestID})
}

func NewSendPasswordResetTask(userID uint64, token, requestID string) (*asynq.Task, error) {
	return newTask(TypeSendPasswordReset, SendEmailPayload{UserID: userID, Token: token, RequestID: requestID})
}
