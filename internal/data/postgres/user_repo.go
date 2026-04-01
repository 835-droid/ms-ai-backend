package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/835-droid/ms-ai-backend/internal/core/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userRepository struct {
	store *PostgresStore
}

func NewUserRepository(store *PostgresStore) user.Repository {
	return &userRepository{store: store}
}

func (r *userRepository) Create(ctx context.Context, u *user.User, details *user.UserDetails) error {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	rolesJSON, err := json.Marshal(u.Roles)
	if err != nil {
		return fmt.Errorf("marshal roles: %w", err)
	}
	detailsRolesJSON, err := json.Marshal(details.Roles)
	if err != nil {
		return fmt.Errorf("marshal details roles: %w", err)
	}

	query := `INSERT INTO users (id, username, password, roles, is_active, created_at, updated_at, refresh_token, refresh_token_expires_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = tx.ExecContext(ctx, query,
		u.ID.Hex(),
		u.Username,
		u.Password,
		string(rolesJSON),
		u.IsActive,
		u.CreatedAt,
		u.UpdatedAt,
		u.RefreshToken,
		u.RefreshTokenExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	detailsQuery := `INSERT INTO user_details (uuid, user_id, roles, is_active, status, created_at, updated_at)
	                 VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.ExecContext(ctx, detailsQuery,
		details.UUID,
		u.ID.Hex(),
		string(detailsRolesJSON),
		details.IsActive,
		details.Status,
		details.CreatedAt,
		details.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert user_details: %w", err)
	}

	return tx.Commit()
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	u := &user.User{}
	var idStr string
	var refreshToken sql.NullString
	var refreshExpiresAt sql.NullTime
	var rolesStr string

	query := `SELECT id, username, password, roles, is_active, created_at, updated_at, refresh_token, refresh_token_expires_at
	          FROM users WHERE username = $1`
	err := r.store.DB.QueryRowContext(ctx, query, username).Scan(
		&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive,
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
		u.Roles = []string{"user"}
	}
	if refreshToken.Valid {
		u.RefreshToken = refreshToken.String
	}
	if refreshExpiresAt.Valid {
		u.RefreshTokenExpiresAt = &refreshExpiresAt.Time
	}
	return u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*user.User, error) {
	u := &user.User{}
	var idStr string
	var refreshToken sql.NullString
	var refreshExpiresAt sql.NullTime
	var rolesStr string

	query := `SELECT id, username, password, roles, is_active, created_at, updated_at, refresh_token, refresh_token_expires_at
	          FROM users WHERE id = $1`
	err := r.store.DB.QueryRowContext(ctx, query, id.Hex()).Scan(
		&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive,
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
		u.Roles = []string{"user"}
	}
	if refreshToken.Valid {
		u.RefreshToken = refreshToken.String
	}
	if refreshExpiresAt.Valid {
		u.RefreshTokenExpiresAt = &refreshExpiresAt.Time
	}
	return u, nil
}

func (r *userRepository) Update(ctx context.Context, u *user.User) error {
	rolesJSON, err := json.Marshal(u.Roles)
	if err != nil {
		return fmt.Errorf("marshal roles: %w", err)
	}
	query := `UPDATE users SET username=$2, password=$3, roles=$4, is_active=$5, updated_at=$6, refresh_token=$7, refresh_token_expires_at=$8
	          WHERE id=$1`
	_, err = r.store.DB.ExecContext(ctx, query,
		u.ID.Hex(),
		u.Username,
		u.Password,
		string(rolesJSON),
		u.IsActive,
		u.UpdatedAt,
		u.RefreshToken,
		u.RefreshTokenExpiresAt,
	)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.store.DB.ExecContext(ctx, query, id.Hex())
	return err
}

