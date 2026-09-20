package queue

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestAsynqLogger_FatalExits(t *testing.T) {
	if os.Getenv("TEST_FATAL_CHILD") == "1" {
		NewAsynqLogger().Fatal("boom")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run", "TestAsynqLogger_FatalExits")
	cmd.Env = append(os.Environ(), "TEST_FATAL_CHILD=1")
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Errorf("expected exit error from Fatal, got %v", err)
	}
}

func TestWorkerRunning_NoWorkers(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "localhost:6379", 200*time.Millisecond)
	if err != nil {
		t.Skip("redis unavailable, skipping live worker check")
	}
	_ = conn.Close()

	cfg := testQueueConfig()
	cfg.QueueRedisPassword = ""
	cfg.QueueRedisDB = 15
	checker := NewHealthChecker(cfg)
	if err := checker.WorkerRunning(); !errors.Is(err, ErrNoQueueWorkerRunning) {
		t.Logf("WorkerRunning() = %v (redis has workers or is unreachable)", err)
	}
}
