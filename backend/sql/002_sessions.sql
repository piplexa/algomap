-- =====================================================
-- Таблица сессий для аутентификации
-- =====================================================

CREATE TABLE main.sessions (
    id TEXT PRIMARY KEY,                                    -- session_key (uuid)
    user_id BIGINT NOT NULL REFERENCES main.users(id),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE main.sessions IS 'Сессии пользователей для аутентификации';
COMMENT ON COLUMN main.sessions.id IS 'Session key (UUID) - используется в Cookie или Bearer token';
COMMENT ON COLUMN main.sessions.expires_at IS 'Время истечения сессии';

-- Индексы
CREATE INDEX idx_sessions_user_id ON main.sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON main.sessions(expires_at);

-- =====================================================
-- TODO: Таблица для OAuth провайдеров (Google, GitHub)
-- Будет добавлена позже для интеграции с внешними провайдерами
-- =====================================================
-- CREATE TABLE main.oauth_providers (
--     id BIGSERIAL PRIMARY KEY,
--     user_id BIGINT NOT NULL REFERENCES main.users(id),
--     provider VARCHAR(50) NOT NULL,           -- google, github, etc
--     provider_user_id VARCHAR(255) NOT NULL,  -- ID пользователя у провайдера
--     email VARCHAR(255),
--     created_at TIMESTAMP NOT NULL DEFAULT NOW(),
--     UNIQUE(provider, provider_user_id)
-- );