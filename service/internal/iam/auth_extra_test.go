package iam

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"golang.org/x/crypto/bcrypt"
)

func extraPasswordSvc() PasswordService {
	return NewPasswordService(&config.Config{BcryptCost: bcrypt.MinCost, Pepper: "test-pepper"})
}

func extraTokenSvc() TokenService {
	return NewTokenService(&config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "test-access-secret",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "test-refresh-secret",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	})
}

func brokenTokenSvc() TokenService {
	svc := extraTokenSvc()
	svc.accessTokenSignMethod = errSigningMethod{}
	svc.refreshTokenSignMethod = errSigningMethod{}
	return svc
}

func extraUser(passwordSvc PasswordService, password string) *User {
	user := UserFixture()
	hash, err := passwordSvc.Hash(password)
	if err != nil {
		panic(err)
	}
	user.Password = hash
	user.Active = true
	return user
}

func TestLogin_TokenFailures(t *testing.T) {
	ctx := context.Background()
	passwordSvc := extraPasswordSvc()
	user := extraUser(passwordSvc, "secret")

	newSvc := func(tokenSvc TokenService, sessions UserSessionDAOMock) AuthService {
		return testAuthActionService(UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return user, nil },
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, sessions)
	}
	_ = newSvc

	t.Run("refresh token issue error", func(t *testing.T) {
		svc := testAuthActionService(UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return user, nil },
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		svc.tokenSvc = brokenTokenSvc()
		svc.passwordSvc = passwordSvc
		_, err := svc.Login(ctx, user.Email, "secret", SessionClientInfo{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("session create error", func(t *testing.T) {
		svc := testAuthActionService(UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return user, nil },
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{
			DAOMock: DAOMock[UserSession]{
				CreateFunc: func(_ context.Context, _ *UserSession) (*UserSession, error) {
					return nil, errors.New("db down")
				},
			},
		})
		svc.tokenSvc = extraTokenSvc()
		svc.passwordSvc = passwordSvc
		_, err := svc.Login(ctx, user.Email, "secret", SessionClientInfo{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRefresh_Failures(t *testing.T) {
	ctx := context.Background()
	passwordSvc := extraPasswordSvc()
	tokenSvc := extraTokenSvc()
	user := extraUser(passwordSvc, "secret")

	issueRefresh := func() string {
		token, _, err := tokenSvc.IssueUserRefreshToken(user)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}

	t.Run("inactive user", func(t *testing.T) {
		inactive := *user
		inactive.Active = false
		svc := testAuthActionService(UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc: func(_ context.Context, _ uint64) (*User, error) { return &inactive, nil },
			},
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		svc.tokenSvc = tokenSvc
		_, err := svc.Refresh(ctx, issueRefresh(), SessionClientInfo{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("rotate error", func(t *testing.T) {
		svc := testAuthActionService(UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc: func(_ context.Context, _ uint64) (*User, error) { return user, nil },
			},
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{
			RotateRefreshTokenFunc: func(_ context.Context, _ string, _ *UserSession) (*UserSession, error) {
				return nil, errors.New("db down")
			},
		})
		svc.tokenSvc = tokenSvc
		_, err := svc.Refresh(ctx, issueRefresh(), SessionClientInfo{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("issue error", func(t *testing.T) {
		svc := testAuthActionService(UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc: func(_ context.Context, _ uint64) (*User, error) { return user, nil },
			},
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		svc.tokenSvc = brokenTokenSvc()
		_, err := svc.Refresh(ctx, issueRefresh(), SessionClientInfo{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRequestEmailVerification_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("lookup error", func(t *testing.T) {
		svc := testAuthActionService(UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return nil, errors.New("db down") },
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		if err := svc.RequestEmailVerification(ctx, "a@b.c"); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("delete error", func(t *testing.T) {
		user := UserFixture()
		svc := testAuthActionService(UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return user, nil },
		}, UserEmailVerificationDAOMock{
			DeleteByUserIDFunc: func(_ context.Context, _ uint64) error { return errors.New("db down") },
		}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		if err := svc.RequestEmailVerification(ctx, "a@b.c"); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("create error", func(t *testing.T) {
		user := UserFixture()
		svc := testAuthActionService(UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return user, nil },
		}, UserEmailVerificationDAOMock{
			DAOMock: DAOMock[UserEmailVerification]{
				CreateFunc: func(_ context.Context, _ *UserEmailVerification) (*UserEmailVerification, error) {
					return nil, errors.New("db down")
				},
			},
		}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		if err := svc.RequestEmailVerification(ctx, "a@b.c"); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestVerifyEmail_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("lookup error", func(t *testing.T) {
		svc := testAuthActionService(UserDAOMock{}, UserEmailVerificationDAOMock{
			FindByTokenFunc: func(_ context.Context, _ string) (*UserEmailVerification, error) { return nil, dbErr },
		}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		helper.AssertError(t, svc.VerifyEmail(ctx, "tok"), true, dbErr)
	})

	t.Run("user lookup error", func(t *testing.T) {
		verification := &UserEmailVerification{Base: model.Base{ID: 1}, UserID: 7, ExpiresAt: time.Now().Add(time.Hour)}
		svc := testAuthActionService(UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc: func(_ context.Context, _ uint64) (*User, error) { return nil, dbErr },
			},
		}, UserEmailVerificationDAOMock{
			FindByTokenFunc: func(_ context.Context, _ string) (*UserEmailVerification, error) { return verification, nil },
		}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		helper.AssertError(t, svc.VerifyEmail(ctx, "tok"), true, dbErr)
	})

	t.Run("update error", func(t *testing.T) {
		verification := &UserEmailVerification{Base: model.Base{ID: 1}, UserID: 7, ExpiresAt: time.Now().Add(time.Hour)}
		svc := testAuthActionService(UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc:   func(_ context.Context, _ uint64) (*User, error) { return &User{Base: model.Base{ID: 7}}, nil },
				UpdateFunc: func(_ context.Context, _ *User) (*User, error) { return nil, dbErr },
			},
		}, UserEmailVerificationDAOMock{
			FindByTokenFunc: func(_ context.Context, _ string) (*UserEmailVerification, error) { return verification, nil },
		}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		helper.AssertError(t, svc.VerifyEmail(ctx, "tok"), true, dbErr)
	})
}

func TestRequestPasswordReset_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("lookup error", func(t *testing.T) {
		svc := testAuthActionService(UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return nil, dbErr },
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		helper.AssertError(t, svc.RequestPasswordReset(ctx, "a@b.c"), true, dbErr)
	})

	t.Run("delete error", func(t *testing.T) {
		user := UserFixture()
		user.ID = 7
		svc := testAuthActionService(UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return user, nil },
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{
			DeleteByUserIDFunc: func(_ context.Context, _ uint64) error { return dbErr },
		}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		helper.AssertError(t, svc.RequestPasswordReset(ctx, "a@b.c"), true, dbErr)
	})

	t.Run("create error", func(t *testing.T) {
		user := UserFixture()
		user.ID = 7
		svc := testAuthActionService(UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return user, nil },
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{
			DAOMock: DAOMock[UserPasswordReset]{
				CreateFunc: func(_ context.Context, _ *UserPasswordReset) (*UserPasswordReset, error) { return nil, dbErr },
			},
		}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		helper.AssertError(t, svc.RequestPasswordReset(ctx, "a@b.c"), true, dbErr)
	})
}

func TestResetPassword_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	reset := &UserPasswordReset{Base: model.Base{ID: 1}, UserID: 7, ExpiresAt: time.Now().Add(time.Hour)}

	newSvc := func(findFn func(_ context.Context, _ string) (*UserPasswordReset, error)) AuthService {
		return testAuthActionService(UserDAOMock{}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{
			FindByTokenHashFunc: findFn,
		}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
	}

	t.Run("lookup error", func(t *testing.T) {
		svc := newSvc(func(_ context.Context, _ string) (*UserPasswordReset, error) { return nil, dbErr })
		helper.AssertError(t, svc.ResetPassword(ctx, "tok", "new-secret"), true, dbErr)
	})

	t.Run("user lookup error", func(t *testing.T) {
		svc := testAuthActionService(UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc: func(_ context.Context, _ uint64) (*User, error) { return nil, dbErr },
			},
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{
			FindByTokenHashFunc: func(_ context.Context, _ string) (*UserPasswordReset, error) { return reset, nil },
		}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		helper.AssertError(t, svc.ResetPassword(ctx, "tok", "new-secret"), true, dbErr)
	})

	t.Run("hash error", func(t *testing.T) {
		svc := testAuthActionService(UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc: func(_ context.Context, _ uint64) (*User, error) { return &User{Base: model.Base{ID: 7}}, nil },
			},
		}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{
			FindByTokenHashFunc: func(_ context.Context, _ string) (*UserPasswordReset, error) { return reset, nil },
		}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{})
		svc.passwordSvc = NewPasswordService(&config.Config{BcryptCost: 100, Pepper: "x"})
		helper.AssertError(t, svc.ResetPassword(ctx, "tok", "new-secret"), true, nil)
	})
}
