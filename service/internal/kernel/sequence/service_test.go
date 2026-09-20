package sequence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
)

func TestService_Next(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		reserveFn func(ctx context.Context, organizationID uint64, code string, now time.Time) (*Reservation, error)
		want      string
		wantErr   bool
		wantErrV  error
	}{
		{
			name: "returns reserved document number",
			reserveFn: func(_ context.Context, _ uint64, code string, now time.Time) (*Reservation, error) {
				return &Reservation{Value: 42, Number: "SO/00042"}, nil
			},
			want: "SO/00042",
		},
		{
			name: "unused sequence code returns not found",
			reserveFn: func(_ context.Context, _ uint64, _ string, _ time.Time) (*Reservation, error) {
				return nil, ErrSequenceNotFound
			},
			wantErr:  true,
			wantErrV: ErrSequenceNotFound,
		},
		{
			name: "dao error propagates",
			reserveFn: func(_ context.Context, _ uint64, _ string, _ time.Time) (*Reservation, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSequenceService(DAOMock{ReserveFunc: tt.reserveFn})
			got, err := svc.Next(ctx, 1, "SALE_ORDER")
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrV) {
				return
			}
			if got != tt.want {
				t.Errorf("Next() = %q, want %q", got, tt.want)
			}
		})
	}
}
