package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type MemberHandler struct {
	memberSvc iam.MemberService
}

func NewMemberHandler(memberSvc iam.MemberService) MemberHandler {
	return MemberHandler{memberSvc: memberSvc}
}

func (h MemberHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	members := api.Group("/members", guards.AuthN)
	members.Get("/", guards.Guard("member", "view"), h.ListMembers)
	members.Post("/", guards.Guard("member", "create"), h.InviteMember)
	members.Delete("/:id", guards.Guard("member", "delete"), h.RemoveMember)

	roles := api.Group("/member-roles", guards.AuthN)
	roles.Get("/", guards.Guard("member_role", "view"), h.ListMemberRoles)
	roles.Post("/", guards.Guard("member_role", "create"), h.CreateMemberRole)

	permissions := api.Group("/permissions", guards.AuthN)
	permissions.Get("/", guards.Guard("permission", "view"), h.ListPermissions)
}
