package helper

import (
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func TestParseBirthday(t *testing.T) {
	tests := []struct {
		name    string
		value   *string
		want    *time.Time
		wantErr bool
	}{
		{
			name:  "nil",
			value: nil,
		},
		{
			name:  "empty string",
			value: ptr(""),
		},
		{
			name:  "valid date",
			value: ptr("2000-01-31"),
			want:  ptr(time.Date(2000, 1, 31, 0, 0, 0, 0, time.UTC)),
		},
		{
			name:    "invalid date",
			value:   ptr("2000-13-45"),
			wantErr: true,
		},
		{
			name:    "not a date",
			value:   ptr("hello"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBirthday(tt.value)
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

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{"empty", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", "hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashToken(tt.token); got != tt.want {
				t.Errorf("HashToken(%q) = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}

func TestRandomHex(t *testing.T) {
	s, err := RandomHex(16)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 32 {
		t.Errorf("expected 32 hex chars, got %d", len(s))
	}
	if _, err := hex.DecodeString(s); err != nil {
		t.Errorf("expected valid hex, got %q", s)
	}
}

func TestRandomHex_Empty(t *testing.T) {
	s, err := RandomHex(0)
	if err != nil {
		t.Fatal(err)
	}
	if s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

func TestParseOS(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want string
	}{
		{"Windows 10", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Windows 10"},
		{"Windows 8.1", "Mozilla/5.0 (Windows NT 6.3)", "Windows 8.1"},
		{"Windows 8", "Mozilla/5.0 (Windows NT 6.2)", "Windows 8"},
		{"Windows 7", "Mozilla/5.0 (Windows NT 6.1)", "Windows 7"},
		{"Windows Vista", "Mozilla/5.0 (Windows NT 6.0)", "Windows Vista"},
		{"Windows XP", "Mozilla/5.0 (Windows NT 5.1)", "Windows XP"},
		{"Windows generic", "Mozilla/5.0 (Windows; U; en)", "Windows"},
		{"macOS", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", "macOS"},
		{"iOS", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0)", "iOS"},
		{"Android", "Mozilla/5.0 (Linux; Android 13; Pixel 7)", "Android"},
		{"Linux", "Mozilla/5.0 (X11; Linux x86_64)", "Linux"},
		{"Unknown", "curl/8.0.0", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseOS(tt.ua); got != tt.want {
				t.Errorf("ParseOS(%q) = %q, want %q", tt.ua, got, tt.want)
			}
		})
	}
}

func TestParseOS_UppercaseInput(t *testing.T) {
	got := ParseOS(strings.ToUpper("Mozilla/5.0 (Windows NT 10.0; Win64; x64)"))
	if got != "Windows 10" {
		t.Errorf("expected Windows 10 for uppercase input, got %q", got)
	}
}

func TestParseDeviceName(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want string
	}{
		{"iPhone", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0)", "iPhone"},
		{"iPad", "Mozilla/5.0 (iPad; CPU OS 16_0)", "iPad"},
		{"Mac", "Mozilla/5.0 (Macintosh; Intel Mac OS X)", "Mac"},
		{"Android", "Mozilla/5.0 (Linux; Android 13; Pixel)", "Android Device"},
		{"Windows Phone", "Mozilla/5.0 (Windows Phone 10.0; ARM; Touch)", "Windows Phone"},
		{"Windows PC", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Windows PC"},
		{"Linux", "Mozilla/5.0 (X11; Linux x86_64)", "Linux Desktop"},
		{"Unknown", "curl/8.0.0", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseDeviceName(tt.ua); got != tt.want {
				t.Errorf("ParseDeviceName(%q) = %q, want %q", tt.ua, got, tt.want)
			}
		})
	}
}

func TestParseBrowser(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want string
	}{
		{"Edge edg", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Edg/120.0.0.0", "Edge"},
		{"Edge generic", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Edge/120.0.0.0", "Edge"},
		{"Opera", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 OPR/105.0.0.0", "Opera"},
		{"Chrome", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36", "Chrome"},
		{"Firefox", "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0", "Firefox"},
		{"Safari", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Version/17.1 Safari/605.1.15", "Safari"},
		{"Unknown", "curl/8.0.0", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseBrowser(tt.ua); got != tt.want {
				t.Errorf("ParseBrowser(%q) = %q, want %q", tt.ua, got, tt.want)
			}
		})
	}
}
