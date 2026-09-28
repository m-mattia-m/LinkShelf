-- 0003_dashboard seeded a dev@linkshelf.local account with a hardcoded
-- password so a fresh instance had someone to log in as. Now that the
-- bootstrap admin (AUTHENTICATION_BOOTSTRAPADMIN_*) covers that role, remove
-- the seeded account so its known credentials aren't left in the database.
DELETE FROM `user` WHERE id = '018f1a3e-0000-7000-8000-000000000001';
