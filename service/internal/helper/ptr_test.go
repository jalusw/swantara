package helper

import "testing"

func TestPtr(t *testing.T) {
	want := 42
	got := Ptr(want)
	if got == nil || *got != want {
		t.Errorf("expected pointer to %d, got %v", want, got)
	}
}

func TestDeref(t *testing.T) {
	value := 42
	if got := Deref(&value, 0); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
	if got := Deref[int](nil, 7); got != 7 {
		t.Errorf("expected 7, got %d", got)
	}
}

func TestStringPtr(t *testing.T) {
	if got := StringPtr(""); got != nil {
		t.Errorf("expected nil for empty string, got %v", got)
	}
	got := StringPtr("value")
	if got == nil || *got != "value" {
		t.Errorf("expected pointer to value, got %v", got)
	}
}
