// internal/data/user/postgres_user_repository.go
package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/user"
	pginfra "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostgresUserRepository struct {
	store *pginfra.PostgresStore
}

func NewPostgresUserRepository(store *pginfra.PostgresStore) user.Repository {
	return &PostgresUserRepository{store: store}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *user.User, details *user.UserDetails) error {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	rolesJSON, err := json.Marshal(u.Roles)
	if err != nil {
		return fmt.Errorf("marshal roles: %w", err)
	}

	query := `INSERT INTO users (id, username, password, roles, is_active, last_login_at, created_at, updated_at, refresh_token, refresh_token_expires_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.ExecContext(ctx, query,
		u.ID.Hex(),
		u.Username,
		u.Password,
		string(rolesJSON),
		u.IsActive,
		u.LastLoginAt,
		u.CreatedAt,
		u.UpdatedAt,
		u.RefreshToken,
		u.RefreshTokenExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	detailsQuery := `INSERT INTO user_details (uuid, user_id, status, created_at, updated_at)
	                 VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, detailsQuery,
		details.UUID,
		u.ID.Hex(),
		details.Status,
		details.CreatedAt,
		details.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert user_details: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	u := &user.User{}
	var idStr string
	var refreshToken sql.NullString
	var refreshExpiresAt sql.NullTime
	var lastLoginAt sql.NullTime
	var rolesStr string

	query := `SELECT id, username, password, roles, is_active, last_login_at, created_at, updated_at, refresh_token, refresh_token_expires_at
	          FROM users WHERE username = $1`
	err := r.store.DB.QueryRowContext(ctx, query, username).Scan(
		&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive, &lastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &refreshToken, &refreshExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find by username: %w", err)
	}
	u.ID, _ = primitive.ObjectIDFromHex(idStr)
	if err := json.Unmarshal([]byte(rolesStr), &u.Roles); err != nil {
		u.Roles = user.FromStrings([]string{"user"})
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	if refreshToken.Valid {
		u.RefreshToken = refreshToken.String
	}
	if refreshExpiresAt.Valid {
		u.RefreshTokenExpiresAt = &refreshExpiresAt.Time
	}
	return u, nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*user.User, error) {
	u := &user.User{}
	var idStr string
	var refreshToken sql.NullString
	var refreshExpiresAt sql.NullTime
	var lastLoginAt sql.NullTime
	var rolesStr string

	query := `SELECT id, username, password, roles, is_active, last_login_at, created_at, updated_at, refresh_token, refresh_token_expires_at
	          FROM users WHERE id = $1`
	err := r.store.DB.QueryRowContext(ctx, query, id.Hex()).Scan(
		&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive, &lastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &refreshToken, &refreshExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find by id: %w", err)
	}
	u.ID, _ = primitive.ObjectIDFromHex(idStr)
	if err := json.Unmarshal([]byte(rolesStr), &u.Roles); err != nil {
		u.Roles = user.FromStrings([]string{"user"})
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	if refreshToken.Valid {
		u.RefreshToken = refreshToken.String
	}
	if refreshExpiresAt.Valid {
		u.RefreshTokenExpiresAt = &refreshExpiresAt.Time
	}
	return u, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *user.User) error {
	rolesJSON, err := json.Marshal(u.Roles)
	if err != nil {
		return fmt.Errorf("marshal roles: %w", err)
	}
	query := `UPDATE users SET username=$2, password=$3, roles=$4, is_active=$5, last_login_at=$6, updated_at=$7, refresh_token=$8, refresh_token_expires_at=$9
	          WHERE id=$1`
	result, err := r.store.DB.ExecContext(ctx, query,
		u.ID.Hex(),
		u.Username,
		u.Password,
		string(rolesJSON),
		u.IsActive,
		u.LastLoginAt,
		u.UpdatedAt,
		u.RefreshToken,
		u.RefreshTokenExpiresAt,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return core.ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.store.DB.ExecContext(ctx, query, id.Hex())
	return err
}

func (r *PostgresUserRepository) FindAll(ctx context.Context, page, limit int) ([]*user.User, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `SELECT id, username, password, roles, is_active, last_login_at, created_at, updated_at, refresh_token, refresh_token_expires_at
	          FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		u := &user.User{}
		var idStr string
		var refreshToken sql.NullString
		var refreshExpiresAt sql.NullTime
		var lastLoginAt sql.NullTime
		var rolesStr string

		err := rows.Scan(
			&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive, &lastLoginAt,
			&u.CreatedAt, &u.UpdatedAt, &refreshToken, &refreshExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		u.ID, _ = primitive.ObjectIDFromHex(idStr)
		if err := json.Unmarshal([]byte(rolesStr), &u.Roles); err != nil {
			u.Roles = user.FromStrings([]string{"user"})
		}
		if lastLoginAt.Valid {
			u.LastLoginAt = &lastLoginAt.Time
		}
		if refreshToken.Valid {
			u.RefreshToken = refreshToken.String
		}
		if refreshExpiresAt.Valid {
			u.RefreshTokenExpiresAt = &refreshExpiresAt.Time
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *PostgresUserRepository) FindAllUsers(ctx context.Context, skip, limit int64) ([]*user.User, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.store.DB.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	// Get users
	query := `SELECT id, username, password, roles, is_active, last_login_at, created_at, updated_at, refresh_token, refresh_token_expires_at
	          FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("find users: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		u := &user.User{}
		var idStr string
		var refreshToken sql.NullString
		var refreshExpiresAt sql.NullTime
		var lastLoginAt sql.NullTime
		var rolesStr string

		err := rows.Scan(
			&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive, &lastLoginAt,
			&u.CreatedAt, &u.UpdatedAt, &refreshToken, &refreshExpiresAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		u.ID, _ = primitive.ObjectIDFromHex(idStr)
		if err := json.Unmarshal([]byte(rolesStr), &u.Roles); err != nil {
			u.Roles = user.FromStrings([]string{"user"})
		}
		if lastLoginAt.Valid {
			u.LastLoginAt = &lastLoginAt.Time
		}
		if refreshToken.Valid {
			u.RefreshToken = refreshToken.String
		}
		if refreshExpiresAt.Valid {
			u.RefreshTokenExpiresAt = &refreshExpiresAt.Time
		}
		users = append(users, u)
	}
	return users, total, nil
}
