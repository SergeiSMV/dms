package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"dms-backend/internal/appctx"
	"dms-backend/internal/model"
)

// OrgRepository выполняет SQL-запросы к таблице organizations.
type OrgRepository struct {
	pool *pgxpool.Pool
}

// NewOrgRepository создаёт репозиторий с пулом соединений.
func NewOrgRepository(pool *pgxpool.Pool) *OrgRepository {
	return &OrgRepository{pool: pool}
}

// Get возвращает организацию по org_id из контекста.
func (r *OrgRepository) Get(ctx context.Context) (*model.Organization, error) {
	orgID := appctx.GetOrgID(ctx)

	query := `
		SELECT id, name, inn, settings, created_at, updated_at
		FROM organizations
		WHERE id = $1`

	o := &model.Organization{}
	err := r.pool.QueryRow(ctx, query, orgID).Scan(
		&o.ID, &o.Name, &o.INN, &o.Settings, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("получение организации: %w", err)
	}
	return o, nil
}

// Update обновляет название, ИНН и настройки организации.
// Обновляет только организацию из контекста (org_id).
func (r *OrgRepository) Update(ctx context.Context, org *model.Organization) error {
	orgID := appctx.GetOrgID(ctx)

	query := `
		UPDATE organizations
		SET name = $1, inn = $2, settings = $3, updated_at = now()
		WHERE id = $4
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		org.Name, org.INN, org.Settings, orgID,
	).Scan(&org.UpdatedAt)
	if err != nil {
		return fmt.Errorf("обновление организации: %w", err)
	}
	return nil
}
