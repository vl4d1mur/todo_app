-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS tasks (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id           UUID NOT NULL,
    title             VARCHAR(200) NOT NULL,
    description       TEXT,
    status            VARCHAR(20) NOT NULL DEFAULT 'todo',
    priority          INTEGER NOT NULL DEFAULT 2 CHECK (priority BETWEEN 1 AND 4),
    deadline          TIMESTAMPTZ,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_tasks_user_id;
DROP TABLE IF EXISTS tasks;