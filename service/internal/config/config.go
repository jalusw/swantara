package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	ApplicationName     string `mapstructure:"APP_NAME" validate:"required"`
	ApplicationHost     string `mapstructure:"APP_HOST" validate:"required"`
	ApplicationPort     int    `mapstructure:"APP_PORT" validate:"required,min=1,max=65535"`
	ApplicationDebug    bool   `mapstructure:"APP_DEBUG"`
	ApplicationLanguage string `mapstructure:"APP_LANG" validate:"required"`
	ApplicationURL      string `mapstructure:"APP_URL" validate:"required,url"`
	ApplicationVersion  string `mapstructure:"APP_VERSION"`
	ApplicationPrefork  bool   `mapstructure:"APP_PREFORK"`

	LogLevel      string `mapstructure:"LOG_LEVEL" validate:"oneof=debug info warn error"`
	LogFile       string `mapstructure:"LOG_FILE"`
	LogFormat     string `mapstructure:"LOG_FORMAT" validate:"omitempty,oneof=text json pretty pretty-json"`
	LogMaxSize    int    `mapstructure:"LOG_MAX_SIZE" validate:"min=1"`
	LogMaxAge     int    `mapstructure:"LOG_MAX_AGE" validate:"min=1"`
	LogMaxBackups int    `mapstructure:"LOG_MAX_BACKUPS" validate:"min=0"`
	LogCompress   bool   `mapstructure:"LOG_COMPRESS"`

	DatabaseName                string        `mapstructure:"DB_NAME" validate:"required"`
	DatabaseHost                string        `mapstructure:"DB_HOST" validate:"required"`
	DatabasePort                int           `mapstructure:"DB_PORT" validate:"required,min=1,max=65535"`
	DatabaseUser                string        `mapstructure:"DB_USER" validate:"required"`
	DatabasePassword            string        `mapstructure:"DB_PASSWORD"`
	DatabaseSSLMode             string        `mapstructure:"DB_SSL_MODE" validate:"oneof=disable require verify-ca verify-full prefer"`
	DatabaseMaxOpenConns        int           `mapstructure:"DB_MAX_OPEN_CONNS" validate:"min=1"`
	DatabaseMaxIdleConns        int           `mapstructure:"DB_MAX_IDLE_CONNS" validate:"min=0"`
	DatabaseConnMaxLife         time.Duration `mapstructure:"DB_CONN_MAX_LIFE" validate:"min=1s"`
	DatabaseConnMaxIdle         time.Duration `mapstructure:"DB_CONN_MAX_IDLE" validate:"min=1s"`
	DatabaseSlowThreshold       time.Duration `mapstructure:"DB_SLOW_THRESHOLD" validate:"omitempty,min=1ms"`
	DatabaseRequireNonSuperuser bool          `mapstructure:"DB_REQUIRE_NON_SUPERUSER"`

	BcryptCost int    `mapstructure:"BCRYPT_PASSWORD_COST" validate:"min=4,max=31"`
	Pepper     string `mapstructure:"PASSWORD_PEPPER"`

	JWTAccessTokenSecret           string `mapstructure:"JWT_ACCESS_TOKEN_SECRET" validate:"required"`
	JWTAccessTokenExpires          string `mapstructure:"JWT_ACCESS_TOKEN_EXPIRES_IN" validate:"required"`
	JWTAccessTokenExpiresDuration  time.Duration
	JWTAccessTokenSigningMethod    jwt.SigningMethod
	JWTRefreshTokenSecret          string `mapstructure:"JWT_REFRESH_TOKEN_SECRET" validate:"required"`
	JWTRefreshTokenExpires         string `mapstructure:"JWT_REFRESH_TOKEN_EXPIRES_IN" validate:"required"`
	JWTRefreshTokenExpiresDuration time.Duration
	JWTRefreshTokenSigningMethod   jwt.SigningMethod

	QueueRedisHost     string `mapstructure:"QUEUE_REDIS_HOST" validate:"required"`
	QueueRedisPort     int    `mapstructure:"QUEUE_REDIS_PORT" validate:"required"`
	QueueRedisPassword string `mapstructure:"QUEUE_REDIS_PASSWORD"`
	QueueRedisDB       int    `mapstructure:"QUEUE_REDIS_DB" validate:"min=0"`

	SMTPHost       string `mapstructure:"SMTP_HOST" validate:"required"`
	SMTPPort       int    `mapstructure:"SMTP_PORT" validate:"required,min=1,max=65535"`
	SMTPUsername   string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword   string `mapstructure:"SMTP_PASSWORD"`
	SMTPEncryption string `mapstructure:"SMTP_ENCRYPTION" validate:"oneof=none tls ssl"`
	SMTPFromEmail  string `mapstructure:"SMTP_FROM_EMAIL" validate:"omitempty,email"`
	SMTPFromName   string `mapstructure:"SMTP_FROM_NAME"`

	ClientWebURL string `mapstructure:"CLIENT_WEB_URL" validate:"required,url"`

	TrustedProxies string `mapstructure:"APP_TRUSTED_PROXIES"`

	MailUserVerificationRedirection string `mapstructure:"MAIL_USER_VERIFICATION_REDIRECTION" validate:"required,url"`
	MailPasswordResetRedirection    string `mapstructure:"MAIL_PASSWORD_RESET_REDIRECTION" validate:"required,url"`

	AuthEmailVerifyTTL   string `mapstructure:"AUTH_EMAIL_VERIFY_TTL"`
	AuthPasswordResetTTL string `mapstructure:"AUTH_PASSWORD_RESET_TTL"`

	AuthEmailVerifyTTLDuration   time.Duration
	AuthPasswordResetTTLDuration time.Duration

	StoragePath          string `mapstructure:"STORAGE_PATH" validate:"required"`
	StorageMaxUploadSize int64  `mapstructure:"STORAGE_MAX_UPLOAD_SIZE" validate:"required,min=1"`

	RateLimiterEnabled    bool   `mapstructure:"RATE_LIMITER_ENABLED"`
	RateLimiterMax        int    `mapstructure:"RATE_LIMITER_MAX" validate:"min=1"`
	RateLimiterExpiration string `mapstructure:"RATE_LIMITER_EXPIRATION" validate:"required"`
	RateLimiterDuration   time.Duration

	RateLimiterStorage       string `mapstructure:"RATE_LIMITER_STORAGE" validate:"oneof=memory redis"`
	RateLimiterRedisHost     string `mapstructure:"RATE_LIMITER_REDIS_HOST"`
	RateLimiterRedisPort     int    `mapstructure:"RATE_LIMITER_REDIS_PORT" validate:"min=0,max=65535"`
	RateLimiterRedisPassword string `mapstructure:"RATE_LIMITER_REDIS_PASSWORD"`
	RateLimiterRedisDB       int    `mapstructure:"RATE_LIMITER_REDIS_DB" validate:"min=0"`
}

