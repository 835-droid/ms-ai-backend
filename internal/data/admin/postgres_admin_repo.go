// ----- START OF FILE: backend/MS-AI/internal/data/admin/postgres_admin_repo.go -----
// internal/data/admin/postgres_admin_repo.go
package admin

import (
	"context"
	"database/sql"
	"errors"

	"github.com/835-droid/ms-ai-backend/internal/core/admin"
	coreuser "github.com/835-droid/ms-ai-backend/internal/core/user"
	"github.com/835-droid/ms-ai-backend/internal/data/postgres"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type adminRepository struct {
	store *postgres.PostgresStore
}

func NewPostgresAdminRepository(store *postgres.PostgresStore) admin.Repository {
	return &adminRepository{store: store}
}

func (r *adminRepository) CreateInvite(ctx context.Context, invite *coreuser.InviteCode) error {
	query := `INSERT INTO invite_codes (id, code, is_used, used_by, expires_at, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.store.DB.ExecContext(ctx, query,
		invite.ID.Hex(),
		invite.Code,
		invite.IsUsed,
		invite.UsedBy.Hex(),
		invite.ExpiresAt,
		invite.CreatedAt,
	)
	return err
}

func (r *adminRepository) ListInvites(ctx context.Context, skip, limit int64) ([]*coreuser.InviteCode, int64, error) {
	var total int64
	err := r.store.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM invite_codes`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, code, is_used, used_by, expires_at, created_at
	          FROM invite_codes ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.store.DB.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invites []*coreuser.InviteCode
	for rows.Next() {
		var inv coreuser.InviteCode
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

func (r *adminRepository) DeleteInvite(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid invite code id")
	}
	query := `DELETE FROM invite_codes WHERE id = $1`
	result, err := r.store.DB.ExecContext(ctx, query, oid.Hex())
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("invite code not found")
	}
	return nil
}

// ----- END OF FILE: backend/MS-AI/internal/data/admin/postgres_admin_repo.go -----
