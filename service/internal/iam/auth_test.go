package iam

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
	"golang.org/x/crypto/bcrypt"
)

func testAuthActionService(
	userDAO UserDAO,
	emailVerificationDAO UserEmailVerificationDAO,
	passwordResetDAO UserPasswordResetDAO,
	queueClient queue.TaskEnqueuer,
	sessionDAO UserSessionDAO,
) AuthService {
	passwordSvc := NewPasswordService(&config.Config{
		BcryptCost: bcrypt.MinCost,
		Pepper:     "test-pepper",
	})
	return NewAuthService(
		userDAO,
		NewUserService(userDAO, passwordSvc),
		sessionDAO,
		passwordSvc,
		TokenService{},
		emailVerificationDAO,
		passwordResetDAO,
		queueClient,
		time.Hour*24,
		time.Minute*30,
	)
}

func TestRequestEmailVerification(t *testing.T) {
	ctx := context.Background()
	user := UserFixture()

	t.Run("success", func(t *testing.T) {
		var created *UserEmailVerification
		var enqueuedToken string
		var deleted bool

		emailVerificationDAO := UserEmailVerificationDAOMock{
			DAOMock: DAOMock[UserEmailVerification]{
				CreateFunc: func(ctx context.Context, entity *UserEmailVerification) (*UserEmailVerification, error) {
					created = entity
					return entity, nil
				},
			},
			DeleteByUserIDFunc: func(ctx context.Context, userID uint64) error {
				deleted = true
				return nil
			},
		}
		queueClient := queue.TaskEnqueuerMock{
			EnqueueFunc: func(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
				var payload tasks.SendEmailPayload
				if err := json.Unmarshal(task.Payload(), &payload); err != nil {
					return nil, err
				}
				enqueuedToken = payload.Token
				return &asynq.TaskInfo{}, nil
			},
		}

		svc := testAuthActionService(
			UserDAOMock{
				FindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
					return user, nil
				},
			},
			emailVerificationDAO,
			UserPasswordResetDAOMock{},
			queueClient,
			UserSessionDAOMock{},
		)

		if err := svc.RequestEmailVerification(ctx, user.Email); err != nil {
			t.Fatal(err)
		}
		if !deleted {
			t.Error("expected DeleteByUserID to be called")
		}
		if created == nil {
			t.Fatal("expected verification to be created")
		}
		if created.UserID != user.ID {
			t.Errorf("expected user id %d, got %d", user.ID, created.UserID)
		}
		if !created.ExpiresAt.After(time.Now()) {
			t.Error("expected future expiry")
		}
		if enqueuedToken == "" {
			t.Error("expected a token to be enqueued")
		}
		if created.VerificationToken != helper.HashToken(enqueuedToken) {
			t.Error("expected stored token to be the hash of the enqueued token")
		}
	})

	t.Run("user not found - no op", func(t *testing.T) {
		var created bool
		emailVerificationDAO := UserEmailVerificationDAOMock{
			DAOMock: DAOMock[UserEmailVerification]{
				CreateFunc: func(ctx context.Context, entity *UserEmailVerification) (*UserEmailVerification, error) {
					created = true
					return entity, nil
				},
			},
		}
		svc := testAuthActionService(
			UserDAOMock{
				FindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
					return nil, nil
				},
			},
			emailVerificationDAO,
			UserPasswordResetDAOMock{},
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)

		if err := svc.RequestEmailVerification(ctx, "missing@test.com"); err != nil {
			t.Fatal(err)
		}
		if created {
			t.Error("expected no verification to be created")
		}
	})

	t.Run("already verified - silent skip", func(t *testing.T) {
		verifiedUser := *user
		now := time.Now()
		verifiedUser.EmailVerifiedAt = &now

		svc := testAuthActionService(
			UserDAOMock{
				FindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
					return &verifiedUser, nil
				},
			},
			UserEmailVerificationDAOMock{},
			UserPasswordResetDAOMock{},
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)

		if err := svc.RequestEmailVerification(ctx, verifiedUser.Email); err != nil {
			t.Errorf("expected nil for already verified email, got %v", err)
		}
	})

	t.Run("search error", func(t *testing.T) {
		svc := testAuthActionService(
			UserDAOMock{
				FindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
					return nil, errors.New("db error")
				},
			},
			UserEmailVerificationDAOMock{},
			UserPasswordResetDAOMock{},
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)

		if err := svc.RequestEmailVerification(ctx, "x@test.com"); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("enqueue error", func(t *testing.T) {
		emailVerificationDAO := UserEmailVerificationDAOMock{
			DAOMock: DAOMock[UserEmailVerification]{
				CreateFunc: func(ctx context.Context, entity *UserEmailVerification) (*UserEmailVerification, error) {
					return entity, nil
				},
			},
		}
		queueClient := queue.TaskEnqueuerMock{
			EnqueueFunc: func(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
				return nil, errors.New("redis down")
			},
		}
		svc := testAuthActionService(
			UserDAOMock{
				FindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
					return user, nil
				},
			},
			emailVerificationDAO,
			UserPasswordResetDAOMock{},
			queueClient,
			UserSessionDAOMock{},
		)

		if err := svc.RequestEmailVerification(ctx, user.Email); err == nil {
			t.Error("expected error")
		}
	})
}

