-- Emails are now stored lowercased. MySQL's default collation already makes
-- uq_user_email case-insensitive, so no second index is needed here (unlike
-- Postgres). If a case-sensitive collation was configured and two accounts
-- have emails that only differ in case, this fails on uq_user_email; one of
-- the two accounts has to be changed or merged by hand before the migration
-- can be run again.
UPDATE `user` SET email = LOWER(email);

-- An email change is only applied once the new address is confirmed through
-- the emailed link. Until then the requested address waits here and the
-- account keeps using (and logging in with) its current, verified email.
ALTER TABLE `user` ADD COLUMN pending_email VARCHAR(255);
