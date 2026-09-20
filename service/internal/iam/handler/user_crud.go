package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

var userQueryAllowlist = map[string]struct{}{
	"username":   {},
	"first_name": {},
	"last_name":  {},
	"created_at": {},
	"updated_at": {},
}

type GetUsersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetUsersResponse `json:"data"`
}
type GetUsersResponse struct {
	Users []PublicUser `json:"users"`
}

// @Summary      List users
// @Description  List public users. Any authenticated user may call it. Pass q to search by username or name like GitHub collaborator invite; email or phone queries return no suggestions.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        q query string false "Search by username or name"
// @Param        page query int false "Page number"
// @Param        size query int false "Page size"
// @Success      200 {object} GetUsersResponseEnvelope
// @Failure      401 {object} httpx.ErrorResponse
// @Security     BearerAuth
// @Router       /users [get]
func (h UserHandler) GetUsers(c fiber.Ctx) error {
	if term := strings.TrimSpace(c.Query("q")); term != "" {
		return h.searchUsers(c, term)
	}

	parsedQuery, err := httpx.ParseQueryParams(c, userQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	clientQuery := *parsedQuery
	parsedQuery.Filters = append([]query.Filter{{Field: "private", Operator: query.Equal, Value: false}}, parsedQuery.Filters...)

	page, err := h.userSvc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("user list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve users.", err)
	}

	users := make([]PublicUser, len(page.Items))
	for i, user := range page.Items {
		users[i] = newPublicUser(user)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "users.csv", users)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Users retrieved successfully.", GetUsersResponse{
		Users: users,
	}, httpx.BuildListMeta(&clientQuery, page.Count))
}

func (h UserHandler) searchUsers(c fiber.Ctx, term string) error {
	found, err := h.userSvc.Search(c, term, iam.UserSearchLimit)
	if err != nil {
		httpx.RequestLog(c).Error("user search failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve users.", err)
	}

	users := make([]PublicUser, len(found))
	for i, user := range found {
		users[i] = newPublicUser(user)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Users retrieved successfully.", GetUsersResponse{
		Users: users,
	}, httpx.BuildListMeta(&query.Query{Pagination: &query.Pagination{Page: 1, Size: iam.UserSearchLimit}}, int64(len(users))))
}

type GetUserResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetUserResponse `json:"data"`
}
type GetUserResponse struct {
	User PublicUser `json:"user"`
}

func buildUserLinks(c fiber.Ctx, userID uint64) []httpx.Link {
	return []httpx.Link{
		{Rel: "self", Method: fiber.MethodGet, Href: fmt.Sprintf("%s/users/%d", httpx.LinkBaseFor(c), userID)},
		{Rel: "update", Method: fiber.MethodPut, Href: fmt.Sprintf("%s/users/%d", httpx.LinkBaseFor(c), userID)},
		{Rel: "delete", Method: fiber.MethodDelete, Href: fmt.Sprintf("%s/users/%d", httpx.LinkBaseFor(c), userID)},
	}
}

func (h UserHandler) GetUser(c fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid user id provided.", nil)
	}

	user, err := h.userSvc.Find(c, userID)
	if err != nil {
		httpx.RequestLog(c).Error("user find failed", "user_id", userID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get user detail.", err)
	}

	if user == nil || user.ID == 0 || user.Private {
		return httpx.CreateNotFoundResponse(c, "User not found.")
	}

	return httpx.CreateSuccessResponseWithLinks(c, "User retrieved successfully.", GetUserResponse{
		User: newPublicUser(user),
	}, buildUserLinks(c, user.ID))
}

func (h UserHandler) GetUserAvatar(c fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || userID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid user id provided.", nil)
	}
	return h.serveAvatar(c, userID)
}

type CreateUserRequest struct {
	Username   string  `json:"username" validate:"required"`
	FirstName  string  `json:"first_name" validate:"required"`
	LastName   *string `json:"last_name"`
	Email      string  `json:"email" validate:"required,email"`
	Phone      *string `json:"phone"`
	Password   string  `json:"password" validate:"required,min=8,max=72"`
	Birthday   *string `json:"birthday"`
	Sex        *string `json:"sex"`
	Address    *string `json:"address"`
	City       *string `json:"city"`
	PostalCode *string `json:"postal_code"`
}

type CreateUserResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateUserResponse `json:"data"`
}
type CreateUserResponse struct {
	User PublicUser `json:"user"`
}

func (h UserHandler) CreateUser(c fiber.Ctx) error {
	var request CreateUserRequest
	if !bindAndValidate(c, &request) {
		return nil
	}

	birthday, ok := parseBirthday(c, request.Birthday)
	if !ok {
		return nil
	}

	user, err := h.userSvc.Create(c, &iam.User{
		Username:   request.Username,
		FirstName:  request.FirstName,
		LastName:   request.LastName,
		Email:      request.Email,
		Phone:      request.Phone,
		Password:   request.Password,
		Birthday:   birthday,
		Sex:        request.Sex,
		Address:    request.Address,
		City:       request.City,
		PostalCode: request.PostalCode,
	})
	if err != nil {
		switch {
		case errors.Is(err, iam.ErrEmailRegistered):
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Email is already registered.", nil)
		case errors.Is(err, iam.ErrUsernameTaken):
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Username is already taken.", nil)
		case errors.Is(err, iam.ErrPhoneTaken):
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Phone is already registered.", nil)
		default:
			httpx.RequestLog(c).Error("user creation failed", "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to create user.", err)
		}
	}

	httpx.RequestLog(c).Info("user created", "user_id", user.ID)

	return httpx.CreateCreatedResponse(c, "User created successfully.", CreateUserResponse{
		User: newPublicUser(user),
	})
}

type UpdateUserRequest struct {
	FirstName  string  `json:"first_name" validate:"required"`
	LastName   *string `json:"last_name"`
	Avatar     *string `json:"avatar"`
	Bio        *string `json:"bio"`
	Birthday   *string `json:"birthday"`
	Active     *bool   `json:"active"`
	Sex        *string `json:"sex" validate:"omitempty,oneof=male female other"`
	Address    *string `json:"address"`
	City       *string `json:"city"`
	PostalCode *string `json:"postal_code"`
}

type UpdateUserResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateUserResponse `json:"data"`
}
type UpdateUserResponse struct {
	User PublicUser `json:"user"`
}

func (h UserHandler) UpdateUser(c fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid user id provided.", nil)
	}

	if err := ensureSelf(c, userID); err != nil {
		return writeAccessError(c, err)
	}

	var request UpdateUserRequest
	if !bindAndValidate(c, &request) {
		return nil
	}

	birthday, ok := parseBirthday(c, request.Birthday)
	if !ok {
		return nil
	}

	user, err := h.userSvc.Update(c, userID, &iam.User{
		FirstName:  request.FirstName,
		LastName:   request.LastName,
		Avatar:     request.Avatar,
		Bio:        request.Bio,
		Birthday:   birthday,
		Sex:        request.Sex,
		Address:    request.Address,
		City:       request.City,
		PostalCode: request.PostalCode,
	}, request.Active)
	if err != nil {
		switch {
		case errors.Is(err, iam.ErrUserNotFound):
			return httpx.CreateNotFoundResponse(c, "User not found.")
		default:
			httpx.RequestLog(c).Error("user update failed", "user_id", userID, "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to update user.", err)
		}
	}

	httpx.RequestLog(c).Info("user updated", "user_id", userID)

	return httpx.CreateSuccessResponse(c, "User updated successfully.", UpdateUserResponse{
		User: newPublicUser(user),
	})
}

func (h UserHandler) DeleteUser(c fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid user id provided.", nil)
	}

	if err := h.userSvc.Delete(c, userID); err != nil {
		httpx.RequestLog(c).Error("user deletion failed", "user_id", userID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete user.", err)
	}

	httpx.RequestLog(c).Info("user deleted", "user_id", userID)

	return httpx.CreateNoContentResponse(c)
}
