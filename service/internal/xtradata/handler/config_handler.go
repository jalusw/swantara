package handler

import "github.com/jalusw/swantara/apps/service/internal/xtradata"

type SystemConfigHandler struct {
	configSvc xtradata.ConfigService
}

func NewSystemConfigHandler(configSvc xtradata.ConfigService) SystemConfigHandler {
	return SystemConfigHandler{configSvc: configSvc}
}
