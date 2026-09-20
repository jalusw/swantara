package iam

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"golang.org/x/crypto/bcrypt"
)

func ptr[T any](v T) *T { return &v }

func newUserTestService(userDao UserDAO, passwordSvc PasswordService) UserService {
	return UserService{userDAO: userDao, passwordSvc: passwordSvc}
}

func validTestPasswordSvc() PasswordService {
	return NewPasswordService(&config.Config{
		BcryptCost: bcrypt.MinCost,
		Pepper:     "test-pepper",
	})
}

func brokenTestPasswordSvc() PasswordService {
	return NewPasswordService(&config.Config{
		BcryptCost: 100,
		Pepper:     "test-pepper",
	})
}

func TestNewUserService(t *testing.T) {
	svc := NewUserService(UserDAOMock{}, validTestPasswordSvc())
	if svc.userDAO == nil {
		t.Error("expected non-nil userDao")
	}
}

func TestUserService_Create(t *testing.T) {
	ctx := context.Background()
	phone := "0812"
	birthday := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	sex := "male"
	address := "addr"
	city := "city"
	postal := "12345"

	data := &User{
		Username:   "johndoe",
		FirstName:  "John",
		LastName:   ptr("Doe"),
		Email:      "john@test.com",
		Phone:      &phone,
		Password:   "test-password",
		Birthday:   &birthday,
		Sex:        &sex,
		Address:    &address,
		City:       &city,
		PostalCode: &postal,
	}

	created := &User{Username: data.Username, Email: data.Email}

	tests := []struct {
		name               string
		data               *User
		passwordSvc        PasswordService
		findByEmailFunc    func(ctx context.Context, email string) (*User, error)
		findByUsernameFunc func(ctx context.Context, username string) (*User, error)
		findByPhoneFunc    func(ctx context.Context, phone string) (*User, error)
		createFunc         func(ctx context.Context, entity *User) (*User, error)
		wantErr            bool
		wantErrVal         error
	}{
		{
			name:        "success",
			data:        data,
			passwordSvc: validTestPasswordSvc(),
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
			createFunc: func(ctx context.Context, entity *User) (*User, error) {
				return created, nil
			},
		},
		{
			name:        "success without phone",
			data:        &User{Username: "u", FirstName: "F", Email: "e@t.com", Password: "p"},
			passwordSvc: validTestPasswordSvc(),
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			createFunc: func(ctx context.Context, entity *User) (*User, error) {
				return entity, nil
			},
		},
		{
			name:        "username taken",
			data:        data,
			passwordSvc: validTestPasswordSvc(),
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return &User{Base: model.Base{ID: 9}, Username: "x"}, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrUsernameTaken,
		},
		{
			name:        "email registered",
			data:        data,
			passwordSvc: validTestPasswordSvc(),
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return &User{Base: model.Base{ID: 9}, Email: "x"}, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrEmailRegistered,
		},
		{
			name:        "phone taken",
			data:        data,
			passwordSvc: validTestPasswordSvc(),
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return &User{Base: model.Base{ID: 9}, Phone: &phone}, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrPhoneTaken,
		},
		{
			name:        "username search error",
			data:        data,
			passwordSvc: validTestPasswordSvc(),
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, errors.New("db error")
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:        "phone search error",
			data:        data,
			passwordSvc: validTestPasswordSvc(),
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, errors.New("db error")
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:        "hash error",
			data:        data,
			passwordSvc: brokenTestPasswordSvc(),
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:        "create error",
			data:        data,
			passwordSvc: validTestPasswordSvc(),
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
			createFunc: func(ctx context.Context, entity *User) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserTestService(
				UserDAOMock{
					FindByEmailFunc:    tt.findByEmailFunc,
					FindByUsernameFunc: tt.findByUsernameFunc,
					FindByPhoneFunc:    tt.findByPhoneFunc,
					DAOMock:            DAOMock[User]{CreateFunc: tt.createFunc},
				},
				tt.passwordSvc,
			)
			got, err := svc.Create(ctx, tt.data)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected user, got nil")
			}
		})
	}
}

