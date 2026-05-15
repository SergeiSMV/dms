package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"dms-backend/internal/auth"
)

// SeedResult — результат начальной инициализации БД.
type SeedResult struct {
	OrgID string // ID созданной/существующей организации
}

// Seed создаёт организацию и admin-пользователя, если БД пуста.
// На on-premise всегда одна организация. Если уже есть — возвращает её ID.
func Seed(ctx context.Context, pool *pgxpool.Pool) (*SeedResult, error) {
	// Проверяем, есть ли хотя бы одна организация.
	var orgID string
	err := pool.QueryRow(ctx, `SELECT id FROM organizations LIMIT 1`).Scan(&orgID)
	if err == nil {
		// Организация уже есть — seed не нужен.
		slog.Info("организация найдена", "org_id", orgID)
		return &SeedResult{OrgID: orgID}, nil
	}

	// Создаём организацию по умолчанию.
	slog.Info("создание начальной организации и admin-пользователя")

	err = pool.QueryRow(ctx,
		`INSERT INTO organizations (name, inn) VALUES ($1, $2) RETURNING id`,
		"Моя компания", "",
	).Scan(&orgID)
	if err != nil {
		return nil, fmt.Errorf("создание организации: %w", err)
	}

	// Хешируем пароль по умолчанию для admin.
	hash, err := auth.HashPassword("admin")
	if err != nil {
		return nil, fmt.Errorf("хеширование пароля admin: %w", err)
	}

	// Создаём admin-пользователя.
	_, err = pool.Exec(ctx,
		`INSERT INTO users (org_id, email, password_hash, role, is_active) VALUES ($1, $2, $3, $4, $5)`,
		orgID, "admin@dms.local", hash, "admin", true,
	)
	if err != nil {
		return nil, fmt.Errorf("создание admin: %w", err)
	}

	slog.Info("seed выполнен", "org_id", orgID, "admin_email", "admin@dms.local")
	return &SeedResult{OrgID: orgID}, nil
}
