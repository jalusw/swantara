package queue

import "errors"

var (
	ErrNoQueueWorkerRunning = errors.New("no queue worker is running")
)