func TestVerifyEmail(t *testing.T) {
	ctx := context.Background()
	user := UserFixture()

	t.Run("success", func(t *testing.T) {
		verification := &UserEmailVerification{
			Base:      model.Base{ID: 1},
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Hour),
		}
		var markedUsed bool
		var updated *User

		emailVerificationDAO := UserEmailVerificationDAOMock{
			FindByTokenFunc: func(ctx context.Context, tokenHash string) (*UserEmailVerification, error) {
				return verification, nil
			},
			MarkUsedFunc: func(ctx context.Context, id uint64) error {
				markedUsed = true
				return nil
			},
		}
		svc := testAuthActionService(
			UserDAOMock{},
			emailVerificationDAO,
			UserPasswordResetDAOMock{},
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)
		svc.userDAO = UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc: func(ctx context.Context, id uint64) (*User, error) {
					return user, nil
				},
				UpdateFunc: func(ctx context.Context, entity *User) (*User, error) {
					updated = entity
					return entity, nil
				},
			},
		}

		if err := svc.VerifyEmail(ctx, "raw-token"); err != nil {
			t.Fatal(err)
		}
		if !markedUsed {
			t.Error("expected MarkUsed to be called")
		}
		if updated == nil || updated.EmailVerifiedAt == nil {
			t.Error("expected EmailVerifiedAt to be set")
		}
	})

	t.Run("token not found", func(t *testing.T) {
		svc := testAuthActionService(
			UserDAOMock{},
			UserEmailVerificationDAOMock{},
			UserPasswordResetDAOMock{},
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)
		if err := svc.VerifyEmail(ctx, "raw-token"); !errors.Is(err, ErrInvalidActionToken) {
			t.Errorf("expected ErrInvalidActionToken, got %v", err)
		}
	})

	t.Run("already used", func(t *testing.T) {
		used := time.Now()
		verification := &UserEmailVerification{
			Base:      model.Base{ID: 1},
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Hour),
			UsedAt:    &used,
		}
		emailVerificationDAO := UserEmailVerificationDAOMock{
			FindByTokenFunc: func(ctx context.Context, tokenHash string) (*UserEmailVerification, error) {
				return verification, nil
			},
		}
		svc := testAuthActionService(
			UserDAOMock{},
			emailVerificationDAO,
			UserPasswordResetDAOMock{},
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)
		if err := svc.VerifyEmail(ctx, "raw-token"); !errors.Is(err, ErrInvalidActionToken) {
			t.Errorf("expected ErrInvalidActionToken, got %v", err)
		}
	})

	t.Run("expired", func(t *testing.T) {
		verification := &UserEmailVerification{
			Base:      model.Base{ID: 1},
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		emailVerificationDAO := UserEmailVerificationDAOMock{
			FindByTokenFunc: func(ctx context.Context, tokenHash string) (*UserEmailVerification, error) {
				return verification, nil
			},
		}
		svc := testAuthActionService(
			UserDAOMock{},
			emailVerificationDAO,
			UserPasswordResetDAOMock{},
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)
		if err := svc.VerifyEmail(ctx, "raw-token"); !errors.Is(err, ErrActionTokenExpired) {
			t.Errorf("expected ErrActionTokenExpired, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		verification := &UserEmailVerification{
			Base:      model.Base{ID: 1},
			UserID:    999,
			ExpiresAt: time.Now().Add(time.Hour),
		}
		emailVerificationDAO := UserEmailVerificationDAOMock{
			FindByTokenFunc: func(ctx context.Context, tokenHash string) (*UserEmailVerification, error) {
				return verification, nil
			},
		}
		svc := testAuthActionService(
			UserDAOMock{
				DAOMock: DAOMock[User]{
					FindFunc: func(ctx context.Context, id uint64) (*User, error) {
						return nil, nil
					},
				},
			},
			emailVerificationDAO,
			UserPasswordResetDAOMock{},
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)
		if err := svc.VerifyEmail(ctx, "raw-token"); !errors.Is(err, ErrInvalidActionToken) {
			t.Errorf("expected ErrInvalidActionToken, got %v", err)
		}
	})
}

func TestRequestPasswordReset(t *testing.T) {
	ctx := context.Background()
	user := UserFixture()

	t.Run("success", func(t *testing.T) {
		var created *UserPasswordReset
		var enqueuedToken string

		passwordResetDAO := UserPasswordResetDAOMock{
			DAOMock: DAOMock[UserPasswordReset]{
				CreateFunc: func(ctx context.Context, entity *UserPasswordReset) (*UserPasswordReset, error) {
					created = entity
					return entity, nil
				},
			},
		}
		queueClient := queue.TaskEnqueuerMock{
			EnqueueFunc: func(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
				var payload tasks.SendEmailPayload
				if err := json.Unmarshal(task.Payload(), &payload); err != nil {
					return nil, err
				}
				enqueuedToken = payload.Token
				return &asynq.TaskInfo{}, nil
			},
		}
		svc := testAuthActionService(
			UserDAOMock{
				FindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
					return user, nil
				},
			},
			UserEmailVerificationDAOMock{},
			passwordResetDAO,
			queueClient,
			UserSessionDAOMock{},
		)

		if err := svc.RequestPasswordReset(ctx, user.Email); err != nil {
			t.Fatal(err)
		}
		if created == nil {
			t.Fatal("expected reset to be created")
		}
		if created.UserID != user.ID {
			t.Errorf("expected user id %d, got %d", user.ID, created.UserID)
		}
		if created.ResetTokenHash != helper.HashToken(enqueuedToken) {
			t.Error("expected stored hash to match enqueued token")
		}
	})

	t.Run("user not found - no op", func(t *testing.T) {
		var created bool
		passwordResetDAO := UserPasswordResetDAOMock{
			DAOMock: DAOMock[UserPasswordReset]{
				CreateFunc: func(ctx context.Context, entity *UserPasswordReset) (*UserPasswordReset, error) {
					created = true
					return entity, nil
				},
			},
		}
		svc := testAuthActionService(
			UserDAOMock{
				FindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
					return nil, nil
				},
			},
			UserEmailVerificationDAOMock{},
			passwordResetDAO,
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)

		if err := svc.RequestPasswordReset(ctx, "missing@test.com"); err != nil {
			t.Fatal(err)
		}
		if created {
			t.Error("expected no reset to be created")
		}
	})
}

