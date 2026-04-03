package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/835-droid/ms-ai-backend/internal/core/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *PostgresUserRepository) GetNextSequence(ctx context.Context, sequenceName string) (int, error) {
	query := `INSERT INTO sequences (name, value) VALUES ($1, 1) ON CONFLICT (name) DO UPDATE SET value = sequences.value + 1 RETURNING value`
	var value int
	err := r.store.DB.QueryRowContext(ctx, query, sequenceName).Scan(&value)
	if err != nil {
		return 0, fmt.Errorf("get next sequence: %w", err)
	}
	return value, nil
}

func (r *PostgresUserRepository) UpdateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt primitive.DateTime) error {
	expires := expiresAt.Time()
	query := `UPDATE users SET refresh_token=$2, refresh_token_expires_at=$3 WHERE id=$1`
	_, err := r.store.DB.ExecContext(ctx, query, userID.Hex(), token, expires)
	return err
}

func (r *PostgresUserRepository) InvalidateRefreshToken(ctx context.Context, userID primitive.ObjectID) error {
	query := `UPDATE users SET refresh_token=NULL, refresh_token_expires_at=NULL WHERE id=$1`
	_, err := r.store.DB.ExecContext(ctx, query, userID.Hex())
	return err
}

func (r *PostgresUserRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*user.User, error) {
	u := &user.User{}
	var idStr string
	var rolesStr string
	var refreshExpiresAt sql.NullTime
	var lastLoginAt sql.NullTime

	query := `SELECT id, username, password, roles, is_active, last_login_at, created_at, updated_at, refresh_token_expires_at
		  FROM users WHERE refresh_token = $1 AND refresh_token_expires_at > NOW() AND is_active = true`
	err := r.store.DB.QueryRowContext(ctx, query, refreshToken).Scan(
		&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive, &lastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &refreshExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u.ID, _ = primitive.ObjectIDFromHex(idStr)
	if err := json.Unmarshal([]byte(rolesStr), &u.Roles); err != nil {
		u.Roles = user.FromStrings([]string{"user"})
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	u.RefreshToken = refreshToken
	if refreshExpiresAt.Valid {
		u.RefreshTokenExpiresAt = &refreshExpiresAt.Time
	}
	return u, nil
}
