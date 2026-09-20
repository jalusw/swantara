package logger

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestFatalExits(t *testing.T) {
	if os.Getenv("TEST_LOGGER_FATAL") == "1" {
		Fatal("boom", errors.New("bang"))
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run", "TestFatalExits")
	cmd.Env = append(os.Environ(), "TEST_LOGGER_FATAL=1")
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Errorf("expected exit error, got %v", err)
	}
}

func TestFatalfExits(t *testing.T) {
	if os.Getenv("TEST_LOGGER_FATALF") == "1" {
		Fatalf("boom %d", 42)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run", "TestFatalfExits")
	cmd.Env = append(os.Environ(), "TEST_LOGGER_FATALF=1")
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Errorf("expected exit error, got %v", err)
	}
}

func TestTerminalColorNonFile(t *testing.T) {
	var buf bytes.Buffer
	if terminalColor(&buf) {
		t.Error("expected false for buffer writer")
	}
}

func TestColorizeAllKinds(t *testing.T) {
	src := []byte(`{"s":"x","e":"a\"b","n":1,"neg":-1.5e3,"b":true,"f":false,"z":null,"o":{"k":"v"},"a":[1]}`)
	out := colorizeJSON(src, true)
	if len(out) == 0 {
		t.Error("expected output")
	}
}