func TestResetPassword(t *testing.T) {
	ctx := context.Background()
	user := UserFixture()
	user.Password, _ = NewPasswordService(&config.Config{
		BcryptCost: bcrypt.MinCost,
		Pepper:     "test-pepper",
	}).Hash("old-password")
	oldHash := user.Password

	t.Run("success", func(t *testing.T) {
		reset := &UserPasswordReset{
			Base:      model.Base{ID: 1},
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Hour),
		}
		var markedUsed bool
		var revoked bool
		var updated *User

		passwordResetDAO := UserPasswordResetDAOMock{
			FindByTokenHashFunc: func(ctx context.Context, tokenHash string) (*UserPasswordReset, error) {
				return reset, nil
			},
			MarkUsedFunc: func(ctx context.Context, id uint64) error {
				markedUsed = true
				return nil
			},
		}
		sessionDAO := UserSessionDAOMock{
			RevokeByUserIDFunc: func(ctx context.Context, userID uint64) error {
				revoked = true
				return nil
			},
		}
		svc := testAuthActionService(
			UserDAOMock{},
			UserEmailVerificationDAOMock{},
			passwordResetDAO,
			queue.TaskEnqueuerMock{},
			sessionDAO,
		)
		svc.userDAO = UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc: func(ctx context.Context, id uint64) (*User, error) {
					return user, nil
				},
				UpdateFunc: func(ctx context.Context, entity *User) (*User, error) {
					updated = entity
					return entity, nil
				},
			},
		}

		if err := svc.ResetPassword(ctx, "raw-token", "new-password"); err != nil {
			t.Fatal(err)
		}
		if !markedUsed {
			t.Error("expected MarkUsed to be called")
		}
		if !revoked {
			t.Error("expected RevokeByUserID to be called")
		}
		if updated == nil {
			t.Fatal("expected user update")
		}
		if updated.Password == oldHash {
			t.Error("expected password hash to change")
		}
		if err := svc.passwordSvc.Verify(updated.Password, "new-password"); err != nil {
			t.Error("expected new password to verify")
		}
	})

	t.Run("token not found", func(t *testing.T) {
		svc := testAuthActionService(
			UserDAOMock{},
			UserEmailVerificationDAOMock{},
			UserPasswordResetDAOMock{},
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)
		if err := svc.ResetPassword(ctx, "raw-token", "new-password"); !errors.Is(err, ErrInvalidActionToken) {
			t.Errorf("expected ErrInvalidActionToken, got %v", err)
		}
	})

	t.Run("already used", func(t *testing.T) {
		used := time.Now()
		reset := &UserPasswordReset{
			Base:      model.Base{ID: 1},
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Hour),
			UsedAt:    &used,
		}
		passwordResetDAO := UserPasswordResetDAOMock{
			FindByTokenHashFunc: func(ctx context.Context, tokenHash string) (*UserPasswordReset, error) {
				return reset, nil
			},
		}
		svc := testAuthActionService(
			UserDAOMock{},
			UserEmailVerificationDAOMock{},
			passwordResetDAO,
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)
		if err := svc.ResetPassword(ctx, "raw-token", "new-password"); !errors.Is(err, ErrInvalidActionToken) {
			t.Errorf("expected ErrInvalidActionToken, got %v", err)
		}
	})

	t.Run("expired", func(t *testing.T) {
		reset := &UserPasswordReset{
			Base:      model.Base{ID: 1},
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		passwordResetDAO := UserPasswordResetDAOMock{
			FindByTokenHashFunc: func(ctx context.Context, tokenHash string) (*UserPasswordReset, error) {
				return reset, nil
			},
		}
		svc := testAuthActionService(
			UserDAOMock{},
			UserEmailVerificationDAOMock{},
			passwordResetDAO,
			queue.TaskEnqueuerMock{},
			UserSessionDAOMock{},
		)
		if err := svc.ResetPassword(ctx, "raw-token", "new-password"); !errors.Is(err, ErrActionTokenExpired) {
			t.Errorf("expected ErrActionTokenExpired, got %v", err)
		}
	})
}

