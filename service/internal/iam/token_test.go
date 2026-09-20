package iam

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jalusw/swantara/apps/service/internal/config"
)

type errSigningMethod struct{}

func (errSigningMethod) Alg() string { return "ERR" }
func (errSigningMethod) Sign(_ string, _ any) ([]byte, error) {
	return nil, errors.New("forced sign error")
}
func (errSigningMethod) Verify(_ string, _ []byte, _ any) error { return nil }

func TestTokenService(t *testing.T) {
	cfg := &config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	tokenSvc := NewTokenService(cfg)
	user := UserFixture()

	t.Run("issue and verify user access token", func(t *testing.T) {
		token, expiresAt, err := tokenSvc.IssueUserAccessToken(user, 1)

		if err != nil {
			t.Error(err)
		}

		if token == "" {
			t.Error("The issue user access token doesn't generate anything")
		}

		if expiresAt == nil {
			t.Error("The issue user access token doesn't provided expires at")
		}

		claims, err := tokenSvc.VerifyUserAccessToken(token)

		if err != nil {
			t.Error(err)
		}

		if claims.UserID != user.ID {
			t.Error("The claims user id doesn't match")
		}

		if claims.Email != user.Email {
			t.Error("The claims user email doesn't match")
		}

		if claims.SessionID != 1 {
			t.Error("The claims session id doesn't match")
		}
	})

	t.Run("issue and verify user refresh token", func(t *testing.T) {
		token, expiresAt, err := tokenSvc.IssueUserRefreshToken(user)

		if err != nil {
			t.Error(err)
		}

		if token == "" {
			t.Error("The issue user refresh token doesn't generate anything")
		}

		if expiresAt == nil {
			t.Error("The issue user refresh token doesn't provide expires at")
		}

		claims, err := tokenSvc.VerifyUserRefreshToken(token)

		if err != nil {
			t.Error(err)
		}

		if claims.UserID != user.ID {
			t.Error("The claims user id doesn't match")
		}

		if claims.Email != user.Email {
			t.Error("The claims user email doesn't match")
		}

	})
}

func TestTokenService_IssuesIssuerAndAudience(t *testing.T) {
	cfg := &config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	tokenSvc := NewTokenService(cfg)
	user := UserFixture()

	accessToken, _, err := tokenSvc.IssueUserAccessToken(user, 1)
	if err != nil {
		t.Error(err)
	}
	accessClaims, err := tokenSvc.VerifyUserAccessToken(accessToken)
	if err != nil {
		t.Error(err)
	}
	if accessClaims.Issuer != cfg.ApplicationURL {
		t.Errorf("expected issuer %q, got %q", cfg.ApplicationURL, accessClaims.Issuer)
	}
	if len(accessClaims.Audience) != 1 || accessClaims.Audience[0] != cfg.ClientWebURL {
		t.Errorf("expected audience %q, got %v", cfg.ClientWebURL, accessClaims.Audience)
	}

	refreshToken, _, err := tokenSvc.IssueUserRefreshToken(user)
	if err != nil {
		t.Error(err)
	}
	refreshClaims, err := tokenSvc.VerifyUserRefreshToken(refreshToken)
	if err != nil {
		t.Error(err)
	}
	if refreshClaims.Issuer != cfg.ApplicationURL {
		t.Errorf("expected issuer %q, got %q", cfg.ApplicationURL, refreshClaims.Issuer)
	}
	if len(refreshClaims.Audience) != 1 || refreshClaims.Audience[0] != cfg.ClientWebURL {
		t.Errorf("expected audience %q, got %v", cfg.ClientWebURL, refreshClaims.Audience)
	}
}

func TestTokenService_VerifyAccessToken_WrongSigningMethod(t *testing.T) {
	cfg := &config.Config{
		JWTAccessTokenSecret:           "model-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "model-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	signer := NewTokenService(cfg)
	user := UserFixture()
	token, _, err := signer.IssueUserAccessToken(user, 1)
	if err != nil {
		t.Error(err)
	}

	verifierCfg := &config.Config{
		JWTAccessTokenSecret:           "model-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS512,
		JWTRefreshTokenSecret:          "model-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	verifier := NewTokenService(verifierCfg)
	_, err = verifier.VerifyUserAccessToken(token)
	if err == nil {
		t.Error("expected error for wrong signing method, got nil")
	}
}

func TestTokenService_VerifyAccessToken_InvalidSignature(t *testing.T) {
	cfg := &config.Config{
		JWTAccessTokenSecret:          "model-secret",
		JWTAccessTokenExpires:         "30m",
		JWTAccessTokenExpiresDuration: 30 * time.Minute,
		JWTAccessTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	signer := NewTokenService(cfg)
	user := UserFixture()
	token, _, err := signer.IssueUserAccessToken(user, 1)
	if err != nil {
		t.Error(err)
	}

	verifierCfg := &config.Config{
		JWTAccessTokenSecret:          "different-secret",
		JWTAccessTokenExpires:         "30m",
		JWTAccessTokenExpiresDuration: 30 * time.Minute,
		JWTAccessTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	verifier := NewTokenService(verifierCfg)
	_, err = verifier.VerifyUserAccessToken(token)
	if err == nil {
		t.Error("expected error for invalid signature, got nil")
	}
}

func TestTokenService_VerifyRefreshToken_WrongSigningMethod(t *testing.T) {
	cfg := &config.Config{
		JWTAccessTokenSecret:           "model-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	signer := NewTokenService(cfg)
	user := UserFixture()
	token, _, err := signer.IssueUserRefreshToken(user)
	if err != nil {
		t.Error(err)
	}

	verifierCfg := &config.Config{
		JWTRefreshTokenSecret:          "refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS512,
	}
	verifier := NewTokenService(verifierCfg)
	_, err = verifier.VerifyUserRefreshToken(token)
	if err == nil {
		t.Error("expected error for wrong signing method, got nil")
	}
}

func TestTokenService_IssueAccessToken_SigningError(t *testing.T) {
	cfg := &config.Config{
		JWTAccessTokenSecret:           "secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    errSigningMethod{},
		JWTRefreshTokenSecret:          "secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	tokenSvc := NewTokenService(cfg)
	user := UserFixture()
	_, _, err := tokenSvc.IssueUserAccessToken(user, 1)
	if err == nil {
		t.Error("expected signing error, got nil")
	}
}

func TestTokenService_IssueRefreshToken_SigningError(t *testing.T) {
	cfg := &config.Config{
		JWTAccessTokenSecret:           "secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   errSigningMethod{},
	}
	tokenSvc := NewTokenService(cfg)
	user := UserFixture()
	_, _, err := tokenSvc.IssueUserRefreshToken(user)
	if err == nil {
		t.Error("expected signing error, got nil")
	}
}

func TestTokenService_VerifyRefreshToken_InvalidSignature(t *testing.T) {
	cfg := &config.Config{
		JWTRefreshTokenSecret:          "refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	signer := NewTokenService(cfg)
	user := UserFixture()
	token, _, err := signer.IssueUserRefreshToken(user)
	if err != nil {
		t.Error(err)
	}

	verifierCfg := &config.Config{
		JWTRefreshTokenSecret:          "different-refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	verifier := NewTokenService(verifierCfg)
	_, err = verifier.VerifyUserRefreshToken(token)
	if err == nil {
		t.Error("expected error for invalid signature, got nil")
	}
}
