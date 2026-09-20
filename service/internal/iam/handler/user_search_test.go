package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestUserHandler_GetUsersSearch(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		users          iam.UserDAOMock
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "search by username",
			path: "/users/?q=jane",
			users: iam.UserDAOMock{
				SearchUsersFunc: func(_ context.Context, _ string, _ int) ([]*iam.User, error) {
					return []*iam.User{sampleUser()}, nil
				},
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "jane.doe",
		},
		{
			name:           "email query returns empty",
			path:           "/users/?q=jane%40example.com",
			users:          iam.UserDAOMock{},
			wantStatusCode: http.StatusOK,
			wantBody:       `"users":[]`,
		},
		{
			name:           "phone query returns empty",
			path:           "/users/?q=%2B6281234567890",
			users:          iam.UserDAOMock{},
			wantStatusCode: http.StatusOK,
			wantBody:       `"users":[]`,
		},
		{
			name: "search error",
			path: "/users/?q=jane",
			users: iam.UserDAOMock{
				SearchUsersFunc: func(_ context.Context, _ string, _ int) ([]*iam.User, error) {
					return nil, errors.New("db down")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "email filter rejected",
			path: "/users/?filter=email%3Aeq%3Ajane%40example.com",
			users: iam.UserDAOMock{
				DAOMock: iam.DAOMock[iam.User]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[iam.User], error) {
						return &query.Page[iam.User]{Items: []*iam.User{sampleUser()}, Count: 1}, nil
					},
				},
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := userHandlerTest(t, tt.users, iam.PermissionDAOMock{}, iam.MemberDAOMock{}, 5)
			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
			if tt.wantBody != "" {
				buf, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("read body: %v", err)
				}
				if !strings.Contains(string(buf), tt.wantBody) {
					t.Errorf("body does not contain %q, got %q", tt.wantBody, string(buf))
				}
			}
		})
	}
}