func TestNewAuthService(t *testing.T) {
	svc := NewAuthService(
		UserDAOMock{},
		NewUserService(UserDAOMock{}, PasswordService{}),
		UserSessionDAOMock{},
		PasswordService{},
		TokenService{},
		UserEmailVerificationDAOMock{},
		UserPasswordResetDAOMock{},
		queue.TaskEnqueuerMock{},
		time.Hour*24,
		time.Minute*30,
	)
	if svc.userDAO == nil {
		t.Error("expected non-nil user dao")
	}
}

func TestSessionModel_HashesRefreshToken(t *testing.T) {
	expires := time.Now().Add(time.Hour)

	session := sessionModel(SessionClientInfo{}, 7, "raw-token", expires)

	if session.RefreshToken == nil {
		t.Fatal("expected refresh token to be set")
	}
	if *session.RefreshToken == "raw-token" {
		t.Error("refresh token must not be stored in plaintext")
	}
	if *session.RefreshToken != helper.HashToken("raw-token") {
		t.Errorf("stored token = %s, want sha256 hash", *session.RefreshToken)
	}
	if session.UserID != 7 || session.ExpiresAt == nil || !session.ExpiresAt.Equal(expires) {
		t.Error("session fields not populated")
	}
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	password := "test-password"

	passwordCfg := config.Config{
		BcryptCost: bcrypt.MinCost,
		Pepper:     "test-pepper",
	}
	passwordSvc := NewPasswordService(&passwordCfg)

	hash, err := passwordSvc.Hash(password)
	if err != nil {
		t.Fatal(err)
	}

	user := UserFixture()
	user.Password = hash

	validTokenCfg := config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "test-access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "test-refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	validTokenSvc := NewTokenService(&validTokenCfg)

	tests := []struct {
		name            string
		email           string
		password        string
		tokenSvc        TokenService
		findByEmailFunc func(ctx context.Context, email string) (*User, error)
		wantErr         bool
	}{
		{
			name:     "success",
			email:    user.Email,
			password: password,
			tokenSvc: validTokenSvc,
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return user, nil
			},
		},
		{
			name:     "user not found - nil user",
			email:    "notfound@test.com",
			password: password,
			tokenSvc: validTokenSvc,
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:     "user not found - search error",
			email:    "error@test.com",
			password: password,
			tokenSvc: validTokenSvc,
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:     "wrong password",
			email:    user.Email,
			password: "wrong-password",
			tokenSvc: validTokenSvc,
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return user, nil
			},
			wantErr: true,
		},
		{
			name:     "inactive user",
			email:    user.Email,
			password: password,
			tokenSvc: validTokenSvc,
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				inactive := *user
				inactive.Active = false
				return &inactive, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc := AuthService{
				userDAO:        UserDAOMock{FindByEmailFunc: tt.findByEmailFunc},
				userSessionDAO: UserSessionDAOMock{},
				passwordSvc:    passwordSvc,
				tokenSvc:       tt.tokenSvc,
			}
			result, err := authSvc.Login(ctx, tt.email, tt.password, SessionClientInfo{})
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected result, got nil")
			}
		})
	}
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	validPasswordCfg := config.Config{
		BcryptCost: bcrypt.MinCost,
		Pepper:     "test-pepper",
	}
	validPasswordSvc := NewPasswordService(&validPasswordCfg)

	brokenPasswordSvc := NewPasswordService(&config.Config{
		BcryptCost: 100,
		Pepper:     "test-pepper",
	})

	registerData := RegisterData{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@test.com",
		Password:  "test-password",
	}

	tests := []struct {
		name                   string
		data                   RegisterData
		passwordSvc            PasswordService
		userFindByEmailFunc    func(ctx context.Context, email string) (*User, error)
		userFindByUsernameFunc func(ctx context.Context, username string) (*User, error)
		userCreateFunc         func(ctx context.Context, user *User) (*User, error)
		wantUsername           string
		wantUsernameSuffixed   bool
		wantErr                bool
		wantErrVal             error
	}{
		{
			name:        "success",
			data:        registerData,
			passwordSvc: validPasswordSvc,
			userFindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			userFindByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			userCreateFunc: func(ctx context.Context, user *User) (*User, error) {
				return user, nil
			},
			wantUsername: "john",
		},
		{
			name:        "username taken - suffixed fallback",
			data:        registerData,
			passwordSvc: validPasswordSvc,
			userFindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			userFindByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return &User{Base: model.Base{ID: 9}, Username: "john"}, nil
			},
			userCreateFunc: func(ctx context.Context, user *User) (*User, error) {
				return user, nil
			},
			wantUsernameSuffixed: true,
		},
		{
			name:        "hash error",
			data:        registerData,
			passwordSvc: brokenPasswordSvc,
			userFindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:        "already registered",
			data:        registerData,
			passwordSvc: validPasswordSvc,
			userFindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return &User{Base: model.Base{ID: 1}, Email: registerData.Email}, nil
			},
			wantErr:    true,
			wantErrVal: ErrEmailRegistered,
		},
		{
			name:        "user search error",
			data:        registerData,
			passwordSvc: validPasswordSvc,
			userFindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:        "create error",
			data:        registerData,
			passwordSvc: validPasswordSvc,
			userFindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			userFindByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			userCreateFunc: func(ctx context.Context, user *User) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:        "username uniqueness search error",
			data:        registerData,
			passwordSvc: validPasswordSvc,
			userFindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			userFindByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userDAOMock := UserDAOMock{
				FindByEmailFunc:    tt.userFindByEmailFunc,
				FindByUsernameFunc: tt.userFindByUsernameFunc,
				DAOMock:            DAOMock[User]{CreateFunc: tt.userCreateFunc},
			}
			authSvc := AuthService{
				userDAO:     userDAOMock,
				userSvc:     UserService{userDAO: userDAOMock, passwordSvc: tt.passwordSvc},
				passwordSvc: tt.passwordSvc,
			}
			user, err := authSvc.Register(ctx, tt.data)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if user == nil {
				t.Error("expected user, got nil")
			}
			if tt.wantUsername != "" && user.Username != tt.wantUsername {
				t.Errorf("expected username %q, got %q", tt.wantUsername, user.Username)
			}
			if tt.wantUsernameSuffixed && (user.Username == "" || user.Username == "john") {
				t.Errorf("expected suffixed username, got %q", user.Username)
			}
		})
	}
}

