package database

import (
	"errors"
	"fmt"
	"log/slog"

	// golang-migrate — библиотека для версионированных SQL-миграций.
	// Читает файлы вида 0001_name.up.sql / 0001_name.down.sql
	// и применяет их по порядку, запоминая текущую версию в таблице schema_migrations.
	"github.com/golang-migrate/migrate/v4"

	// Драйвер для применения миграций к PostgreSQL.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Драйвер для чтения миграций из локальных файлов.
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations применяет все новые миграции из указанной директории к БД.
// dsn — строка подключения к PostgreSQL.
// migrationsPath — путь к папке с .sql файлами (например, "file://migrations").
func RunMigrations(dsn string, migrationsPath string) error {
	// migrate.New создаёт экземпляр мигратора: откуда читать SQL + куда применять.
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("инициализация мигратора: %w", err)
	}
	// defer m.Close() освобождает ресурсы (соединение к БД, открытые файлы).
	defer m.Close()

	// Up() применяет все миграции, которые ещё не были применены.
	// Если все миграции уже на месте — возвращает ErrNoChange (не ошибка).
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("применение миграций: %w", err)
	}

	// Читаем текущую версию схемы для логирования.
	version, dirty, _ := m.Version()
	slog.Info("миграции применены", "версия", version, "dirty", dirty)

	return nil
}
