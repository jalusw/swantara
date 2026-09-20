package helper

import (
	"testing"
	"time"
)

func TestParseDateStr(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    time.Time
		wantErr bool
	}{
		{name: "valid", raw: "2000-01-31", want: time.Date(2000, 1, 31, 0, 0, 0, 0, time.UTC)},
		{name: "invalid", raw: "hello", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateStr(tt.raw)
			if AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "minutes", value: "30m", want: 30 * time.Minute},
		{name: "hours", value: "48h", want: 48 * time.Hour},
		{name: "seconds", value: "30s", want: 30 * time.Second},
		{name: "invalid", value: "not-a-duration", wantErr: true},
		{name: "empty", value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration("TEST_DURATION", tt.value)
			if AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
