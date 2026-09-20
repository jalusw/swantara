package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

func isolateConfigEnv(t *testing.T) {
	t.Helper()
	val := reflect.ValueOf(validBaseConfig()).Elem()
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		tag, _, _ := strings.Cut(typ.Field(i).Tag.Get("mapstructure"), ",")
		if tag == "" || tag == "-" {
			continue
		}
		if prev, ok := os.LookupEnv(tag); ok {
			_ = os.Unsetenv(tag)
			t.Cleanup(func() { _ = os.Setenv(tag, prev) })
		}
	}
}

func validBaseConfig() *Config {
	return &Config{
		ApplicationName: "t", ApplicationHost: "localhost", ApplicationPort: 8080,
		ApplicationLanguage: "en", ApplicationURL: "http://localhost",
		LogLevel:   "info",
		LogMaxSize: 100, LogMaxAge: 30, LogMaxBackups: 3,
		DatabaseName: "d", DatabaseHost: "h", DatabasePort: 5432, DatabaseUser: "u",
		DatabaseSSLMode:              "disable",
		DatabaseMaxOpenConns:         50,
		DatabaseMaxIdleConns:         10,
		DatabaseConnMaxLife:          time.Hour,
		DatabaseConnMaxIdle:          30 * time.Minute,
		BcryptCost:                   12,
		JWTAccessTokenExpires:        "30m",
		JWTRefreshTokenExpires:       "48h",
		JWTAccessTokenSigningMethod:  jwt.SigningMethodHS256,
		JWTRefreshTokenSigningMethod: jwt.SigningMethodHS256,
		QueueRedisHost:               "h", QueueRedisPort: 6379,
		SMTPHost: "h", SMTPPort: 25, SMTPEncryption: "none",
		ClientWebURL:                    "http://localhost:3000",
		MailUserVerificationRedirection: "http://localhost:3000/email-verify",
		MailPasswordResetRedirection:    "http://localhost:3000/password-reset",
		AuthEmailVerifyTTL:              "24h",
		AuthPasswordResetTTL:            "30m",
		RateLimiterMax:                  60, RateLimiterStorage: "memory", RateLimiterExpiration: "1m",
		StoragePath: "s", StorageMaxUploadSize: 1,
	}
}

func TestConfig_Validate_RejectsMissingSecrets(t *testing.T) {
	c := validBaseConfig()
	if err := c.Validate(); err == nil {
		t.Fatal("expected validation to reject empty JWT secrets (no hardcoded fallback allowed)")
	}
}

func TestConfig_Validate_AcceptsExplicitSecrets(t *testing.T) {
	c := validBaseConfig()
	c.JWTAccessTokenSecret = "access-secret-from-env"
	c.JWTRefreshTokenSecret = "refresh-secret-from-env"
	if err := c.Validate(); err != nil {
		t.Fatalf("expected validation to pass with explicit secrets, got %v", err)
	}
}

func TestConfig_Validate_RejectsInvalidDuration(t *testing.T) {
	c := validBaseConfig()
	c.JWTAccessTokenSecret = "access-secret-from-env"
	c.JWTRefreshTokenSecret = "refresh-secret-from-env"
	c.RateLimiterExpiration = "not-a-duration"
	if err := c.Validate(); err == nil {
		t.Fatal("expected validation to reject invalid rate limiter duration")
	}
}

func TestConfig_Validate_RejectsBadPort(t *testing.T) {
	c := validBaseConfig()
	c.JWTAccessTokenSecret = "access-secret-from-env"
	c.JWTRefreshTokenSecret = "refresh-secret-from-env"
	c.ApplicationPort = 0
	if err := c.Validate(); err == nil {
		t.Fatal("expected validation to reject port 0")
	}
}

func TestConfig_Validate_RejectsBadURL(t *testing.T) {
	c := validBaseConfig()
	c.JWTAccessTokenSecret = "access-secret-from-env"
	c.JWTRefreshTokenSecret = "refresh-secret-from-env"
	c.ApplicationURL = "not-a-url"
	if err := c.Validate(); err == nil {
		t.Fatal("expected validation to reject invalid url")
	}
}

func TestConfig_HTTPAddress(t *testing.T) {
	c := validBaseConfig()
	c.ApplicationHost = "api.example.com"
	c.ApplicationPort = 8443
	if got := c.HTTPAddress(); got != "api.example.com:8443" {
		t.Errorf("HTTPAddress() = %q, want api.example.com:8443", got)
	}
}

func TestConfig_QueueRedisAddress(t *testing.T) {
	c := validBaseConfig()
	c.QueueRedisHost = "redis.internal"
	c.QueueRedisPort = 7000
	if got := c.QueueRedisAddress(); got != "redis.internal:7000" {
		t.Errorf("QueueRedisAddress() = %q, want redis.internal:7000", got)
	}
}

