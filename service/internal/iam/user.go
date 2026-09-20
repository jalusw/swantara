package iam

import (
	"context"
	"strings"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

const UserSearchLimit = 10

type UserService struct {
	userDAO     UserDAO
	passwordSvc PasswordService
}

func NewUserService(userDAO UserDAO, passwordSvc PasswordService) UserService {
	return UserService{userDAO: userDAO, passwordSvc: passwordSvc}
}

func (s UserService) List(ctx context.Context, q *query.Query) (*query.Page[User], error) {
	return s.userDAO.List(ctx, q)
}

func (s UserService) Find(ctx context.Context, id uint64) (*User, error) {
	return s.userDAO.Find(ctx, id)
}

func (s UserService) FindByEmail(ctx context.Context, email string) (*User, error) {
	return s.userDAO.FindByEmail(ctx, email)
}

func IsEmailOrPhoneQuery(term string) bool {
	trimmed := strings.TrimSpace(term)
	if strings.Contains(trimmed, "@") {
		return true
	}
	digits := 0
	for _, r := range trimmed {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '+' || r == '-' || r == ' ' || r == '(' || r == ')':
			continue
		default:
			return false
		}
	}
	return digits >= 6
}

func (s UserService) Search(ctx context.Context, term string, limit int) ([]*User, error) {
	trimmed := strings.TrimSpace(term)
	if trimmed == "" || IsEmailOrPhoneQuery(trimmed) {
		return []*User{}, nil
	}
	if limit <= 0 || limit > 20 {
		limit = UserSearchLimit
	}
	return s.userDAO.SearchUsers(ctx, trimmed, limit)
}

func (s UserService) Delete(ctx context.Context, id uint64) error {
	return s.userDAO.Delete(ctx, id)
}

func (s UserService) Create(ctx context.Context, user *User) (*User, error) {
	usernameTaken, err := s.IsUsernameTaken(ctx, user.Username)
	if err != nil {
		return nil, err
	}
	if usernameTaken {
		return nil, ErrUsernameTaken
	}

	emailTaken, err := s.IsEmailTaken(ctx, user.Email)
	if err != nil {
		return nil, err
	}
	if emailTaken {
		return nil, ErrEmailRegistered
	}

	if user.Phone != nil && *user.Phone != "" {
		phoneTaken, err := s.IsPhoneTaken(ctx, *user.Phone)
		if err != nil {
			return nil, err
		}
		if phoneTaken {
			return nil, ErrPhoneTaken
		}
	}

	hashedPassword, err := s.passwordSvc.Hash(user.Password)
	if err != nil {
		return nil, err
	}

	user.Password = hashedPassword
	user.Active = true

	return s.userDAO.Create(ctx, user)
}

func (s UserService) Update(
	ctx context.Context,
	userID uint64,
	user *User,
	active *bool,
) (*User, error) {
	existing, err := s.userDAO.Find(ctx, userID)
	if err != nil {
		return nil, err
	}

	if existing == nil || existing.ID == 0 {
		return nil, ErrUserNotFound
	}
	if user.FirstName != "" {
		existing.FirstName = user.FirstName
	}
	existing.LastName = user.LastName
	existing.Avatar = user.Avatar
	existing.Bio = user.Bio
	existing.Birthday = user.Birthday
	existing.Sex = user.Sex
	existing.Address = user.Address
	existing.City = user.City
	existing.PostalCode = user.PostalCode

	if active != nil {
		existing.Active = *active
	}

	return s.userDAO.Update(ctx, existing)
}

func (s UserService) UpdateUsername(ctx context.Context, userID uint64, username string) (*User, error) {
	user, err := s.userDAO.Find(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil || user.ID == 0 {
		return nil, ErrUserNotFound
	}

	usernameTaken, err := s.IsUsernameTaken(ctx, username)
	if err != nil {
		return nil, err
	}
	if usernameTaken && user.Username != username {
		return nil, ErrUsernameTaken
	}

	user.Username = username
	return s.userDAO.Update(ctx, user)
}

func (s UserService) UpdateEmail(ctx context.Context, userID uint64, email string) (*User, error) {
	user, err := s.userDAO.Find(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil || user.ID == 0 {
		return nil, ErrUserNotFound
	}

	emailTaken, err := s.IsEmailTaken(ctx, email)
	if err != nil {
		return nil, err
	}
	if emailTaken && user.Email != email {
		return nil, ErrEmailRegistered
	}

	user.Email = email
	user.EmailVerifiedAt = nil
	return s.userDAO.Update(ctx, user)
}

func (s UserService) UpdatePhone(ctx context.Context, userID uint64, phone string) (*User, error) {
	user, err := s.userDAO.Find(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil || user.ID == 0 {
		return nil, ErrUserNotFound
	}

	if phone != "" {
		phoneTaken, err := s.IsPhoneTaken(ctx, phone)
		if err != nil {
			return nil, err
		}
		if phoneTaken && (user.Phone == nil || *user.Phone != phone) {
			return nil, ErrPhoneTaken
		}
		user.Phone = &phone
	} else {
		user.Phone = nil
	}

	user.PhoneVerifiedAt = nil
	return s.userDAO.Update(ctx, user)
}

func (s UserService) IsUsernameTaken(ctx context.Context, username string) (bool, error) {
	existing, err := s.userDAO.FindByUsername(ctx, username)
	if err != nil {
		return false, err
	}

	return existing != nil && existing.ID != 0, nil
}

func (s UserService) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	existing, err := s.userDAO.FindByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	return existing != nil && existing.ID != 0, nil
}

func (s UserService) IsPhoneTaken(ctx context.Context, phone string) (bool, error) {
	existing, err := s.userDAO.FindByPhone(ctx, phone)
	if err != nil {
		return false, err
	}

	return existing != nil && existing.ID != 0, nil
}
