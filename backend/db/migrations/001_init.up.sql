CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL,
    xp_total INT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE transactions (
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

CREATE TABLE xp_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delta INT NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE user_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    earned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, achievement_id)
);

CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    target_amount NUMERIC(12,2) NOT NULL,
    current_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    deadline DATE NOT NULL
);

CREATE TABLE challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    xp_reward INT NOT NULL DEFAULT 50,
    condition_json JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE user_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- System categories
INSERT INTO categories (user_id, name, is_system) VALUES
    (NULL, 'Еда', TRUE),
    (NULL, 'Транспорт', TRUE),
    (NULL, 'ЖКХ', TRUE),
    (NULL, 'Здоровье', TRUE),
    (NULL, 'Развлечения', TRUE),
    (NULL, 'Одежда', TRUE),
    (NULL, 'Зарплата', TRUE),
    (NULL, 'Прочее', TRUE);

-- Achievements
INSERT INTO achievements (code, name, description) VALUES
    ('first_transaction', 'Первый шаг', 'Добавьте первую транзакцию'),
    ('first_import', 'Импортёр', 'Импортируйте транзакции из CSV'),
    ('ten_transactions', 'Десятка', 'Добавьте 10 транзакций'),
    ('hundred_transactions', 'Сотня', 'Добавьте 100 транзакций'),
    ('level_5', 'Опытный', 'Достигните 5 уровня'),
    ('first_goal', 'Мечтатель', 'Создайте первую цель');
