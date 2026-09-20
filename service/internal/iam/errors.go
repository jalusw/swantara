package iam

import "errors"

var (
	ErrEmailRegistered      = errors.New("email is registered")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotActive        = errors.New("user is not active")
	ErrUserNotFound         = errors.New("user not found")
	ErrUsernameTaken        = errors.New("username is taken")
	ErrPhoneTaken           = errors.New("phone is taken")
	ErrForbidden            = errors.New("forbidden")
	ErrInvalidRefreshToken  = errors.New("invalid or expired refresh token")
	ErrInvalidActionToken   = errors.New("invalid or already used action token")
	ErrActionTokenExpired   = errors.New("action token has expired")
	ErrTokenReused          = errors.New("refresh token already used")
	ErrTokenNotFound        = errors.New("refresh token not found")
	ErrAlreadyMember        = errors.New("user is already a member of the organization")
	ErrMemberNotFound       = errors.New("member not found")
	ErrMemberRoleNotFound   = errors.New("member role not found")
	ErrMemberRoleExists     = errors.New("member role already exists")
	ErrAvatarEmpty          = errors.New("avatar content cannot be empty")
	ErrAvatarInvalidType    = errors.New("avatar must be an image")
	ErrAvatarNotFound       = errors.New("avatar not found")
	ErrUnexpectedSignMethod = errors.New("unexpected signing method")
)
