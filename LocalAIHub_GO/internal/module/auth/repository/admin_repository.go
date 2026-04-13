package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Admin struct {
	ID           int64
	Username     string
	PasswordHash string
	Status       string
	LastLoginAt  sql.NullTime
}

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) GetByUsername(ctx context.Context, username string) (*Admin, error) {
	query := `SELECT id, username, password_hash, status, last_login_at FROM admin_user WHERE username = ? LIMIT 1`
	var admin Admin
	if err := r.db.QueryRowContext(ctx, query, username).Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Status, &admin.LastLoginAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query admin by username: %w", err)
	}
	return &admin, nil
}

func (r *AdminRepository) CreateAdmin(ctx context.Context, username, passwordHash string) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO admin_user (username, password_hash, password_algo, status, created_at, updated_at) VALUES (?, ?, 'bcrypt', 'active', ?, ?)`, username, passwordHash, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create admin: %w", err)
	}
	return result.LastInsertId()
}

func (r *AdminRepository) GetByID(ctx context.Context, id int64) (*Admin, error) {
	query := `SELECT id, username, password_hash, status, last_login_at FROM admin_user WHERE id = ? LIMIT 1`
	var admin Admin
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Status, &admin.LastLoginAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query admin by id: %w", err)
	}
	return &admin, nil
}

func (r *AdminRepository) HasAnyAdmin(ctx context.Context) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_user`).Scan(&count); err != nil {
		return false, fmt.Errorf("count admins: %w", err)
	}
	return count > 0, nil
}

func (r *AdminRepository) UpdateLastLogin(ctx context.Context, adminID int64, ip string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE admin_user SET last_login_at = ?, last_login_ip = ?, updated_at = ? WHERE id = ?`, time.Now().UTC(), ip, time.Now().UTC(), adminID)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	return nil
}

func (r *AdminRepository) CreateLoginLog(ctx context.Context, adminID *int64, username, action, ip, userAgent, result string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO admin_login_log (admin_user_id, username, action, ip_address, user_agent, result_message, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, nullableInt64(adminID), username, action, ip, userAgent, result, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert admin login log: %w", err)
	}
	return nil
}

func (r *AdminRepository) EnsureDefaultAdmin(ctx context.Context, username, passwordHash string) error {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_user WHERE username = ?`, username).Scan(&count); err != nil {
		return fmt.Errorf("count default admin: %w", err)
	}
	if count > 0 {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO admin_user (username, password_hash, password_algo, status, created_at, updated_at) VALUES (?, ?, 'bcrypt', 'active', ?, ?)`, username, passwordHash, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert default admin: %w", err)
	}
	return nil
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