func ConfigFilePath() string {
	if f := os.Getenv("CONFIG_FILE"); f != "" {
		return f
	}
	return ".env"
}

func New(path, filename string) (*Config, error) {
	var config Config

	if path == "" {
		path = "."
	}

	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := bindEnvKeys(&config); err != nil {
		return nil, err
	}

	if filename != "" {
		viper.SetConfigName(filename)
		viper.AddConfigPath(path)
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	viper.SetDefault("APP_NAME", "Swantara API Service")
	viper.SetDefault("APP_HOST", "localhost")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_DEBUG", false)
	viper.SetDefault("APP_LANG", "en")
	viper.SetDefault("APP_URL", "http://localhost:8080")
	viper.SetDefault("APP_VERSION", "dev")
	viper.SetDefault("APP_PREFORK", false)

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FILE", "")
	viper.SetDefault("LOG_FORMAT", "")
	viper.SetDefault("LOG_MAX_SIZE", 100)
	viper.SetDefault("LOG_MAX_AGE", 30)
	viper.SetDefault("LOG_MAX_BACKUPS", 3)
	viper.SetDefault("LOG_COMPRESS", false)

	viper.SetDefault("DB_NAME", "swantara")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 50)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_CONN_MAX_LIFE", time.Hour)
	viper.SetDefault("DB_CONN_MAX_IDLE", 30*time.Minute)
	viper.SetDefault("DB_SLOW_THRESHOLD", "200ms")
	viper.SetDefault("DB_REQUIRE_NON_SUPERUSER", false)

	viper.SetDefault("BCRYPT_PASSWORD_COST", bcrypt.DefaultCost)

	viper.SetDefault("JWT_ACCESS_TOKEN_EXPIRES_IN", "30m")
	viper.SetDefault("JWT_REFRESH_TOKEN_EXPIRES_IN", "48h")

	viper.SetDefault("QUEUE_REDIS_HOST", "localhost")
	viper.SetDefault("QUEUE_REDIS_PORT", 6379)
	viper.SetDefault("QUEUE_REDIS_DB", 0)

	viper.SetDefault("SMTP_HOST", "localhost")
	viper.SetDefault("SMTP_PORT", 1025)
	viper.SetDefault("SMTP_USERNAME", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_ENCRYPTION", "none")
	viper.SetDefault("SMTP_FROM_EMAIL", "noreply@swantara.com")
	viper.SetDefault("SMTP_FROM_NAME", "Swantara")

	viper.SetDefault("CLIENT_WEB_URL", "http://localhost:3000")

	viper.SetDefault("MAIL_USER_VERIFICATION_REDIRECTION", "http://localhost:3000/email-verify")
	viper.SetDefault("MAIL_PASSWORD_RESET_REDIRECTION", "http://localhost:3000/password-reset")

	viper.SetDefault("AUTH_EMAIL_VERIFY_TTL", "24h")
	viper.SetDefault("AUTH_PASSWORD_RESET_TTL", "30m")

	viper.SetDefault("RATE_LIMITER_ENABLED", true)
	viper.SetDefault("RATE_LIMITER_MAX", 60)
	viper.SetDefault("RATE_LIMITER_EXPIRATION", "1m")

	viper.SetDefault("RATE_LIMITER_STORAGE", "memory")
	viper.SetDefault("RATE_LIMITER_REDIS_HOST", "localhost")
	viper.SetDefault("RATE_LIMITER_REDIS_PORT", 6379)
	viper.SetDefault("RATE_LIMITER_REDIS_DB", 0)

	viper.SetDefault("STORAGE_PATH", "./storage")
	viper.SetDefault("STORAGE_MAX_UPLOAD_SIZE", 10*1024*1024)

	if err := viper.Unmarshal(&config); err != nil {
		log.Println("Failed to parse the configuration", err)
		return nil, err
	}

	config.JWTAccessTokenSigningMethod = jwt.SigningMethodHS256
	config.JWTRefreshTokenSigningMethod = jwt.SigningMethodHS256

	if err := config.Validate(); err != nil {
		return nil, err
	}

	duration, err := helper.ParseDuration("RATE_LIMITER_EXPIRATION", config.RateLimiterExpiration)
	if err != nil {
		return nil, err
	}
	config.RateLimiterDuration = duration

	emailVerifyTTL, err := helper.ParseDuration("AUTH_EMAIL_VERIFY_TTL", config.AuthEmailVerifyTTL)
	if err != nil {
		return nil, err
	}
	config.AuthEmailVerifyTTLDuration = emailVerifyTTL

	passwordResetTTL, err := helper.ParseDuration("AUTH_PASSWORD_RESET_TTL", config.AuthPasswordResetTTL)
	if err != nil {
		return nil, err
	}
	config.AuthPasswordResetTTLDuration = passwordResetTTL

	accessTokenExpires, err := helper.ParseDuration("JWT_ACCESS_TOKEN_EXPIRES_IN", config.JWTAccessTokenExpires)
	if err != nil {
		return nil, err
	}
	config.JWTAccessTokenExpiresDuration = accessTokenExpires

	refreshTokenExpires, err := helper.ParseDuration("JWT_REFRESH_TOKEN_EXPIRES_IN", config.JWTRefreshTokenExpires)
	if err != nil {
		return nil, err
	}
	config.JWTRefreshTokenExpiresDuration = refreshTokenExpires

	return &config, nil
}

func (c *Config) HTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.ApplicationHost, c.ApplicationPort)
}

