package iam

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestUserDAO_FindByEmail(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, user *User)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 LIMIT $2`)).
					WithArgs("john@test.com", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "john@test.com"))
			},
			wantErr: false,
			validate: func(t *testing.T, user *User) {
				if user == nil || user.Email != "john@test.com" {
					t.Errorf("user = %+v, want email john@test.com", user)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 LIMIT $2`)).
					WithArgs("john@test.com", 1).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			userDAO := NewUserDAO(db)

			user, err := userDAO.FindByEmail(ctx, "john@test.com")
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, user)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserDAO_FindByUsername(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, user *User)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 LIMIT $2`)).
					WithArgs("johndoe", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "johndoe"))
			},
			wantErr: false,
			validate: func(t *testing.T, user *User) {
				if user == nil || user.Username != "johndoe" {
					t.Errorf("user = %+v, want username johndoe", user)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			userDAO := NewUserDAO(db)

			user, err := userDAO.FindByUsername(ctx, "johndoe")
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, user)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserDAO_FindByPhone(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, user *User)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE phone = $1 LIMIT $2`)).
					WithArgs("0812", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "phone"}).AddRow(1, "0812"))
			},
			wantErr: false,
			validate: func(t *testing.T, user *User) {
				if user == nil || user.Phone == nil || *user.Phone != "0812" {
					t.Errorf("user = %+v, want phone 0812", user)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			userDAO := NewUserDAO(db)

			user, err := userDAO.FindByPhone(ctx, "0812")
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, user)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserSessionDAO_FindByRefreshToken(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, session *UserSession)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE refresh_token = $1`)).
					WithArgs("token").
					WillReturnRows(sqlmock.NewRows([]string{"id", "refresh_token"}).AddRow(1, "token"))
			},
			wantErr: false,
			validate: func(t *testing.T, session *UserSession) {
				if session == nil || session.ID != 1 {
					t.Errorf("session = %+v, want id 1", session)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE refresh_token = $1`)).
					WithArgs("token").
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			sessions := NewUserSessionDAO(db)

			session, err := sessions.FindByRefreshToken(ctx, "token")
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, session)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserSessionDAO_ListUserSessions(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, items []*UserSession)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE user_id = $1`)).
					WithArgs(3).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, 3).AddRow(2, 3))
			},
			wantErr: false,
			validate: func(t *testing.T, items []*UserSession) {
				if len(items) != 2 || items[0].ID != 1 || items[1].ID != 2 {
					t.Errorf("items = %+v, want two sessions", items)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE user_id = $1`)).
					WithArgs(3).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			sessions := NewUserSessionDAO(db)

			items, err := sessions.ListUserSessions(ctx, 3)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, items)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserSessionDAO_RevokeByUserID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "user_sessions" SET "revoked_at"=$1`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 3).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "user_sessions" SET "revoked_at"=$1`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 3).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			sessions := NewUserSessionDAO(db)

			err := sessions.RevokeByUserID(ctx, 3)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserSessionDAO_RotateRefreshToken(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, rotated *UserSession)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE refresh_token = $1 ORDER BY "user_sessions"."id" LIMIT $2 FOR UPDATE`)).
					WithArgs("old", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "refresh_token"}).AddRow(1, "old"))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "user_sessions" SET "created_at"=$1`)).
					WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "user_sessions"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				mock.ExpectCommit()
			},
			wantErr: false,
			validate: func(t *testing.T, rotated *UserSession) {
				if rotated.UserID != 3 {
					t.Errorf("rotated = %+v, want user id 3", rotated)
				}
			},
		},
		{
			name: "token not found",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE refresh_token = $1 ORDER BY "user_sessions"."id" LIMIT $2 FOR UPDATE`)).
					WithArgs("old", 1).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			wantErr:    true,
			wantErrVal: ErrTokenNotFound,
		},
		{
			name: "lock error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE refresh_token = $1 ORDER BY "user_sessions"."id" LIMIT $2 FOR UPDATE`)).
					WithArgs("old", 1).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "token reused",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE refresh_token = $1 ORDER BY "user_sessions"."id" LIMIT $2 FOR UPDATE`)).
					WithArgs("old", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "refresh_token", "revoked_at"}).AddRow(1, "old", time.Now().Add(-2*time.Minute)))
				mock.ExpectRollback()
			},
			wantErr:    true,
			wantErrVal: ErrTokenReused,
		},
		{
			name: "token reused within grace creates new session",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE refresh_token = $1 ORDER BY "user_sessions"."id" LIMIT $2 FOR UPDATE`)).
					WithArgs("old", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "refresh_token", "revoked_at"}).AddRow(1, "old", time.Now()))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "user_sessions"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				mock.ExpectCommit()
			},
			wantErr: false,
			validate: func(t *testing.T, rotated *UserSession) {
				if rotated.UserID != 3 {
					t.Errorf("rotated = %+v, want user id 3", rotated)
				}
			},
		},
		{
			name: "save error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE refresh_token = $1 ORDER BY "user_sessions"."id" LIMIT $2 FOR UPDATE`)).
					WithArgs("old", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "refresh_token"}).AddRow(1, "old"))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "user_sessions" SET "created_at"=$1`)).
					WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "create error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions" WHERE refresh_token = $1 ORDER BY "user_sessions"."id" LIMIT $2 FOR UPDATE`)).
					WithArgs("old", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "refresh_token"}).AddRow(1, "old"))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "user_sessions" SET "created_at"=$1`)).
					WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "user_sessions"`)).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			sessions := NewUserSessionDAO(db)

			rotated, err := sessions.RotateRefreshToken(ctx, "old", &UserSession{UserID: 3})
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, rotated)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserEmailVerificationDAO_FindByToken(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, entity *UserEmailVerification)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_email_verifications" WHERE verification_token = $1`)).
					WithArgs("hash").
					WillReturnRows(sqlmock.NewRows([]string{"id", "verification_token"}).AddRow(1, "hash"))
			},
			wantErr: false,
			validate: func(t *testing.T, entity *UserEmailVerification) {
				if entity == nil || entity.ID != 1 {
					t.Errorf("entity = %+v, want id 1", entity)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_email_verifications" WHERE verification_token = $1`)).
					WithArgs("hash").
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			verifications := NewUserEmailVerificationDAO(db)

			entity, err := verifications.FindByToken(ctx, "hash")
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, entity)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserEmailVerificationDAO_MarkUsed(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "user_email_verifications" SET "used_at"=$1`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 9).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "user_email_verifications" SET "used_at"=$1`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 9).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			verifications := NewUserEmailVerificationDAO(db)

			err := verifications.MarkUsed(ctx, 9)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserEmailVerificationDAO_DeleteByUserID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "user_email_verifications" WHERE user_id = $1`)).
					WithArgs(9).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "user_email_verifications" WHERE user_id = $1`)).
					WithArgs(9).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			verifications := NewUserEmailVerificationDAO(db)

			err := verifications.DeleteByUserID(ctx, 9)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserPasswordResetDAO_FindByTokenHash(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, entity *UserPasswordReset)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_password_resets" WHERE reset_token_hash = $1`)).
					WithArgs("hash").
					WillReturnRows(sqlmock.NewRows([]string{"id", "reset_token_hash"}).AddRow(1, "hash"))
			},
			wantErr: false,
			validate: func(t *testing.T, entity *UserPasswordReset) {
				if entity == nil || entity.ID != 1 {
					t.Errorf("entity = %+v, want id 1", entity)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_password_resets" WHERE reset_token_hash = $1`)).
					WithArgs("hash").
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			resets := NewUserPasswordResetDAO(db)

			entity, err := resets.FindByTokenHash(ctx, "hash")
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, entity)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserPasswordResetDAO_MarkUsed(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "user_password_resets" SET "used_at"=$1`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 9).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			resets := NewUserPasswordResetDAO(db)

			err := resets.MarkUsed(ctx, 9)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUserPasswordResetDAO_DeleteByUserID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "user_password_resets" WHERE user_id = $1`)).
					WithArgs(9).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			resets := NewUserPasswordResetDAO(db)

			err := resets.DeleteByUserID(ctx, 9)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestPermissionDAO_FindByCode(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, permission *Permission)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions" WHERE code = $1`)).
					WithArgs("user.view").
					WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "user.view"))
			},
			wantErr: false,
			validate: func(t *testing.T, permission *Permission) {
				if permission == nil || permission.Code != "user.view" {
					t.Errorf("permission = %+v, want code user.view", permission)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			permissions := NewPermissionDAO(db)

			permission, err := permissions.FindByCode(ctx, "user.view")
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, permission)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestPermissionDAO_UserHasOrganizationPermission(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, allowed bool)
	}{
		{
			name: "allowed",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "members" JOIN member_role_assignments ON member_role_assignments.member_id = members.id JOIN member_roles ON member_roles.id = member_role_assignments.role_id JOIN member_role_permissions ON member_role_permissions.member_role_id = member_roles.id JOIN permissions ON permissions.id = member_role_permissions.permission_id WHERE members.user_id = $1 AND members.organization_id = $2 AND member_roles.organization_id = $3 AND permissions.resource = $4 AND permissions.action = $5`)).
					WithArgs(7, 3, 3, "user", "view").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			wantErr: false,
			validate: func(t *testing.T, allowed bool) {
				if !allowed {
					t.Error("expected allowed, got false")
				}
			},
		},
		{
			name: "denied",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "members"`)).
					WithArgs(7, 3, 3, "user", "view").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			wantErr: false,
			validate: func(t *testing.T, allowed bool) {
				if allowed {
					t.Error("expected denied, got true")
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "members"`)).
					WithArgs(7, 3, 3, "user", "view").
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			permissions := NewPermissionDAO(db)

			allowed, err := permissions.UserHasOrganizationPermission(ctx, 7, 3, "user", "view")
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, allowed)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestPermissionDAO_UserIsOrganizationOwner(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, isOwner bool)
	}{
		{
			name: "owner",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "members" JOIN member_role_assignments ON member_role_assignments.member_id = members.id JOIN member_roles ON member_roles.id = member_role_assignments.role_id WHERE members.user_id = $1 AND members.organization_id = $2 AND member_roles.code = $3`)).
					WithArgs(7, 3, MemberRoleOwner).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			wantErr: false,
			validate: func(t *testing.T, isOwner bool) {
				if !isOwner {
					t.Error("expected owner, got false")
				}
			},
		},
		{
			name: "not owner",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "members"`)).
					WithArgs(7, 3, MemberRoleOwner).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			wantErr: false,
			validate: func(t *testing.T, isOwner bool) {
				if isOwner {
					t.Error("expected not owner, got true")
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "members"`)).
					WithArgs(7, 3, MemberRoleOwner).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			permissions := NewPermissionDAO(db)

			isOwner, err := permissions.UserIsOrganizationOwner(ctx, 7, 3)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, isOwner)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestPermissionDAO_ListUserOrganizationPermissions(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, items []*Permission)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT "permissions"."id","permissions"."created_at","permissions"."updated_at","permissions"."deleted_at","permissions"."name","permissions"."code","permissions"."description","permissions"."action","permissions"."resource" FROM "permissions" JOIN member_role_permissions ON member_role_permissions.permission_id = permissions.id JOIN member_roles ON member_roles.id = member_role_permissions.member_role_id JOIN member_role_assignments ON member_role_assignments.role_id = member_roles.id JOIN members ON members.id = member_role_assignments.member_id WHERE members.user_id = $1 AND members.organization_id = $2 AND member_roles.organization_id = $3`)).
					WithArgs(7, 3, 3).
					WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "user.view"))
			},
			wantErr: false,
			validate: func(t *testing.T, items []*Permission) {
				if len(items) != 1 || items[0].Code != "user.view" {
					t.Errorf("items = %+v, want one permission", items)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT "permissions"."id"`)).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			permissions := NewPermissionDAO(db)

			items, err := permissions.ListUserOrganizationPermissions(ctx, 7, 3)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, items)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestMemberDAO_ListOrganizationsByUser(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, orgs []*reference.Organization)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "members" WHERE user_id = $1`)).
					WithArgs(3).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "organization_id"}).AddRow(1, 3, 5))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE "organizations"."id" = $1`)).
					WithArgs(5).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "Acme"))
			},
			wantErr: false,
			validate: func(t *testing.T, orgs []*reference.Organization) {
				if len(orgs) != 1 || orgs[0].Name != "Acme" {
					t.Errorf("orgs = %+v, want one organization", orgs)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "members" WHERE user_id = $1`)).
					WithArgs(3).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			members := NewMemberDAO(db)

			orgs, err := members.ListOrganizationsByUser(ctx, 3)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, orgs)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestMemberDAO_ListByOrganization(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, items []*Member)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "members" WHERE organization_id = $1`)).
					WithArgs(5).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "organization_id"}).AddRow(1, 3, 5))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_role_assignments" WHERE "member_role_assignments"."member_id" = $1`)).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"member_id", "role_id"}).AddRow(1, 7))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_roles" WHERE "member_roles"."id" = $1`)).
					WithArgs(7).
					WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "code"}).AddRow(7, 5, "owner"))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
					WithArgs(3).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(3, "John"))
			},
			wantErr: false,
			validate: func(t *testing.T, items []*Member) {
				if len(items) != 1 || items[0].ID != 1 {
					t.Errorf("items = %+v, want one member", items)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "members" WHERE organization_id = $1`)).
					WithArgs(5).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			members := NewMemberDAO(db)

			items, err := members.ListByOrganization(ctx, 5)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, items)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestMemberDAO_FindByUserAndOrganization(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, member *Member)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "members" WHERE user_id = $1 AND organization_id = $2 ORDER BY "members"."id" LIMIT $3`)).
					WithArgs(3, 5, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "organization_id"}).AddRow(1, 3, 5))
			},
			wantErr: false,
			validate: func(t *testing.T, member *Member) {
				if member == nil || member.ID != 1 {
					t.Errorf("member = %+v, want id 1", member)
				}
			},
		},
		{
			name: "not found",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "members" WHERE user_id = $1 AND organization_id = $2 ORDER BY "members"."id" LIMIT $3`)).
					WithArgs(3, 5, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			validate: func(t *testing.T, member *Member) {
				if member != nil {
					t.Errorf("member = %+v, want nil", member)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "members" WHERE user_id = $1 AND organization_id = $2 ORDER BY "members"."id" LIMIT $3`)).
					WithArgs(3, 5, 1).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			members := NewMemberDAO(db)

			member, err := members.FindByUserAndOrganization(ctx, 3, 5)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, member)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestMemberDAO_AssignRole(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "member_role_assignments" ("member_id","role_id") VALUES ($1,$2)`)).
					WithArgs(3, 5).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "member_role_assignments" ("member_id","role_id") VALUES ($1,$2)`)).
					WithArgs(3, 5).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			members := NewMemberDAO(db)

			err := members.AssignRole(ctx, 3, 5)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestMemberRoleDAO_FindByOrganizationAndCode(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, role *MemberRole)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_roles" WHERE organization_id = $1 AND code = $2 ORDER BY "member_roles"."id" LIMIT $3`)).
					WithArgs(5, "owner", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "code"}).AddRow(1, 5, "owner"))
			},
			wantErr: false,
			validate: func(t *testing.T, role *MemberRole) {
				if role == nil || role.Code != "owner" {
					t.Errorf("role = %+v, want code owner", role)
				}
			},
		},
		{
			name: "not found",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_roles" WHERE organization_id = $1 AND code = $2 ORDER BY "member_roles"."id" LIMIT $3`)).
					WithArgs(5, "owner", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: false,
			validate: func(t *testing.T, role *MemberRole) {
				if role != nil {
					t.Errorf("role = %+v, want nil", role)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_roles" WHERE organization_id = $1 AND code = $2 ORDER BY "member_roles"."id" LIMIT $3`)).
					WithArgs(5, "owner", 1).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			roles := NewMemberRoleDAO(db)

			role, err := roles.FindByOrganizationAndCode(ctx, 5, "owner")
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, role)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestMemberRoleDAO_ListByOrganization(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
		validate   func(t *testing.T, items []*MemberRole)
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_roles" WHERE organization_id = $1`)).
					WithArgs(5).
					WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id"}).AddRow(1, 5))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_role_permissions" WHERE "member_role_permissions"."member_role_id" = $1`)).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"member_role_id", "permission_id"}).AddRow(1, 2))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions" WHERE "permissions"."id" = $1`)).
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(2, "user.view"))
			},
			wantErr: false,
			validate: func(t *testing.T, items []*MemberRole) {
				if len(items) != 1 || items[0].ID != 1 {
					t.Errorf("items = %+v, want one role", items)
				}
			},
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_roles" WHERE organization_id = $1`)).
					WithArgs(5).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			roles := NewMemberRoleDAO(db)

			items, err := roles.ListByOrganization(ctx, 5)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.validate != nil {
				tt.validate(t, items)
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestMemberRoleDAO_BindPermission(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "member_role_permissions" ("member_role_id","permission_id") VALUES ($1,$2)`)).
					WithArgs(3, 5).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "propagates error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "member_role_permissions" ("member_role_id","permission_id") VALUES ($1,$2)`)).
					WithArgs(3, 5).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			roles := NewMemberRoleDAO(db)

			err := roles.BindPermission(ctx, 3, 5)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestMemberRoleDAO_SetPermissions(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, mock sqlmock.Sqlmock)
		wantErr       bool
		wantErrVal    error
		permissionIDs []uint64
	}{
		{
			name: "success",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "member_role_permissions" WHERE member_role_id = $1`)).
					WithArgs(3).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "member_role_permissions" ("member_role_id","permission_id") VALUES ($1,$2)`)).
					WithArgs(3, 5).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "member_role_permissions" ("member_role_id","permission_id") VALUES ($1,$2)`)).
					WithArgs(3, 6).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr:       false,
			permissionIDs: []uint64{5, 6},
		},
		{
			name: "empty permissions",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "member_role_permissions" WHERE member_role_id = $1`)).
					WithArgs(3).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr:       false,
			permissionIDs: nil,
		},
		{
			name: "delete error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "member_role_permissions" WHERE member_role_id = $1`)).
					WithArgs(3).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr:       true,
			permissionIDs: []uint64{5},
		},
		{
			name: "bind error",
			setup: func(t *testing.T, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "member_role_permissions" WHERE member_role_id = $1`)).
					WithArgs(3).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "member_role_permissions"`)).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr:       true,
			permissionIDs: []uint64{5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			ctx := context.Background()

			tt.setup(t, mock)

			roles := NewMemberRoleDAO(db)

			err := roles.SetPermissions(ctx, 3, tt.permissionIDs)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)

			query.AssertDBMockDone(t, mock)
		})
	}
}
