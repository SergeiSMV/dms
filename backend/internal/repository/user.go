// Пакет repository — слой доступа к данным (SQL-запросы через pgx).
// Каждый метод извлекает org_id из контекста для автоматической фильтрации по организации.
package repository

import (
	"context"
	"fmt"

	// pgxpool — пул соединений PostgreSQL.
	"github.com/jackc/pgx/v5/pgxpool"

	"dms-backend/internal/appctx"
	"dms-backend/internal/model"
)

// UserRepository выполняет SQL-запросы к таблице users.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository создаёт репозиторий с пулом соединений.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create вставляет нового пользователя. OrgID берётся из контекста.
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	orgID := appctx.GetOrgID(ctx)

	query := `
		INSERT INTO users (org_id, email, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		orgID, user.Email, user.PasswordHash, user.Role, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByEmail ищет пользователя по email в рамках организации из контекста.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	orgID := appctx.GetOrgID(ctx)

	query := `
		SELECT id, org_id, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE org_id = $1 AND email = $2`

	u := &model.User{}
	err := r.pool.QueryRow(ctx, query, orgID, email).Scan(
		&u.ID, &u.OrgID, &u.Email, &u.PasswordHash,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("поиск пользователя по email: %w", err)
	}
	return u, nil
}

// GetByID возвращает пользователя по UUID с фильтрацией по org_id из контекста.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	orgID := appctx.GetOrgID(ctx)

	query := `
		SELECT id, org_id, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND org_id = $2`

	u := &model.User{}
	err := r.pool.QueryRow(ctx, query, id, orgID).Scan(
		&u.ID, &u.OrgID, &u.Email, &u.PasswordHash,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("поиск пользователя по id: %w", err)
	}
	return u, nil
}

// List возвращает всех пользователей организации из контекста.
func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	orgID := appctx.GetOrgID(ctx)

	query := `
		SELECT id, org_id, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE org_id = $1
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("список пользователей: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID, &u.OrgID, &u.Email, &u.PasswordHash,
			&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("чтение строки пользователя: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

// UpdatePassword обновляет хеш пароля пользователя с фильтрацией по org_id из контекста.
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	orgID := appctx.GetOrgID(ctx)

	query := `UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2 AND org_id = $3`
	tag, err := r.pool.Exec(ctx, query, passwordHash, userID, orgID)
	if err != nil {
		return fmt.Errorf("обновление пароля: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("пользователь не найден: %s", userID)
	}
	return nil
}

// Count возвращает количество пользователей в организации из контекста.
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	orgID := appctx.GetOrgID(ctx)

	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE org_id = $1`, orgID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("подсчёт пользователей: %w", err)
	}
	return count, nil
}
