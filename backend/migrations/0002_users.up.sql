-- Таблица пользователей системы (сотрудники компании-клиента, НЕ наша админка).
-- org_id — привязка к организации. На on-premise всегда один и тот же,
-- но колонка нужна для будущей облачной мультитенантности (см. 0001_organizations).
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50) NOT NULL DEFAULT 'viewer',
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Email уникален в рамках одной организации.
    CONSTRAINT users_org_email_unique UNIQUE (org_id, email)
);

-- Индекс для быстрого поиска пользователя по email при логине.
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Проверка допустимых ролей на уровне БД.
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'manager', 'viewer'));