func (c *Config) QueueRedisAddress() string {
	return fmt.Sprintf("%s:%d", c.QueueRedisHost, c.QueueRedisPort)
}

func (c *Config) RateLimiterRedisAddress() string {
	return fmt.Sprintf("%s:%d", c.RateLimiterRedisHost, c.RateLimiterRedisPort)
}

func (c *Config) TrustedProxyList() []string {
	if c.TrustedProxies == "" {
		return nil
	}

	parts := strings.Split(c.TrustedProxies, ",")
	proxies := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			proxies = append(proxies, trimmed)
		}
	}
	return proxies
}

func (c *Config) Validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("mapstructure"), ",")
		if name == "-" {
			return ""
		}
		return name
	})

	if err := v.Struct(c); err != nil {
		return err
	}

	for _, d := range []struct {
		name  string
		value string
	}{
		{name: "RATE_LIMITER_EXPIRATION", value: c.RateLimiterExpiration},
		{name: "AUTH_EMAIL_VERIFY_TTL", value: c.AuthEmailVerifyTTL},
		{name: "AUTH_PASSWORD_RESET_TTL", value: c.AuthPasswordResetTTL},
	} {
		if _, err := helper.ParseDuration(d.name, d.value); err != nil {
			return err
		}
	}

	return nil
}

func bindEnvKeys(config *Config) error {
	value := reflect.ValueOf(config).Elem()
	typ := value.Type()
	for i := 0; i < typ.NumField(); i++ {
		tag, _, _ := strings.Cut(typ.Field(i).Tag.Get("mapstructure"), ",")
		if tag != "" {
			if err := viper.BindEnv(tag); err != nil {
				return err
			}
		}
	}
	return nil
}