func TestUserService_Update(t *testing.T) {
	ctx := context.Background()
	phone := "0812"

	existing := &User{Base: model.Base{ID: 7}, Username: "old", FirstName: "Old", Email: "old@test.com"}

	validData := &User{
		Username:   "newname",
		FirstName:  "New",
		LastName:   ptr("Last"),
		Email:      "new@test.com",
		Phone:      &phone,
		Avatar:     ptr("avatar.png"),
		Bio:        ptr("bio"),
		Birthday:   ptr(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)),
		Sex:        ptr("male"),
		Address:    ptr("addr"),
		City:       ptr("city"),
		PostalCode: ptr("123"),
	}

	tests := []struct {
		name               string
		userID             uint64
		data               *User
		active             *bool
		passwordSvc        PasswordService
		findFunc           func(ctx context.Context, id uint64) (*User, error)
		findByEmailFunc    func(ctx context.Context, email string) (*User, error)
		findByUsernameFunc func(ctx context.Context, username string) (*User, error)
		findByPhoneFunc    func(ctx context.Context, phone string) (*User, error)
		updateFunc         func(ctx context.Context, entity *User) (*User, error)
		wantFirstName      string
		wantErr            bool
		wantErrVal         error
	}{
		{
			name:          "success",
			userID:        7,
			data:          validData,
			active:        ptr(true),
			passwordSvc:   validTestPasswordSvc(),
			wantFirstName: "New",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:          "success without password and active",
			userID:        7,
			data:          &User{Username: "newname", FirstName: "New", Email: "new@test.com"},
			passwordSvc:   validTestPasswordSvc(),
			wantFirstName: "New",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:          "empty first name keeps existing value",
			userID:        7,
			data:          &User{Username: "newname", FirstName: "", Email: "new@test.com"},
			passwordSvc:   validTestPasswordSvc(),
			wantFirstName: "Old",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:        "find error",
			userID:      7,
			data:        validData,
			passwordSvc: validTestPasswordSvc(),
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:        "user not found",
			userID:      99,
			data:        validData,
			passwordSvc: validTestPasswordSvc(),
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrUserNotFound,
		},
		{
			name:        "update error",
			userID:      7,
			data:        validData,
			passwordSvc: validTestPasswordSvc(),
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
			updateFunc: func(ctx context.Context, entity *User) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserTestService(
				UserDAOMock{
					DAOMock:            DAOMock[User]{FindFunc: tt.findFunc, UpdateFunc: tt.updateFunc},
					FindByEmailFunc:    tt.findByEmailFunc,
					FindByUsernameFunc: tt.findByUsernameFunc,
					FindByPhoneFunc:    tt.findByPhoneFunc,
				},
				tt.passwordSvc,
			)
			got, err := svc.Update(ctx, tt.userID, tt.data, tt.active)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected user, got nil")
			}
			if tt.active != nil && got.Active != *tt.active {
				t.Errorf("expected active %v, got %v", *tt.active, got.Active)
			}
			if tt.wantFirstName != "" && got.FirstName != tt.wantFirstName {
				t.Errorf("expected first name %q, got %q", tt.wantFirstName, got.FirstName)
			}
		})
	}
}

func TestUserService_IsUsernameTaken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		username           string
		findByUsernameFunc func(ctx context.Context, username string) (*User, error)
		wantTaken          bool
		wantErr            bool
	}{
		{
			name:     "no existing user",
			username: "value",
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:     "existing zero id",
			username: "value",
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return &User{}, nil
			},
		},
		{
			name:     "username taken",
			username: "value",
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return &User{Base: model.Base{ID: 7}, Username: "v"}, nil
			},
			wantTaken: true,
		},
		{
			name:     "search error",
			username: "value",
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserTestService(
				UserDAOMock{FindByUsernameFunc: tt.findByUsernameFunc},
				validTestPasswordSvc(),
			)
			taken, err := svc.IsUsernameTaken(ctx, tt.username)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if taken != tt.wantTaken {
				t.Errorf("expected taken %v, got %v", tt.wantTaken, taken)
			}
		})
	}
}

