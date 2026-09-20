package httpx

import (
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gofiber/contrib/v3/monitor"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jalusw/swantara/apps/service/internal/config"
)

func NewApp(cfg *config.Config, limiterStorage fiber.Storage) *fiber.App {
	server := fiber.New(newFiberConfig(cfg))
	server.Use(recover.New(recover.Config{
		EnableStackTrace:  cfg.ApplicationDebug,
		StackTraceHandler: panicStackTraceHandler,
	}))
	server.Use(requestid.New())
	server.Use(RequestLogMiddleware())
	server.Use(securityHeaders())
	server.Use(cors.New(cors.Config{
		AllowOrigins: clientOrigins(cfg.ClientWebURL),
	}))
	server.Use(etag.New())
	server.Use(compress.New())

	if cfg.ApplicationDebug {
		server.Use(pprof.New())
	}

	if cfg.RateLimiterEnabled && !cfg.ApplicationDebug {
		server.Use(limiter.New(limiter.Config{
			Max:          cfg.RateLimiterMax,
			Expiration:   cfg.RateLimiterDuration,
			Storage:      limiterStorage,
			LimitReached: RateLimitReached,
		}))
	}

	if cfg.ApplicationDebug {
		server.Get("/metrics", monitor.New())
	}

	return server
}

func clientOrigins(clientWebURL string) []string {
	origins := []string{"http://localhost:3000"}
	if clientWebURL != "" {
		origins = append(origins, clientWebURL)
	}
	return origins
}

func securityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		return c.Next()
	}
}

func AuthRateLimitGuard() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          10,
		Expiration:   1 * time.Minute,
		LimitReached: RateLimitReached,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
	})
}

func panicStackTraceHandler(c fiber.Ctx, r any) {
	slog.Error("panic", "method", c.Method(), "path", c.Path(), "request_id", requestid.FromContext(c), "panic", r, "stack", string(debug.Stack()))
}

func newFiberConfig(cfg *config.Config) fiber.Config {
	fiberConfig := fiber.Config{
		AppName:       cfg.ApplicationName,
		CaseSensitive: true,
		BodyLimit:     int(cfg.StorageMaxUploadSize),
	}

	if proxies := cfg.TrustedProxyList(); len(proxies) > 0 {
		fiberConfig.TrustProxy = true
		fiberConfig.TrustProxyConfig = fiber.DefaultTrustProxyConfig
		fiberConfig.TrustProxyConfig.Proxies = proxies
		fiberConfig.ProxyHeader = fiber.HeaderXForwardedFor
	}

	return fiberConfig
}
