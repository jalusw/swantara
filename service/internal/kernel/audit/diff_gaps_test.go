package audit

import (
	"reflect"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type diffNested struct {
	Inner diffInner
	Note  string
}

type diffInner struct {
	Level int
}

type diffTagged struct {
	Hidden  string `json:"-"`
	Code    string `json:"code"`
	NoTag   string
	private string
}

func TestDiff_EdgeCases(t *testing.T) {
	if got := Diff(5, 6); got["value"] != 6 {
		t.Errorf("non-struct diff = %v", got)
	}
	if got := Diff((*saleOrder)(nil), (*saleOrder)(nil)); len(got) != 0 {
		t.Errorf("nil ptr diff = %v", got)
	}
	if got := Diff(nil, 7); got["value"] != 7 {
		t.Errorf("nil before scalar = %v", got)
	}
	if got := Diff("x", nil); got["value"] != "x" {
		t.Errorf("nil after scalar = %v", got)
	}
	if got := Diff("x", 7); got["value"] != 7 {
		t.Errorf("mismatched types = %v", got)
	}

	tagged := Diff(diffTagged{}, diffTagged{Hidden: "a", Code: "b", NoTag: "c", private: "d"})
	if _, ok := tagged["Hidden"]; ok {
		t.Errorf("json:- field leaked: %v", tagged)
	}
	if tagged["code"] != "b" {
		t.Errorf("tagged code = %v", tagged)
	}
	if tagged["NoTag"] != "c" {
		t.Errorf("untagged field = %v", tagged)
	}

	nested := Diff(diffNested{}, diffNested{Inner: diffInner{Level: 2}, Note: "hi"})
	inner, ok := nested["Inner"].(map[string]any)
	if !ok || inner["Level"] != 2 {
		t.Errorf("nested diff = %v", nested)
	}
	if nested["Note"] != "hi" {
		t.Errorf("note diff = %v", nested)
	}

	_ = reflect.TypeOf(model.Base{})
}
