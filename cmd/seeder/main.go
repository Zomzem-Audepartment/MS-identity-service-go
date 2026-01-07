package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		// Fallback for local run without env vars if needed, or just fail
		if len(os.Getenv("DATABASE_URL")) > 0 {
			cfg.DatabaseURL = os.Getenv("DATABASE_URL")
		} else {
             // Default for container
			cfg.DatabaseURL = "postgres://postgres:postgres@identity-db:5432/identity?sslmode=disable"
        }
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

    // 1. Permissions
	perms := []struct {
		Code   string
		Module string
		Action string
		Name   string
	}{
		{"organization.view", "organization", "view", "View Organization"},
		{"organization.create", "organization", "create", "Create Organization Units"},
		{"organization.update", "organization", "update", "Update Organization Units"},
		{"organization.delete", "organization", "delete", "Delete Organization Units"},
		{"users.view", "users", "view", "View Users"},
		{"users.manage", "users", "manage", "Manage Users"},
	}

	permIDs := make(map[string]int)

	for _, p := range perms {
		var id int
		err := db.QueryRowContext(ctx, `
			INSERT INTO permissions (code, module, action, name) 
			VALUES ($1, $2, $3, $4) 
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
			RETURNING id`, p.Code, p.Module, p.Action, p.Name).Scan(&id)
		if err != nil {
			log.Printf("Error inserting permission %s: %v", p.Code, err)
            // Try fetch if error
            db.QueryRowContext(ctx, "SELECT id FROM permissions WHERE code = $1", p.Code).Scan(&id)
		}
		permIDs[p.Code] = id
	}

    // 2. Roles
    var adminRoleId int
    err = db.QueryRowContext(ctx, `
        INSERT INTO roles (code, name, is_system, level) VALUES ('SUPER_ADMIN', 'Super Admin', true, 0)
        ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
        RETURNING id`).Scan(&adminRoleId)
    if err != nil {
         db.QueryRowContext(ctx, "SELECT id FROM roles WHERE code = 'SUPER_ADMIN'").Scan(&adminRoleId)
    }

    // 3. Role Permissions
    for _, pid := range permIDs {
        _, err := db.ExecContext(ctx, `
            INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)
            ON CONFLICT (role_id, permission_id) DO NOTHING`, adminRoleId, pid)
        if err != nil {
            log.Printf("Error linking role-perm: %v", err)
        }
    }

    // 4. Assign to 'admin' user (or first user if admin missing)
    // Try update user 'admin' or 'hieu.tran' or whatever exists
    res, err := db.ExecContext(ctx, `UPDATE users SET role_id = $1 WHERE username = 'admin'`, adminRoleId)
    rows, _ := res.RowsAffected()
    log.Printf("Assigned Admin role to %d users (username='admin')", rows)
    
    if rows == 0 {
         // Assign to first user
         db.ExecContext(ctx, `UPDATE users SET role_id = $1 WHERE id = (SELECT id FROM users ORDER BY id ASC LIMIT 1)`, adminRoleId)
         log.Println("Assigned Admin role to first user found")
    }
}
