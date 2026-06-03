-- Create users table.
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(32) NOT NULL UNIQUE,
    nickname VARCHAR(64),
    password VARCHAR(255) NOT NULL,
    email VARCHAR(128) NOT NULL,
    phone VARCHAR(20),
    is_admin INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1,
    logined_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Upgrade existing databases created before soft delete support.
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Create indexes.
CREATE UNIQUE INDEX IF NOT EXISTS idx_name ON users(name);
CREATE INDEX IF NOT EXISTS idx_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Seed test users. Password: IamGo2026!Aa
INSERT INTO users (name, nickname, password, email, phone, is_admin, status, logined_at)
VALUES
    ('admin', 'admin', '$2a$10$YFtGAw4eInOKgOHBOwdrS.1.fJYnYEEpBiULwEPpHHTducg.nmvV2', 'admin@example.com', '13800138000', 1, 1, CURRENT_TIMESTAMP),
    ('user1', 'user1', '$2a$10$YFtGAw4eInOKgOHBOwdrS.1.fJYnYEEpBiULwEPpHHTducg.nmvV2', 'user1@example.com', '13800138001', 0, 1, CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;

-- casbin_rule is auto-created by gorm-adapter on application startup.
-- The application seeds default admin policies and binds admin to the admin role.
