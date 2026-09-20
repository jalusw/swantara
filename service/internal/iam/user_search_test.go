package iam

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestIsEmailOrPhoneQuery(t *testing.T) {
	tests := []struct {
		name string
		term string
		want bool
	}{
		{"username allowed", "janedoe", false},
		{"name allowed", "Jane Doe", false},
		{"short digits allowed", "12345", false},
		{"email blocked", "jane@example.com", true},
		{"partial email blocked", "jane@", true},
		{"phone blocked", "+6281234567890", true},
		{"phone dashes blocked", "0812-345-678", true},
		{"phone spaces blocked", "0812 345 678", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmailOrPhoneQuery(tt.term); got != tt.want {
				t.Errorf("IsEmailOrPhoneQuery(%q) = %v, want %v", tt.term, got, tt.want)
			}
		})
	}
}

func TestUserService_Search(t *testing.T) {
	ctx := context.Background()
	found := []*User{{Base: model.Base{ID: 1}, Username: "janedoe", FirstName: "Jane"}}

	tests := []struct {
		name      string
		term      string
		searchFn  func(ctx context.Context, term string, limit int) ([]*User, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty term returns empty",
			term:      "   ",
			wantCount: 0,
		},
		{
			name: "email term returns empty without dao call",
			term: "jane@example.com",
			searchFn: func(ctx context.Context, term string, limit int) ([]*User, error) {
				return nil, errors.New("must not be called")
			},
			wantCount: 0,
		},
		{
			name: "phone term returns empty without dao call",
			term: "+6281234567890",
			searchFn: func(ctx context.Context, term string, limit int) ([]*User, error) {
				return nil, errors.New("must not be called")
			},
			wantCount: 0,
		},
		{
			name: "username term searches",
			term: "jane",
			searchFn: func(ctx context.Context, term string, limit int) ([]*User, error) {
				if limit != UserSearchLimit {
					t.Errorf("limit = %d, want %d", limit, UserSearchLimit)
				}
				return found, nil
			},
			wantCount: 1,
		},
		{
			name: "dao error propagates",
			term: "jane",
			searchFn: func(ctx context.Context, term string, limit int) ([]*User, error) {
				return nil, errors.New("db down")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserTestService(UserDAOMock{SearchUsersFunc: tt.searchFn}, validTestPasswordSvc())
			got, err := svc.Search(ctx, tt.term, 0)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}
