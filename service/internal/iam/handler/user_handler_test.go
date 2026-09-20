package handler

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/storage"
)

func TestUserHandler_GetMe(t *testing.T) {
	tests := []struct {
		name           string
		users          iam.UserDAOMock
		useNoCaller    bool
		wantStatusCode int
	}{
		{
			name: "returns profile",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "returns unauthorized",
			useNoCaller:    true,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "returns not found",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return nil, nil
					},
				},
			},
			wantStatusCode: http.StatusNotFound,
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
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var app *fiber.App
			if tt.useNoCaller {
				app = userHandlerTestNoCaller(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{})
			} else {
				app = userHandlerTest(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
			}
			resp, err := doRequest(app, http.MethodGet, "/me", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestUserHandler_MeOrganizations(t *testing.T) {
	tests := []struct {
		name           string
		members        iam.MemberDAOMock
		useNoCaller    bool
		wantStatusCode int
	}{
		{
			name: "returns organizations",
			members: iam.MemberDAOMock{
				ListOrganizationsByUserFunc: func(_ context.Context, _ uint64) ([]*reference.Organization, error) {
					return []*reference.Organization{sampleOrganization()}, nil
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "returns unauthorized",
			useNoCaller:    true,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "returns server error",
			members: iam.MemberDAOMock{
				ListOrganizationsByUserFunc: func(_ context.Context, _ uint64) ([]*reference.Organization, error) {
					return nil, errors.New("db down")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var app *fiber.App
			if tt.useNoCaller {
				app = userHandlerTestNoCaller(t, iam.UserDAOMock{}, iam.PermissionDAOMock{}, iam.MemberDAOMock{})
			} else {
				app = userHandlerTest(t, iam.UserDAOMock{}, iam.PermissionDAOMock{}, tt.members, 5)
			}
			resp, err := doRequest(app, http.MethodGet, "/me/organizations", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestUserHandler_MeOrganizationPermissions(t *testing.T) {
	tests := []struct {
		name           string
		permissions    iam.PermissionDAOMock
		members        iam.MemberDAOMock
		path           string
		useNoCaller    bool
		wantStatusCode int
	}{
		{
			name: "returns permissions",
			permissions: iam.PermissionDAOMock{
				DAOMock: iam.DAOMock[iam.Permission]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[iam.Permission], error) {
						return &query.Page[iam.Permission]{Items: []*iam.Permission{samplePermission()}, Count: 1}, nil
					},
				},
			},
			members: iam.MemberDAOMock{
				FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
					return sampleMember(), nil
				},
			},
			path:           "/me/organizations/10/permissions",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "returns unauthorized",
			useNoCaller:    true,
			path:           "/me/organizations/10/permissions",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "rejects invalid organization id",
			path:           "/me/organizations/abc/permissions",
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "returns not found when not member",
			members: iam.MemberDAOMock{
				FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
					return nil, nil
				},
			},
			path:           "/me/organizations/10/permissions",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "returns server error on membership lookup",
			members: iam.MemberDAOMock{
				FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
					return nil, errors.New("db down")
				},
			},
			path:           "/me/organizations/10/permissions",
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "returns server error on authz",
			permissions: iam.PermissionDAOMock{
				UserIsOrganizationOwnerFunc: func(_ context.Context, _, _ uint64) (bool, error) {
					return false, errors.New("db down")
				},
			},
			members: iam.MemberDAOMock{
				FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
					return sampleMember(), nil
				},
			},
			path:           "/me/organizations/10/permissions",
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var app *fiber.App
			if tt.useNoCaller {
				app = userHandlerTestNoCaller(t, iam.UserDAOMock{}, iam.PermissionDAOMock{}, iam.MemberDAOMock{})
			} else {
				app = userHandlerTest(t, iam.UserDAOMock{}, tt.permissions, tt.members, 5)
			}
			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestUserHandler_GetUsers(t *testing.T) {
	tests := []struct {
		name           string
		users          iam.UserDAOMock
		path           string
		wantStatusCode int
	}{
		{
			name: "returns users",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[iam.User], error) {
						return &query.Page[iam.User]{Items: []*iam.User{sampleUser()}, Count: 1}, nil
					},
				},
			},
			path:           "/users/",
			wantStatusCode: http.StatusOK,
		},
		{
			name: "returns csv",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[iam.User], error) {
						return &query.Page[iam.User]{Items: []*iam.User{sampleUser()}, Count: 1}, nil
					},
				},
			},
			path:           "/users/?format=csv",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "rejects invalid query",
			path:           "/users/?sort=bogus:asc",
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[iam.User], error) {
						return nil, errors.New("db down")
					},
				},
			},
			path:           "/users/",
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := userHandlerTest(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		users          iam.UserDAOMock
		path           string
		wantStatusCode int
	}{
		{
			name: "returns user",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
				},
			},
			path:           "/users/1",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "rejects invalid id",
			path:           "/users/abc",
			wantStatusCode: http.StatusUnprocessableEntity,
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
			path:           "/users/1",
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "returns not found",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return nil, nil
					},
				},
			},
			path:           "/users/1",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "returns not found when private",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						user := sampleUser()
						user.Private = true
						return user, nil
					},
				},
			},
			path:           "/users/1",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := userHandlerTest(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		users          iam.UserDAOMock
		actorID        uint64
		body           string
		useNoCaller    bool
		wantStatusCode int
	}{
		{
			name: "creates user",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					CreateFunc: func(_ context.Context, entity *iam.User) (*iam.User, error) {
						entity.ID = 9
						return entity, nil
					},
				},
			},
			actorID:        5,
			body:           `{"username":"john.doe","first_name":"John","email":"john@example.com","password":"password123"}`,
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "rejects validation",
			actorID:        5,
			body:           `{"username":"","first_name":"","email":"john@example.com","password":"password123"}`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "rejects invalid birthday",
			actorID:        5,
			body:           `{"username":"john.doe","first_name":"John","email":"john@example.com","password":"password123","birthday":"not-a-date"}`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "maps email registered",
			users: iam.UserDAOMock{
				FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return sampleUser(), nil
				},
			},
			actorID:        5,
			body:           `{"username":"john.doe","first_name":"John","email":"john@example.com","password":"password123"}`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "maps username taken",
			users: iam.UserDAOMock{
				FindByUsernameFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return sampleUser(), nil
				},
			},
			actorID:        5,
			body:           `{"username":"john.doe","first_name":"John","email":"john@example.com","password":"password123"}`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "maps phone taken",
			users: iam.UserDAOMock{
				FindByPhoneFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return sampleUser(), nil
				},
			},
			actorID:        5,
			body:           `{"username":"john.doe","first_name":"John","email":"john@example.com","phone":"+6281234567890","password":"password123"}`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error on username check",
			users: iam.UserDAOMock{
				FindByUsernameFunc: func(_ context.Context, _ string) (*iam.User, error) {
					return nil, errors.New("db down")
				},
			},
			actorID:        5,
			body:           `{"username":"john.doe","first_name":"John","email":"john@example.com","password":"password123"}`,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "returns server error on create",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					CreateFunc: func(_ context.Context, _ *iam.User) (*iam.User, error) {
						return nil, errors.New("db down")
					},
				},
			},
			actorID:        5,
			body:           `{"username":"john.doe","first_name":"John","email":"john@example.com","password":"password123"}`,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var app *fiber.App
			if tt.useNoCaller {
				app = userHandlerTestNoCaller(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{})
			} else {
				app = userHandlerTest(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, tt.actorID)
			}
			resp, err := doRequest(app, http.MethodPost, "/users/", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		users          iam.UserDAOMock
		actorID        uint64
		path           string
		body           string
		useNoCaller    bool
		wantStatusCode int
	}{
		{
			name: "updates self",
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
			actorID:        5,
			path:           "/users/5",
			body:           `{"first_name":"Janet"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "returns forbidden for non-self",
			actorID:        5,
			path:           "/users/2",
			body:           `{"first_name":"Janet"}`,
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "returns unauthorized",
			useNoCaller:    true,
			path:           "/users/5",
			body:           `{"first_name":"Janet"}`,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "rejects invalid id",
			actorID:        5,
			path:           "/users/abc",
			body:           `{"first_name":"Janet"}`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "rejects invalid birthday",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return sampleUser(), nil
					},
				},
			},
			actorID:        5,
			path:           "/users/5",
			body:           `{"first_name":"Janet","birthday":"not-a-date"}`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "returns not found",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return nil, nil
					},
				},
			},
			actorID:        5,
			path:           "/users/5",
			body:           `{"first_name":"Janet"}`,
			wantStatusCode: http.StatusNotFound,
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
			actorID:        5,
			path:           "/users/5",
			body:           `{"first_name":"Janet"}`,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "rejects invalid body",
			actorID:        5,
			path:           "/users/5",
			body:           `{"first_name":`,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var app *fiber.App
			if tt.useNoCaller {
				app = userHandlerTestNoCaller(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{})
			} else {
				app = userHandlerTest(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, tt.actorID)
			}
			resp, err := doRequest(app, http.MethodPut, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		users          iam.UserDAOMock
		path           string
		wantStatusCode int
	}{
		{
			name: "deletes user",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					DeleteFunc: func(_ context.Context, _ uint64) error {
						return nil
					},
				},
			},
			path:           "/users/1",
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:           "rejects invalid id",
			path:           "/users/abc",
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					DeleteFunc: func(_ context.Context, _ uint64) error {
						return errors.New("db down")
					},
				},
			},
			path:           "/users/1",
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := userHandlerTest(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
			resp, err := doRequest(app, http.MethodDelete, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestUserHandler_UpdateMeAvatar(t *testing.T) {
	tests := []struct {
		name           string
		users          iam.UserDAOMock
		useNoCaller    bool
		field          string
		filename       string
		content        string
		customAvatar   bool
		wantStatusCode int
	}{
		{
			name: "accepts upload",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return &iam.User{Base: model.Base{ID: 5}}, nil
					},
					UpdateFunc: func(_ context.Context, u *iam.User) (*iam.User, error) {
						return u, nil
					},
				},
			},
			field:          "file",
			filename:       "avatar.png",
			content:        "\x89PNG\r\n\x1a\n",
			customAvatar:   true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing file field",
			field:          "not_file",
			filename:       "avatar.png",
			content:        "content",
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "rejects non-image content",
			field:          "file",
			filename:       "notes.txt",
			content:        "just plain text",
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "rejects empty content",
			field:          "file",
			filename:       "empty.png",
			content:        "",
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "returns unauthorized",
			useNoCaller:    true,
			field:          "file",
			filename:       "avatar.png",
			content:        "\x89PNG\r\n\x1a\n",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "returns server error when update fails",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) {
						return &iam.User{Base: model.Base{ID: 5}}, nil
					},
					UpdateFunc: func(_ context.Context, _ *iam.User) (*iam.User, error) {
						return nil, errors.New("db down")
					},
				},
			},
			field:          "file",
			filename:       "avatar.png",
			content:        "\x89PNG\r\n\x1a\n",
			customAvatar:   true,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var app *fiber.App
			if tt.useNoCaller {
				app = userHandlerTestNoCaller(t, iam.UserDAOMock{}, iam.PermissionDAOMock{}, iam.MemberDAOMock{})
			} else if tt.customAvatar {
				avatarSvc := iam.NewAvatarService(tt.users, storage.NewLocalStore(t.TempDir()))
				app = userHandlerTestApp(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5, avatarSvc)
			} else {
				app = userHandlerTest(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
			}
			resp, err := doMultipartRequest(app, http.MethodPut, "/me/avatar", tt.field, tt.filename, tt.content)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func doMultipartRequest(app *fiber.App, method, path, field, filename, content string) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write([]byte(content)); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return app.Test(req)
}
