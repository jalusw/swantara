package seeders

import (
	"context"
	"log/slog"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type Seeder struct {
	db          *gorm.DB
	passwordSvc iam.PasswordService
}

func New(db *gorm.DB, passwordSvc iam.PasswordService) (*Seeder, error) {
	if db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	return &Seeder{db: db, passwordSvc: passwordSvc}, nil
}

func (s *Seeder) ctx() context.Context {
	return context.Background()
}

func (s *Seeder) SeedPermissions() error {
	ctx := s.ctx()

	seed, err := loadIamSeed()
	if err != nil {
		return err
	}

	permissionDAO := iam.NewPermissionDAO(s.db)

	for _, permission := range seed.permissions {
		existing, err := permissionDAO.FindByCode(ctx, permission.Code)
		if err != nil {
			slog.Error("Failed to check permission", "code", permission.Code, "error", err)
			continue
		}
		if existing != nil && existing.ID != 0 {
			continue
		}

		if _, err := permissionDAO.Create(ctx, permission); err != nil {
			slog.Error("Failed to create permission", "code", permission.Code, "error", err)
		}
	}

	slog.Info("Permissions Seeded.")

	return nil
}

func (s *Seeder) SeedDemoUsers() error {
	ctx := s.ctx()

	seed, err := loadSystemSeed()
	if err != nil {
		return err
	}

	existing, err := iam.NewUserDAO(s.db).FindByEmail(ctx, seed.AdminUser.Email)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != 0 {
		slog.Info("Demo users already seeded.")
		return nil
	}

	hashed, err := s.passwordSvc.Hash(seed.AdminUser.Password)
	if err != nil {
		return err
	}

	admin := &iam.User{
		Username:  seed.AdminUser.Username,
		FirstName: seed.AdminUser.FirstName,
		LastName:  helper.Ptr(seed.AdminUser.LastName),
		Email:     seed.AdminUser.Email,
		Password:  hashed,
		Active:    seed.AdminUser.Active,
	}

	if _, err := iam.NewUserDAO(s.db).Create(ctx, admin); err != nil {
		return err
	}

	slog.Info("Demo users seeded.", "user_id", admin.ID)

	ctx2 := s.ctx()
	org, err := dao.NewBase[reference.Organization](s.db).Search(ctx2, "name", seed.DefaultOrg.Name)
	if err != nil {
		return err
	}
	if org != nil {
		memberSvc := iam.NewMemberService(iam.NewMemberDAO(s.db), iam.NewMemberRoleDAO(s.db), iam.NewPermissionDAO(s.db))
		if err := memberSvc.ProvisionOwner(ctx2, org.ID, admin.ID); err != nil {
			return err
		}

		if err := s.seedApproverIDs(ctx2, org.ID, []uint64{admin.ID}); err != nil {
			slog.Error("Failed to seed purchase approvers", "error", err)
		}
	}

	return nil
}
