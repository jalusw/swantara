package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h AuthHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	auth := api.Group("/auth")

	auth.Post("/login", httpx.AuthRateLimitGuard(), h.LoginHandler)
	auth.Post("/register", httpx.AuthRateLimitGuard(), h.RegisterHandler)
	auth.Post("/email/check", httpx.AuthRateLimitGuard(), h.CheckEmailHandler)
	auth.Post("/refresh", h.RefreshHandler)
	auth.Post("/logout", guards.AuthN, h.LogoutHandler)
	auth.Post("/email-verification/request", httpx.AuthRateLimitGuard(), h.RequestEmailVerificationHandler)
	auth.Post("/email-verification/verify", h.VerifyEmailHandler)
	auth.Post("/password-reset/request", httpx.AuthRateLimitGuard(), h.RequestPasswordResetHandler)
	auth.Post("/password-reset", httpx.AuthRateLimitGuard(), h.ResetPasswordHandler)

	sessions := auth.Group("", guards.AuthN)
	sessions.Get("/sessions", h.GetSessionsHandler)
	sessions.Delete("/sessions/:session_id", h.RevokeSessionHandler)
}

func (h UserHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	api.Get("/me", guards.AuthN, h.GetMe)
	api.Put("/me/avatar", guards.AuthN, h.UpdateMeAvatar)
	api.Get("/me/avatar", guards.AuthN, h.GetMeAvatar)
	api.Get("/me/organizations", guards.AuthN, h.MeOrganizations)
	api.Get("/me/organizations/:organization_id/permissions", guards.AuthN, h.MeOrganizationPermissions)

	users := api.Group("/users", guards.AuthN)
	users.Get("/", h.GetUsers)
	users.Get("/:id", h.GetUser)
	users.Get("/:id/avatar", h.GetUserAvatar)

	users.Post("/", h.CreateUser)
	users.Put("/:id", h.UpdateUser)
	users.Delete("/:id", h.DeleteUser)
}
