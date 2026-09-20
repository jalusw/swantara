package iam

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type User struct {
	model.Base
	Username        string     `json:"username"`
	FirstName       string     `json:"first_name"`
	LastName        *string    `json:"last_name"`
	Email           string     `json:"email" audit:"redact"`
	Phone           *string    `json:"phone" audit:"redact"`
	Password        string     `json:"-"`
	Avatar          *string    `json:"avatar"`
	Bio             *string    `json:"bio"`
	Birthday        *time.Time `json:"birthday"`
	Active          bool       `json:"active"`
	Private         bool       `json:"private"`
	Sex             *string    `json:"sex"`
	Address         *string    `json:"address"`
	City            *string    `json:"city"`
	PostalCode      *string    `json:"postal_code"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at"`
}

type UserSession struct {
	model.Base
	DeviceName   *string    `json:"device_name"`
	OS           *string    `json:"os"`
	Browser      *string    `json:"browser"`
	UserAgent    *string    `json:"user_agent"`
	RefreshToken *string    `json:"-"`
	IPAddress    *string    `json:"ip_address"`
	Country      *string    `json:"country"`
	City         *string    `json:"city"`
	ExpiresAt    *time.Time `json:"expires_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	UserID       uint64     `json:"user_id"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type UserEmailVerification struct {
	model.Base
	UserID            uint64     `json:"user_id"`
	VerificationToken string     `json:"-" audit:"redact"`
	ExpiresAt         time.Time  `json:"expires_at"`
	UsedAt            *time.Time `json:"used_at"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type UserPasswordReset struct {
	model.Base
	UserID         uint64     `json:"user_id"`
	ResetTokenHash string     `json:"-" audit:"redact"`
	ExpiresAt      time.Time  `json:"expires_at"`
	UsedAt         *time.Time `json:"used_at"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type Permission struct {
	model.Base
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
	Action      string  `json:"action"`
	Resource    string  `json:"resource"`
}

type Member struct {
	model.Base
	UserID         uint64  `json:"user_id"`
	OrganizationID uint64  `json:"organization_id"`
	Position       *string `json:"position"`

	User         *User                   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Organization *reference.Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Roles        []*MemberRole           `json:"roles,omitempty" gorm:"many2many:member_role_assignments;joinForeignKey:member_id;joinReferences:role_id"`
}

type MemberRole struct {
	model.Base
	OrganizationID uint64  `json:"organization_id"`
	Name           string  `json:"name"`
	Code           string  `json:"code"`
	Description    *string `json:"description"`

	Permissions []*Permission `json:"permissions,omitempty" gorm:"many2many:member_role_permissions;joinForeignKey:member_role_id;joinReferences:permission_id"`
}

type MemberRoleAssignment struct {
	model.Base
	MemberID uint64 `json:"member_id"`
	RoleID   uint64 `json:"role_id"`
}

type MemberRolePermission struct {
	model.Base
	MemberRoleID uint64 `json:"member_role_id"`
	PermissionID uint64 `json:"permission_id"`
}
