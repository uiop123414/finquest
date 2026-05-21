CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL,
    xp_total INT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(12,2) NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('income','expense')),
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    external_id TEXT,
    ai_confidence NUMERIC(4,3),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, external_id)
);

CREATE TABLE IF NOT EXISTS xp_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delta INT NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    earned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, achievement_id)
);

CREATE TABLE IF NOT EXISTS goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    target_amount NUMERIC(12,2) NOT NULL,
    current_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    deadline DATE NOT NULL
);

CREATE TABLE IF NOT EXISTS challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    xp_reward INT NOT NULL DEFAULT 50,
    condition_json JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS user_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Seed: system categories
INSERT INTO categories (id, user_id, name, is_system) VALUES
    (gen_random_uuid(), NULL, 'Еда', TRUE),
    (gen_random_uuid(), NULL, 'Транспорт', TRUE),
    (gen_random_uuid(), NULL, 'ЖКХ', TRUE),
    (gen_random_uuid(), NULL, 'Здоровье', TRUE),
    (gen_random_uuid(), NULL, 'Развлечения', TRUE),
    (gen_random_uuid(), NULL, 'Одежда', TRUE),
    (gen_random_uuid(), NULL, 'Зарплата', TRUE),
    (gen_random_uuid(), NULL, 'Прочее', TRUE)
ON CONFLICT DO NOTHING;

-- Seed: achievements
INSERT INTO achievements (id, code, name, description) VALUES
    (gen_random_uuid(), 'first_transaction', 'Первый шаг', 'Добавьте первую транзакцию'),
    (gen_random_uuid(), 'first_import', 'Импортёр', 'Импортируйте транзакции из CSV'),
    (gen_random_uuid(), 'ten_transactions', 'Десятка', 'Добавьте 10 транзакций'),
    (gen_random_uuid(), 'hundred_transactions', 'Сотня', 'Добавьте 100 транзакций'),
    (gen_random_uuid(), 'level_5', 'Опытный', 'Достигните 5 уровня'),
    (gen_random_uuid(), 'first_goal', 'Мечтатель', 'Создайте первую цель')
ON CONFLICT DO NOTHING;
