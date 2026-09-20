package iam

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
)

var (
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionRevoked        = errors.New("session revoked")
	ErrSessionExpired        = errors.New("session has expired")
	ErrSessionNotOwnedByUser = errors.New("session does not belong to user")
)

type SessionClientInfo struct {
	IPAddress  string
	DeviceName string
	OS         string
	Browser    string
}

func sessionModel(client SessionClientInfo, userID uint64, refreshToken string, expiresAt time.Time) *UserSession {
	ipAddress := client.IPAddress
	deviceName := client.DeviceName
	os := client.OS
	browser := client.Browser
	hashedToken := helper.HashToken(refreshToken)

	return &UserSession{
		IPAddress:    &ipAddress,
		DeviceName:   &deviceName,
		OS:           &os,
		Browser:      &browser,
		RefreshToken: &hashedToken,
		UserID:       userID,
		ExpiresAt:    &expiresAt,
	}
}

func (s AuthService) ListUserSessions(ctx context.Context, userID uint64) ([]*UserSession, error) {
	return s.userSessionDAO.ListUserSessions(ctx, userID)
}

func (s AuthService) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	return s.userSvc.IsEmailTaken(ctx, email)
}

func (s AuthService) Logout(ctx context.Context, refreshToken string) error {
	session, err := s.userSessionDAO.FindByRefreshToken(ctx, helper.HashToken(refreshToken))
	if err != nil {
		return err
	}

	if session == nil || session.RevokedAt != nil {
		return nil
	}

	now := time.Now()
	session.RevokedAt = &now
	_, err = s.userSessionDAO.Update(ctx, session)
	return err
}

func (s AuthService) RevokeSession(ctx context.Context, userID, sessionID uint64) error {
	session, err := s.userSessionDAO.Find(ctx, sessionID)
	if err != nil {
		return err
	}

	if session == nil {
		return ErrSessionNotFound
	}

	if session.UserID != userID {
		return ErrSessionNotOwnedByUser
	}

	now := time.Now()
	session.RevokedAt = &now
	_, err = s.userSessionDAO.Update(ctx, session)
	return err
}
