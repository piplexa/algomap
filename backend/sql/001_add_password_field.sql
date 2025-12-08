-- =====================================================
-- Migration: Add password hash field to users table
-- =====================================================

-- Добавляем поле для хранения хеша пароля
ALTER TABLE main.users
ADD COLUMN hashPassword TEXT;

COMMENT ON COLUMN main.users.hashPassword IS 'Хеш пароля (bcrypt через pgcrypto)';
