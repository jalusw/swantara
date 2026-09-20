package config

import (
	"os"
	"testing"
)

func TestConfigFilePath(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/tmp/custom.env")
	if got := ConfigFilePath(); got != "/tmp/custom.env" {
		t.Errorf("path = %q", got)
	}
	_ = os.Unsetenv("CONFIG_FILE")
	if got := ConfigFilePath(); got != ".env" {
		t.Errorf("path = %q, want .env", got)
	}
}
