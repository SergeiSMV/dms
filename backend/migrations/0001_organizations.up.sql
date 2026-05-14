-- Расширение для генерации UUID (v4) на стороне БД.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица организаций — корневая сущность мультитенантности.
-- На on-premise инсталляции здесь всегда одна запись.
-- Зачем тогда отдельная таблица? Чтобы при переходе на облачную версию
-- (несколько клиентов в одной БД) не переписывать схему с нуля.
-- org_id будет во всех таблицах — это «бесплатная» страховка на будущее.
CREATE TABLE IF NOT EXISTS organizations (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(255) NOT NULL,
    inn        VARCHAR(12),
    settings   JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
