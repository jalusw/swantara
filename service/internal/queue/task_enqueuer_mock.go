package queue

import "github.com/hibiken/asynq"

type TaskEnqueuerMock struct {
	EnqueueFunc func(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

func (m TaskEnqueuerMock) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if m.EnqueueFunc != nil {
		return m.EnqueueFunc(task, opts...)
	}
	return nil, nil
}
