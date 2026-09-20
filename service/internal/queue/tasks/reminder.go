package tasks

import "github.com/hibiken/asynq"

const TypeSendReminderEmail = "send-reminder-email"

type SendReminderPayload struct {
	ActionID  uint64 `json:"action_id"`
	ContactID uint64 `json:"contact_id"`
	InvoiceID uint64 `json:"invoice_id"`
}

func NewSendReminderTask(actionID, contactID, invoiceID uint64) (*asynq.Task, error) {
	return newTask(TypeSendReminderEmail, SendReminderPayload{ActionID: actionID, ContactID: contactID, InvoiceID: invoiceID})
}
