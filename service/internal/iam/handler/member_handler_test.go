package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestMemberHandler_ListMembers(t *testing.T) {
	tests := []struct {
		name       string
		members    iam.MemberDAOMock
		withTenant bool
		wantStatus int
	}{
		{
			name: "returns members",
			members: iam.MemberDAOMock{
				ListByOrganizationFunc: func(_ context.Context, _ uint64) ([]*iam.Member, error) {
					return []*iam.Member{sampleMember()}, nil
				},
			},
			withTenant: true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns unauthorized",
			members:    iam.MemberDAOMock{},
			withTenant: false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns server error",
			members: iam.MemberDAOMock{
				ListByOrganizationFunc: func(_ context.Context, _ uint64) ([]*iam.Member, error) {
					return nil, errors.New("db down")
				},
			},
			withTenant: true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := memberHandlerTest(t, tt.members, iam.MemberRoleDAOMock{}, iam.PermissionDAOMock{}, tt.withTenant)
			resp, err := doRequest(app, http.MethodGet, "/members/", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestMemberHandler_InviteMember(t *testing.T) {
	tests := []struct {
		name       string
		members    iam.MemberDAOMock
		roles      iam.MemberRoleDAOMock
		withTenant bool
		body       string
		wantStatus int
	}{
		{
			name: "invites member",
			members: iam.MemberDAOMock{
				FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
					return nil, nil
				},
				CRUDMock: dao.CRUDMock[iam.Member]{
					CreateFunc: func(_ context.Context, entity *iam.Member) (*iam.Member, error) {
						entity.ID = 1
						return entity, nil
					},
				},
			},
			roles: iam.MemberRoleDAOMock{
				FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*iam.MemberRole, error) {
					return sampleMemberRole(), nil
				},
			},
			withTenant: true,
			body:       `{"user_id":7}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns unauthorized",
			members:    iam.MemberDAOMock{},
			roles:      iam.MemberRoleDAOMock{},
			withTenant: false,
			body:       `{"user_id":7}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "rejects validation",
			members:    iam.MemberDAOMock{},
			roles:      iam.MemberRoleDAOMock{},
			withTenant: true,
			body:       `{"user_id":0}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns conflict when already member",
			members: iam.MemberDAOMock{
				FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
					return sampleMember(), nil
				},
			},
			roles:      iam.MemberRoleDAOMock{},
			withTenant: true,
			body:       `{"user_id":7}`,
			wantStatus: http.StatusConflict,
		},
		{
			name: "maps missing default role",
			members: iam.MemberDAOMock{
				FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
					return nil, nil
				},
			},
			roles: iam.MemberRoleDAOMock{
				FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*iam.MemberRole, error) {
					return nil, nil
				},
			},
			withTenant: true,
			body:       `{"user_id":7}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			members: iam.MemberDAOMock{
				FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
					return nil, errors.New("db down")
				},
			},
			roles:      iam.MemberRoleDAOMock{},
			withTenant: true,
			body:       `{"user_id":7}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "returns server error on create",
			members: iam.MemberDAOMock{
				FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
					return nil, nil
				},
				CRUDMock: dao.CRUDMock[iam.Member]{
					CreateFunc: func(_ context.Context, _ *iam.Member) (*iam.Member, error) {
						return nil, errors.New("db down")
					},
				},
			},
			roles: iam.MemberRoleDAOMock{
				FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*iam.MemberRole, error) {
					return sampleMemberRole(), nil
				},
			},
			withTenant: true,
			body:       `{"user_id":7}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := memberHandlerTest(t, tt.members, tt.roles, iam.PermissionDAOMock{}, tt.withTenant)
			resp, err := doRequest(app, http.MethodPost, "/members/", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestMemberHandler_RemoveMember(t *testing.T) {
	tests := []struct {
		name       string
		members    iam.MemberDAOMock
		withTenant bool
		path       string
		wantStatus int
	}{
		{
			name: "removes member",
			members: iam.MemberDAOMock{
				CRUDMock: dao.CRUDMock[iam.Member]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.Member, error) {
						return sampleMember(), nil
					},
					HardDeleteFunc: func(_ context.Context, _ uint64) error {
						return nil
					},
				},
			},
			withTenant: true,
			path:       "/members/1",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns unauthorized",
			members:    iam.MemberDAOMock{},
			withTenant: false,
			path:       "/members/1",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "rejects invalid id",
			members:    iam.MemberDAOMock{},
			withTenant: true,
			path:       "/members/abc",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns not found",
			members: iam.MemberDAOMock{
				CRUDMock: dao.CRUDMock[iam.Member]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.Member, error) {
						return nil, nil
					},
				},
			},
			withTenant: true,
			path:       "/members/1",
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns server error",
			members: iam.MemberDAOMock{
				CRUDMock: dao.CRUDMock[iam.Member]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.Member, error) {
						return nil, errors.New("db down")
					},
				},
			},
			withTenant: true,
			path:       "/members/1",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "returns server error on delete",
			members: iam.MemberDAOMock{
				CRUDMock: dao.CRUDMock[iam.Member]{
					FindFunc: func(_ context.Context, _ uint64) (*iam.Member, error) {
						member := sampleMember()
						member.ID = 1
						return member, nil
					},
					HardDeleteFunc: func(_ context.Context, _ uint64) error {
						return errors.New("db down")
					},
				},
			},
			withTenant: true,
			path:       "/members/1",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := memberHandlerTest(t, tt.members, iam.MemberRoleDAOMock{}, iam.PermissionDAOMock{}, tt.withTenant)
			resp, err := doRequest(app, http.MethodDelete, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestMemberHandler_ListMemberRoles(t *testing.T) {
	tests := []struct {
		name       string
		roles      iam.MemberRoleDAOMock
		withTenant bool
		wantStatus int
	}{
		{
			name: "returns roles",
			roles: iam.MemberRoleDAOMock{
				ListByOrganizationFunc: func(_ context.Context, _ uint64) ([]*iam.MemberRole, error) {
					return []*iam.MemberRole{sampleMemberRole()}, nil
				},
			},
			withTenant: true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns unauthorized",
			roles:      iam.MemberRoleDAOMock{},
			withTenant: false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns server error",
			roles: iam.MemberRoleDAOMock{
				ListByOrganizationFunc: func(_ context.Context, _ uint64) ([]*iam.MemberRole, error) {
					return nil, errors.New("db down")
				},
			},
			withTenant: true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := memberHandlerTest(t, iam.MemberDAOMock{}, tt.roles, iam.PermissionDAOMock{}, tt.withTenant)
			resp, err := doRequest(app, http.MethodGet, "/member-roles/", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestMemberHandler_ListPermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions iam.PermissionDAOMock
		wantStatus  int
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
			wantStatus: http.StatusOK,
		},
		{
			name: "returns server error",
			permissions: iam.PermissionDAOMock{
				DAOMock: iam.DAOMock[iam.Permission]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[iam.Permission], error) {
						return nil, errors.New("db down")
					},
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := memberHandlerTest(t, iam.MemberDAOMock{}, iam.MemberRoleDAOMock{}, tt.permissions, true)
			resp, err := doRequest(app, http.MethodGet, "/permissions/", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestMemberHandler_CreateMemberRole(t *testing.T) {
	tests := []struct {
		name       string
		roles      iam.MemberRoleDAOMock
		withTenant bool
		body       string
		wantStatus int
	}{
		{
			name: "creates role",
			roles: iam.MemberRoleDAOMock{
				FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*iam.MemberRole, error) {
					return nil, nil
				},
				CRUDMock: dao.CRUDMock[iam.MemberRole]{
					CreateFunc: func(_ context.Context, entity *iam.MemberRole) (*iam.MemberRole, error) {
						entity.ID = 1
						return entity, nil
					},
				},
				SetPermissionsFunc: func(_ context.Context, _ uint64, _ []uint64) error {
					return nil
				},
			},
			withTenant: true,
			body:       `{"name":"Manager","code":"manager","permission_ids":[1,2]}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns unauthorized",
			roles:      iam.MemberRoleDAOMock{},
			withTenant: false,
			body:       `{"name":"Manager","code":"manager"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "rejects validation",
			roles:      iam.MemberRoleDAOMock{},
			withTenant: true,
			body:       `{"name":"","code":""}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns conflict when code exists",
			roles: iam.MemberRoleDAOMock{
				FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*iam.MemberRole, error) {
					return sampleMemberRole(), nil
				},
			},
			withTenant: true,
			body:       `{"name":"Manager","code":"member"}`,
			wantStatus: http.StatusConflict,
		},
		{
			name: "returns server error",
			roles: iam.MemberRoleDAOMock{
				FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*iam.MemberRole, error) {
					return nil, errors.New("db down")
				},
			},
			withTenant: true,
			body:       `{"name":"Manager","code":"manager"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "returns server error on set permissions",
			roles: iam.MemberRoleDAOMock{
				FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*iam.MemberRole, error) {
					return nil, nil
				},
				CRUDMock: dao.CRUDMock[iam.MemberRole]{
					CreateFunc: func(_ context.Context, entity *iam.MemberRole) (*iam.MemberRole, error) {
						entity.ID = 1
						return entity, nil
					},
				},
				SetPermissionsFunc: func(_ context.Context, _ uint64, _ []uint64) error {
					return errors.New("db down")
				},
			},
			withTenant: true,
			body:       `{"name":"Manager","code":"manager"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := memberHandlerTest(t, iam.MemberDAOMock{}, tt.roles, iam.PermissionDAOMock{}, tt.withTenant)
			resp, err := doRequest(app, http.MethodPost, "/member-roles/", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}
