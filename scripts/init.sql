-- 创建用户表
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_name ON users(name);
CREATE INDEX IF NOT EXISTS idx_email ON users(email);

-- 插入测试数据（密码是 123456 的 bcrypt 哈希）
INSERT INTO users (name, nickname, password, email, phone, is_admin, status, logined_at)
VALUES
    ('admin', '管理员', '$2a$10$KRZ6wgW8E8HnFaKE.RQzYeXC1GOXqgLANBxZhCJgB6H4QWvJWxQB2', 'admin@example.com', '13800138000', 1, 1, CURRENT_TIMESTAMP),
    ('user1', '用户1', '$2a$10$KRZ6wgW8E8HnFaKE.RQzYeXC1GOXqgLANBxZhCJgB6H4QWvJWxQB2', 'user1@example.com', '13800138001', 0, 1, CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;

-- casbin_rule 表由应用启动时的 gorm-adapter 自动创建/迁移，无需手动建表。
-- 首个管理员授权示例（创建用户后执行）：
--   INSERT INTO casbin_rule (ptype, v0, v1) VALUES ('g', '<your-username>', 'admin');
