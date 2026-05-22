-- Demo deposits and credits for demo@finquest.ru
-- Only inserted if demo user exists

INSERT INTO deposits (id, user_id, bank_name, amount, interest_rate, start_date, end_date, note)
SELECT
    gen_random_uuid(),
    'a0000000-0000-0000-0000-000000000001',
    bank_name, amount, interest_rate, start_date::date, end_date::date, note
FROM (VALUES
    ('Сбербанк',  200000, 16.5, '2026-01-15', '2026-07-15', 'Вклад «Лучший старт»'),
    ('Т-Банк',    150000, 17.0, '2026-02-01', '2026-08-01', 'Накопительный счёт'),
    ('ВТБ',       100000, 15.8, '2025-12-01', '2026-06-01', 'Срочный вклад 6 мес.')
) AS t(bank_name, amount, interest_rate, start_date, end_date, note)
WHERE EXISTS (SELECT 1 FROM users WHERE id = 'a0000000-0000-0000-0000-000000000001')
ON CONFLICT DO NOTHING;

INSERT INTO credits (id, user_id, type, bank_name, total_amount, remaining_balance, interest_rate, monthly_payment, note)
SELECT
    gen_random_uuid(),
    'a0000000-0000-0000-0000-000000000001',
    type, bank_name, total_amount, remaining_balance, interest_rate, monthly_payment, note
FROM (VALUES
    ('consumer', 'Сбербанк',  500000, 320000, 14.9, 12500, 'Потребительский кредит на ремонт'),
    ('card',     'Т-Банк',    100000,  35000, 29.9,  3500, 'Кредитная карта Платинум')
) AS t(type, bank_name, total_amount, remaining_balance, interest_rate, monthly_payment, note)
WHERE EXISTS (SELECT 1 FROM users WHERE id = 'a0000000-0000-0000-0000-000000000001')
ON CONFLICT DO NOTHING;
