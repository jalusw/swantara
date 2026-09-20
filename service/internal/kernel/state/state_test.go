package state

import (
	"reflect"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestMachine_TryTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    model.Status
		to      model.Status
		wantErr error
	}{
		{name: "allows defined transition", from: "draft", to: "confirmed"},
		{name: "allows cancellation from draft", from: "draft", to: "cancelled"},
		{name: "allows cancellation from confirmed", from: "confirmed", to: "cancelled"},
		{name: "rejects illegal transition", from: "draft", to: "done", wantErr: ErrIllegalTransition},
		{name: "rejects unknown from state", from: "posted", to: "done", wantErr: ErrIllegalTransition},
		{name: "rejects empty state", from: "", to: "confirmed", wantErr: ErrIllegalTransition},
		{name: "rejects transition from terminal state", from: "done", to: "draft", wantErr: ErrIllegalTransition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine := newOrderMachine()

			err := machine.TryTransition(tt.from, tt.to)

			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
		})
	}
}

func TestMachine_IsTerminal(t *testing.T) {
	machine := newOrderMachine()

	if !machine.IsTerminal("done") {
		t.Error("expected done to be terminal")
	}
	if machine.IsTerminal("draft") {
		t.Error("expected draft not to be terminal")
	}
	if !machine.IsTerminal("cancelled") {
		t.Error("expected cancelled to be terminal")
	}
}

func TestMachine_ListsOnlyAllowedTargets(t *testing.T) {
	machine := newOrderMachine()

	got := machine.AllowedTransitions("draft")

	want := []model.Status{"cancelled", "confirmed"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}