// CreateInvite - تعديل used_by
func (r *userRepository) CreateInvite(ctx context.Context, invite *user.InviteCode) error {
	usedBy := interface{}(nil)
	if !invite.UsedBy.IsZero() {
		usedBy = invite.UsedBy.Hex()
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

func (r *userRepository) FindCode(ctx context.Context, code string) (*user.InviteCode, error) {
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

func (r *userRepository) UseCode(ctx context.Context, codeID primitive.ObjectID, userID primitive.ObjectID) error {
	query := `UPDATE invite_codes SET is_used = true, used_by = $2 WHERE id = $1`
	_, err := r.store.DB.ExecContext(ctx, query, codeID.Hex(), userID.Hex())
	return err
}

func (r *userRepository) FindAllInvites(ctx context.Context, page, limit int) ([]*user.InviteCode, error) {
	skip := (page - 1) * limit
	query := `SELECT id, code, is_used, used_by, expires_at, created_at
	          FROM invite_codes LIMIT $1 OFFSET $2`
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

func (r *userRepository) DeleteInvite(ctx context.Context, codeID primitive.ObjectID) error {
	query := `DELETE FROM invite_codes WHERE id = $1`
	_, err := r.store.DB.ExecContext(ctx, query, codeID.Hex())
	return err
}

func (r *userRepository) UpdateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt primitive.DateTime) error {
	expires := expiresAt.Time()
	query := `UPDATE users SET refresh_token=$2, refresh_token_expires_at=$3 WHERE id=$1`
	_, err := r.store.DB.ExecContext(ctx, query, userID.Hex(), token, expires)
	return err
}

func (r *userRepository) InvalidateRefreshToken(ctx context.Context, userID primitive.ObjectID) error {
	query := `UPDATE users SET refresh_token=NULL, refresh_token_expires_at=NULL WHERE id=$1`
	_, err := r.store.DB.ExecContext(ctx, query, userID.Hex())
	return err
}

func (r *userRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*user.User, error) {
	u := &user.User{}
	var idStr string
	var rolesStr string
	var refreshExpiresAt sql.NullTime

	query := `SELECT id, username, password, roles, is_active, created_at, updated_at, refresh_token_expires_at
	          FROM users WHERE refresh_token = $1 AND refresh_token_expires_at > NOW()`
	err := r.store.DB.QueryRowContext(ctx, query, refreshToken).Scan(
		&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive,
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
		u.Roles = []string{"user"}
	}
	u.RefreshToken = refreshToken
	if refreshExpiresAt.Valid {
		u.RefreshTokenExpiresAt = &refreshExpiresAt.Time
	}
	return u, nil
}

func (r *userRepository) GetNextSequence(ctx context.Context, sequenceName string) (int, error) {
	var seq int
	query := `SELECT nextval($1)`
	err := r.store.DB.QueryRowContext(ctx, query, sequenceName).Scan(&seq)
	if err != nil {
		return 0, err
	}
	return seq, nil
}

func (r *userRepository) FindAll(ctx context.Context, page, limit int) ([]*user.User, error) {
	skip := (page - 1) * limit
	query := `SELECT id, username, password, roles, is_active, created_at, updated_at, refresh_token, refresh_token_expires_at
	          FROM users LIMIT $1 OFFSET $2`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*user.User
	for rows.Next() {
		u := &user.User{}
		var idStr string
		var refreshToken sql.NullString
		var refreshExpiresAt sql.NullTime
		var rolesStr string
		err := rows.Scan(
			&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive,
			&u.CreatedAt, &u.UpdatedAt, &refreshToken, &refreshExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		u.ID, _ = primitive.ObjectIDFromHex(idStr)
		if err := json.Unmarshal([]byte(rolesStr), &u.Roles); err != nil {
			u.Roles = []string{"user"}
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

func (r *userRepository) FindAllUsers(ctx context.Context, skip, limit int64) ([]*user.User, int64, error) {
	var total int64
	err := r.store.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	query := `SELECT id, username, password, roles, is_active, created_at, updated_at, refresh_token, refresh_token_expires_at
	          FROM users LIMIT $1 OFFSET $2`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var users []*user.User
	for rows.Next() {
		u := &user.User{}
		var idStr string
		var refreshToken sql.NullString
		var refreshExpiresAt sql.NullTime
		var rolesStr string
		err := rows.Scan(
			&idStr, &u.Username, &u.Password, &rolesStr, &u.IsActive,
			&u.CreatedAt, &u.UpdatedAt, &refreshToken, &refreshExpiresAt,
		)
		if err != nil {
			return nil, 0, err
		}
		u.ID, _ = primitive.ObjectIDFromHex(idStr)
		if err := json.Unmarshal([]byte(rolesStr), &u.Roles); err != nil {
			u.Roles = []string{"user"}
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

func (r *userRepository) UpdateUserRole(ctx context.Context, userID primitive.ObjectID, role string, add bool) error {
	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var rolesStr string
	query := `SELECT roles FROM users WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, userID.Hex()).Scan(&rolesStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return fmt.Errorf("select user: %w", err)
	}

	var roles []string
	if err := json.Unmarshal([]byte(rolesStr), &roles); err != nil {
		roles = []string{}
	}

	if add {
		found := false
		for _, r := range roles {
			if r == role {
				found = true
				break
			}
		}
		if !found {
			roles = append(roles, role)
		}
	} else {
		newRoles := make([]string, 0, len(roles))
		for _, r := range roles {
			if r != role {
				newRoles = append(newRoles, r)
			}
		}
		roles = newRoles
	}

	newRolesJSON, err := json.Marshal(roles)
	if err != nil {
		return fmt.Errorf("marshal roles: %w", err)
	}
	updateQuery := `UPDATE users SET roles = $2, updated_at = NOW() WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, userID.Hex(), string(newRolesJSON))
	if err != nil {
		return fmt.Errorf("update roles: %w", err)
	}
	return tx.Commit()

}
