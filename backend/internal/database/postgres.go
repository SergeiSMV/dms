// Пакет database управляет подключением к PostgreSQL.
package database

import (
	"context"
	"fmt"

	// pgxpool — пул соединений к PostgreSQL от jackc/pgx.
	// Пул автоматически управляет открытием/закрытием коннектов,
	// переиспользует свободные соединения и ограничивает их количество.
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect создаёт пул соединений к PostgreSQL по DSN-строке.
// DSN (Data Source Name) — строка вида "postgres://user:pass@host:port/db?sslmode=disable".
// Возвращает *pgxpool.Pool, который потокобезопасен и используется во всём приложении.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	// pgxpool.New парсит DSN, создаёт пул и проверяет подключение.
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("создание пула соединений: %w", err)
	}

	// Ping проверяет, что БД действительно доступна (а не только DSN валиден).
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("подключение к БД: %w", err)
	}

	return pool, nil
}
