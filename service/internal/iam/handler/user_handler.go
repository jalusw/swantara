package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type UserHandler struct {
	userSvc   iam.UserService
	authzSvc  iam.AuthzService
	memberSvc iam.MemberService
	avatarSvc iam.AvatarService
}

func NewUserHandler(
	userSvc iam.UserService,
	authzSvc iam.AuthzService,
	memberSvc iam.MemberService,
	avatarSvc iam.AvatarService,
) UserHandler {
	return UserHandler{
		userSvc:   userSvc,
		authzSvc:  authzSvc,
		memberSvc: memberSvc,
		avatarSvc: avatarSvc,
	}
}

type PublicUser struct {
	ID        uint64  `json:"id"`
	Username  string  `json:"username"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name"`
	Avatar    *string `json:"avatar"`
	Bio       *string `json:"bio"`
}

func newPublicUser(user *iam.User) PublicUser {
	return PublicUser{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
	}
}

type Profile struct {
	ID              uint64     `json:"id"`
	Username        string     `json:"username"`
	FirstName       string     `json:"first_name"`
	LastName        *string    `json:"last_name"`
	Email           string     `json:"email"`
	Phone           *string    `json:"phone"`
	Avatar          *string    `json:"avatar"`
	Bio             *string    `json:"bio"`
	Birthday        *time.Time `json:"birthday"`
	Active          bool       `json:"active"`
	Sex             *string    `json:"sex"`
	Address         *string    `json:"address"`
	City            *string    `json:"city"`
	PostalCode      *string    `json:"postal_code"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func newProfile(user *iam.User) Profile {
	return Profile{
		ID:              user.ID,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Email:           user.Email,
		Phone:           user.Phone,
		Avatar:          user.Avatar,
		Bio:             user.Bio,
		Birthday:        user.Birthday,
		Active:          user.Active,
		Sex:             user.Sex,
		Address:         user.Address,
		City:            user.City,
		PostalCode:      user.PostalCode,
		EmailVerifiedAt: user.EmailVerifiedAt,
		PhoneVerifiedAt: user.PhoneVerifiedAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}
