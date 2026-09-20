package iam

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

type TokenService struct {
	issuer   string
	audience []string

	accessTokenSecret     string
	accessTokenExpires    time.Duration
	accessTokenSignMethod jwt.SigningMethod

	refreshTokenSecret     string
	refreshTokenExpires    time.Duration
	refreshTokenSignMethod jwt.SigningMethod
}

type TokenClaims struct {
	UserID    uint64 `json:"user_id"`
	Email     string `json:"email"`
	SessionID uint64 `json:"session_id"`
	Jti       string `json:"jti,omitempty"`

	jwt.RegisteredClaims
}

func NewTokenService(cfg *config.Config) TokenService {
	return TokenService{
		issuer:                 cfg.ApplicationURL,
		audience:               []string{cfg.ClientWebURL},
		accessTokenSecret:      cfg.JWTAccessTokenSecret,
		accessTokenExpires:     cfg.JWTAccessTokenExpiresDuration,
		accessTokenSignMethod:  cfg.JWTAccessTokenSigningMethod,
		refreshTokenSecret:     cfg.JWTRefreshTokenSecret,
		refreshTokenExpires:    cfg.JWTRefreshTokenExpiresDuration,
		refreshTokenSignMethod: cfg.JWTRefreshTokenSigningMethod,
	}
}

func (s TokenService) IssueUserAccessToken(user *User, sessionID uint64) (string, *time.Time, error) {
	return s.issue(user, sessionID, s.accessTokenSecret, s.accessTokenExpires, s.accessTokenSignMethod, "")
}

func (s TokenService) IssueUserRefreshToken(user *User) (string, *time.Time, error) {
	jti, err := helper.RandomHex(16)
	if err != nil {
		return "", nil, err
	}
	return s.issue(user, 0, s.refreshTokenSecret, s.refreshTokenExpires, s.refreshTokenSignMethod, jti)
}

func (s TokenService) issue(user *User, sessionID uint64, secret string, expires time.Duration, signMethod jwt.SigningMethod, jti string) (string, *time.Time, error) {
	expiresAt := jwt.NewNumericDate(time.Now().Add(expires))

	claims := TokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		SessionID: sessionID,
		Jti:       jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(user.ID)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: expiresAt,
			Issuer:    s.issuer,
			Audience:  s.audience,
		},
	}

	token := jwt.NewWithClaims(signMethod, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", nil, err
	}

	return signed, &expiresAt.Time, nil
}

func (s TokenService) VerifyUserAccessToken(tokenString string) (*TokenClaims, error) {
	return s.verify(s.accessTokenSecret, s.accessTokenSignMethod, tokenString)
}

func (s TokenService) VerifyUserRefreshToken(tokenString string) (*TokenClaims, error) {
	return s.verify(s.refreshTokenSecret, s.refreshTokenSignMethod, tokenString)
}

func (s TokenService) verify(secret string, signMethod jwt.SigningMethod, tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	_, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != signMethod {
				return nil, fmt.Errorf("%w: %v", ErrUnexpectedSignMethod, token.Header["alg"])
			}
			return []byte(secret), nil
		},
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience...),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
