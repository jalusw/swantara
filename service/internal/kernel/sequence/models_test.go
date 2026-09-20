package sequence

import (
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
)

func TestDocumentSequence_Format(t *testing.T) {
	tests := []struct {
		name string
		seq  DocumentSequence
		val  int64
		want string
	}{
		{name: "padded sales order", seq: DocumentSequence{Prefix: "SO/", Padding: 5}, val: 42, want: "SO/00042"},
		{name: "year prefixed invoice", seq: DocumentSequence{Prefix: "INV/2026/", Padding: 4}, val: 11, want: "INV/2026/0011"},
		{name: "suffixed", seq: DocumentSequence{Prefix: "PO/", Suffix: "/26", Padding: 5}, val: 301, want: "PO/00301/26"},
		{name: "no padding", seq: DocumentSequence{Prefix: "REF/", Padding: 0}, val: 7, want: "REF/7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.seq.Format(tt.val); got != tt.want {
				t.Errorf("Format(%d) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

func TestShouldReset(t *testing.T) {
	january := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	february := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	nextYear := time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		period   string
		lastUsed time.Time
		now      time.Time
		want     bool
	}{
		{name: "never does not reset", period: "never", lastUsed: january, now: nextYear, want: false},
		{name: "empty period does not reset", period: "", lastUsed: january, now: nextYear, want: false},
		{name: "yearly same year", period: "yearly", lastUsed: january, now: february, want: false},
		{name: "yearly different year", period: "yearly", lastUsed: january, now: nextYear, want: true},
		{name: "monthly same month", period: "monthly", lastUsed: january, now: time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC), want: false},
		{name: "monthly different month", period: "monthly", lastUsed: january, now: february, want: true},
		{name: "monthly different year", period: "monthly", lastUsed: january, now: nextYear, want: true},
		{name: "unknown period does not reset", period: "custom", lastUsed: january, now: nextYear, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldReset(tt.period, tt.lastUsed, tt.now); got != tt.want {
				t.Errorf("shouldReset(%q, %v, %v) = %v, want %v", tt.period, tt.lastUsed, tt.now, got, tt.want)
			}
		})
	}
}

func TestParseResetPeriod(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		want     ResetPeriod
		wantErr  bool
		wantErrV error
	}{
		{name: "never", value: "never", want: ResetNever},
		{name: "yearly", value: "yearly", want: ResetYearly},
		{name: "monthly", value: "monthly", want: ResetMonthly},
		{name: "empty defaults to never", value: "", want: ResetNever},
		{name: "whitespace trimmed", value: " yearly ", want: ResetYearly},
		{name: "invalid rejected", value: "daily", wantErr: true, wantErrV: ErrInvalidResetPeriod},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseResetPeriod(tt.value)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrV) {
				return
			}
			if got != tt.want {
				t.Errorf("ParseResetPeriod(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