func TestConfig_RateLimiterRedisAddress(t *testing.T) {
	c := validBaseConfig()
	c.RateLimiterRedisHost = "ratelimit.internal"
	c.RateLimiterRedisPort = 6380
	if got := c.RateLimiterRedisAddress(); got != "ratelimit.internal:6380" {
		t.Errorf("RateLimiterRedisAddress() = %q, want ratelimit.internal:6380", got)
	}
}

func TestConfig_TrustedProxyList(t *testing.T) {
	c := validBaseConfig()
	if got := c.TrustedProxyList(); got != nil {
		t.Errorf("TrustedProxyList() = %v, want nil for empty", got)
	}

	c.TrustedProxies = "10.0.0.1, 10.0.0.2,10.0.0.3"
	got := c.TrustedProxyList()
	want := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	if len(got) != len(want) {
		t.Fatalf("TrustedProxyList() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("TrustedProxyList()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseDuration(t *testing.T) {
	d, err := helper.ParseDuration("RATE_LIMITER_EXPIRATION", "2m")
	if err != nil {
		t.Fatalf("ParseDuration() error = %v", err)
	}
	if d != 2*time.Minute {
		t.Errorf("duration = %v, want 2m", d)
	}

	if _, err := helper.ParseDuration("RATE_LIMITER_EXPIRATION", "oops"); err == nil {
		t.Fatal("ParseDuration() expected error for invalid value")
	}
}

func TestNew_ReadsEnvFile(t *testing.T) {
	isolateConfigEnv(t)
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := strings.Join([]string{
		`APP_NAME="Swantara Test"`,
		`APP_HOST="localhost"`,
		`APP_PORT=8081`,
		`APP_LANG="en"`,
		`APP_URL="http://localhost:8081"`,
		`LOG_LEVEL="debug"`,
		`DB_NAME="swantara"`,
		`DB_HOST="localhost"`,
		`DB_PORT=5432`,
		`DB_USER="postgres"`,
		`DB_SSL_MODE="disable"`,
		`JWT_ACCESS_TOKEN_SECRET="access"`,
		`JWT_ACCESS_TOKEN_EXPIRES_IN="15m"`,
		`JWT_REFRESH_TOKEN_SECRET="refresh"`,
		`JWT_REFRESH_TOKEN_EXPIRES_IN="24h"`,
		`QUEUE_REDIS_HOST="localhost"`,
		`QUEUE_REDIS_PORT=6379`,
		`SMTP_HOST="localhost"`,
		`SMTP_PORT=1025`,
		`SMTP_ENCRYPTION="none"`,
		`SMTP_FROM_EMAIL="noreply@swantara.com"`,
		`CLIENT_WEB_URL="http://localhost:3000"`,
		`MAIL_USER_VERIFICATION_REDIRECTION="http://localhost:3000/email-verify"`,
		`MAIL_PASSWORD_RESET_REDIRECTION="http://localhost:3000/password-reset"`,
		`STORAGE_PATH="./storage"`,
		`STORAGE_MAX_UPLOAD_SIZE=1048576`,
		`RATE_LIMITER_STORAGE="memory"`,
		`RATE_LIMITER_EXPIRATION="1m"`,
		`APP_TRUSTED_PROXIES="10.0.0.1, 10.0.0.2"`,
	}, "\n")
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file error = %v", err)
	}

	cfg, err := New(dir, ".env")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if cfg.ApplicationName != "Swantara Test" {
		t.Errorf("ApplicationName = %q, want Swantara Test", cfg.ApplicationName)
	}
	if cfg.ApplicationPort != 8081 {
		t.Errorf("ApplicationPort = %d, want 8081", cfg.ApplicationPort)
	}
	if cfg.JWTAccessTokenSigningMethod != jwt.SigningMethodHS256 {
		t.Error("expected HS256 signing method to be configured")
	}
	if cfg.RateLimiterDuration != time.Minute {
		t.Errorf("RateLimiterDuration = %v, want 1m", cfg.RateLimiterDuration)
	}
	if cfg.AuthEmailVerifyTTLDuration != 24*time.Hour {
		t.Errorf("AuthEmailVerifyTTLDuration = %v, want 24h", cfg.AuthEmailVerifyTTLDuration)
	}
	if cfg.JWTAccessTokenExpiresDuration != 15*time.Minute {
		t.Errorf("JWTAccessTokenExpiresDuration = %v, want 15m", cfg.JWTAccessTokenExpiresDuration)
	}
	if cfg.JWTRefreshTokenExpiresDuration != 24*time.Hour {
		t.Errorf("JWTRefreshTokenExpiresDuration = %v, want 24h", cfg.JWTRefreshTokenExpiresDuration)
	}
	if got := cfg.TrustedProxyList(); len(got) != 2 {
		t.Errorf("TrustedProxyList() = %v, want 2 proxies", got)
	}
}

func TestNew_MissingEnvFile(t *testing.T) {
	if _, err := New(t.TempDir(), ".env"); err == nil {
		t.Fatal("New() expected error when env file is missing")
	}
}

func TestNew_EnvOnlyWithoutFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `APP_PORT=9999`
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file error = %v", err)
	}
	t.Setenv("APP_NAME", "Swantara Prod")
	t.Setenv("APP_HOST", "0.0.0.0")
	t.Setenv("APP_PORT", "8080")
	t.Setenv("APP_LANG", "en")
	t.Setenv("APP_URL", "https://api.swantara.com")
	t.Setenv("DB_NAME", "swantara")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "swantara_app")
	t.Setenv("JWT_ACCESS_TOKEN_SECRET", "access")
	t.Setenv("JWT_ACCESS_TOKEN_EXPIRES_IN", "15m")
	t.Setenv("JWT_REFRESH_TOKEN_SECRET", "refresh")
	t.Setenv("JWT_REFRESH_TOKEN_EXPIRES_IN", "24h")
	t.Setenv("QUEUE_REDIS_HOST", "redis.internal")
	t.Setenv("QUEUE_REDIS_PORT", "6379")
	t.Setenv("SMTP_HOST", "smtp.internal")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_ENCRYPTION", "tls")
	t.Setenv("SMTP_FROM_EMAIL", "noreply@swantara.com")
	t.Setenv("CLIENT_WEB_URL", "https://swantara.com")
	t.Setenv("MAIL_USER_VERIFICATION_REDIRECTION", "https://swantara.com/email-verify")
	t.Setenv("MAIL_PASSWORD_RESET_REDIRECTION", "https://swantara.com/password-reset")
	t.Setenv("STORAGE_PATH", "/var/lib/swantara/storage")
	t.Setenv("STORAGE_MAX_UPLOAD_SIZE", "1048576")
	t.Setenv("RATE_LIMITER_STORAGE", "memory")
	t.Setenv("RATE_LIMITER_EXPIRATION", "1m")

	cfg, err := New(t.TempDir(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if cfg.ApplicationName != "Swantara Prod" {
		t.Errorf("ApplicationName = %q, want Swantara Prod", cfg.ApplicationName)
	}
	if cfg.DatabaseHost != "db.internal" {
		t.Errorf("DatabaseHost = %q, want db.internal", cfg.DatabaseHost)
	}
	if cfg.JWTAccessTokenSecret != "access" {
		t.Errorf("JWTAccessTokenSecret = %q, want access", cfg.JWTAccessTokenSecret)
	}
	if cfg.ApplicationPort != 8080 {
		t.Errorf("ApplicationPort = %d, want 8080 from defaults; a stray .env must not be loaded", cfg.ApplicationPort)
	}
}

func TestNew_InvalidDuration(t *testing.T) {
	isolateConfigEnv(t)
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := strings.Join([]string{
		`APP_NAME="Swantara Test"`,
		`APP_HOST="localhost"`,
		`APP_PORT=8081`,
		`APP_LANG="en"`,
		`APP_URL="http://localhost:8081"`,
		`DB_NAME="swantara"`,
		`DB_HOST="localhost"`,
		`DB_PORT=5432`,
		`DB_USER="postgres"`,
		`DB_SSL_MODE="disable"`,
		`JWT_ACCESS_TOKEN_SECRET="access"`,
		`JWT_ACCESS_TOKEN_EXPIRES_IN="15m"`,
		`JWT_REFRESH_TOKEN_SECRET="refresh"`,
		`JWT_REFRESH_TOKEN_EXPIRES_IN="24h"`,
		`QUEUE_REDIS_HOST="localhost"`,
		`QUEUE_REDIS_PORT=6379`,
		`SMTP_HOST="localhost"`,
		`SMTP_PORT=1025`,
		`SMTP_ENCRYPTION="none"`,
		`SMTP_FROM_EMAIL="noreply@swantara.com"`,
		`CLIENT_WEB_URL="http://localhost:3000"`,
		`MAIL_USER_VERIFICATION_REDIRECTION="http://localhost:3000/email-verify"`,
		`MAIL_PASSWORD_RESET_REDIRECTION="http://localhost:3000/password-reset"`,
		`STORAGE_PATH="./storage"`,
		`STORAGE_MAX_UPLOAD_SIZE=1048576`,
		`RATE_LIMITER_STORAGE="memory"`,
		`RATE_LIMITER_EXPIRATION="nonsense"`,
	}, "\n")
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file error = %v", err)
	}

	if _, err := New(dir, ".env"); err == nil {
		t.Fatal("New() expected error for invalid duration")
	}
}
