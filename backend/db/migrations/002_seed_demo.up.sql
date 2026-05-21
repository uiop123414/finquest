-- Demo user: demo@finquest.ru / demo123
-- Password hashed via pgcrypto bcrypt (compatible with Go's bcrypt.CompareHashAndPassword)
INSERT INTO users (id, email, hashed_password, xp_total, level)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'demo@finquest.ru',
    crypt('demo123', gen_salt('bf', 12)),
    1150,
    12
) ON CONFLICT (email) DO NOTHING;

-- Demo goals
INSERT INTO goals (user_id, name, target_amount, current_amount, deadline)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'Отпуск в Турции',  150000, 42000, '2026-08-01'),
    ('a0000000-0000-0000-0000-000000000001', 'Новый ноутбук',     90000, 27500, '2026-07-15'),
    ('a0000000-0000-0000-0000-000000000001', 'Подушка безопасности', 300000, 84000, '2026-12-31')
ON CONFLICT DO NOTHING;

-- ~100 transactions spread over Nov 2025 – Apr 2026
-- Uses system categories by name lookup
DO $$
DECLARE
    uid  UUID := 'a0000000-0000-0000-0000-000000000001';
    cat_food     UUID;
    cat_transport UUID;
    cat_utilities UUID;
    cat_health    UUID;
    cat_fun       UUID;
    cat_clothes   UUID;
    cat_salary    UUID;
    cat_other     UUID;
