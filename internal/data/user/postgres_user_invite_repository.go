// internal/data/user/postgres_user_invite_repository.go
package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/835-droid/ms-ai-backend/internal/core/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *PostgresUserRepository) CreateInvite(ctx context.Context, invite *user.InviteCode) error {
	usedBy := interface{}(nil)
	if !invite.UsedBy.IsZero() {
		usedBy = invite.UsedBy.Hex()
	}

	if invite.ID.IsZero() {
		invite.ID = primitive.NewObjectID()
	}

	query := `INSERT INTO invite_codes (id, code, is_used, used_by, expires_at, created_at)
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.store.DB.ExecContext(ctx, query,
		invite.ID.Hex(),
		invite.Code,
		invite.IsUsed,
		usedBy,
		invite.ExpiresAt,
		invite.CreatedAt,
	)
	return err
}

func (r *PostgresUserRepository) FindCode(ctx context.Context, code string) (*user.InviteCode, error) {
	var inv user.InviteCode
	var idStr string
	var usedByNull sql.NullString

	query := `SELECT id, code, is_used, used_by, expires_at, created_at
	          FROM invite_codes WHERE code = $1`
	err := r.store.DB.QueryRowContext(ctx, query, code).Scan(
		&idStr, &inv.Code, &inv.IsUsed, &usedByNull, &inv.ExpiresAt, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	inv.ID, _ = primitive.ObjectIDFromHex(idStr)
	if usedByNull.Valid {
		inv.UsedBy, _ = primitive.ObjectIDFromHex(usedByNull.String)
	}
	return &inv, nil
}

func (r *PostgresUserRepository) UseCode(ctx context.Context, codeID primitive.ObjectID, userID primitive.ObjectID) error {
	query := `UPDATE invite_codes SET is_used = true, used_by = $2 WHERE id = $1`
	_, err := r.store.DB.ExecContext(ctx, query, codeID.Hex(), userID.Hex())
	return err
}

func (r *PostgresUserRepository) FindAllInvites(ctx context.Context, page, limit int) ([]*user.InviteCode, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	skip := (page - 1) * limit
	query := `SELECT id, code, is_used, used_by, expires_at, created_at
	          FROM invite_codes ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*user.InviteCode
	for rows.Next() {
		var inv user.InviteCode
		var idStr string
		var usedByNull sql.NullString
		err := rows.Scan(&idStr, &inv.Code, &inv.IsUsed, &usedByNull, &inv.ExpiresAt, &inv.CreatedAt)
		if err != nil {
			return nil, err
		}
		inv.ID, _ = primitive.ObjectIDFromHex(idStr)
		if usedByNull.Valid {
			inv.UsedBy, _ = primitive.ObjectIDFromHex(usedByNull.String)
		}
		invites = append(invites, &inv)
	}
	return invites, nil
}

func (r *PostgresUserRepository) FindAllInvitesWithTotal(ctx context.Context, skip, limit int64) ([]*user.InviteCode, int64, error) {
	if skip < 0 {
		skip = 0
	}
	if limit <= 0 {
		limit = 20
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM invite_codes`
	if err := r.store.DB.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, code, is_used, used_by, expires_at, created_at
	          FROM invite_codes ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invites []*user.InviteCode
	for rows.Next() {
		var inv user.InviteCode
		var idStr string
		var usedByNull sql.NullString
		err := rows.Scan(&idStr, &inv.Code, &inv.IsUsed, &usedByNull, &inv.ExpiresAt, &inv.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		inv.ID, _ = primitive.ObjectIDFromHex(idStr)
		if usedByNull.Valid {
			inv.UsedBy, _ = primitive.ObjectIDFromHex(usedByNull.String)
		}
		invites = append(invites, &inv)
	}
	return invites, total, nil
}

func (r *PostgresUserRepository) DeleteInvite(ctx context.Context, codeID primitive.ObjectID) error {
	query := `DELETE FROM invite_codes WHERE id = $1`
	_, err := r.store.DB.ExecContext(ctx, query, codeID.Hex())
	return err
}

func (r *PostgresUserRepository) CreateUserWithInvite(ctx context.Context, u *user.User, details *user.UserDetails, inviteCode string) error {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var inviteID string
	var isUsed bool
	var expiresAt time.Time
	query := `SELECT id, is_used, expires_at FROM invite_codes WHERE code = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, query, inviteCode).Scan(&inviteID, &isUsed, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.ErrInvalidInviteCode
		}
		return fmt.Errorf("select invite code: %w", err)
	}

	if isUsed {
		return core.ErrInviteCodeUsed
	}
	if time.Now().After(expiresAt) {
		return core.ErrInviteCodeExpired
	}

	if err := r.createUserInTx(ctx, tx, u, details); err != nil {
		return err
	}

	updateQuery := `UPDATE invite_codes SET is_used = true, used_by = $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateQuery, u.ID.Hex(), inviteID); err != nil {
		return fmt.Errorf("mark invite used: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresUserRepository) createUserInTx(ctx context.Context, tx *sql.Tx, u *user.User, details *user.UserDetails) error {
	ownTx := false
	if tx == nil {
		var err error
		tx, err = r.store.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		ownTx = true
		defer tx.Rollback()
	}

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

	if ownTx {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}
	}
	return nil
}
