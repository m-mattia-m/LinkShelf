-- Emails are now stored lowercased. Every lookup already compared them with
-- LOWER(), but the plain UNIQUE(email) is case-sensitive on Postgres, so
-- "Victim@corp.com" and "victim@corp.com" could both exist and login, OIDC
-- matching and password reset would hit whichever row came back first.
--
-- This fails on uq_user_email if two accounts have emails that only differ in
-- case. The error names the value; one of the two accounts has to be changed
-- or merged by hand before the migration can be run again. That is on
-- purpose: merging accounts automatically could hand one person's data to
-- another.
UPDATE "user" SET email = LOWER(email) WHERE email <> LOWER(email);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_email_lower ON "user" (LOWER(email));

-- An email change is only applied once the new address is confirmed through
-- the emailed link. Until then the requested address waits here and the
-- account keeps using (and logging in with) its current, verified email.
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS pending_email VARCHAR(255);
