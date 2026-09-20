package queue

import (
	"testing"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/config"
)

func testQueueConfig() *config.Config {
	return &config.Config{
		QueueRedisHost:     "localhost",
		QueueRedisPort:     6379,
		QueueRedisPassword: "secret",
		QueueRedisDB:       2,
	}
}

func TestRedisClientOpt(t *testing.T) {
	opt := redisClientOpt(testQueueConfig())
	if opt.Addr != "localhost:6379" {
		t.Errorf("Addr = %q, want localhost:6379", opt.Addr)
	}
	if opt.Password != "secret" {
		t.Errorf("Password = %q, want secret", opt.Password)
	}
	if opt.DB != 2 {
		t.Errorf("DB = %d, want 2", opt.DB)
	}
}

func TestNewQueueClient(t *testing.T) {
	client := NewQueueClient(testQueueConfig())
	if client == nil {
		t.Fatal("NewQueueClient() returned nil")
	}
	_ = client.Close()
}

func TestNewQueueServer(t *testing.T) {
	server := NewQueueServer(testQueueConfig())
	if server == nil {
		t.Fatal("NewQueueServer() returned nil")
	}
}

func TestNewHealthChecker(t *testing.T) {
	checker := NewHealthChecker(testQueueConfig())
	if checker == nil {
		t.Fatal("NewHealthChecker() returned nil")
	}
}

func TestWorkerRunning_PropagatesRedisError(t *testing.T) {
	cfg := testQueueConfig()
	cfg.QueueRedisPort = 1
	checker := NewHealthChecker(cfg)

	if err := checker.WorkerRunning(); err == nil {
		t.Error("WorkerRunning() expected error when redis unreachable")
	}
}

func TestAsynqLogger(t *testing.T) {
	logger := NewAsynqLogger()
	if logger == nil {
		t.Fatal("NewAsynqLogger() returned nil")
	}
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
}

func TestTaskEnqueuerMock(t *testing.T) {
	mock := TaskEnqueuerMock{}
	info, err := mock.Enqueue(nil)
	if err != nil || info != nil {
		t.Errorf("Enqueue() = (%v, %v), want (nil, nil)", info, err)
	}

	mock = TaskEnqueuerMock{EnqueueFunc: func(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
		return &asynq.TaskInfo{ID: "1"}, nil
	}}
	info, err = mock.Enqueue(nil)
	if err != nil || info == nil || info.ID != "1" {
		t.Errorf("Enqueue() = (%v, %v), want id 1", info, err)
	}
}
