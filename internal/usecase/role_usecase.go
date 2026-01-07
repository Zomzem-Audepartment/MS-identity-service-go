package usecase

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zomzem/identity-service/internal/repository"
)

type RoleUseCase interface {
	ListRoles(ctx context.Context) ([]RoleResponse, error)
	GetRoleByID(ctx context.Context, id int32) (*RoleResponse, error)
	CreateRole(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error)
	UpdateRole(ctx context.Context, id int32, req UpdateRoleRequest) (*RoleResponse, error)
	DeleteRole(ctx context.Context, id int32) error

	ListPermissions(ctx context.Context) ([]PermissionResponse, error)
	AssignPermission(ctx context.Context, roleID int32, req AssignPermissionRequest) error
	RemovePermission(ctx context.Context, roleID int32, permissionID int32) error
}

type roleUseCase struct {
	store repository.Store
}

func NewRoleUseCase(store repository.Store) RoleUseCase {
	return &roleUseCase{store: store}
}

type RoleResponse struct {
	ID          int32                    `json:"id"`
	Code        string                   `json:"code"`
	Name        string                   `json:"name"`
	Description *string                  `json:"description"`
	Level       int32                    `json:"level"`
	IsSystem    bool                     `json:"isSystem"`
	Status      string                   `json:"status"`
	Permissions []RolePermissionResponse `json:"permissions,omitempty"`
}

type PermissionResponse struct {
	ID     int32  `json:"id"`
	Module string `json:"module"`
	Action string `json:"action"`
	Code   string `json:"code"`
	Name   string `json:"name"`
}

type RolePermissionResponse struct {
	ID           int32               `json:"id"`
	RoleID       int32               `json:"roleId"`
	PermissionID int32               `json:"permissionId"`
	Permission   *PermissionResponse `json:"permission,omitempty"`
	DataScope    string              `json:"dataScope"`
}

type CreateRoleRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Level       *int32  `json:"level"`
	Status      *string `json:"status"`
}

type UpdateRoleRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Level       *int32  `json:"level"`
	Status      *string `json:"status"`
}

type AssignPermissionRequest struct {
	PermissionID int32  `json:"permissionId"`
	DataScope    string `json:"dataScope"`
}

func (u *roleUseCase) ListRoles(ctx context.Context) ([]RoleResponse, error) {
	roles, err := u.store.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		roleRes := u.mapRoleToResponse(r)
		
		// Load permissions for each role
		rp, err := u.store.ListRolePermissions(ctx, r.ID)
		if err == nil {
			roleRes.Permissions = make([]RolePermissionResponse, 0, len(rp))
			for _, p := range rp {
				roleRes.Permissions = append(roleRes.Permissions, RolePermissionResponse{
					ID:           p.ID,
					RoleID:       p.RoleID,
					PermissionID: p.PermissionID,
					DataScope:    p.DataScope.String,
					Permission: &PermissionResponse{
						ID:   p.PermissionID,
						Code: p.PermissionCode,
					},
				})
			}
		}
		
		res = append(res, roleRes)
	}
	return res, nil
}

func (u *roleUseCase) GetRoleByID(ctx context.Context, id int32) (*RoleResponse, error) {
	r, err := u.store.GetRoleById(ctx, id)
	if err != nil {
		return nil, err
	}

	roleRes := u.mapRoleToResponse(r)
	
	rp, err := u.store.ListRolePermissions(ctx, r.ID)
	if err == nil {
		roleRes.Permissions = make([]RolePermissionResponse, 0, len(rp))
		for _, p := range rp {
			roleRes.Permissions = append(roleRes.Permissions, RolePermissionResponse{
				ID:           p.ID,
				RoleID:       p.RoleID,
				PermissionID: p.PermissionID,
				DataScope:    p.DataScope.String,
				Permission: &PermissionResponse{
					ID:   p.PermissionID,
					Code: p.PermissionCode,
				},
			})
		}
	}

	return &roleRes, nil
}

func (u *roleUseCase) CreateRole(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error) {
	level := int32(100)
	if req.Level != nil {
		level = *req.Level
	}
	status := "ACTIVE"
	if req.Status != nil {
		status = *req.Status
	}

	r, err := u.store.CreateRole(ctx, repository.CreateRoleParams{
		Code:        req.Code,
		Name:        req.Name,
		Description: pgtype.Text{String: getString(req.Description), Valid: req.Description != nil},
		Level:       pgtype.Int4{Int32: level, Valid: true},
		IsSystem:    pgtype.Bool{Bool: false, Valid: true},
		Status:      pgtype.Text{String: status, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	res := u.mapRoleToResponse(r)
	return &res, nil
}

func (u *roleUseCase) UpdateRole(ctx context.Context, id int32, req UpdateRoleRequest) (*RoleResponse, error) {
	// Check if role exists
	existing, err := u.store.GetRoleById(ctx, id)
	if err != nil {
		return nil, err
	}

	level := existing.Level.Int32
	if req.Level != nil {
		level = *req.Level
	}
	status := existing.Status.String
	if req.Status != nil {
		status = *req.Status
	}

	r, err := u.store.UpdateRole(ctx, repository.UpdateRoleParams{
		ID:          id,
		Name:        req.Name,
		Description: pgtype.Text{String: getString(req.Description), Valid: req.Description != nil},
		Level:       pgtype.Int4{Int32: level, Valid: true},
		Status:      pgtype.Text{String: status, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	res := u.mapRoleToResponse(r)
	return &res, nil
}

func (u *roleUseCase) DeleteRole(ctx context.Context, id int32) error {
	return u.store.DeleteRole(ctx, id)
}

func (u *roleUseCase) ListPermissions(ctx context.Context) ([]PermissionResponse, error) {
	perms, err := u.store.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		res = append(res, PermissionResponse{
			ID:     p.ID,
			Module: p.Module,
			Action: p.Action,
			Code:   p.Code,
			Name:   p.Name,
		})
	}
	return res, nil
}

func (u *roleUseCase) AssignPermission(ctx context.Context, roleID int32, req AssignPermissionRequest) error {
	_, err := u.store.AssignPermissionToRole(ctx, repository.AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: req.PermissionID,
		DataScope:    pgtype.Text{String: req.DataScope, Valid: true},
	})
	return err
}

func (u *roleUseCase) RemovePermission(ctx context.Context, roleID int32, permissionID int32) error {
	return u.store.RemovePermissionFromRole(ctx, repository.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (u *roleUseCase) mapRoleToResponse(r repository.Role) RoleResponse {
	return RoleResponse{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Description: stringPtr(r.Description.String, r.Description.Valid),
		Level:       r.Level.Int32,
		IsSystem:    r.IsSystem.Bool,
		Status:      r.Status.String,
	}
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
