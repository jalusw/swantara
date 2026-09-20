package iam

import (
	"context"
	"strings"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
)

type AuthService struct {
	userDAO              UserDAO
	userSvc              UserService
	userSessionDAO       UserSessionDAO
	passwordSvc          PasswordService
	tokenSvc             TokenService
	emailVerificationDAO UserEmailVerificationDAO
	passwordResetDAO     UserPasswordResetDAO
	queueClient          queue.TaskEnqueuer
	emailVerifyTTL       time.Duration
	passwordResetTTL     time.Duration
}

func NewAuthService(
	userDAO UserDAO,
	userSvc UserService,
	userSessionDAO UserSessionDAO,
	passwordSvc PasswordService,
	tokenSvc TokenService,
	emailVerificationDAO UserEmailVerificationDAO,
	passwordResetDAO UserPasswordResetDAO,
	queueClient queue.TaskEnqueuer,
	emailVerifyTTL time.Duration,
	passwordResetTTL time.Duration,
) AuthService {
	return AuthService{
		userDAO:              userDAO,
		userSvc:              userSvc,
		userSessionDAO:       userSessionDAO,
		passwordSvc:          passwordSvc,
		tokenSvc:             tokenSvc,
		emailVerificationDAO: emailVerificationDAO,
		passwordResetDAO:     passwordResetDAO,
		queueClient:          queueClient,
		emailVerifyTTL:       emailVerifyTTL,
		passwordResetTTL:     passwordResetTTL,
	}
}

