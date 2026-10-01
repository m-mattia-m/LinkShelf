-- Store emails lowercased. The default collation already makes uq_user_email case-insensitive.
UPDATE `user` SET email = LOWER(email);

-- Requested email change, applied once the new address is confirmed.
ALTER TABLE `user` ADD COLUMN pending_email VARCHAR(255);
