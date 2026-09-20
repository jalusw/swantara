package helper

import (
	"database/sql/driver"
	"testing"
)

func TestInt64ArrayValue(t *testing.T) {
	tests := []struct {
		name  string
		value Int64Array
		want  driver.Value
	}{
		{name: "empty", value: Int64Array{}, want: "{}"},
		{name: "single", value: Int64Array{1}, want: "{1}"},
		{name: "multiple", value: Int64Array{1, 2, 3}, want: "{1,2,3}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.Value()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestInt64ArrayScan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    Int64Array
		wantErr bool
	}{
		{name: "nil", value: nil},
		{name: "empty", value: "{}", want: Int64Array{}},
		{name: "single", value: "{1}", want: Int64Array{1}},
		{name: "multiple", value: "{1,2,3}", want: Int64Array{1, 2, 3}},
		{name: "wrong type", value: int64(42), wantErr: true},
		{name: "invalid number", value: "{abc}", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Int64Array
			err := got.Scan(tt.value)
			if AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("expected %v, got %v", tt.want, got)
				}
			}
		})
	}
}
