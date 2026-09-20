package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type mapAvatarStore struct {
	objects map[string][]byte
}

func newAvatarStoreWith(id string, content []byte) *mapAvatarStore {
	return &mapAvatarStore{objects: map[string][]byte{id: content}}
}

func (m *mapAvatarStore) Save(_ context.Context, id string, reader io.Reader) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.objects[id] = content
	return nil
}

func (m *mapAvatarStore) Open(_ context.Context, id string) (io.ReadCloser, error) {
	content, ok := m.objects[id]
	if !ok {
		return nil, errors.New("object not found")
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}

func (m *mapAvatarStore) Delete(_ context.Context, id string) error {
	delete(m.objects, id)
	return nil
}

func TestUserHandler_Avatar(t *testing.T) {
	avatarID := "avatars/7/abc"

	newUsers := func(user *iam.User, err error) iam.UserDAOMock {
		return iam.UserDAOMock{
			DAOMock: iam.DAOMock[iam.User]{
				FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) { return user, err },
			},
		}
	}

	t.Run("me avatar requires caller", func(t *testing.T) {
		app := userHandlerTestNoCaller(t, newUsers(nil, nil), iam.PermissionDAOMock{}, iam.MemberDAOMock{})
		resp, err := doIamRequest(app, http.MethodGet, "/me/avatar", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)
	})

	t.Run("get avatar rejects invalid id", func(t *testing.T) {
		app := userHandlerTest(t, newUsers(nil, nil), iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
		for _, path := range []string{"/users/abc/avatar", "/users/0/avatar"} {
			resp, err := doIamRequest(app, http.MethodGet, path, "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("avatar not found", func(t *testing.T) {
		app := userHandlerTest(t, newUsers(nil, nil), iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
		resp, err := doIamRequest(app, http.MethodGet, "/users/5/avatar", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		app = userHandlerTest(t, newUsers(&iam.User{Base: model.Base{ID: 5}}, nil), iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
		resp, err = doIamRequest(app, http.MethodGet, "/users/5/avatar", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("avatar lookup error", func(t *testing.T) {
		dbErr := errors.New("db down")
		app := userHandlerTest(t, newUsers(nil, dbErr), iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
		resp, err := doIamRequest(app, http.MethodGet, "/users/5/avatar", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("serves stored avatar", func(t *testing.T) {
		users := newUsers(&iam.User{Base: model.Base{ID: 5}, Avatar: &avatarID}, nil)
		app := userHandlerTestApp(t, users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5, iam.NewAvatarService(users, newAvatarStoreWith(avatarID, []byte("img-bytes"))))
		resp, err := doIamRequest(app, http.MethodGet, "/users/5/avatar", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})
}
