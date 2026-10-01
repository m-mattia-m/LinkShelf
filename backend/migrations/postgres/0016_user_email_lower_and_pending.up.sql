-- Store emails lowercased. Fails if two emails differ only in case; resolve those by hand.
UPDATE "user" SET email = LOWER(email) WHERE email <> LOWER(email);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_email_lower ON "user" (LOWER(email));

-- Requested email change, applied once the new address is confirmed.
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS pending_email VARCHAR(255);
