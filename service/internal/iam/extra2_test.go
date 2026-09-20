package iam

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"regexp"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type errAvatarStore struct{ err error }

func (m errAvatarStore) Save(_ context.Context, _ string, _ io.Reader) error { return m.err }

func (m errAvatarStore) Open(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, m.err
}

func (m errAvatarStore) Delete(_ context.Context, _ string) error {
	return m.err
}

func TestAvatarService_Download(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	avatar := "avatars/7/abc"

	t.Run("downloads avatar", func(t *testing.T) {
		users := UserDAOMock{
			DAOMock: DAOMock[User]{
				FindFunc: func(_ context.Context, _ uint64) (*User, error) {
					return &User{Base: model.Base{ID: 7}, Avatar: &avatar}, nil
				},
			},
		}
		store := newAvatarStoreMock()
		store.opened[avatar] = []byte("img")
		svc := NewAvatarService(users, store)
		reader, err := svc.Download(ctx, 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		content, _ := io.ReadAll(reader)
		if !bytes.Equal(content, []byte("img")) {
			t.Errorf("content = %q", content)
		}
	})

	t.Run("download failures", func(t *testing.T) {
		users := func(user *User, err error) UserDAOMock {
			return UserDAOMock{
				DAOMock: DAOMock[User]{
					FindFunc: func(_ context.Context, _ uint64) (*User, error) { return user, err },
				},
			}
		}

		svc := NewAvatarService(users(nil, dbErr), newAvatarStoreMock())
		_, err := svc.Download(ctx, 7)
		helper.AssertError(t, err, true, dbErr)

		svc = NewAvatarService(users(nil, nil), newAvatarStoreMock())
		_, err = svc.Download(ctx, 7)
		helper.AssertError(t, err, true, ErrAvatarNotFound)

		svc = NewAvatarService(users(&User{Base: model.Base{ID: 7}}, nil), newAvatarStoreMock())
		_, err = svc.Download(ctx, 7)
		helper.AssertError(t, err, true, ErrAvatarNotFound)

		svc = NewAvatarService(users(&User{Base: model.Base{ID: 7}, Avatar: &avatar}, nil), errAvatarStore{err: fs.ErrNotExist})
		_, err = svc.Download(ctx, 7)
		helper.AssertError(t, err, true, ErrAvatarNotFound)

		svc = NewAvatarService(users(&User{Base: model.Base{ID: 7}, Avatar: &avatar}, nil), errAvatarStore{err: dbErr})
		_, err = svc.Download(ctx, 7)
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestMemberService_Roles(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	provisionSvc := func(roles MemberRoleDAOMock, permissions PermissionDAOMock) MemberService {
		return NewMemberService(MemberDAOMock{}, roles, permissions)
	}
	emptyRole := MemberRoleDAOMock{
		FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*MemberRole, error) {
			return nil, nil
		},
	}

	t.Run("ensure role create error", func(t *testing.T) {
		roles := emptyRole
		roles.CRUDMock = dao.CRUDMock[MemberRole]{
			CreateFunc: func(_ context.Context, _ *MemberRole) (*MemberRole, error) { return nil, dbErr },
		}
		svc := provisionSvc(roles, PermissionDAOMock{})
		helper.AssertError(t, svc.ProvisionOwner(ctx, 1, 10), true, dbErr)
	})

	t.Run("bind permissions list error", func(t *testing.T) {
		roles := emptyRole
		roles.CRUDMock = dao.CRUDMock[MemberRole]{
			CreateFunc: func(_ context.Context, r *MemberRole) (*MemberRole, error) {
				r.ID = 5
				return r, nil
			},
		}
		svc := provisionSvc(roles, PermissionDAOMock{
			DAOMock: DAOMock[Permission]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Permission], error) {
					return nil, dbErr
				},
			},
		})
		helper.AssertError(t, svc.ProvisionOwner(ctx, 1, 10), true, dbErr)
	})

	t.Run("bind permissions bind error", func(t *testing.T) {
		roles := emptyRole
		roles.CRUDMock = dao.CRUDMock[MemberRole]{
			CreateFunc: func(_ context.Context, r *MemberRole) (*MemberRole, error) {
				r.ID = 5
				return r, nil
			},
		}
		roles.BindPermissionFunc = func(_ context.Context, _, _ uint64) error { return dbErr }
		svc := provisionSvc(roles, PermissionDAOMock{
			DAOMock: DAOMock[Permission]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Permission], error) {
					return &query.Page[Permission]{Items: []*Permission{{Base: model.Base{ID: 1}, Action: "view"}}}, nil
				},
			},
		})
		helper.AssertError(t, svc.ProvisionOwner(ctx, 1, 10), true, dbErr)
	})

	t.Run("create role set permissions error", func(t *testing.T) {
		svc := NewMemberService(MemberDAOMock{}, MemberRoleDAOMock{
			FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*MemberRole, error) {
				return nil, nil
			},
			CRUDMock: dao.CRUDMock[MemberRole]{
				CreateFunc: func(_ context.Context, r *MemberRole) (*MemberRole, error) {
					r.ID = 6
					return r, nil
				},
			},
			SetPermissionsFunc: func(_ context.Context, _ uint64, _ []uint64) error { return dbErr },
		}, PermissionDAOMock{})
		_, err := svc.CreateRole(ctx, 1, &MemberRole{Code: "x"}, []uint64{1})
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestIAMDAO_Queries(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("refresh token lookup error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_sessions"`)).WillReturnError(dbErr)
		_, err := NewUserSessionDAO(db).FindByRefreshToken(ctx, "tok")
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("permission code error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions"`)).WillReturnError(dbErr)
		_, err := NewPermissionDAO(db).FindByCode(ctx, "x")
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}