func TestUsernameFromEmail(t *testing.T) {
	tests := []struct {
		email    string
		wantUser string
	}{
		{email: "john@test.com", wantUser: "john"},
		{email: "John.Doe@test.com", wantUser: "john.doe"},
		{email: "jane_doe@test.com", wantUser: "jane_doe"},
		{email: "user+tag@test.com", wantUser: "usertag"},
		{email: "@test.com", wantUser: "user"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := usernameFromEmail(tt.email); got != tt.wantUser {
				t.Errorf("usernameFromEmail(%q) = %q, want %q", tt.email, got, tt.wantUser)
			}
		})
	}
}

func TestRefresh_UserNotFound(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{
		JWTAccessTokenSecret:           "test-access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "test-refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	tokenSvc := NewTokenService(&cfg)

	user := UserFixture()
	refreshToken, _, err := tokenSvc.IssueUserRefreshToken(user)
	if err != nil {
		t.Fatal(err)
	}

	authSvc := AuthService{
		userDAO: UserDAOMock{DAOMock: DAOMock[User]{FindFunc: func(ctx context.Context, id uint64) (*User, error) {
			return nil, nil
		}}},
		userSessionDAO: UserSessionDAOMock{},
		tokenSvc:       tokenSvc,
	}

	_, err = authSvc.Refresh(ctx, refreshToken, SessionClientInfo{})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{
		JWTAccessTokenSecret:           "test-access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "test-refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	tokenSvc := NewTokenService(&cfg)

	authSvc := AuthService{
		userDAO: UserDAOMock{DAOMock: DAOMock[User]{FindFunc: func(ctx context.Context, id uint64) (*User, error) {
			return UserFixture(), nil
		}}},
		userSessionDAO: UserSessionDAOMock{},
		tokenSvc:       tokenSvc,
	}

	_, err := authSvc.Refresh(ctx, "not-a-token", SessionClientInfo{})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefresh(t *testing.T) {
	ctx := context.Background()

	validTokenCfg := config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "test-access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "test-refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	}
	validTokenSvc := NewTokenService(&validTokenCfg)

	user := UserFixture()

	refreshToken, _, err := validTokenSvc.IssueUserRefreshToken(user)
	if err != nil {
		t.Fatal(err)
	}

	differentSecretSvc := NewTokenService(&config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "test-access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "different-refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	})

	tests := []struct {
		name     string
		token    string
		tokenSvc TokenService
		findFunc func(ctx context.Context, id uint64) (*User, error)
		wantErr  bool
	}{
		{
			name:     "success",
			token:    refreshToken,
			tokenSvc: validTokenSvc,
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return user, nil
			},
		},
		{
			name:     "verify token error",
			token:    refreshToken,
			tokenSvc: differentSecretSvc,
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return user, nil
			},
			wantErr: true,
		},
		{
			name:     "find user error",
			token:    refreshToken,
			tokenSvc: validTokenSvc,
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:     "inactive user",
			token:    refreshToken,
			tokenSvc: validTokenSvc,
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				inactive := *user
				inactive.Active = false
				return &inactive, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc := AuthService{
				userDAO:        UserDAOMock{DAOMock: DAOMock[User]{FindFunc: tt.findFunc}},
				userSessionDAO: UserSessionDAOMock{},
				tokenSvc:       tt.tokenSvc,
			}
			result, err := authSvc.Refresh(ctx, tt.token, SessionClientInfo{})
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected result, got nil")
			}
		})
	}
}