type LoginResult struct {
	UserID                uint64
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

func (s AuthService) Login(ctx context.Context, email, password string, client SessionClientInfo) (result *LoginResult, err error) {
	defer helper.RecordOp(ctx, "auth.login", time.Now(), &err)

	user, err := s.userDAO.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !user.Active {
		return nil, ErrInvalidCredentials
	}

	if err := s.passwordSvc.Verify(user.Password, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	refreshToken, refreshTokenExpiresAt, err := s.tokenSvc.IssueUserRefreshToken(user)
	if err != nil {
		return nil, err
	}

	session, err := s.userSessionDAO.Create(ctx, sessionModel(client, user.ID, refreshToken, *refreshTokenExpiresAt))
	if err != nil {
		return nil, err
	}

	accessToken, accessTokenExpiresAt, err := s.tokenSvc.IssueUserAccessToken(user, session.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID:                user.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  *accessTokenExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: *refreshTokenExpiresAt,
	}, nil
}

type RegisterData struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

func (s AuthService) Register(ctx context.Context, registrationData RegisterData) (result *User, err error) {
	defer helper.RecordOp(ctx, "auth.register", time.Now(), &err)

	existingUser, err := s.userDAO.FindByEmail(ctx, registrationData.Email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil && existingUser.ID != 0 {
		return nil, ErrEmailRegistered
	}

	hashedPassword, err := s.passwordSvc.Hash(registrationData.Password)
	if err != nil {
		return nil, err
	}

	username, err := s.generateUsername(ctx, registrationData.Email)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:  username,
		FirstName: registrationData.FirstName,
		LastName:  &registrationData.LastName,
		Email:     registrationData.Email,
		Password:  hashedPassword,
		Active:    true,
	}

	return s.userDAO.Create(ctx, user)
}

func (s AuthService) generateUsername(ctx context.Context, email string) (string, error) {
	base := usernameFromEmail(email)

	taken, err := s.userSvc.IsUsernameTaken(ctx, base)
	if err != nil {
		return "", err
	}

	if taken {
		suffix, err := helper.RandomHex(4)
		if err != nil {
			return "", err
		}
		base = base + "_" + suffix
	}

	return base, nil
}

func usernameFromEmail(email string) string {
	local := strings.ToLower(strings.Split(email, "@")[0])

	var builder strings.Builder
	for _, r := range local {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			builder.WriteRune(r)
		}
	}

	if builder.Len() == 0 {
		return "user"
	}
	return builder.String()
}

type RefreshResult = LoginResult

func (s AuthService) Refresh(ctx context.Context, refreshToken string, client SessionClientInfo) (result *RefreshResult, err error) {
	defer helper.RecordOp(ctx, "auth.refresh", time.Now(), &err)

	claims, err := s.tokenSvc.VerifyUserRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userDAO.Find(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidRefreshToken
	}

	if !user.Active {
		return nil, ErrInvalidRefreshToken
	}

	newRefreshToken, newRefreshTokenExpiresAt, err := s.tokenSvc.IssueUserRefreshToken(user)
	if err != nil {
		return nil, err
	}

	newSession, err := s.userSessionDAO.RotateRefreshToken(ctx, helper.HashToken(refreshToken), sessionModel(client, user.ID, newRefreshToken, *newRefreshTokenExpiresAt))
	if err != nil {
		return nil, err
	}

	accessToken, accessTokenExpiresAt, err := s.tokenSvc.IssueUserAccessToken(user, newSession.ID)
	if err != nil {
		return nil, err
	}

	return &RefreshResult{
		UserID:                user.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  *accessTokenExpiresAt,
		RefreshToken:          newRefreshToken,
		RefreshTokenExpiresAt: *newRefreshTokenExpiresAt,
	}, nil
}

func (s AuthService) RequestEmailVerification(ctx context.Context, email string) (err error) {
	defer helper.RecordOp(ctx, "auth.request_email_verification", time.Now(), &err)

	user, err := s.userDAO.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user == nil || user.ID == 0 || user.EmailVerifiedAt != nil {
		return nil
	}

	token, err := helper.RandomHex(32)
	if err != nil {
		return err
	}

	if err := s.emailVerificationDAO.DeleteByUserID(ctx, user.ID); err != nil {
		return err
	}

	if _, err := s.emailVerificationDAO.Create(ctx, &UserEmailVerification{
		UserID:            user.ID,
		VerificationToken: helper.HashToken(token),
		ExpiresAt:         time.Now().Add(s.emailVerifyTTL),
	}); err != nil {
		return err
	}

	task, err := tasks.NewSendEmailVerificationTask(user.ID, token, helper.RequestID(ctx))
	if err != nil {
		return err
	}

	if _, err := s.queueClient.Enqueue(task); err != nil {
		return err
	}

	return nil
}

func (s AuthService) VerifyEmail(ctx context.Context, token string) (err error) {
	defer helper.RecordOp(ctx, "auth.verify_email", time.Now(), &err)

	verification, err := s.emailVerificationDAO.FindByToken(ctx, helper.HashToken(token))
	if err != nil {
		return err
	}

	if verification == nil || verification.ID == 0 || verification.UsedAt != nil {
		return ErrInvalidActionToken
	}

	if time.Now().After(verification.ExpiresAt) {
		return ErrActionTokenExpired
	}

	user, err := s.userDAO.Find(ctx, verification.UserID)
	if err != nil {
		return err
	}

	if user == nil || user.ID == 0 {
		return ErrInvalidActionToken
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	if _, err := s.userDAO.Update(ctx, user); err != nil {
		return err
	}

	return s.emailVerificationDAO.MarkUsed(ctx, verification.ID)
}

func (s AuthService) RequestPasswordReset(ctx context.Context, email string) (err error) {
	defer helper.RecordOp(ctx, "auth.request_password_reset", time.Now(), &err)

	user, err := s.userDAO.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user == nil || user.ID == 0 {
		return nil
	}

	token, err := helper.RandomHex(32)
	if err != nil {
		return err
	}

	if err := s.passwordResetDAO.DeleteByUserID(ctx, user.ID); err != nil {
		return err
	}

	if _, err := s.passwordResetDAO.Create(ctx, &UserPasswordReset{
		UserID:         user.ID,
		ResetTokenHash: helper.HashToken(token),
		ExpiresAt:      time.Now().Add(s.passwordResetTTL),
	}); err != nil {
		return err
	}

	task, err := tasks.NewSendPasswordResetTask(user.ID, token, helper.RequestID(ctx))
	if err != nil {
		return err
	}

	if _, err := s.queueClient.Enqueue(task); err != nil {
		return err
	}

	return nil
}

func (s AuthService) ResetPassword(ctx context.Context, token, newPassword string) (err error) {
	defer helper.RecordOp(ctx, "auth.reset_password", time.Now(), &err)

	reset, err := s.passwordResetDAO.FindByTokenHash(ctx, helper.HashToken(token))
	if err != nil {
		return err
	}

	if reset == nil || reset.ID == 0 || reset.UsedAt != nil {
		return ErrInvalidActionToken
	}

	if time.Now().After(reset.ExpiresAt) {
		return ErrActionTokenExpired
	}

	user, err := s.userDAO.Find(ctx, reset.UserID)
	if err != nil {
		return err
	}

	if user == nil || user.ID == 0 {
		return ErrInvalidActionToken
	}

	hashed, err := s.passwordSvc.Hash(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashed
	if _, err := s.userDAO.Update(ctx, user); err != nil {
		return err
	}

	if err := s.passwordResetDAO.MarkUsed(ctx, reset.ID); err != nil {
		return err
	}

	return s.userSessionDAO.RevokeByUserID(ctx, user.ID)
}