BEGIN
    SELECT id INTO cat_food      FROM categories WHERE name = 'Еда'           AND is_system;
    SELECT id INTO cat_transport FROM categories WHERE name = 'Транспорт'     AND is_system;
    SELECT id INTO cat_utilities FROM categories WHERE name = 'ЖКХ'           AND is_system;
    SELECT id INTO cat_health    FROM categories WHERE name = 'Здоровье'      AND is_system;
    SELECT id INTO cat_fun       FROM categories WHERE name = 'Развлечения'   AND is_system;
    SELECT id INTO cat_clothes   FROM categories WHERE name = 'Одежда'        AND is_system;
    SELECT id INTO cat_salary    FROM categories WHERE name = 'Зарплата'      AND is_system;
    SELECT id INTO cat_other     FROM categories WHERE name = 'Прочее'        AND is_system;

    INSERT INTO transactions (user_id, amount, type, category_id, date, note) VALUES

    -- ── Ноябрь 2025 ──────────────────────────────────────────────────────────
    (uid, 85000,  'income',  cat_salary,    '2025-11-05', 'Зарплата'),
    (uid,  5000,  'income',  cat_other,     '2025-11-12', 'Фриланс — верстка'),
    (uid,  7200,  'expense', cat_food,      '2025-11-03', 'Пятёрочка — продукты'),
    (uid,  3800,  'expense', cat_food,      '2025-11-10', 'ВкусВилл'),
    (uid,  4600,  'expense', cat_food,      '2025-11-17', 'Магнит — закупка'),
    (uid,  2100,  'expense', cat_food,      '2025-11-24', 'Доставка еды'),
    (uid,  1800,  'expense', cat_transport, '2025-11-04', 'Яндекс.Метро — карта'),
    (uid,   950,  'expense', cat_transport, '2025-11-14', 'Такси до аэропорта'),
    (uid,  7500,  'expense', cat_utilities, '2025-11-08', 'Квартплата ноябрь'),
    (uid,  1200,  'expense', cat_health,    '2025-11-20', 'Аптека — витамины'),
    (uid,  3500,  'expense', cat_fun,       '2025-11-09', 'Кино + ужин'),
    (uid,  2800,  'expense', cat_fun,       '2025-11-22', 'Spotify + Netflix'),
    (uid,  8900,  'expense', cat_clothes,   '2025-11-15', 'Куртка зимняя'),
    (uid,  1400,  'expense', cat_other,     '2025-11-28', 'Подписка iCloud'),

    -- ── Декабрь 2025 ─────────────────────────────────────────────────────────
    (uid, 85000,  'income',  cat_salary,    '2025-12-05', 'Зарплата'),
    (uid, 15000,  'income',  cat_other,     '2025-12-20', '13-я зарплата'),
    (uid,  9500,  'expense', cat_food,      '2025-12-02', 'Ашан — продукты'),
    (uid,  5200,  'expense', cat_food,      '2025-12-09', 'Праздничный стол'),
    (uid,  6800,  'expense', cat_food,      '2025-12-22', 'Новогодние продукты'),
    (uid,  3200,  'expense', cat_food,      '2025-12-28', 'Доставка суши'),
    (uid,  1800,  'expense', cat_transport, '2025-12-07', 'Проездной'),
    (uid,  1500,  'expense', cat_transport, '2025-12-24', 'Такси — Новый год'),
    (uid,  7500,  'expense', cat_utilities, '2025-12-08', 'Квартплата декабрь'),
    (uid,  4200,  'expense', cat_health,    '2025-12-15', 'Стоматолог'),
    (uid,  6500,  'expense', cat_fun,       '2025-12-12', 'Новогодний корпоратив'),
    (uid, 12000,  'expense', cat_fun,       '2025-12-30', 'Новогодняя ночь'),
    (uid,  4500,  'expense', cat_clothes,   '2025-12-18', 'Свитер и джинсы'),
    (uid,  3800,  'expense', cat_other,     '2025-12-10', 'Подарки коллегам'),
    (uid,  8700,  'expense', cat_other,     '2025-12-23', 'Подарки семье'),

    -- ── Январь 2026 ──────────────────────────────────────────────────────────
    (uid, 85000,  'income',  cat_salary,    '2026-01-10', 'Зарплата (задержка)'),
    (uid,  8000,  'income',  cat_other,     '2026-01-15', 'Фриланс — лого'),
    (uid,  6400,  'expense', cat_food,      '2026-01-05', 'Пятёрочка'),
    (uid,  4100,  'expense', cat_food,      '2026-01-12', 'ВкусВилл — фрукты'),
    (uid,  5300,  'expense', cat_food,      '2026-01-19', 'Закупка на месяц'),
    (uid,  2400,  'expense', cat_food,      '2026-01-26', 'Самокат — доставка'),
    (uid,  1800,  'expense', cat_transport, '2026-01-06', 'Проездной январь'),
    (uid,   750,  'expense', cat_transport, '2026-01-20', 'Каршеринг'),
    (uid,  7500,  'expense', cat_utilities, '2026-01-09', 'Квартплата январь'),
    (uid,  1100,  'expense', cat_health,    '2026-01-17', 'Аптека'),
    (uid,  2200,  'expense', cat_health,    '2026-01-28', 'Спортзал — абонемент'),
    (uid,  1500,  'expense', cat_fun,       '2026-01-08', 'Кино'),
    (uid,  3400,  'expense', cat_fun,       '2026-01-16', 'Боулинг с друзьями'),
    (uid,  1400,  'expense', cat_other,     '2026-01-22', 'Книги'),

    -- ── Февраль 2026 ─────────────────────────────────────────────────────────
    (uid, 85000,  'income',  cat_salary,    '2026-02-05', 'Зарплата'),
    (uid,  6000,  'income',  cat_other,     '2026-02-18', 'Возврат долга'),
    (uid,  5800,  'expense', cat_food,      '2026-02-03', 'Магнит'),
    (uid,  3600,  'expense', cat_food,      '2026-02-10', 'Пятёрочка'),
    (uid,  4900,  'expense', cat_food,      '2026-02-17', 'Продукты + напитки'),
    (uid,  7200,  'expense', cat_food,      '2026-02-14', 'Ресторан — 14 февраля'),
    (uid,  1800,  'expense', cat_transport, '2026-02-06', 'Проездной февраль'),
    (uid,  2200,  'expense', cat_transport, '2026-02-20', 'Поезд в Питер'),
    (uid,  7500,  'expense', cat_utilities, '2026-02-07', 'Квартплата февраль'),
    (uid,  2200,  'expense', cat_health,    '2026-02-13', 'Спортзал'),
    (uid,  1600,  'expense', cat_health,    '2026-02-25', 'Аптека — простуда'),
    (uid,  4800,  'expense', cat_fun,       '2026-02-21', 'Театр + ужин'),
    (uid,  6500,  'expense', cat_clothes,   '2026-02-11', 'Кроссовки'),
    (uid,  1900,  'expense', cat_other,     '2026-02-27', 'Канцелярия'),

    -- ── Март 2026 ────────────────────────────────────────────────────────────
    (uid, 85000,  'income',  cat_salary,    '2026-03-05', 'Зарплата'),
    (uid, 12000,  'income',  cat_other,     '2026-03-08', 'Фриланс — сайт'),
    (uid,  6200,  'expense', cat_food,      '2026-03-02', 'ВкусВилл'),
    (uid,  4400,  'expense', cat_food,      '2026-03-09', 'Пятёрочка'),
    (uid,  5100,  'expense', cat_food,      '2026-03-16', 'Ашан — закупка'),
    (uid,  3300,  'expense', cat_food,      '2026-03-23', 'Доставка Яндекс.Еда'),
    (uid,  1800,  'expense', cat_transport, '2026-03-06', 'Проездной март'),
    (uid,  1200,  'expense', cat_transport, '2026-03-14', 'Каршеринг BelkaCar'),
    (uid,  7800,  'expense', cat_utilities, '2026-03-07', 'Квартплата март'),
    (uid,  2200,  'expense', cat_health,    '2026-03-11', 'Спортзал'),
    (uid,  3500,  'expense', cat_health,    '2026-03-19', 'Врач — терапевт'),
    (uid,  5500,  'expense', cat_fun,       '2026-03-08', '8 марта — цветы и ресторан'),
    (uid,  2100,  'expense', cat_fun,       '2026-03-28', 'Подписки'),
    (uid,  7200,  'expense', cat_clothes,   '2026-03-22', 'Весенняя куртка'),
    (uid,  2600,  'expense', cat_other,     '2026-03-30', 'Хозтовары IKEA'),

    -- ── Апрель 2026 ──────────────────────────────────────────────────────────
    (uid, 85000,  'income',  cat_salary,    '2026-04-05', 'Зарплата'),
    (uid,  9000,  'income',  cat_other,     '2026-04-20', 'Бонус за проект'),
    (uid,  5600,  'expense', cat_food,      '2026-04-01', 'Магнит'),
    (uid,  4200,  'expense', cat_food,      '2026-04-08', 'ВкусВилл'),
    (uid,  6100,  'expense', cat_food,      '2026-04-15', 'Пятёрочка — шашлык'),
    (uid,  3800,  'expense', cat_food,      '2026-04-22', 'Продукты на дачу'),
    (uid,  1800,  'expense', cat_transport, '2026-04-06', 'Проездной апрель'),
    (uid,  4500,  'expense', cat_transport, '2026-04-18', 'Поезд на дачу'),
    (uid,  7800,  'expense', cat_utilities, '2026-04-07', 'Квартплата апрель'),
    (uid,  2200,  'expense', cat_health,    '2026-04-10', 'Спортзал'),
    (uid,  1800,  'expense', cat_health,    '2026-04-24', 'Аптека'),
    (uid,  3200,  'expense', cat_fun,       '2026-04-12', 'Кино + кафе'),
    (uid,  5800,  'expense', cat_fun,       '2026-04-26', 'Концерт'),
    (uid,  4300,  'expense', cat_clothes,   '2026-04-17', 'Летняя одежда'),
    (uid,  2100,  'expense', cat_other,     '2026-04-29', 'Хозтовары'),

    -- ── Май 2026 (текущий месяц) ──────────────────────────────────────────────
    (uid, 85000,  'income',  cat_salary,    '2026-05-05', 'Зарплата'),
    (uid,  5800,  'expense', cat_food,      '2026-05-02', 'Пятёрочка'),
    (uid,  4100,  'expense', cat_food,      '2026-05-09', 'Ашан'),
    (uid,  6200,  'expense', cat_food,      '2026-05-12', 'Шашлык — майские'),
    (uid,  1800,  'expense', cat_transport, '2026-05-06', 'Проездной май'),
    (uid,  3200,  'expense', cat_transport, '2026-05-10', 'Поездка на дачу'),
    (uid,  7800,  'expense', cat_utilities, '2026-05-07', 'Квартплата май'),
    (uid,  2200,  'expense', cat_health,    '2026-05-13', 'Спортзал'),
    (uid,  4500,  'expense', cat_fun,       '2026-05-08', 'Майские праздники'),
    (uid,  3100,  'expense', cat_other,     '2026-05-15', 'Инструменты для дачи'),
    (uid,  2700,  'expense', cat_food,      '2026-05-19', 'ВкусВилл'),
    (uid,  1600,  'expense', cat_health,    '2026-05-21', 'Аптека'),
    (uid,  8500,  'income',  cat_other,     '2026-05-20', 'Фриланс — мобилка'),
    (uid,  2900,  'expense', cat_clothes,   '2026-05-17', 'Летняя обувь');

END $$;

-- Award first_transaction + ten_transactions + hundred_transactions achievements
INSERT INTO user_achievements (user_id, achievement_id)
SELECT 'a0000000-0000-0000-0000-000000000001', id
FROM achievements
WHERE code IN ('first_transaction', 'ten_transactions', 'hundred_transactions', 'first_goal')
ON CONFLICT DO NOTHING;
