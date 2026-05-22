-- Remove demo deposits and credits
DELETE FROM deposits WHERE user_id = 'a0000000-0000-0000-0000-000000000001';
DELETE FROM credits  WHERE user_id = 'a0000000-0000-0000-0000-000000000001';