func TestUserService_IsEmailTaken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		email           string
		findByEmailFunc func(ctx context.Context, email string) (*User, error)
		wantTaken       bool
		wantErr         bool
	}{
		{
			name:  "no existing user",
			email: "value",
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:  "existing zero id",
			email: "value",
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return &User{}, nil
			},
		},
		{
			name:  "email registered",
			email: "value",
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return &User{Base: model.Base{ID: 7}, Email: "v"}, nil
			},
			wantTaken: true,
		},
		{
			name:  "search error",
			email: "value",
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserTestService(
				UserDAOMock{FindByEmailFunc: tt.findByEmailFunc},
				validTestPasswordSvc(),
			)
			taken, err := svc.IsEmailTaken(ctx, tt.email)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if taken != tt.wantTaken {
				t.Errorf("expected taken %v, got %v", tt.wantTaken, taken)
			}
		})
	}
}

func TestUserService_IsPhoneTaken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		phone           string
		findByPhoneFunc func(ctx context.Context, phone string) (*User, error)
		wantTaken       bool
		wantErr         bool
	}{
		{
			name:  "no existing user",
			phone: "value",
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:  "existing zero id",
			phone: "value",
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return &User{}, nil
			},
		},
		{
			name:  "phone taken",
			phone: "value",
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return &User{Base: model.Base{ID: 9}, Phone: ptr("v")}, nil
			},
			wantTaken: true,
		},
		{
			name:  "search error",
			phone: "value",
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserTestService(
				UserDAOMock{FindByPhoneFunc: tt.findByPhoneFunc},
				validTestPasswordSvc(),
			)
			taken, err := svc.IsPhoneTaken(ctx, tt.phone)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if taken != tt.wantTaken {
				t.Errorf("expected taken %v, got %v", tt.wantTaken, taken)
			}
		})
	}
}

func TestUserService_UpdateUsername(t *testing.T) {
	ctx := context.Background()
	existing := &User{Base: model.Base{ID: 7}, Username: "old", Email: "u@test.com"}

	tests := []struct {
		name               string
		username           string
		findFunc           func(ctx context.Context, id uint64) (*User, error)
		findByUsernameFunc func(ctx context.Context, username string) (*User, error)
		updateFunc         func(ctx context.Context, entity *User) (*User, error)
		wantErr            bool
		wantErrVal         error
	}{
		{
			name:     "success",
			username: "newusername",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:     "find error",
			username: "newusername",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:     "user not found",
			username: "newusername",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrUserNotFound,
		},
		{
			name:     "username taken",
			username: "newusername",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return &User{Base: model.Base{ID: 9}, Username: "x"}, nil
			},
			wantErr:    true,
			wantErrVal: ErrUsernameTaken,
		},
		{
			name:     "unchanged username allowed",
			username: "old",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return existing, nil
			},
		},
		{
			name:     "update error",
			username: "newusername",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByUsernameFunc: func(ctx context.Context, username string) (*User, error) {
				return nil, nil
			},
			updateFunc: func(ctx context.Context, entity *User) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserTestService(
				UserDAOMock{
					DAOMock:            DAOMock[User]{FindFunc: tt.findFunc, UpdateFunc: tt.updateFunc},
					FindByUsernameFunc: tt.findByUsernameFunc,
				},
				validTestPasswordSvc(),
			)
			got, err := svc.UpdateUsername(ctx, 7, tt.username)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Username != tt.username {
				t.Errorf("expected username %q, got %q", tt.username, got.Username)
			}
		})
	}
}

