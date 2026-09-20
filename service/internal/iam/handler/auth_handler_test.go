package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/queue"
)

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		users      iam.UserDAOMock
		sessions   iam.UserSessionDAOMock
		body       string
		wantStatus int
	}{
		{
			name: "returns tokens",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return sampleUserWithPassword(), nil
				},
			},
			sessions: iam.UserSessionDAOMock{
				DAOMock: iam.DAOMock[iam.UserSession]{
					CreateFunc: func(_ context.Context, entity *iam.UserSession) (*iam.UserSession, error) {
						entity.ID = 1
						return entity, nil
					},
				},
			},
			body:       `{"email":"jane@example.com","password":"password123"}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "rejects invalid credentials",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return nil, nil
				},
			},
			body:       `{"email":"jane@example.com","password":"password123"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "rejects validation",
			users:      iam.UserDAOMock{},
			body:       `{"email":"not-an-email","password":""}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return nil, errors.New("db down")
				},
			},
			body:       `{"email":"jane@example.com","password":"password123"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := authTestService(tt.users, iam.UserSessionDAOMock{}, iam.UserEmailVerificationDAOMock{}, iam.UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{})
			app := authHandlerTest(t, svc, tt.users, iam.UserSessionDAOMock{})
			resp, err := doRequest(app, http.MethodPost, "/auth/login", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name       string
		users      iam.UserDAOMock
		body       string
		wantStatus int
	}{
		{
			name: "creates user",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					CreateFunc: func(_ context.Context, entity *iam.User) (*iam.User, error) {
						entity.ID = 1
						return entity, nil
					},
				},
			},
			body:       `{"first_name":"Jane","last_name":"Doe","email":"jane@example.com","password":"password123"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name: "rejects email registered",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return sampleUser(), nil
				},
			},
			body:       `{"first_name":"Jane","email":"jane@example.com","password":"password123"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "rejects validation",
			users:      iam.UserDAOMock{},
			body:       `{"first_name":"","email":"jane@example.com","password":"short"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return nil, errors.New("db down")
				},
			},
			body:       `{"first_name":"Jane","email":"jane@example.com","password":"password123"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := authTestService(tt.users, iam.UserSessionDAOMock{}, iam.UserEmailVerificationDAOMock{}, iam.UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{})
			app := authHandlerTest(t, svc, tt.users, iam.UserSessionDAOMock{})
			resp, err := doRequest(app, http.MethodPost, "/auth/register", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	tests := []struct {
		name       string
		users      iam.UserDAOMock
		sessions   iam.UserSessionDAOMock
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name: "returns tokens",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
				},
			},
			sessions: iam.UserSessionDAOMock{
				RotateRefreshTokenFunc: func(_ context.Context, _ string, newSession *iam.UserSession) (*iam.UserSession, error) {
					newSession.ID = 2
					return newSession, nil
				},
			},
			body:       `{"refresh_token":"` + validRefreshToken() + `"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "rejects invalid token",
			users:      iam.UserDAOMock{},
			sessions:   iam.UserSessionDAOMock{},
			body:       `{"refresh_token":"garbage"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "rejects reused token",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
				},
			},
			sessions: iam.UserSessionDAOMock{
				RotateRefreshTokenFunc: func(_ context.Context, _ string, _ *iam.UserSession) (*iam.UserSession, error) {
					return nil, iam.ErrTokenReused
				},
			},
			body:       `{"refresh_token":"` + validRefreshToken() + `"}`,
			wantStatus: http.StatusUnauthorized,
			wantCode:   httpx.ErrTokenReusedCode,
		},
		{
			name: "rejects token not found",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
				},
			},
			sessions: iam.UserSessionDAOMock{
				RotateRefreshTokenFunc: func(_ context.Context, _ string, _ *iam.UserSession) (*iam.UserSession, error) {
					return nil, iam.ErrTokenNotFound
				},
			},
			body:       `{"refresh_token":"` + validRefreshToken() + `"}`,
			wantStatus: http.StatusUnauthorized,
			wantCode:   httpx.ErrUnauthorizedCode,
		},
		{
			name:       "rejects validation",
			users:      iam.UserDAOMock{},
			sessions:   iam.UserSessionDAOMock{},
			body:       `{"refresh_token":""}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return nil, errors.New("db down")
					},
				},
			},
			sessions:   iam.UserSessionDAOMock{},
			body:       `{"refresh_token":"` + validRefreshToken() + `"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := authTestService(tt.users, tt.sessions, iam.UserEmailVerificationDAOMock{}, iam.UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{})
			app := authHandlerTest(t, svc, tt.users, tt.sessions)
			resp, err := doRequest(app, http.MethodPost, "/auth/refresh", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if tt.wantCode != "" {
				var out struct {
					ErrorCode string `json:"error_code"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
					t.Fatal(err)
				}
				if out.ErrorCode != tt.wantCode {
					t.Fatalf("error_code = %q, want %q", out.ErrorCode, tt.wantCode)
				}
			}
		})
	}
}

func TestAuthHandler_RequestEmailVerification(t *testing.T) {
	tests := []struct {
		name               string
		users              iam.UserDAOMock
		emailVerifications iam.UserEmailVerificationDAOMock
		queueClient        queue.TaskEnqueuer
		body               string
		wantStatus         int
	}{
		{
			name: "accepted",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return sampleUser(), nil
				},
			},
			emailVerifications: iam.UserEmailVerificationDAOMock{
				DAOMock: iam.DAOMock[iam.UserEmailVerification]{
					CreateFunc: func(_ context.Context, entity *iam.UserEmailVerification) (*iam.UserEmailVerification, error) {
						return entity, nil
					},
				},
				DeleteByUserIDFunc: func(_ context.Context, _ uint64) error {
					return nil
				},
			},
			queueClient: queue.TaskEnqueuerMock{
				EnqueueFunc: func(_ *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
					return &asynq.TaskInfo{}, nil
				},
			},
			body:       `{"email":"jane@example.com"}`,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "rejects validation",
			body:       `{"email":"not-an-email"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return nil, errors.New("db down")
				},
			},
			body:       `{"email":"jane@example.com"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := authTestService(tt.users, iam.UserSessionDAOMock{}, tt.emailVerifications, iam.UserPasswordResetDAOMock{}, tt.queueClient)
			app := authHandlerTest(t, svc, tt.users, iam.UserSessionDAOMock{})
			resp, err := doRequest(app, http.MethodPost, "/auth/email-verification/request", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_VerifyEmail(t *testing.T) {
	tests := []struct {
		name               string
		users              iam.UserDAOMock
		emailVerifications iam.UserEmailVerificationDAOMock
		body               string
		wantStatus         int
	}{
		{
			name: "verifies",
			emailVerifications: iam.UserEmailVerificationDAOMock{
				DAOMock: iam.DAOMock[iam.UserEmailVerification]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.UserEmailVerification, error) {
						return &iam.UserEmailVerification{Base: model.Base{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}, nil
					},
				},
				FindByTokenFunc: func(_ context.Context, _ string) (*iam.UserEmailVerification, error) {
					return &iam.UserEmailVerification{Base: model.Base{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				MarkUsedFunc: func(_ context.Context, _ uint64) error {
					return nil
				},
			},
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
					UpdateFunc: func(_ context.Context, entity *iam.User) (*iam.User, error) {
						return entity, nil
					},
				},
			},
			body:       `{"token":"abc123"}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "rejects invalid token",
			emailVerifications: iam.UserEmailVerificationDAOMock{
				FindByTokenFunc: func(_ context.Context, _ string) (*iam.UserEmailVerification, error) {
					return nil, nil
				},
			},
			body:       `{"token":"abc123"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "rejects expired token",
			emailVerifications: iam.UserEmailVerificationDAOMock{
				FindByTokenFunc: func(_ context.Context, _ string) (*iam.UserEmailVerification, error) {
					return &iam.UserEmailVerification{Base: model.Base{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(-time.Hour)}, nil
				},
			},
			body:       `{"token":"abc123"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects validation",
			body:       `{"token":""}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			emailVerifications: iam.UserEmailVerificationDAOMock{
				FindByTokenFunc: func(_ context.Context, _ string) (*iam.UserEmailVerification, error) {
					return nil, errors.New("db down")
				},
			},
			body:       `{"token":"abc123"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "returns server error on update",
			emailVerifications: iam.UserEmailVerificationDAOMock{
				FindByTokenFunc: func(_ context.Context, _ string) (*iam.UserEmailVerification, error) {
					return &iam.UserEmailVerification{Base: model.Base{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
			},
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
					UpdateFunc: func(_ context.Context, _ *iam.User) (*iam.User, error) {
						return nil, errors.New("db down")
					},
				},
			},
			body:       `{"token":"abc123"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := authTestService(tt.users, iam.UserSessionDAOMock{}, tt.emailVerifications, iam.UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{})
			app := authHandlerTest(t, svc, tt.users, iam.UserSessionDAOMock{})
			resp, err := doRequest(app, http.MethodPost, "/auth/email-verification/verify", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_RequestPasswordReset(t *testing.T) {
	tests := []struct {
		name           string
		users          iam.UserDAOMock
		passwordResets iam.UserPasswordResetDAOMock
		queueClient    queue.TaskEnqueuer
		body           string
		wantStatus     int
	}{
		{
			name: "accepted",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return sampleUser(), nil
				},
			},
			passwordResets: iam.UserPasswordResetDAOMock{
				DAOMock: iam.DAOMock[iam.UserPasswordReset]{
					CreateFunc: func(_ context.Context, entity *iam.UserPasswordReset) (*iam.UserPasswordReset, error) {
						return entity, nil
					},
				},
				DeleteByUserIDFunc: func(_ context.Context, _ uint64) error {
					return nil
				},
			},
			queueClient: queue.TaskEnqueuerMock{
				EnqueueFunc: func(_ *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
					return &asynq.TaskInfo{}, nil
				},
			},
			body:       `{"email":"jane@example.com"}`,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "rejects validation",
			body:       `{"email":"not-an-email"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return nil, errors.New("db down")
				},
			},
			body:       `{"email":"jane@example.com"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := authTestService(tt.users, iam.UserSessionDAOMock{}, iam.UserEmailVerificationDAOMock{}, tt.passwordResets, tt.queueClient)
			app := authHandlerTest(t, svc, tt.users, iam.UserSessionDAOMock{})
			resp, err := doRequest(app, http.MethodPost, "/auth/password-reset/request", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_ResetPassword(t *testing.T) {
	tests := []struct {
		name           string
		users          iam.UserDAOMock
		sessions       iam.UserSessionDAOMock
		passwordResets iam.UserPasswordResetDAOMock
		body           string
		wantStatus     int
	}{
		{
			name: "resets",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
					UpdateFunc: func(_ context.Context, entity *iam.User) (*iam.User, error) {
						return entity, nil
					},
				},
			},
			sessions: iam.UserSessionDAOMock{
				RevokeByUserIDFunc: func(_ context.Context, _ uint64) error {
					return nil
				},
			},
			passwordResets: iam.UserPasswordResetDAOMock{
				DAOMock: iam.DAOMock[iam.UserPasswordReset]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.UserPasswordReset, error) {
						return &iam.UserPasswordReset{Base: model.Base{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}, nil
					},
				},
				FindByTokenHashFunc: func(_ context.Context, _ string) (*iam.UserPasswordReset, error) {
					return &iam.UserPasswordReset{Base: model.Base{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				MarkUsedFunc: func(_ context.Context, _ uint64) error {
					return nil
				},
			},
			body:       `{"token":"abc123","password":"newpassword"}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "rejects invalid token",
			passwordResets: iam.UserPasswordResetDAOMock{
				FindByTokenHashFunc: func(_ context.Context, _ string) (*iam.UserPasswordReset, error) {
					return nil, nil
				},
			},
			body:       `{"token":"abc123","password":"newpassword"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "rejects expired token",
			passwordResets: iam.UserPasswordResetDAOMock{
				FindByTokenHashFunc: func(_ context.Context, _ string) (*iam.UserPasswordReset, error) {
					return &iam.UserPasswordReset{Base: model.Base{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(-time.Hour)}, nil
				},
			},
			body:       `{"token":"abc123","password":"newpassword"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects validation",
			body:       `{"token":"abc123","password":"short"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			passwordResets: iam.UserPasswordResetDAOMock{
				FindByTokenHashFunc: func(_ context.Context, _ string) (*iam.UserPasswordReset, error) {
					return nil, errors.New("db down")
				},
			},
			body:       `{"token":"abc123","password":"newpassword"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "returns server error on revoke",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
					UpdateFunc: func(_ context.Context, entity *iam.User) (*iam.User, error) {
						return entity, nil
					},
				},
			},
			sessions: iam.UserSessionDAOMock{
				RevokeByUserIDFunc: func(_ context.Context, _ uint64) error {
					return errors.New("db down")
				},
			},
			passwordResets: iam.UserPasswordResetDAOMock{
				FindByTokenHashFunc: func(_ context.Context, _ string) (*iam.UserPasswordReset, error) {
					return &iam.UserPasswordReset{Base: model.Base{ID: 1}, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}, nil
				},
				MarkUsedFunc: func(_ context.Context, _ uint64) error {
					return nil
				},
			},
			body:       `{"token":"abc123","password":"newpassword"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := authTestService(tt.users, tt.sessions, iam.UserEmailVerificationDAOMock{}, tt.passwordResets, queue.TaskEnqueuerMock{})
			app := authHandlerTest(t, svc, tt.users, tt.sessions)
			resp, err := doRequest(app, http.MethodPost, "/auth/password-reset", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_GetSessions(t *testing.T) {
	tests := []struct {
		name       string
		sessions   iam.UserSessionDAOMock
		noCaller   bool
		wantStatus int
	}{
		{
			name: "returns sessions",
			sessions: iam.UserSessionDAOMock{
				ListUserSessionsFunc: func(_ context.Context, _ uint64) ([]*iam.UserSession, error) {
					return []*iam.UserSession{sampleSession()}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns unauthorized",
			sessions:   iam.UserSessionDAOMock{},
			noCaller:   true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns server error",
			sessions: iam.UserSessionDAOMock{
				ListUserSessionsFunc: func(_ context.Context, _ uint64) ([]*iam.UserSession, error) {
					return nil, errors.New("db down")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := authTestService(iam.UserDAOMock{}, tt.sessions, iam.UserEmailVerificationDAOMock{}, iam.UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{})
			var app *fiber.App
			if tt.noCaller {
				app = authHandlerTestNoCaller(t, svc, iam.UserDAOMock{}, tt.sessions)
			} else {
				app = authHandlerTest(t, svc, iam.UserDAOMock{}, tt.sessions)
			}
			resp, err := doRequest(app, http.MethodGet, "/auth/sessions", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	tests := []struct {
		name       string
		sessions   iam.UserSessionDAOMock
		body       string
		wantStatus int
	}{
		{
			name: "logs out",
			sessions: iam.UserSessionDAOMock{
				FindByRefreshTokenFunc: func(_ context.Context, _ string) (*iam.UserSession, error) {
					return sampleSession(), nil
				},
				DAOMock: iam.DAOMock[iam.UserSession]{
					UpdateFunc: func(_ context.Context, entity *iam.UserSession) (*iam.UserSession, error) {
						return entity, nil
					},
				},
			},
			body:       `{"refresh_token":"abc"}`,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "rejects validation",
			sessions:   iam.UserSessionDAOMock{},
			body:       `{"refresh_token":""}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			sessions: iam.UserSessionDAOMock{
				FindByRefreshTokenFunc: func(_ context.Context, _ string) (*iam.UserSession, error) {
					return nil, errors.New("db down")
				},
			},
			body:       `{"refresh_token":"abc"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := authTestService(iam.UserDAOMock{}, tt.sessions, iam.UserEmailVerificationDAOMock{}, iam.UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{})
			app := authHandlerTest(t, svc, iam.UserDAOMock{}, tt.sessions)
			resp, err := doRequest(app, http.MethodPost, "/auth/logout", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_RevokeSession(t *testing.T) {
	tests := []struct {
		name       string
		sessions   iam.UserSessionDAOMock
		noCaller   bool
		path       string
		wantStatus int
	}{
		{
			name: "revokes",
			sessions: iam.UserSessionDAOMock{
				DAOMock: iam.DAOMock[iam.UserSession]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.UserSession, error) {
						return sampleSession(), nil
					},
					UpdateFunc: func(_ context.Context, entity *iam.UserSession) (*iam.UserSession, error) {
						return entity, nil
					},
				},
			},
			path:       "/auth/sessions/1",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "rejects invalid id",
			sessions:   iam.UserSessionDAOMock{},
			path:       "/auth/sessions/abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns unauthorized",
			sessions:   iam.UserSessionDAOMock{},
			noCaller:   true,
			path:       "/auth/sessions/1",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns not found",
			sessions: iam.UserSessionDAOMock{
				DAOMock: iam.DAOMock[iam.UserSession]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.UserSession, error) {
						return nil, nil
					},
				},
			},
			path:       "/auth/sessions/1",
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns unauthorized when not owned",
			sessions: iam.UserSessionDAOMock{
				DAOMock: iam.DAOMock[iam.UserSession]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.UserSession, error) {
						session := sampleSession()
						session.UserID = 99
						return session, nil
					},
				},
			},
			path:       "/auth/sessions/1",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns server error",
			sessions: iam.UserSessionDAOMock{
				DAOMock: iam.DAOMock[iam.UserSession]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.UserSession, error) {
						return nil, errors.New("db down")
					},
				},
			},
			path:       "/auth/sessions/1",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var app *fiber.App
			svc := authTestService(iam.UserDAOMock{}, tt.sessions, iam.UserEmailVerificationDAOMock{}, iam.UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{})
			if tt.noCaller {
				app = authHandlerTestNoCaller(t, svc, iam.UserDAOMock{}, tt.sessions)
			} else {
				app = authHandlerTest(t, svc, iam.UserDAOMock{}, tt.sessions)
			}
			resp, err := doRequest(app, http.MethodDelete, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}