func TestLogin_AccessTokenBoundToSession(t *testing.T) {
	ctx := context.Background()

	passwordCfg := config.Config{
		BcryptCost: bcrypt.MinCost,
		Pepper:     "test-pepper",
	}
	passwordSvc := NewPasswordService(&passwordCfg)

	hash, err := passwordSvc.Hash("test-password")
	if err != nil {
		t.Fatal(err)
	}

	user := UserFixture()
	user.Password = hash

	tokenSvc := NewTokenService(&config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "test-access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "test-refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	})

	const sessionID = uint64(42)

	authSvc := AuthService{
		userDAO: UserDAOMock{FindByEmailFunc: func(ctx context.Context, email string) (*User, error) {
			return user, nil
		}},
		userSessionDAO: UserSessionDAOMock{DAOMock: DAOMock[UserSession]{CreateFunc: func(ctx context.Context, entity *UserSession) (*UserSession, error) {
			entity.ID = sessionID
			return entity, nil
		}}},
		passwordSvc: passwordSvc,
		tokenSvc:    tokenSvc,
	}

	result, err := authSvc.Login(ctx, user.Email, "test-password", SessionClientInfo{})
	if err != nil {
		t.Fatal(err)
	}

	claims, err := tokenSvc.VerifyUserAccessToken(result.AccessToken)
	if err != nil {
		t.Fatal(err)
	}

	if claims.SessionID != sessionID {
		t.Errorf("access token session id = %d, want %d", claims.SessionID, sessionID)
	}
}