func TestUserService_UpdateEmail(t *testing.T) {
	ctx := context.Background()
	existing := &User{Base: model.Base{ID: 7}, Username: "user", Email: "old@test.com"}

	tests := []struct {
		name            string
		email           string
		findFunc        func(ctx context.Context, id uint64) (*User, error)
		findByEmailFunc func(ctx context.Context, email string) (*User, error)
		updateFunc      func(ctx context.Context, entity *User) (*User, error)
		wantErr         bool
		wantErrVal      error
	}{
		{
			name:  "success",
			email: "new@test.com",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:  "find error",
			email: "new@test.com",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:  "user not found",
			email: "new@test.com",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrUserNotFound,
		},
		{
			name:  "email registered",
			email: "new@test.com",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return &User{Base: model.Base{ID: 9}, Email: "x"}, nil
			},
			wantErr:    true,
			wantErrVal: ErrEmailRegistered,
		},
		{
			name:  "unchanged email allowed",
			email: "old@test.com",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return existing, nil
			},
		},
		{
			name:  "update error",
			email: "new@test.com",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByEmailFunc: func(ctx context.Context, email string) (*User, error) {
				return nil, nil
			},
			updateFunc: func(ctx context.Context, entity *User) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserTestService(
				UserDAOMock{
					DAOMock:         DAOMock[User]{FindFunc: tt.findFunc, UpdateFunc: tt.updateFunc},
					FindByEmailFunc: tt.findByEmailFunc,
				},
				validTestPasswordSvc(),
			)
			got, err := svc.UpdateEmail(ctx, 7, tt.email)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Email != tt.email {
				t.Errorf("expected email %q, got %q", tt.email, got.Email)
			}
			if got.EmailVerifiedAt != nil {
				t.Error("expected email_verified_at to be cleared")
			}
		})
	}
}

func TestUserService_UpdatePhone(t *testing.T) {
	ctx := context.Background()
	existingPhone := "9999"
	existing := &User{Base: model.Base{ID: 7}, Username: "user", Email: "u@test.com", Phone: &existingPhone}

	tests := []struct {
		name            string
		phone           string
		findFunc        func(ctx context.Context, id uint64) (*User, error)
		findByPhoneFunc func(ctx context.Context, phone string) (*User, error)
		updateFunc      func(ctx context.Context, entity *User) (*User, error)
		wantErr         bool
		wantErrVal      error
	}{
		{
			name:  "success",
			phone: "0812",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:  "success clears phone",
			phone: "",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
		},
		{
			name:  "find error",
			phone: "0812",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:  "user not found",
			phone: "0812",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrUserNotFound,
		},
		{
			name:  "phone taken",
			phone: "0812",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return &User{Base: model.Base{ID: 9}, Phone: ptr("x")}, nil
			},
			wantErr:    true,
			wantErrVal: ErrPhoneTaken,
		},
		{
			name:  "unchanged phone allowed",
			phone: "9999",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return existing, nil
			},
		},
		{
			name:  "update error",
			phone: "0812",
			findFunc: func(ctx context.Context, id uint64) (*User, error) {
				user := *existing
				return &user, nil
			},
			findByPhoneFunc: func(ctx context.Context, phone string) (*User, error) {
				return nil, nil
			},
			updateFunc: func(ctx context.Context, entity *User) (*User, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserTestService(
				UserDAOMock{
					DAOMock:         DAOMock[User]{FindFunc: tt.findFunc, UpdateFunc: tt.updateFunc},
					FindByPhoneFunc: tt.findByPhoneFunc,
				},
				validTestPasswordSvc(),
			)
			got, err := svc.UpdatePhone(ctx, 7, tt.phone)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.phone == "" && got.Phone != nil {
				t.Error("expected phone to be cleared")
			}
			if tt.phone != "" && (got.Phone == nil || *got.Phone != tt.phone) {
				t.Errorf("expected phone %q, got %v", tt.phone, got.Phone)
			}
			if got.PhoneVerifiedAt != nil {
				t.Error("expected phone_verified_at to be cleared")
			}
		})
	}
}
