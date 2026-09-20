package helper

import (
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		value   *string
		want    *time.Time
		wantErr bool
	}{
		{name: "nil"},
		{name: "empty string", value: ptr("")},
		{name: "valid timestamp", value: ptr("2024-01-31T10:00:00Z"), want: ptr(time.Date(2024, 1, 31, 10, 0, 0, 0, time.UTC))},
		{name: "invalid timestamp", value: ptr("not-a-timestamp"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimestamp(tt.value)
			if AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if got == nil || !got.Equal(*tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestFormatDatePtr(t *testing.T) {
	tests := []struct {
		name  string
		value *time.Time
		want  *string
	}{
		{name: "nil"},
		{name: "zero value", value: ptr(time.Time{})},
		{name: "valid date", value: ptr(time.Date(2000, 1, 31, 0, 0, 0, 0, time.UTC)), want: ptr("2000-01-31")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDatePtr(tt.value)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if got == nil || *got != *tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
