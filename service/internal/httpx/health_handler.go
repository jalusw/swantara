package httpx

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"gorm.io/gorm"
)

type QueueHealthChecker interface {
	WorkerRunning() error
}

type HealthHandler struct {
	cfg         *config.Config
	db          *gorm.DB
	queueHealth QueueHealthChecker
}

func NewHealthHandler(
	cfg *config.Config,
	db *gorm.DB,
	queueHealth QueueHealthChecker,
) HealthHandler {
	return HealthHandler{
		cfg:         cfg,
		db:          db,
		queueHealth: queueHealth,
	}
}

type HealthResponseEnvelope struct {
	EnvelopeBase
	Data HealthResponse `json:"data"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

// @Summary Health check
// @Description Reports service liveness by checking database connectivity and queue worker availability, responding 200 with an ok status when healthy. Returns 503 when the database is unreachable or the queue worker is not running.
// @Tags System
// @Produce json
// @Success 200 {object} HealthResponseEnvelope "Service is healthy."
// @Failure 503 {object} httpx.ErrorResponse "Service unavailable"
// @Router /health [get]
func (h HealthHandler) Health(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	sqlDB, err := h.db.DB()
	if err != nil {
		return CreateServiceUnavailableErrorResponse(c, "Database is unavailable.", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return CreateServiceUnavailableErrorResponse(c, "Database is unavailable.", err)
	}

	if err := h.queueHealth.WorkerRunning(); err != nil {
		return CreateServiceUnavailableErrorResponse(c, "Queue worker is unavailable.", err)
	}

	return CreateSuccessResponse(c, "Service is healthy.", HealthResponse{
		Status:  "ok",
		Service: h.cfg.ApplicationName,
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}
