package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zomzem/identity-service/internal/config"
	"github.com/zomzem/identity-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type AuthUseCase interface {
	Login(ctx context.Context, username, password string) (*LoginResponse, error)
	LoginGoogle(ctx context.Context, idToken string) (*LoginResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*LoginResponse, error)
}

type authUseCase struct {
	store  repository.Store
	config *config.Config
}

func NewAuthUseCase(store repository.Store, cfg *config.Config) AuthUseCase {
	return &authUseCase{store: store, config: cfg}
}

type LoginResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID          int32    `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"fullName"`
	Avatar      *string  `json:"avatar"`
	Email       *string  `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

func (u *authUseCase) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	// 1. Get User
	user, err := u.store.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials") // Don't leak exists or not
	}

	// 2. Verify Password
	if user.PasswordHash.Valid {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)); err != nil {
			return nil, errors.New("invalid credentials")
		}
	} else {
		// External login or no password
		return nil, errors.New("password not set")
	}

	// 3. Get Permissions
	var permissions []string
	perms, err := u.store.GetUserPermissions(ctx, user.ID)
	if err == nil {
		for _, p := range perms {
			permissions = append(permissions, p.PermissionCode)
		}
	}

	// 4. Generate Token

	accessToken, refreshToken, err := u.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	// 4. Update Last Login
	_ = u.store.UpdateUserLastLogin(ctx, repository.UpdateUserLastLoginParams{
		ID:          user.ID,
		LastLoginIp: pgtype.Text{Valid: false},
	})

	// 5. Get Role Code
	roleCode := ""
	if user.RoleID.Valid {
		role, err := u.store.GetRoleById(ctx, user.RoleID.Int32)
		if err == nil {
			roleCode = role.Code
		}
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			FullName:    user.FullName,
			Avatar:      stringPtr(user.Avatar.String, user.Avatar.Valid),
			Email:       stringPtr(user.Email.String, user.Email.Valid),
			Role:        roleCode,
			Permissions: permissions,
		},
	}, nil
}

func (u *authUseCase) LoginGoogle(ctx context.Context, idTokenStr string) (*LoginResponse, error) {
	// 1. Verify Google Token
	payload, err := idtoken.Validate(ctx, idTokenStr, u.config.GoogleClientID)
	if err != nil {
		log.Printf("[Auth] Google ID Token validation failed: %v (ClientID: %s)", err, u.config.GoogleClientID)
		return nil, errors.New("invalid google token")
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	// 2. Check if user exists
	user, err := u.store.GetUserByEmail(ctx, pgtype.Text{String: email, Valid: true})
	if err != nil {
		// Create user if not found
		username := email
		user, err = u.store.CreateUser(ctx, repository.CreateUserParams{
			Username:     username,
			PasswordHash: pgtype.Text{Valid: false},
			FullName:     name,
			Email:        pgtype.Text{String: email, Valid: true},
			Avatar:       pgtype.Text{String: picture, Valid: true},
			Status:       pgtype.Text{String: "ACTIVE", Valid: true},
		})
		if err != nil {
			return nil, err
		}
		
		// Set as external login
		_ = u.store.UpdateUserExternalLogin(ctx, repository.UpdateUserExternalLoginParams{
			ID:            user.ID,
			ExternalLogin: pgtype.Bool{Bool: true, Valid: true},
		})
	}

	// 3. Generate Tokens
	accessToken, refreshToken, err := u.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	// 5. Update Last Login
	_ = u.store.UpdateUserLastLogin(ctx, repository.UpdateUserLastLoginParams{
		ID:          user.ID,
		LastLoginIp: pgtype.Text{Valid: false},
	})

	// 6. Get Permissions & Role
	var permissions []string
	perms, _ := u.store.GetUserPermissions(ctx, user.ID)
	for _, p := range perms {
		permissions = append(permissions, p.PermissionCode)
	}

	roleCode := ""
	if user.RoleID.Valid {
		role, err := u.store.GetRoleById(ctx, user.RoleID.Int32)
		if err == nil {
			roleCode = role.Code
		}
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			FullName:    user.FullName,
			Avatar:      stringPtr(user.Avatar.String, user.Avatar.Valid),
			Email:       stringPtr(user.Email.String, user.Email.Valid),
			Role:        roleCode,
			Permissions: permissions,
		},
	}, nil
}

func (u *authUseCase) Refresh(ctx context.Context, refreshTokenStr string) (*LoginResponse, error) {
	// 1. Verify Refresh Token in DB
	rt, err := u.store.GetRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// 2. Get User
	user, err := u.store.GetUserById(ctx, rt.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 3. Revoke current token
	_ = u.store.RevokeRefreshToken(ctx, refreshTokenStr)

	// 4. Generate new tokens
	accessToken, newRefreshToken, err := u.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	// 5. Get Permissions & Role
	var permissions []string
	perms, _ := u.store.GetUserPermissions(ctx, user.ID)
	for _, p := range perms {
		permissions = append(permissions, p.PermissionCode)
	}

	roleCode := ""
	if user.RoleID.Valid {
		role, err := u.store.GetRoleById(ctx, user.RoleID.Int32)
		if err == nil {
			roleCode = role.Code
		}
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User: UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			FullName:    user.FullName,
			Avatar:      stringPtr(user.Avatar.String, user.Avatar.Valid),
			Email:       stringPtr(user.Email.String, user.Email.Valid),
			Role:        roleCode,
			Permissions: permissions,
		},
	}, nil
}

func (u *authUseCase) generateTokens(ctx context.Context, user repository.User) (string, string, error) {
	// 1. Get Permissions
	var permissions []string
	perms, err := u.store.GetUserPermissions(ctx, user.ID)
	if err == nil {
		for _, p := range perms {
			permissions = append(permissions, p.PermissionCode)
		}
	}

	// 2. Access Token
	claims := jwt.MapClaims{
		"userId":      user.ID,
		"sub":         user.Username,
		"permissions": permissions,
		"exp":         time.Now().Add(15 * time.Minute).Unix(),
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := at.SignedString([]byte(u.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	// 3. Refresh Token
	refreshTokenStr := fmt.Sprintf("%d-%d", user.ID, time.Now().UnixNano()) // Simple token for now
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = u.store.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenStr, nil
}

func stringPtr(s string, valid bool) *string {
	if !valid {
		return nil
	}
	return &s
}
