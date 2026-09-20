package helper

import (
	"errors"
	"testing"
)

func AssertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected %d, got %d", want, got)
	}
}

func AssertError(t *testing.T, err error, wantErr bool, wantErrVal error) bool {
	t.Helper()

	if !wantErr {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return false
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if wantErrVal != nil && !errors.Is(err, wantErrVal) {
		t.Errorf("expected error %v, got %v", wantErrVal, err)
	}
	return true
}
