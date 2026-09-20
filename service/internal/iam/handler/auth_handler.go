package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type AuthHandler struct {
	authSvc iam.AuthService
}

func NewAuthHandler(authSvc iam.AuthService) AuthHandler {
	return AuthHandler{
		authSvc: authSvc,
	}
}
