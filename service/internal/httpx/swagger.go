package httpx

import (
	"strconv"

	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/httpx/docs"
)

func RegisterDocs(app *fiber.App, cfg *config.Config) {
	docs.SwaggerInfo.Host = cfg.ApplicationHost + ":" + strconv.Itoa(cfg.ApplicationPort)
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Version = cfg.ApplicationVersion

	app.Get("/docs/*", swaggo.HandlerDefault)
}
