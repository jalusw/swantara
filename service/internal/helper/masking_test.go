package helper

import "testing"

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"empty", "", ""},
		{"single char", "a", "*"},
		{"no at", "not-an-email", "n**********l"},
		{"no at single", "x", "*"},
		{"typical", "johndoe@example.com", "j***@***.com"},
		{"short local", "j@example.com", "j@***.com"},
		{"single label domain", "johndoe@example", "j***@***"},
		{"multi subdomain", "johndoe@mail.example.co.uk", "j***@***.***.***.uk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskEmail(tt.value); got != tt.want {
				t.Errorf("MaskEmail(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"empty", "", ""},
		{"one rune", "a", "*"},
		{"two runes", "ab", "**"},
		{"three runes", "abc", "a*c"},
		{"typical", "John", "J**n"},
		{"space inside", "a b c", "a***c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskValue(tt.value); got != tt.want {
				t.Errorf("MaskValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
