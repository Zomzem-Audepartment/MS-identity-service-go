package usecase

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zomzem/identity-service/internal/repository"
)

type UserUseCase interface {
	ListUsers(ctx context.Context) ([]UserResponseWithRole, error)
	GetUserByID(ctx context.Context, id int32) (*UserResponseWithRole, error)
	CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponseWithRole, error)
	UpdateUser(ctx context.Context, id int32, req UpdateUserRequest) (*UserResponseWithRole, error)
	DeleteUser(ctx context.Context, id int32) error
}

type userUseCase struct {
	store repository.Store
}

func NewUserUseCase(store repository.Store) UserUseCase {
	return &userUseCase{store: store}
}

type UserResponseWithRole struct {
	ID           int32   `json:"id"`
	Username     string  `json:"username"`
	FullName     string  `json:"fullName"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	Avatar       *string `json:"avatar"`
	Status       string  `json:"status"`
	RoleID       *int32  `json:"roleId"`
	RoleName     *string `json:"roleName"`
	RoleCode     *string `json:"roleCode"`
	EmployeeCode *string `json:"employeeCode"`
}

type CreateUserRequest struct {
	Username string  `json:"username"`
	FullName string  `json:"fullName"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Avatar   *string `json:"avatar"`
	RoleID   *int32  `json:"roleId"`
	Status   *string `json:"status"`
}

type UpdateUserRequest struct {
	FullName string  `json:"fullName"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Avatar   *string `json:"avatar"`
	RoleID   *int32  `json:"roleId"`
	Status   *string `json:"status"`
}

func (u *userUseCase) ListUsers(ctx context.Context) ([]UserResponseWithRole, error) {
	users, err := u.store.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]UserResponseWithRole, 0, len(users))
	for _, user := range users {
		res = append(res, UserResponseWithRole{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
			Email:    stringPtr(user.Email.String, user.Email.Valid),
			Phone:    stringPtr(user.Phone.String, user.Phone.Valid),
			Avatar:   stringPtr(user.Avatar.String, user.Avatar.Valid),
			Status:   user.Status.String,
			RoleID:   int32Ptr(user.RoleID.Int32, user.RoleID.Valid),
			RoleName: stringPtr(user.RoleName.String, user.RoleName.Valid),
			RoleCode: stringPtr(user.RoleCode.String, user.RoleCode.Valid),
		})
	}
	return res, nil
}

func (u *userUseCase) GetUserByID(ctx context.Context, id int32) (*UserResponseWithRole, error) {
	user, err := u.store.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}

	// For single user, we might want role info, but GetUserById only returns User model.
	// We could call GetRoleById or update the query. Let's keep it simple for now or update query.
	// Actually ListUsers is usually enough for the UI.
	
	res := UserResponseWithRole{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Email:    stringPtr(user.Email.String, user.Email.Valid),
		Phone:    stringPtr(user.Phone.String, user.Phone.Valid),
		Avatar:   stringPtr(user.Avatar.String, user.Avatar.Valid),
		Status:   user.Status.String,
		RoleID:   int32Ptr(user.RoleID.Int32, user.RoleID.Valid),
	}
	return &res, nil
}

func (u *userUseCase) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponseWithRole, error) {
	status := "ACTIVE"
	if req.Status != nil {
		status = *req.Status
	}

	user, err := u.store.CreateUser(ctx, repository.CreateUserParams{
		Username: req.Username,
		FullName: req.FullName,
		Email:    pgtype.Text{String: getString(req.Email), Valid: req.Email != nil},
		Phone:    pgtype.Text{String: getString(req.Phone), Valid: req.Phone != nil},
		Avatar:   pgtype.Text{String: getString(req.Avatar), Valid: req.Avatar != nil},
		RoleID:   pgtype.Int4{Int32: getInt32(req.RoleID), Valid: req.RoleID != nil},
		Status:   pgtype.Text{String: status, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &UserResponseWithRole{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Status:   user.Status.String,
	}, nil
}

func (u *userUseCase) UpdateUser(ctx context.Context, id int32, req UpdateUserRequest) (*UserResponseWithRole, error) {
	user, err := u.store.UpdateUser(ctx, repository.UpdateUserParams{
		ID:       id,
		FullName: req.FullName,
		Email:    pgtype.Text{String: getString(req.Email), Valid: req.Email != nil},
		Phone:    pgtype.Text{String: getString(req.Phone), Valid: req.Phone != nil},
		Avatar:   pgtype.Text{String: getString(req.Avatar), Valid: req.Avatar != nil},
		RoleID:   pgtype.Int4{Int32: getInt32(req.RoleID), Valid: req.RoleID != nil},
		Status:   pgtype.Text{String: getString(req.Status), Valid: req.Status != nil},
	})
	if err != nil {
		return nil, err
	}

	return &UserResponseWithRole{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Status:   user.Status.String,
	}, nil
}

func (u *userUseCase) DeleteUser(ctx context.Context, id int32) error {
	return u.store.DeleteUser(ctx, id)
}

func int32Ptr(i int32, valid bool) *int32 {
	if !valid {
		return nil
	}
	return &i
}

func getInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
