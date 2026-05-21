-- Removes demo user and all their data (cascades to transactions, goals, achievements)
DELETE FROM users WHERE email = 'demo@finquest.ru';
