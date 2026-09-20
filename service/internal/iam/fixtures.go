package iam

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func UserFixture(opts ...func(*User) *User) *User {
	lastName := gofakeit.LastName()
	now := time.Now()
	user := &User{
		Base:      model.Base{ID: uint64(gofakeit.Number(1, 100)), CreatedAt: now, UpdatedAt: now},
		FirstName: gofakeit.FirstName(),
		LastName:  &lastName,
		Email:     gofakeit.Email(),
		Active:    true,
		Password:  gofakeit.Password(true, true, true, true, true, 8),
	}
	for _, opt := range opts {
		opt(user)
	}
	return user
}

func UserSessionFixture(opts ...func(*UserSession) *UserSession) *UserSession {
	deviceName := gofakeit.Word()
	os := gofakeit.Word()
	browser := gofakeit.Word()
	userAgent := gofakeit.UserAgent()
	refreshToken := gofakeit.Password(true, true, true, true, true, 32)
	ipAddress := gofakeit.IPv4Address()
	country := gofakeit.Country()
	city := gofakeit.City()
	expiresAt := time.Now().Add(time.Hour)
	lastUsedAt := time.Now()
	now := time.Now()
	session := &UserSession{
		Base:         model.Base{ID: uint64(gofakeit.Number(1, 100)), CreatedAt: now, UpdatedAt: now},
		DeviceName:   &deviceName,
		OS:           &os,
		Browser:      &browser,
		UserAgent:    &userAgent,
		RefreshToken: &refreshToken,
		IPAddress:    &ipAddress,
		Country:      &country,
		City:         &city,
		ExpiresAt:    &expiresAt,
		LastUsedAt:   &lastUsedAt,
		UserID:       uint64(gofakeit.Number(1, 100)),
	}
	for _, opt := range opts {
		opt(session)
	}
	return session
}

func PermissionFixture(opts ...func(*Permission) *Permission) *Permission {
	description := gofakeit.Sentence(2)
	now := time.Now()
	permission := &Permission{
		Base:        model.Base{ID: uint64(gofakeit.Number(1, 100)), CreatedAt: now, UpdatedAt: now},
		Name:        gofakeit.Word(),
		Code:        gofakeit.Word(),
		Description: &description,
		Action:      gofakeit.Word(),
		Resource:    gofakeit.Word(),
	}
	for _, opt := range opts {
		opt(permission)
	}
	return permission
}
