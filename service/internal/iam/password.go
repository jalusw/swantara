package iam

import (
	"github.com/jalusw/swantara/apps/service/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type PasswordService struct {
	cost   int
	pepper string
}

func NewPasswordService(cfg *config.Config) PasswordService {
	return PasswordService{
		cost:   cfg.BcryptCost,
		pepper: cfg.Pepper,
	}
}

func (s PasswordService) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password+s.pepper), s.cost)
	return string(hashed), err
}

func (s PasswordService) Verify(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+s.pepper))
}