func TestRefresh_AccessTokenBoundToSession(t *testing.T) {
	ctx := context.Background()

	user := UserFixture()

	tokenSvc := NewTokenService(&config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "test-access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "test-refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	})

	refreshToken, _, err := tokenSvc.IssueUserRefreshToken(user)
	if err != nil {
		t.Fatal(err)
	}

	const sessionID = uint64(99)

	authSvc := AuthService{
		userDAO: UserDAOMock{DAOMock: DAOMock[User]{FindFunc: func(ctx context.Context, id uint64) (*User, error) {
			return user, nil
		}}},
		userSessionDAO: UserSessionDAOMock{RotateRefreshTokenFunc: func(ctx context.Context, oldToken string, newSession *UserSession) (*UserSession, error) {
			newSession.ID = sessionID
			return newSession, nil
		}},
		tokenSvc: tokenSvc,
	}

	result, err := authSvc.Refresh(ctx, refreshToken, SessionClientInfo{})
	if err != nil {
		t.Fatal(err)
	}

	claims, err := tokenSvc.VerifyUserAccessToken(result.AccessToken)
	if err != nil {
		t.Fatal(err)
	}

	if claims.SessionID != sessionID {
		t.Errorf("access token session id = %d, want %d", claims.SessionID, sessionID)
	}
}
